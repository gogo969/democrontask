package upgrade

import (
	"cronTask/contrib/helper"
	"cronTask/modules/common"
	"fmt"
	g "github.com/doug-martin/goqu/v9"
	"github.com/shopspring/decimal"
	"time"
)

type mWaterFlow struct {
	UID                 string
	Username            string
	AgencyType          string
	ParentUID           string
	ParentName          string
	TopUID              string
	TopName             string
	IsDowngrade         int
	TotalDeposit        string
	TotalWaterFlow      string
	RelegationWaterFlow string
	ReturnDeposit       string
	ReturnWaterFlow     string
	EarlyMonthPacket    string
}

func mWaterFlowToMap(m mWaterFlow) map[string]interface{} {

	data := map[string]interface{}{}

	data["uid"] = m.UID
	data["is_downgrade"] = m.IsDowngrade
	data["username"] = m.Username
	data["total_deposit"] = m.TotalDeposit
	data["total_water_flow"] = m.TotalWaterFlow
	data["return_deposit"] = m.ReturnDeposit
	data["return_water_flow"] = m.ReturnWaterFlow
	data["relegation_water_flow"] = m.RelegationWaterFlow

	return data
}

func getParameters(mb common.Member, mtp map[string]mValidBetAmount) (decimal.Decimal, decimal.Decimal) {

	var (
		fmValidBet     decimal.Decimal //累计总流水
		fmValidDeposit decimal.Decimal //累计总存款
	)

	// 无累计流水记录，默认为0
	mt, ok := mtp[mb.Username]
	if !ok {
		fmValidBet = decimal.NewFromFloat(0.0000)
		fmValidDeposit = decimal.NewFromFloat(0.0000)
	}
	// 会员累计流水
	fmValidBet = decimal.NewFromFloat(mt.ValidBetAmount)
	// 会员累计存款
	fmValidDeposit = decimal.NewFromFloat(mt.DepositAmount)

	return fmValidBet, fmValidDeposit
}

// vip1-9处理
func processVIP(mb common.Member, mtp map[string]mValidBetAmount) error {

	// 获取会员信息
	fmValidBet, fmValidDeposit := getParameters(mb, mtp)

	m := mWaterFlow{
		UID:                 mb.UID,
		Username:            mb.Username,
		AgencyType:          mb.AgencyType,
		ParentUID:           mb.ParentUid,
		ParentName:          mb.ParentName,
		TopUID:              mb.TopUid,
		TopName:             mb.TopName,
		IsDowngrade:         0,
		TotalWaterFlow:      fmValidBet.Truncate(4).String(),
		TotalDeposit:        fmValidDeposit.Truncate(4).String(),
		RelegationWaterFlow: "0.0000",
		ReturnDeposit:       "0.0000",
		ReturnWaterFlow:     "0.0000",
		EarlyMonthPacket:    ml[mb.Level].EarlyMonthPacket,
	}

	defer func() {
		vAll = append(vAll, m)
	}()

	for level := mb.Level; level < common.Vip10; level++ {

		// 升级所需累计流水
		fnuBet, err := decimal.NewFromString(ml[level+1].UpgradeRecord)
		if err != nil {
			common.Log("upgrade", "error : %v", err)
			return err
		}

		// 升级所需累计存款
		fnuDeposit, err := decimal.NewFromString(ml[level+1].UpgradeDeposit)
		if err != nil {
			common.Log("upgrade", "error : %v", err)
			return err
		}

		common.Log("upgrade", "会员名: %s VIP[%d],累计投注: %s, 升级累计投注: %s, 累计存款: %s, 升级累计存款: %s\n",
			mb.Username, level, fmValidBet.String(), fnuBet.String(), fmValidDeposit.String(), fnuDeposit.String())
		// 累计流水和累计存款 大于等于下一等级升级流水和升级存款，会员升级
		if fmValidBet.Cmp(fnuBet) >= 0 && fmValidDeposit.Cmp(fnuDeposit) >= 0 {
			common.Log("upgrade", "会员名: %s升级,累计投注: %s\n", mb.Username)
			membersUpgrade(m, level)
		} else {
			return nil
		}
	}

	return nil
}

// 会员升级
func membersUpgrade(m mWaterFlow, level int) {

	fmt.Printf("up name:%s,vip%d,index:%d\n", m.Username, level, i)
	i++
	if level >= common.Vip10 {
		return
	}

	uniqueKey := fmt.Sprintf("%s:%d", m.Username, level)
	// 去重判断
	if _, ok := unique[uniqueKey]; ok {
		common.Log("upgrade", "重复升级 : %v", m.Username)
		return
	}

	// 加入去重map
	unique[uniqueKey] = true

	balance, err := common.MemberBalance(db, m.UID, prefix)
	if err != nil {
		common.Log("upgrade", "MemberBalance error : %v", err.Error())
	}

	bonus, _ := decimal.NewFromString(ml[level+1].UpgradeGift)
	balanceAfter := balance.Add(bonus)

	id := helper.GenLongId()
	rd := common.MemberLevelRecord{
		ID:                  id,
		UID:                 m.UID,
		Username:            m.Username,                     //会员账号
		BeforeLevel:         level,                          //调整前会员等级
		AfterLevel:          level + 1,                      //调整后会员等级
		TotalDeposit:        m.TotalDeposit,                 //升级存款
		TotalWaterFlow:      m.TotalWaterFlow,               //升级流水
		RelegationWaterFlow: m.RelegationWaterFlow,          //保级流水
		Ty:                  common.MemberLevelUpgrade,      //会员等级调整类型 201升级 202保级 203降级 204回归
		CreatedAt:           uint64(time.Now().UnixMilli()), //操作时间
		CreatedUid:          "0",                            //操作人uid
		CreatedName:         "admin",                        //操作人名
	}

	dividend := g.Record{
		"id":          id,
		"uid":         m.UID,
		"pid":         0,
		"username":    m.Username,
		"prefix":      prefix,
		"ty":          common.DividendUpgrade,
		"water_limit": 2,
		"water_flow":  ml[level+1].UpgradeGift,
		"amount":      ml[level+1].UpgradeGift,
		"level":       level + 1,
		"parent_uid":  m.ParentUID,
		"parent_name": m.ParentName,
		"top_uid":     m.TopUID,
		"top_name":    m.TopName,
		"remark":      fmt.Sprintf("vip%dupgrade", level+1),
		"apply_at":    time.Now().UnixMilli(),
		"apply_uid":   "0",
		"apply_name":  "taskUpgrade",
		"automatic":   0, //自动发放
		"review_at":   time.Now().Unix(),
		"review_uid":  "0",
		"review_name": "taskUpgrade",
		"state":       common.DividendReviewPass,
	}

	tx, err := db.Begin()
	if err != nil {
		common.Log("upgrade", "error : %v", err)
		return
	}

	ex := g.Ex{
		"uid":   m.UID,
		"level": g.Op{"lt": common.Vip10}, //会员等级小于vip10
	}
	record := g.Record{
		"level": g.L("level+1"),
	}
	query, _, _ := dialect.Update("tbl_members").Set(record).Where(ex).ToSQL()
	res, err := tx.Exec(query)
	if err != nil {
		_ = tx.Rollback()
		common.Log("upgrade", "query: %s error: %v", query, err)
		return
	}

	if af, _ := res.RowsAffected(); af == 0 {
		_ = tx.Rollback()
		common.Log("upgrade", "query: %s ,affected 0 row !", query)
		return
	}

	query, _, _ = dialect.Insert("tbl_member_dividend").Rows(dividend).ToSQL()
	_, err = tx.Exec(query)
	if err != nil {
		_ = tx.Rollback()
		common.Log("upgrade", "error : %v", err)
		return
	}

	record = g.Record{
		"balance": g.L(fmt.Sprintf("balance+%s", ml[level+1].UpgradeGift)),
	}
	query, _, _ = dialect.Update("tbl_members").Set(record).Where(g.Ex{"uid": m.UID}).ToSQL()
	fmt.Println(query)
	_, err = tx.Exec(query)
	if err != nil {
		_ = tx.Rollback()
		common.Log("upgrade", "error : %v", err)
		return
	}

	trans := common.MemberTransaction{
		AfterAmount:  balanceAfter.String(),
		Amount:       bonus.String(),
		BeforeAmount: balance.String(),
		BillNo:       id,
		CreatedAt:    time.Now().UnixMilli(),
		ID:           id,
		CashType:     common.DividendPromo,
		UID:          m.UID,
		Username:     m.Username,
		Prefix:       prefix,
	}
	query, _, _ = dialect.Insert("tbl_balance_transaction").Rows(trans).ToSQL()
	_, err = tx.Exec(query)
	if err != nil {
		_ = tx.Rollback()
		common.Log("upgrade", "query: %s, error: %v", query, err)
		return
	}

	query, _, _ = dialect.Insert("tbl_member_level_record").Rows(rd).ToSQL()
	_, err = tx.Exec(query)
	if err != nil {
		_ = tx.Rollback()
		common.Log("upgrade", "query: %s, error: %v", query, err)
		return
	}

	err = tx.Commit()
	if err != nil {
		common.Log("upgrade", "error : %v", err)
		return
	}

	title := "Thông Báo Khuyến Mãi Thăng Cấp VIP"
	content := fmt.Sprintf("Quý Khách Của P3 Thân Mến :\n    Chúc Mừng Bạn Đã Thăng Cấp VIP: %d !, Khuyến Mãi Thăng Cấp Đã Được Tặng Vào Tài Khoản Của Bạn,Vui Lòng Kiểm Tra Ngay ,Nếu Bạn Có Bất Cứ Thắc Mắc Vấn Đề Gì  Vui Lòng Liên Hệ CSKH Để Biết Thêm Chi Tiết .【P3】Chúc Bạn Thăng Cấp Mạnh Mẽ.", level+1)
	err = messageSend("0", title, "", content, "system", prefix, 0, 0, 1, []string{m.Username})
	if err != nil {
		common.Log("upgrade", "error : %v", err)
	}
}
