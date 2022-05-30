package event

import (
	"errors"
	"fmt"
	"github.com/valyala/fastjson"
	"cronTask/contrib/helper"
	"cronTask/modules/common"
	"time"
)

func getToken() string {

	v, ok := pCfg[ftPlatformID]["public_token"].(string)
	if !ok {
		return ""
	}

	return "Token " + v
}

// 获取包括今天 十五天的数据

func ftStart() {

	headers := map[string]string{
		"Authorization": getToken(),
		"Content-Type":  "application/json",
		"Accept":        "application/json",
		"Connection":    "Keep-Alive",
	}
	dates := afterTodayFourteenDay()
	addr := ftURL + "api/v4/events/paginate?page_size=100&market_option=match&lang=%s&timezone=%s&sort_by_popular=true&date=%s"
	for k, v := range dates {

		currentAddr := fmt.Sprintf(addr, ftLang, locStr, v)
		common.Log("event", "req url: %s", currentAddr)
		status, data, err := HttpPostJsonProxy(nil, currentAddr, headers)
		if err != nil {
			common.Log("event", "statusCode:%s, err:%v", status, err)
			continue
		}

		// 清除老数据
		if k == 0 {
			clearData(ftPlatformID)
		}

		next, rs, err := ftFormatData(data)
		if err != nil {
			common.Log("event", "formatData err:%v", err)
			continue
		}
		err = insert(esCli, rs)
		if err != nil {
			common.Log("event", "ftdj insert err: %v", err)
			continue
		}

		for {

			if next == "" {
				break
			}
			common.Log("event", "req url: %s", ftURL+next)
			status, data, err = HttpPostJsonProxy(nil, ftURL+next, headers)
			if err != nil {
				common.Log("event", "statusCode:%d, err:%v", status, err)
				break
			}

			next, rs, err = ftFormatData(data)
			if err != nil {
				common.Log("event", "formatData err:%v", err)
				break
			}

			err = insert(esCli, rs)
			if err != nil {
				common.Log("event", "ftdj insert err: %v", err)
				continue
			}
		}
	}
}

// 处理im体育 获取的数据 足球
func ftFormatData(b []byte) (string, map[string]popularEvents, error) {

	bStr, err := fastjson.ParseBytes(b)
	if err != nil {
		return "", nil, err
	}

	//赛事id 赛事名称 场馆 跳转详情 主队  客队 项目 开赛时间  赔率 主 和 客
	rcs := bStr.GetArray("results")
	if len(rcs) == 0 {
		return "", nil, errors.New("no data found")
	}

	data := make(map[string]popularEvents)
	for _, v := range rcs {

		lines := v.GetArray("markets")
		if len(lines) == 0 {
			continue
		}

		var (
			homeOdds float64
			awayOdds float64
		)
		selection := lines[0].GetArray("selection")
		for _, vv := range selection {

			name := string(vv.GetStringBytes("name"))
			if name == "home" { // 主
				homeOdds = vv.GetFloat64("malay_odds")
			}

			if name == "away" { // 客
				awayOdds = vv.GetFloat64("malay_odds")
			}
		}

		data[v.Get("event_id").String()] = popularEvents{
			EventName:    string(v.GetStringBytes("competition_name")),
			PlatformName: "雷火电竞",
			HomeTeam:     string(v.GetStringBytes("home", "team_name")),
			AwayTeam:     string(v.GetStringBytes("away", "team_name")),
			Project:      string(v.GetStringBytes("game_name")),
			EventAt:      helper.StrToTime(string(v.GetStringBytes("start_datetime")), loc).Unix(),
			HomeOdds:     homeOdds,
			AwayOdds:     awayOdds,
			EqOdds:       0,
			PlatformID:   ftPlatformID,
		}
	}

	return string(bStr.GetStringBytes("next")), data, nil
}

// 获取包含今日在内的15天
func afterTodayFourteenDay() []string {

	now := time.Now().In(loc)

	dates := []string{now.Format("2006-01-02")}
	for i := 0; i < 14; i++ {
		dates = append(dates, now.Add(time.Duration(24*(i+1))*time.Hour).Format("2006-01-02"))
	}

	return dates
}
