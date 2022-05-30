package withdraw

import (
	"fmt"
	"time"

	"cronTask/contrib/conn"
	"cronTask/modules/common"

	g "github.com/doug-martin/goqu/v9"
	"github.com/jmoiron/sqlx"
)

var (
	db      *sqlx.DB
	dialect = g.Dialect("mysql")
)

func Parse(endpoints []string, path string) {

	conf := common.ConfParse(endpoints, path)
	// 初始化db
	db = conn.InitDB(conf.Db.Master.Addr, conf.Db.Master.MaxIdleConn, conf.Db.Master.MaxIdleConn)

	// 查询在处理中的订单
	var data []withdraw
	ex := g.Ex{
		"pid":        g.Op{"neq": "0"},
		"created_at": g.Op{"lt": time.Now().Unix() - 600},
		"state":      common.WithdrawDealing,
	}
	query, _, _ := dialect.From("tbl_withdraw").Select("id", "pid", "oid").Where(ex).Limit(1).ToSQL()
	err := db.Select(&data, query)
	if err != nil {
		common.Log("withdraw", "query withdrawal order error : %v", err)
		return
	}

	// 初始化财务配置
	InitWithdraw()

	// 查询订单状态
	var ids []string
	for _, v := range data {
		fn, ok := payment[v.PID]
		if !ok {
			fmt.Println("payment not match, pid: ", v.PID)
			continue
		}

		state, err := fn.query(v)
		if err != nil {
			fmt.Println("query withdraw order error: ", err.Error())
			continue
		}

		// 将订单状态为代付失败的, 保存到ids中, 后面直接一次修改订单状态
		if state == common.WithdrawAutoPayFailed {
			ids = append(ids, v.ID)
		}
	}

	if len(ids) == 0 {
		return
	}

	record := g.Record{"state": common.WithdrawAutoPayFailed}
	ex = g.Ex{"id": ids, "state": common.WithdrawDealing}
	query, _, _ = dialect.Update("tbl_withdraw").Set(record).Where(ex).ToSQL()
	_, err = db.Exec(query)
	if err != nil {
		fmt.Println("update withdraw order state error: ", err.Error())
	}
}
