package event

import (
	"context"
	"cronTask/contrib/conn"
	"cronTask/contrib/helper"
	"cronTask/modules/common"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/olivere/elastic/v7"
	"time"
)

var (
	cli       *redis.ClusterClient
	esCli     *elastic.Client
	pCfg      map[string]map[string]interface{}
	lang      string
	loc       *time.Location
	locStr    string
	ctx       = context.Background()
	httpProxy string
	esPrefix  string
)

func Parse(endpoints []string, path, proxy string) {

	conf, platConf, err := common.ConfPlatParse(endpoints, path)
	if err != nil {
		fmt.Println("plat config parse error: ", err)
		return
	}
	// 获取语言
	lang = conf.Lang
	esPrefix = conf.EsPrefix

	if lang == "cn" {
		locStr = "Asia/Shanghai"
		loc, _ = time.LoadLocation(locStr)

	} else if lang == "vn" || lang == "th" {
		locStr = "Asia/Bangkok"
		loc, _ = time.LoadLocation(locStr)
	}

	// 初始化redis
	cli = conn.InitRedisCluster(conf.Redis.Addr, conf.Redis.Password)
	// 初始化es
	esCli = conn.InitES(conf.Es.Host, conf.Es.Username, conf.Es.Password)

	pCfg = platConf

	httpProxy = proxy

	eventHandle()
}

func eventHandle() {

	// im体育
	err := imSportStart()
	if err != nil {
		common.Log("event", "im error: %s", err.Error())
		return
	}
	// 雷火电竞
	ftStart()

	// 更新展示数据
	common.Log("event", "更新展示数据")
	time.Sleep(3 * time.Second) // 等待es刷新

	// 获取开启中的最新数据
	before, err := popularEventsShowData()
	if err != nil {
		common.Log("event", "not found no need update to last error: %s", err.Error())
		return
	}

	var p []popularEventRedis
	for k, v := range before {
		if _, ok := allData[k]; ok {
			fmt.Println("更新：", k)
			d := before[k]
			d.EqOdds = v.EqOdds
			d.HomeOdds = v.HomeOdds
			d.AwayOdds = v.AwayOdds
			d.EventAt = v.EventAt
			before[k] = d

			p = append(p, before[k])
			continue
		}
		delete(before, k)
		// 后台关闭
		// 删除
		fmt.Println("del:", k)
		_ = popularCloseUpdate(k)
	}

	// sort 升序排序 排序 赛事时间 赛事名称
	orderedBy(sortBySort, sortByEventAt, sortByEventName).Sort(p)
	b, err := helper.JsonMarshal(p)
	if err != nil {
		common.Log("event", "jsonMarshal error: %s", err.Error())
		return
	}
	_, err = cli.Set(ctx, popularRedisKey, string(b), 100*time.Hour).Result()
	if err != nil {
		common.Log("event", "redis set error: %s", err.Error())
		return
	}

	cli.Persist(ctx, popularRedisKey)
}

// 获取redis中展示的数据
func popularEventsShowData() (map[string]popularEventRedis, error) {

	rs, err := cli.Get(ctx, popularRedisKey).Result()
	if err != nil {
		return nil, err
	}

	var data []popularEventRedis
	err = helper.JsonUnmarshal([]byte(rs), &data)
	if err != nil {
		return nil, err
	}

	result := make(map[string]popularEventRedis)
	for _, v := range data {
		result[v.ID] = v
	}

	return result, nil
}

// 更新
func popularCloseUpdate(id string) error {
	_, err := esCli.Update().Index(esPrefix + "popular_events").Id(id).
		Doc(map[string]interface{}{"state": 0}).Refresh("wait_for").Do(ctx) // 执行ES查询
	return err
}

// 数据入库
func insert(es *elastic.Client, data map[string]popularEvents) error {

	if len(data) == 0 {
		common.Log("event", "not data need insert")
		return nil
	}

	req := es.Bulk().Index(esPrefix + "popular_events_cache")
	for k, v := range data {
		req.Add(elastic.NewBulkIndexRequest().Id(k).Doc(v))
		allData[k] = latestData{
			EventAt:  v.EventAt,
			HomeOdds: v.HomeOdds,
			AwayOdds: v.AwayOdds,
			EqOdds:   v.EqOdds,
		}
	}

	if req.NumberOfActions() < 0 {
		return nil
	}

	_, err := req.Do(ctx)
	return err
}

// 清空场馆的数据
func clearData(platformID string) {
	query := elastic.NewTermQuery("platform_id", platformID)
	res, err := elastic.NewDeleteByQueryService(esCli).Index(esPrefix + "popular_events_cache").Query(query).Do(ctx)
	if err != nil {
		return
	}

	common.Log("event", "clear old data platform_id= %s, number= %d", platformID, res.Total)
}
