package withdraw

var payment = map[string]Payment{}

type Payment interface {
	query(withdraw) (int, error)
}

func InitWithdraw() {
	payment = map[string]Payment{
		"902366599681860880": newUZ(),  // uz withdraw
		"902678228220935843": newPN(),  // pn withdraw
		"902872228948461756": newG7(),  // g7 withdraw
		"100821864946101590": newW(),   // w withdraw
		"347716135648705292": newF(),   // fpay payout
		"678749588326497691": newYFB(), // 优付宝 payout
	}
}

// withdraw 提现订单查询structure
type withdraw struct {
	ID  string `db:"id" json:"id"`   // 我方订单号
	PID string `db:"pid" json:"pid"` // payment id
	OID string `db:"oid" json:"oid"` // 3方订单号
}
