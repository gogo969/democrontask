package reports

import (
	"cronTask/contrib/conn"
	"cronTask/modules/common"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/olivere/elastic/v7"
	"time"
)

//报表
var (
	pullDb   *sqlx.DB
	reportEs *elastic.Client
	lang     string
	loc      *time.Location
	esPrefix string
)

type Member struct {
	Uid                string  `json:"uid" db:"uid"`
	Prefix             string  `json:"prefix" db:"prefix"`
	AgencyType         string  `json:"agency_type" db:"agency_type"`
	Username           string  `json:"username" db:"username"`
	AgencyUid          int64   `json:"agency_uid" db:"agency_uid"`
	AgencyName         string  `json:"agency_name" db:"agency_name"`
	SourceId           string  `json:"source_id" db:"source_id"`
	RegUrl             string  `json:"reg_url" db:"reg_url"`
	Level              string  `json:"level" db:"level"`
	IsAgent            string  `json:"is_agent" db:"is_agent"`
	State              string  `json:"state" db:"state"`
	FirstDepositAt     int     `json:"first_deposit_at" db:"first_deposit_at"`
	FirstDepositAmount float64 `json:"first_deposit_amount" db:"first_deposit_amount"`
	LastLoginAt        int     `json:"last_login_at" db:"last_login_at"`
	CreatedAt          int     `json:"created_at" db:"created_at"`
}

func Parse(endpoints []string, path, flag string) {

	conf := common.ConfReportParse(endpoints, path)

	lang = conf.Lang
	esPrefix = conf.EsPrefix
	if lang == "cn" {
		loc, _ = time.LoadLocation("Asia/Shanghai")
	} else if lang == "vn" || lang == "th" {
		loc, _ = time.LoadLocation("Asia/Bangkok")
	}

	// 初始化注单db
	pullDb = conn.InitDB(conf.Db.Bet.Addr, conf.Db.Bet.MaxIdleConn, conf.Db.Bet.MaxIdleConn)
	// 初始化es
	reportEs = conn.InitES(conf.Es.Host, conf.Es.Username, conf.Es.Password)

	timeSt := time.Now().UnixMilli()
	reportTask(flag)
	fmt.Println("report use milliseconds", time.Now().UnixMilli()-timeSt)
}

func reportTask(flag string) {

	switch flag {
	case "vaild": //有效投注
		vaildReport()
	default:
		break
	}
}
