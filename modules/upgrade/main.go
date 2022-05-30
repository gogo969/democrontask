package upgrade

import (
	"context"
	"cronTask/contrib/conn"
	"cronTask/contrib/helper"
	"cronTask/modules/common"
	"errors"
	"fmt"
	g "github.com/doug-martin/goqu/v9"
	"github.com/go-redis/redis/v8"
	"github.com/jmoiron/sqlx"
	"github.com/olivere/elastic/v7"
	"strings"
	"sync"
	"time"
)

// 全局参数
var (
	db       *sqlx.DB
	reportDb *sqlx.DB
	cli      *redis.Client
	esCli    *elastic.Client
	prefix   string
	esPrefix string
	loc      *time.Location
	dialect  = g.Dialect("mysql")
	ml       = map[int]common.LevelInfo{}
)

// 活跃会员参数
var (
	ctx         = context.Background()
	size        = 1000     //单次聚合查询记录数
	tSize       = 50       //事务并行处理记录数
	tEarliest   int64      //91天前的日期 + 00:00:00 时间戳
	tEarliestMs int64      //91天前的日期 + 00:00:00 ms时间戳
	tYesterday  int64      //1天前的日期 + 23:59:59 时间戳
	vvv         [][]string //90天内有登陆过的当前等级为vip10的会员名切片
	unique      = map[string]bool{}
	vAll        []mWaterFlow
	i           = 0
)

type Fn struct {
	Level int
	Wg    *sync.WaitGroup
}

type mValidBetAmount struct {
	UserName       string  `db:"user_name" json:"user_name"`
	ValidBetAmount float64 `db:"valid_bet_amount" json:"valid_bet_amount"`
	DepositAmount  float64 `db:"deposit_amount" json:"deposit_amount"`
}

func Parse(endpoints []string, path, usernames string) {

	conf := common.ConfParse(endpoints, path)

	fmt.Println("upgrade", conf)

	// 场馆账号前缀
	prefix = conf.Prefix
	esPrefix = conf.EsPrefix
	//pullPrefix = conf.PullPrefix
	loc, _ = time.LoadLocation("Asia/Bangkok")

	// 初始化db
	db = conn.InitDB(conf.Db.Master.Addr, conf.Db.Master.MaxIdleConn, conf.Db.Master.MaxIdleConn)
	reportDb = conn.InitDB(conf.Db.Report.Addr, conf.Db.Report.MaxIdleConn, conf.Db.Report.MaxIdleConn)
	// 初始化redis
	cli = conn.InitRedisSentinel(conf.Redis.Addr, conf.Redis.Password, conf.Redis.Sentinel, conf.Redis.Db)
	// 初始化es
	esCli = conn.InitES(conf.Es.Host, conf.Es.Username, conf.Es.Password)
	// 初始化td
	td := conn.InitTD(conf.Td.Addr, conf.Td.MaxIdleConn, conf.Td.MaxOpenConn)
	common.InitTD(td)

	var names []string
	if usernames == "repair" {
		levelTask(names, true)
	} else {
		if usernames != "0.0.0.0" {
			names = strings.Split(usernames, ",")
		}
		levelTask(names, false)
	}

	// 清空去重map
	unique = make(map[string]bool)
}

// 会员自动升级/降级
func levelTask(names []string, repair bool) {

	var err error
	// 获取当前时间
	tm := time.Now()
	// 全部执行升级，需要判断是否重复执行
	if len(names) == 0 && !repair {
		// 获取当前日期
		date := tm.Format("2006-01-02")
		expireAt := time.Date(tm.Year(), tm.Month(), tm.Day(), 0, 0, 0, 0, time.Local).AddDate(0, 0, 1)
		err = common.LockExpireAt(cli, "upgrade:"+date, expireAt)
		if err != nil {
			common.Log("upgrade", "升级脚本重复执行")
			return
		}
	}

	// 获取会员等级配置信息
	ml, err = common.MemberLevelList(db)
	if err != nil {
		fmt.Println(err)
		return
	}

	// 获取保级流水计算的上下限日期 90天内，截止昨天
	dEarliest := tm.AddDate(0, 0, -90).Format("2006-01-02")
	dYesterday := tm.AddDate(0, 0, 0).Format("2006-01-02")
	tEarliest = helper.StrToTime(dEarliest+" 00:00:00", loc).Unix()
	tEarliestMs = helper.StrToTime(dEarliest+" 00:00:00", loc).UnixMilli()
	tYesterday = helper.StrToTime(dYesterday+" 23:59:59", loc).Unix()

	fmt.Println(dEarliest+" 00:00:00", dYesterday+" 23:59:59", tEarliestMs)

	// 获取会员名切片
	err = getMembers(names)
	if err != nil {
		common.Log("upgrade", "error: %v", err)
		return
	}

	// 执行活动会员的升降级任务
	process()

	common.Log("upgrade", "升降级完成，耗时 ： %v", time.Since(tm))

	// todo 发送处理完通知
}

func process() {

	// 循环获取
	for _, v := range vvv {

		err := task(v)
		if err != nil {
			fmt.Println(err)
			continue
		}
	}
}

func task(names []string) error {

	ex := g.Ex{
		"username":    names,
		"report_time": g.Op{"between": g.Range(tEarliest, tYesterday)},
		"report_type": "2",
		"data_type":   "2",
	}
	mtp, err := memberReportValidBetAmounts(ex)
	if err != nil {
		fmt.Println(err)
		return err
	}

	mbs, _, err := common.MemberMCache(db, names, prefix)
	if err != nil {
		common.Log("upgrade", "error : %v", err)
		return err
	}

	fmt.Printf("-----mb[%d],mtp[%d]-----\n", len(mbs), len(mtp))

	for k, v := range names {

		// 获取会员id
		mb, ok := mbs[v]
		if !ok || mb.UID == "" {
			continue
		}

		switch mb.Level {
		case common.Vip10:
			continue
		default:
			err = processVIP(mb, mtp)
			if err != nil {
				common.Log("upgrade", "error : %v", err)
			}
		}

		fmt.Printf("----------------%d|%d|%d---------------\n", k, (i-1)%size, i-1)
	}

	fmt.Printf("---vvv[%d]---m[%d]---\n", len(vvv), len(vAll))

	// 保存会员vip信息
	if len(vAll) > 0 {
		p := len(vAll) / tSize
		for j := 0; j < p; j++ {
			offset := j * tSize
			vipInfoToCache(vAll[offset : offset+tSize])
		}

		offset := p * tSize
		vipInfoToCache(vAll[offset:])
	}

	return nil
}

func vipInfoToCache(m []mWaterFlow) {

	pipe := cli.TxPipeline()
	defer pipe.Close()

	for _, v := range m {
		mp := mWaterFlowToMap(v)
		key := fmt.Sprintf("%s:vip:%s", prefix, v.Username)
		pipe.Unlink(ctx, key)
		pipe.HMSet(ctx, key, mp)
		pipe.Persist(ctx, key)
	}
	_, _ = pipe.Exec(ctx)
}

func getMembers(names []string) error {

	// 获取活跃会员usernames
	ex := g.Ex{}
	if len(names) > 0 {
		ex["username"] = names
	}
	count, err := common.MembersCount(db, ex)
	if err != nil {
		common.Log("upgrade", "error : %v", err)
		return err
	}

	fmt.Printf("count : %d", count)

	if count == 0 {
		return errors.New("no members")
	}

	p := count / size
	l := count % size
	if l > 0 {
		p += 1
	}

	for j := 1; j <= p; j++ {
		ns, err := common.MembersPageNames(db, j, size, ex)
		if err != nil {
			common.Log("upgrade", "error : %v", err)
			return err
		}

		vvv = append(vvv, ns)
	}

	return nil
}
