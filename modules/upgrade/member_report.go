package upgrade

import (
	"fmt"
	g "github.com/doug-martin/goqu/v9"
)

// 会员日报表累计流水查询
func memberReportValidBetAmounts(ex g.Ex) (map[string]mValidBetAmount, error) {

	data := map[string]mValidBetAmount{}
	var num int
	query, _, _ := dialect.From("tbl_report_agency").
		Select(g.COUNT("uid").As("num")).Where(ex).Order(g.C("uid").Desc()).ToSQL()
	fmt.Println(query)
	err := reportDb.Get(&num, query)
	if num > 0 {
		var list []mValidBetAmount
		query, _, _ = dialect.From("tbl_report_agency").
			Select(g.C("username").As("user_name"), g.SUM("deposit_amount").As("deposit_amount"),
				g.SUM("valid_bet_amount").As("valid_bet_amount"),
			).Where(ex).GroupBy("username").Order(g.C("username").Desc()).ToSQL()
		fmt.Println(query)
		err = reportDb.Select(&list, query)
		if err != nil {
			return data, err
		}
		for _, v := range list {
			obj := mValidBetAmount{
				DepositAmount:  v.DepositAmount,
				ValidBetAmount: v.ValidBetAmount,
			}
			data[v.UserName] = obj
		}
		return data, nil
	}
	return data, nil
}
