package event

import (
	"bytes"
	"cronTask/contrib/helper"
	"cronTask/modules/common"
	"crypto/aes"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fastjson"
	"strconv"
	"time"
)

func imSportStart() error {

	now := time.Now()
	timestamp := imSportTimestamp(now)
	fTotal, bTotal, err := imSportCount(timestamp) // 足球 篮球数量总计
	if err != nil {
		return err
	}

	common.Log("event", "\n足球:%#v \n篮球: %#v", fTotal, bTotal)

	// 清除老数据
	clearData(imPlatformID)

	// 1 = 早盘 2 = 今日 3 = 滚球
	markets := []int{1, 2, 3}
	for _, v := range markets {

		var ft int
		var bt int
		// 1 = 早盘
		if v == 1 {
			ft = fTotal.EarlyFECount
			bt = bTotal.EarlyFECount
		}
		//  2 = 今日
		if v == 2 {
			ft = fTotal.TodayFECount
			bt = bTotal.TodayFECount
		}
		//  3 = 滚球
		if v == 3 {
			ft = fTotal.RBFECount
			bt = bTotal.RBFECount
		}

		// 足球
		fPageSize := ft / 100
		if ft%100 > 0 {
			fPageSize += 1
		}

		// 篮球
		bPageSize := bt / 100
		if bt%100 > 0 {
			bPageSize += 1
		}

		// 足球 篮球
		sportIds := []int{1, 2}
		for _, sportId := range sportIds {
			pageSize := 0
			if sportId == 1 {
				pageSize = fPageSize
			} else {
				pageSize = bPageSize
			}
			for i := 0; i < pageSize; i++ {
				common.Log("event", "page=%d, sportId=%d, market=%d, pageSize=%d", i+1, sportId, v, pageSize)
				err = imSportProcess(i+1, sportId, imOddsTypeHangKong, v, timestamp)
				if err != nil {
					common.Log("event", "err: %s", err.Error())
					continue
				}
			}
		}
	}
	return nil
}

// im体育场馆
func imSportProcess(page, sportId, oddsType, market int, timestamp string) error {

	p := imSportParam{
		SportId:      sportId,  // 足球
		Market:       market,   // 1 = 早盘 2 = 今日 3 = 滚球
		OddsType:     oddsType, // 1 = 马来盘 2 = 香港盘 3 = 欧洲盘 4 = 印尼盘
		IsCombo:      false,
		Page:         page,
		TimeStamp:    timestamp,
		BetTypeIds:   []int{3},
		LanguageCode: imLang,   // VN 越南文
		PeriodIds:    []int{1}, // 1 全场 上半场 2
		PageRecords:  100,
	}

	body, _ := helper.JsonMarshal(p)
	reqURL := fmt.Sprintf("%s%s", imURL, "/api/mobile/getEventInfoByPage")
	headers := map[string]string{
		"Content-Type": "application/json",
		"Accept":       "application/json",
		"Connection":   "Keep-Alive",
	}
	status, b, err := HttpPostJsonProxy(body, reqURL, headers)
	if err != nil {
		common.Log("event", "request path: %s, error: %s", reqURL, err.Error())
		return err
	}

	if status != fasthttp.StatusOK {
		common.Log("event", "request status code: %d", status)
		return err
	}

	var data map[string]popularEvents
	if sportId == 1 {
		data, err = imFFormatData(b)
	}

	if sportId == 2 {
		data, err = imBFormatData(b)
	}

	if data == nil || err != nil {
		common.Log("event", "formatData error:%v", err)
		return err
	}

	// 写入es
	return insert(esCli, data)
}

// 统计总条数 用户分页查询
func imSportCount(timestamp string) (imCount, imCount, error) {

	p := imSportCountParam{
		IsCombo:      false,
		TimeStamp:    timestamp,
		LanguageCode: "CHS", // VN 越南文
	}

	body, _ := helper.JsonMarshal(p)
	reqURL := fmt.Sprintf("%s%s", imURL, "/api/mobile/getAllSportCount")
	headers := map[string]string{
		"Content-Type": "application/json",
		"Accept":       "application/json",
		"Connection":   "Keep-Alive",
	}
	status, b, err := HttpPostJsonProxy(body, reqURL, headers)
	if err != nil {
		return imCount{}, imCount{}, err
	}

	if status != fasthttp.StatusOK {
		return imCount{}, imCount{}, errors.New(fmt.Sprintf("request status code: %d", status))
	}

	values, err := fastjson.ParseBytes(b)
	if err != nil {
		return imCount{}, imCount{}, err
	}

	if values.GetInt("StatusCode") != 100 {
		return imCount{}, imCount{}, errors.New(fmt.Sprintf("im StatusCode code: %d", values.GetInt("StatusCode")))
	}

	vArr := values.GetArray("SportCount")
	if len(vArr) == 0 {
		return imCount{}, imCount{}, errors.New("count number 0")
	}

	var (
		fTotal imCount
		bTotal imCount
	)
	for _, v := range vArr {
		// 足球
		if v.GetInt("SportId") == 1 {
			fTotal.EarlyFECount = v.GetInt("EarlyFECount")
			fTotal.TodayFECount = v.GetInt("TodayFECount")
			fTotal.RBFECount = v.GetInt("RBFECount")
		}
		// 篮球
		if v.GetInt("SportId") == 2 {
			bTotal.EarlyFECount = v.GetInt("EarlyFECount")
			bTotal.TodayFECount = v.GetInt("TodayFECount")
			bTotal.RBFECount = v.GetInt("RBFECount")
		}
	}

	return fTotal, bTotal, nil
}

// 处理im体育 获取的数据 足球
func imFFormatData(b []byte) (map[string]popularEvents, error) {

	v, err := fastjson.ParseBytes(b)
	if err != nil {
		return nil, err
	}

	//赛事id 赛事名称 场馆 跳转详情 主队  客队 项目 开赛时间  赔率 主 和 客
	rcs := v.GetArray("Sports")
	if len(rcs) == 0 {
		return nil, errors.New("no data found")
	}

	data := make(map[string]popularEvents)
	events := rcs[0].GetArray("Events")
	for _, v := range events {

		lines := v.GetArray("MarketLines")
		if len(lines) == 0 || lines[0].GetInt("PeriodId") != 1 {
			continue
		}

		var (
			homeOdds float64
			awayOdds float64
			eqOdds   float64
		)
		wagers := lines[0].GetArray("WagerSelections")
		for _, vv := range wagers {

			sid := vv.GetInt("SelectionId")
			if sid == 5 { // 主
				homeOdds = vv.GetFloat64("Odds")
			}

			if sid == 6 { // 客
				awayOdds = vv.GetFloat64("Odds")
			}

			if sid == 7 { // 和
				eqOdds = vv.GetFloat64("Odds")
			}
		}

		data[v.Get("EventId").String()] = popularEvents{
			EventName:    string(v.GetStringBytes("Competition", "CompetitionName")),
			PlatformName: "IM体育",
			HomeTeam:     string(v.GetStringBytes("HomeTeam")),
			AwayTeam:     string(v.GetStringBytes("AwayTeam")),
			Project:      string(rcs[0].GetStringBytes("SportName")),
			EventAt:      helper.StrToTime(string(v.GetStringBytes("EventDate")), loc).Unix(),
			HomeOdds:     homeOdds,
			AwayOdds:     awayOdds,
			EqOdds:       eqOdds,
			PlatformID:   imPlatformID,
		}
	}

	return data, nil
}

// 处理im体育 获取的数据 篮球
func imBFormatData(b []byte) (map[string]popularEvents, error) {

	v, err := fastjson.ParseBytes(b)
	if err != nil {
		return nil, err
	}

	//赛事id 赛事名称 场馆 跳转详情 主队  客队 项目 开赛时间  赔率 主 和 客
	rcs := v.GetArray("Sports")
	if len(rcs) == 0 {
		return nil, errors.New("no data found")
	}

	data := make(map[string]popularEvents)
	events := rcs[0].GetArray("Events")
	for _, v := range events {

		lines := v.GetArray("MarketLines")
		if len(lines) == 0 || lines[0].GetInt("PeriodId") != 1 {
			continue
		}

		var (
			homeOdds float64
			awayOdds float64
		)
		wagers := lines[0].GetArray("WagerSelections")
		for _, vv := range wagers {

			sid := vv.GetInt("SelectionId")
			if sid == 8 { // 主
				homeOdds = vv.GetFloat64("Odds")
			}

			if sid == 9 { // 客
				awayOdds = vv.GetFloat64("Odds")
			}
		}

		data[v.Get("EventId").String()] = popularEvents{
			EventName:    string(v.GetStringBytes("Competition", "CompetitionName")),
			PlatformName: "IM体育",
			HomeTeam:     string(v.GetStringBytes("HomeTeam")),
			AwayTeam:     string(v.GetStringBytes("AwayTeam")),
			Project:      string(rcs[0].GetStringBytes("SportName")),
			EventAt:      helper.StrToTime(string(v.GetStringBytes("EventDate")), loc).Unix(),
			HomeOdds:     homeOdds,
			AwayOdds:     awayOdds,
			PlatformID:   imPlatformID,
		}
	}

	return data, nil
}

// 获取im体育 timestamp
func imSportTimestamp(now time.Time) string {

	param := map[string]interface{}{
		"pid":    imPlatformID,
		"tester": "1", // 0 正式用户  1 测试用户
		"s8":     strconv.FormatInt(now.Unix(), 10),
		"ms8":    strconv.FormatInt(now.UnixMilli(), 10),
	}

	for k, v := range pCfg[imPlatformID] {
		param[k] = v
	}

	return imtyPack(param)
}

func imtyPack(param map[string]interface{}) string {

	has := md5.Sum([]byte(param["key"].(string)))
	key := fmt.Sprintf("%x", has)

	lens := len(key) / 2
	md5raw := ""
	for i := 0; i < lens; i++ {
		hexByte, _ := hex.DecodeString(substring(key, i*2, (i*2)+2))
		md5raw = md5raw + string(hexByte)
	}

	nt, _ := strconv.ParseInt(param["s8"].(string), 10, 64)
	preDayTime := nt - 12*3600

	nanoStr := param["ms8"].(string)
	timeStr := time.Unix(preDayTime, 0).Format("2006-01-02 15:04:05")

	timeStamp := timeStr + "." + substring(nanoStr, len(nanoStr)-3, len(nanoStr))
	aesByte := aesEcbEncrypt([]byte(timeStamp), []byte(md5raw))

	return base64.StdEncoding.EncodeToString(aesByte)
}

//获取source的子串,如果start小于0或者end大于source长度则返回""
//start:开始index，从0开始，包括0
//end:结束index，以end结束，但不包括end
func substring(source string, start int, end int) string {

	var r = []rune(source)
	length := len(r)

	if start < 0 || end > length || start > end {
		return ""
	}

	if start == 0 && end == length {
		return source
	}

	return string(r[start:end])
}

// aes ecd
func aesEcbEncrypt(data, key []byte) []byte {

	block, _ := aes.NewCipher(key)
	data = pkcs5Padding(data, block.BlockSize())
	decrypted := make([]byte, len(data))
	size := block.BlockSize()

	for bs, be := 0, size; bs < len(data); bs, be = bs+size, be+size {
		block.Encrypt(decrypted[bs:be], data[bs:be])
	}

	return decrypted
}

func pkcs5Padding(ciphertext []byte, blockSize int) []byte {

	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)

	return append(ciphertext, padtext...)
}
