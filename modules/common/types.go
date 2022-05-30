package common

import "github.com/shopspring/decimal"

type Member struct {
	UID        string `db:"uid" json:"uid"`
	Username   string `db:"username" json:"username"`       //会员名
	Tester     string `db:"tester" json:"tester"`           //1正式 0测试
	AgencyType string `db:"agency_type" json:"agency_type"` //代理类型
	Level      int    `db:"level" json:"level"`             //会员等级
	TopUid     string `db:"top_uid" json:"top_uid"`         //总代uid
	TopName    string `db:"top_name" json:"top_name"`       //总代代理
	ParentUid  string `db:"parent_uid" json:"parent_uid"`   //上级uid
	ParentName string `db:"parent_name" json:"parent_name"` //上级代理
}

// 会员等级调整记录
type MemberLevelRecord struct {
	ID                  string `db:"id" json:"id"`
	UID                 string `db:"uid" json:"uid"`                                     //会员id
	Username            string `db:"username" json:"username"`                           //会员账号
	BeforeLevel         int    `db:"before_level" json:"before_level"`                   //调整前会员等级
	AfterLevel          int    `db:"after_level" json:"after_level"`                     //调整后会员等级
	TotalDeposit        string `db:"total_deposit" json:"total_deposit"`                 //累计存款
	TotalWaterFlow      string `db:"total_water_flow" json:"total_water_flow"`           //累计流水
	RelegationWaterFlow string `db:"relegation_water_flow" json:"relegation_water_flow"` //累计保级流水
	Ty                  int    `db:"ty" json:"ty"`                                       //会员等级调整类型
	CreatedAt           uint64 `db:"created_at" json:"created_at"`                       //操作时间
	CreatedUid          string `db:"created_uid" json:"created_uid"`                     //操作人uid
	CreatedName         string `db:"created_name" json:"created_name"`                   //操作人名
}

//账变表
type MemberTransaction struct {
	AfterAmount  string `db:"after_amount"`  //账变后的金额
	Amount       string `db:"amount"`        //用户填写的转换金额
	BeforeAmount string `db:"before_amount"` //账变前的金额
	BillNo       string `db:"bill_no"`       //转账|充值|提现ID
	CashType     int    `db:"cash_type"`     //0:转入1:转出2:转入失败补回3:转出失败扣除4:存款5:提现
	CreatedAt    int64  `db:"created_at"`    //
	ID           string `db:"id"`            //
	UID          string `db:"uid"`           //用户ID
	Username     string `db:"username"`      //用户名
	Prefix       string `db:"prefix"`        //站点前缀
}

type Message struct {
	ID       string `json:"id"`        //会员站内信id
	MsgID    string `json:"msg_id"`    //站内信id
	Username string `json:"username"`  //会员名
	Title    string `json:"title"`     //标题
	SubTitle string `json:"sub_title"` //标题
	Content  string `json:"content"`   //内容
	IsTop    int    `json:"is_top"`    //0不置顶 1置顶
	IsVip    int    `json:"is_vip"`    //0非vip站内信 1vip站内信
	Ty       int    `json:"ty"`        //1站内消息 2活动消息
	IsRead   int    `json:"is_read"`   //是否已读 0未读 1已读
	SendName string `json:"send_name"` //发送人名
	SendAt   int64  `json:"send_at"`   //发送时间
	Prefix   string `json:"prefix"`    //商户前缀
}

type MemberRebate struct {
	UID              string `db:"uid" json:"uid"`
	ZR               string `db:"zr" json:"zr"`                                 //真人返水
	QP               string `db:"qp" json:"qp"`                                 //棋牌返水
	TY               string `db:"ty" json:"ty"`                                 //体育返水
	DJ               string `db:"dj" json:"dj"`                                 //电竞返水
	DZ               string `db:"dz" json:"dz"`                                 //电游返水
	CP               string `db:"cp" json:"cp"`                                 //彩票返水
	FC               string `db:"fc" json:"fc"`                                 //斗鸡返水
	BY               string `db:"by" json:"by"`                                 //捕鱼返水
	CgOfficialRebate string `db:"cg_official_rebate" json:"cg_official_rebate"` //CG官方彩返点
	CgHighRebate     string `db:"cg_high_rebate" json:"cg_high_rebate"`         //CG高频彩返点
	CreatedAt        uint32 `db:"created_at" json:"created_at"`
	ParentUID        string `db:"parent_uid" json:"parent_uid"`
	Prefix           string `db:"prefix" json:"prefix"`
}

type MemberRebateResult_t struct {
	ZR               decimal.Decimal
	QP               decimal.Decimal
	TY               decimal.Decimal
	DZ               decimal.Decimal
	DJ               decimal.Decimal
	CP               decimal.Decimal
	FC               decimal.Decimal
	BY               decimal.Decimal
	CGOfficialRebate decimal.Decimal
	CGHighRebate     decimal.Decimal
}
