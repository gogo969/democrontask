package reports

import (
	"context"
	"cronTask/contrib/helper"
	"fmt"
	"github.com/olivere/elastic/v7"
	"time"
)

type VaildOddsReport struct {
	Id             string  `json:"id" db:"id"`
	ReportMonth    int64   `json:"report_month" db:"report_month"`
	Uid            string  `json:"uid" db:"uid"`
	ApiName        string  `json:"api_name" db:"api_name"`
	ApiType        string  `json:"api_type" db:"api_type"`
	HandicapType   string  `json:"handicap_type" db:"handicap_type"`
	Odds           float64 `json:"odds" db:"odds"`
	ValidBetAmount float64 `json:"valid_bet_amount" db:"valid_bet_amount"`
}

var (
	vaildOddsReports []VaildOddsReport
)

func vaildReport() {

	thisMonSt := helper.MonthTST(0, loc).UnixMilli()
	thisMonEt := helper.MonthTET(0, loc).UnixMilli()
	stNow := time.Unix(thisMonSt/1000, 0).In(loc)  //2017-08-30 16:19:19 +0800 CST
	timeStr := stNow.Format("2006-01-02 15:04:05") //2015-06-15 08:52:

	//游戏报表-投注时间-日报
	query := fmt.Sprintf(`SELECT CONCAT_WS( '|', uid, api_type, odds) AS id, 0 AS report_month, uid, api_name,
api_type, handicap_type, odds, sum( valid_bet_amount ) valid_bet_amount FROM ( SELECT id, settle_time, uid, api_name, 
api_type, CASE WHEN handicap_type IN ( 'E', 'EU', 'EURO', 'Decimal Odds(Đĩa Châu Âu)', 'kèo châu Âu', 'Decimal Odds(欧洲盘)', '欧盘', 'DE' ) 
OR handicap = 'EURO' THEN 'EU' WHEN handicap_type IN ( 'MY', 'M', 'Malay Odds(马来盘)', 'Malay Odds(Đĩa Malay)', 'MALAY' ) 
THEN 'MY' WHEN handicap_type IN ( 'HK', 'China Odds(Đĩa Trung Quốc)', 'H' ) THEN 'HK' WHEN handicap_type IN ( 'ID', 'I' ) 
THEN 'ID' WHEN handicap_type IN ( 'American Odds(Kèo Mỹ)' ) THEN 'AM' ELSE 'OTHER' END handicap_type, odds, valid_bet_amount 
FROM ( SELECT id, settle_time / 1000 AS settle_time, uid, valid_bet_amount, api_type, api_name, game_type, handicap, handicap_type, 
IF ( game_type IN ( 6, 8 ), odds, 0.0000 ) odds FROM ym_game_record WHERE flag = '1' AND settle_time >= %d and settle_time < %d GROUP BY id ) c ) d GROUP BY
handicap_type, odds, api_name, uid, api_type`, thisMonSt, thisMonEt)
	fmt.Println(query)
	err := pullDb.Select(&vaildOddsReports, query)
	if err != nil {
		fmt.Println("t_vaild_odds_month select error = ", err)
		return
	}

	bulkRequest := reportEs.Bulk()
	for _, v := range vaildOddsReports {

		v.Id = v.Id + "|" + timeStr
		v.ReportMonth = thisMonSt / 1000
		fmt.Println(v)
		req := elastic.NewBulkUpdateRequest().Index(esPrefix + "t_vaild_odds_month").Id(v.Id).Doc(v).DocAsUpsert(true)
		bulkRequest = bulkRequest.Add(req)
	}
	//批量执行es更新或插入
	bulkResponse, err := bulkRequest.Do(context.Background())
	if err != nil {
		fmt.Println(err)
	}
	if bulkResponse != nil {
		//fmt.Println("t_vaild_odds_month response", bulkResponse)
	}
}
