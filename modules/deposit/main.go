package deposit

import (
	"time"

	"cronTask/contrib/conn"
	"cronTask/modules/common"

	g "github.com/doug-martin/goqu/v9"
	"github.com/fluent/fluent-logger-golang/fluent"
	"github.com/jmoiron/sqlx"
)

var (
	db       *sqlx.DB
	zlog     *fluent.Fluent
	esPrefix string
	dialect  = g.Dialect("mysql")
)

func Parse(endpoints []string, path string) {

	conf := common.ConfParse(endpoints, path)
	esPrefix = conf.EsPrefix
	// 初始化db
	db = conn.InitDB(conf.Db.Master.Addr, conf.Db.Master.MaxIdleConn, conf.Db.Master.MaxIdleConn)

	// 处理中的存款订单1天之后变为失败
	now := time.Now().Unix()
	ex := g.Ex{
		"created_at": g.Op{"lt": now - 86400},
		"state":      common.DepositConfirming,
	}
	record := g.Record{"state": common.DepositCancelled}
	query, _, _ := dialect.Update("tbl_deposit").Set(record).Where(ex).ToSQL()
	_, err := db.Exec(query)
	if err != nil {
		common.Log("deposit", "finance execute query [%s] error: [%v]", query, err)
	}

	// 失败的存款订单15天之后删除
	ex = g.Ex{
		"created_at": g.Op{"lt": now - 15*86400},
		"state":      common.DepositCancelled,
	}
	query, _, _ = dialect.Delete("tbl_deposit").Where(ex).ToSQL()
	_, err = db.Exec(query)
	if err != nil {
		common.Log("deposit", "finance execute query [%s] error: [%v]", query, err)
	}
}
