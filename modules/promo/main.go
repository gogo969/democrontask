package promo

import (
	"context"
	"cronTask/contrib/conn"
	"cronTask/modules/common"
	g "github.com/doug-martin/goqu/v9"
	"github.com/go-redis/redis/v8"
	"github.com/jmoiron/sqlx"
	"github.com/olivere/elastic/v7"
	"time"
)

// 活动流水计算脚本
var (
	db         *sqlx.DB
	cli        *redis.ClusterClient
	prefix     string
	esPrefix   string
	pullPrefix string
	lang       string
	loc        *time.Location
	esCli      *elastic.Client
	ctx        = context.Background()
	dialect    = g.Dialect("mysql")
)

func Parse(endpoints []string, path, flag string) {

	conf := common.ConfParse(endpoints, path)
	// 获取语言
	lang = conf.Lang
	if lang == "cn" {
		loc, _ = time.LoadLocation("Asia/Shanghai")
	} else if lang == "vn" || lang == "th" {
		loc, _ = time.LoadLocation("Asia/Bangkok")
	}

	prefix = conf.Prefix
	esPrefix = conf.EsPrefix
	pullPrefix = conf.PullPrefix
	// 初始化redis
	cli = conn.InitRedisCluster(conf.Redis.Addr, conf.Redis.Password)
	// 初始化db
	db = conn.InitDB(conf.Db.Master.Addr, conf.Db.Master.MaxIdleConn, conf.Db.Master.MaxIdleConn)
	// 初始化es
	esCli = conn.InitES(conf.Es.Host, conf.Es.Username, conf.Es.Password)

	if flag == "signday" {

		tm := time.Now()
		//计算当天数据
		signReport(tm.Unix())

		h := tm.Hour()
		//如果当前时间1点以前需计算前一天数据
		if h == 0 {
			tm = tm.AddDate(0, 0, -1)
			signReport(tm.Unix())
		}

		return
	}

	if flag == "signweek" {

		tm := time.Now()
		d := tm.Weekday()
		//周一计算上周签到奖金
		if d == time.Monday {
			signHandoutAward(tm.Unix())
		}

		return
	}

	if flag == "lastsignload" {

		tm := time.Now()
		SignLoadRecord(tm.Unix())

		return
	}
}
