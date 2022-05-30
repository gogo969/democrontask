package transfer

import (
	"cronTask/contrib/conn"
	"cronTask/contrib/helper"
	"cronTask/modules/common"
	"errors"
	"fmt"
	g "github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/shopspring/decimal"
	"time"
)

var (
	db                 *sqlx.DB
	prefix             string
	dialect            = g.Dialect("mysql")
	colsAgencyTransfer = helper.EnumFields(AgencyTransfer{})
	colsMembersTree    = helper.EnumFields(MembersTree{})
	colsMember         = helper.EnumFields(common.Member{})
	colsMemberRebate   = helper.EnumFields(common.MemberRebate{})
)

// 代理团队转代
type AgencyTransfer struct {
	ID           string `json:"id" db:"id"`
	Prefix       string `json:"prefix" db:"prefix"`
	UID          string `json:"uid" db:"uid"`
	Username     string `json:"username" db:"username"`
	BeforeUid    string `json:"before_uid" db:"before_uid"`
	BeforeName   string `json:"before_name" db:"before_name"`
	AfterUid     string `json:"after_uid" db:"after_uid"`
	AfterName    string `json:"after_name" db:"after_name"`
	Status       int    `json:"status" db:"status"`
	ApplyAt      uint32 `json:"apply_at" db:"apply_at"`
	ApplyUid     string `json:"apply_uid" db:"apply_uid"`
	ApplyName    string `json:"apply_name" db:"apply_name"`
	ReviewAt     uint32 `json:"review_at" db:"review_at"`
	ReviewUid    string `json:"review_uid" db:"review_uid"`
	ReviewName   string `json:"review_name" db:"review_name"`
	Remark       string `json:"remark" db:"remark"`
	ReviewRemark string `json:"review_remark" db:"review_remark"`
}

type MembersTree struct {
	Ancestor   string `db:"ancestor"`
	Descendant string `db:"descendant"`
	Lvl        int    `db:"lvl"`
}

type AgencyTransferRecord struct {
	Id            string `json:"id" db:"id"`
	Flag          int    `json:"flag" db:"flag"`
	Uid           string `json:"uid" db:"uid"`
	Username      string `json:"username" db:"username"`
	Type          string `json:"type" db:"type"`
	BeforeUid     string `json:"before_uid" db:"before_uid"`
	BeforeName    string `json:"before_name" db:"before_name"`
	AfterUid      string `json:"after_uid" db:"after_uid"`
	AfterName     string `json:"after_name" db:"after_name"`
	Remark        string `json:"remark" db:"remark"`
	UpdatedAt     int64  `json:"updated_at" db:"updated_at"`
	UpdatedUid    string `json:"updated_uid" db:"updated_uid"`
	UpdatedName   string `json:"updated_name" db:"updated_name"`
	BeforeTopUid  string `json:"before_top_uid" db:"before_top_uid"`
	BeforeTopName string `json:"before_top_name" db:"before_top_name"`
	AfterTopUid   string `json:"after_top_uid" db:"after_top_uid"`
	AfterTopName  string `json:"after_top_name" db:"after_top_name"`
	Prefix        string `json:"prefix" db:"prefix"`
}

func Parse(endpoints []string, path string) {

	conf := common.ConfParse(endpoints, path)
	prefix = conf.Prefix
	// 初始化db
	db = conn.InitDB(conf.Db.Master.Addr, conf.Db.Master.MaxIdleConn, conf.Db.Master.MaxIdleConn)

	transferWork()
}

func transferWork() {

	var data []AgencyTransfer
	ex := g.Ex{
		"status": 2,
		"prefix": prefix,
	}
	query, _, _ := dialect.From("tbl_agency_transfer_apply").Select(colsAgencyTransfer...).Where(ex).ToSQL()
	fmt.Println(query)
	err := db.Select(&data, query)
	if err != nil {
		fmt.Printf("query : %s \n error : %s \n", query, err.Error())
		return
	}

	if len(data) == 0 {
		fmt.Println("no transfer apply records")
		return
	}

	for _, v := range data {
		mb, err := common.MemberFindOne(db, v.Username, prefix)
		if err != nil {
			fmt.Printf("MemberFindOne  error : %s \n", err.Error())
			return
		}

		destMb, err := common.MemberFindOne(db, v.AfterName, prefix)
		if err != nil {
			fmt.Printf("MemberFindOne  error : %s \n", err.Error())
			return
		}

		err = transferRebateRateCheck(mb, destMb)
		if err != nil {
			fmt.Printf("transferRebateRateCheck src [%s] dest [%s] error : %s \n", mb.Username, destMb.Username, err.Error())
			return
		}

		// 已经在该代理下
		if mb.ParentName == destMb.Username {
			continue
		}

		transferAgs(v.UID, destMb)
	}
}

func transferRebateRateCheck(mb, destMb common.Member) error {

	src, err := memberRebateFindOne(mb.UID)
	if err != nil {
		return err
	}

	dest, err := memberRebateFindOne(destMb.UID)
	if err != nil {
		return err
	}

	if src.TY.GreaterThan(dest.TY) || //体育返水比例
		src.ZR.GreaterThan(dest.ZR) || //真人返水比例
		src.QP.GreaterThan(dest.QP) || //棋牌返水比例
		src.DJ.GreaterThan(dest.DJ) || //电竞返水比例
		src.DZ.GreaterThan(dest.DZ) || //电子返水比例
		src.CP.GreaterThan(dest.CP) || //彩票返水比例
		src.BY.GreaterThan(dest.BY) || //捕鱼返水比例
		src.CGHighRebate.GreaterThan(dest.CGHighRebate) || //cg彩票高频彩返点比例
		src.CGOfficialRebate.GreaterThan(dest.CGOfficialRebate) { //cg彩票官方彩返点比例
		return errors.New(helper.RebateOutOfRange)
	}

	return nil
}

func memberRebateFindOne(uid string) (common.MemberRebateResult_t, error) {

	data := common.MemberRebate{}
	res := common.MemberRebateResult_t{}

	t := dialect.From("tbl_member_rebate_info")
	query, _, _ := t.Select(colsMemberRebate...).Where(g.Ex{"uid": uid}).Limit(1).ToSQL()
	err := db.Get(&data, query)
	if err != nil {
		fmt.Printf("memberRebateFindOne  error : %s, sql : %s \n", err.Error(), query)
		return res, err
	}

	res.ZR, _ = decimal.NewFromString(data.ZR)
	res.QP, _ = decimal.NewFromString(data.QP)
	res.TY, _ = decimal.NewFromString(data.TY)
	res.DJ, _ = decimal.NewFromString(data.DJ)
	res.DZ, _ = decimal.NewFromString(data.DZ)
	res.CP, _ = decimal.NewFromString(data.CP)
	res.FC, _ = decimal.NewFromString(data.FC)
	res.BY, _ = decimal.NewFromString(data.BY)
	res.CGOfficialRebate, _ = decimal.NewFromString(data.CgOfficialRebate)
	res.CGHighRebate, _ = decimal.NewFromString(data.CgHighRebate)

	res.ZR = res.ZR.Truncate(1)
	res.QP = res.QP.Truncate(1)
	res.TY = res.TY.Truncate(1)
	res.DJ = res.DJ.Truncate(1)
	res.DZ = res.DZ.Truncate(1)
	res.CP = res.CP.Truncate(1)
	res.FC = res.FC.Truncate(1)
	res.BY = res.BY.Truncate(1)

	res.CGOfficialRebate = res.CGOfficialRebate.Truncate(2)
	res.CGHighRebate = res.CGHighRebate.Truncate(2)

	return res, nil
}

func transferAgs(uid string, destMb common.Member) {

	var data []MembersTree
	ex := g.Ex{
		"ancestor": uid,
		"prefix":   prefix,
	}
	query, _, _ := dialect.From("tbl_members_tree").
		Select(colsMembersTree...).Where(ex).Order(g.C("lvl").Asc()).ToSQL()
	fmt.Println(query)
	err := db.Select(&data, query)
	if err != nil {
		fmt.Printf("query : %s \n error : %s \n", query, err.Error())
		return
	}

	var uids []string
	lvMp := map[int][]string{}
	for _, v := range data {
		uids = append(uids, v.Descendant)
		lvMp[v.Lvl] = append(lvMp[v.Lvl], v.Descendant)
	}

	//fmt.Println(uids)
	//fmt.Println(lvMp)

	mbs, err := membersByUid(uids)
	if err != nil {
		fmt.Printf("membersByUid error , uids : %v \n error : %s \n", query, err.Error())
		return
	}

	//fmt.Println(mbs)

	l := len(lvMp)
	for i := 0; i < l; i++ {
		if i == 0 {
			ids := lvMp[0]
			if len(ids) != 1 {
				fmt.Printf("member tree error , ids : %v \n", ids)
				return
			}

			mb := mbs[ids[0]]
			if mb.ParentName == destMb.Username {
				return
			}

			fmt.Println("transferAg : ", mb.UID, mb.Username, mb.ParentUid, mb.ParentName, destMb.TopUid, destMb.TopName)
			err = transferAg(mb.UID, destMb.UID, destMb.Username, destMb.TopUid, destMb.TopName, mb.ParentUid, mb.ParentName,
				mb.TopUid, mb.TopName, mb.Username, mb.AgencyType, destMb.Tester)
			if err != nil {
				fmt.Printf("transferAg error : %s \n", err.Error())
				return
			}

			continue
		}

		ids := lvMp[i]
		if len(ids) == 0 {
			fmt.Printf("member tree error , ids : %v \n", ids)
			return
		}

		for _, vv := range ids {
			mb := mbs[vv]
			fmt.Println("transferAg : ", mb.UID, mb.Username, mb.ParentUid, mb.ParentName, destMb.TopUid, destMb.TopName)
			err = transferAg(mb.UID, mb.ParentUid, mb.ParentName, destMb.TopUid, destMb.TopName, mb.ParentUid, mb.ParentName,
				mb.TopUid, mb.TopName, mb.Username, mb.AgencyType, destMb.Tester)
			if err != nil {
				fmt.Printf("transferAg error : %s \n", err.Error())
				return
			}
		}
	}
}

func transferAg(uid, parentUid, parentName, topUid, topName, oldParentUid, oldParentName, oldTopUid, oldTopName, username, agencyType, tester string) error {

	tx, err := db.Begin() // 开启事务
	if err != nil {
		return err
	}

	ex := g.Ex{
		"uid":    uid,
		"prefix": prefix,
	}
	record := g.Record{
		"parent_uid":  parentUid,
		"parent_name": parentName,
		"top_uid":     topUid,
		"top_name":    topName,
		"tester":      tester,
	}
	query, _, _ := dialect.Update("tbl_members").Set(record).Where(ex).ToSQL()
	fmt.Println(query)
	_, err = tx.Exec(query)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	query = fmt.Sprintf("delete from tbl_members_tree where descendant = %s and prefix = '%s'", uid, prefix)
	fmt.Println(query)
	_, err = tx.Exec(query)
	if err != nil {
		fmt.Printf("query : %s \n error : %s \n", query, err.Error())
		_ = tx.Rollback()
		return err
	}

	treeNode := memberClosureInsert(uid, parentUid)
	fmt.Println(treeNode)
	_, err = tx.Exec(treeNode)
	if err != nil {
		fmt.Printf("query : %s \n error : %s \n", query, err.Error())
		_ = tx.Rollback()
		return err
	}

	_ = tx.Commit()

	// 记录转代日志
	transRecord := AgencyTransferRecord{
		Id:            helper.GenLongId(),
		Flag:          551,
		Uid:           uid,
		Username:      username,
		Type:          agencyType,
		BeforeUid:     oldParentUid,
		BeforeName:    oldParentName,
		AfterUid:      parentUid,
		AfterName:     parentName,
		Remark:        "团队转代",
		UpdatedAt:     time.Now().Unix(),
		UpdatedUid:    "0",
		UpdatedName:   "bot",
		BeforeTopUid:  oldTopUid,
		BeforeTopName: oldTopName,
		AfterTopUid:   topUid,
		AfterTopName:  topName,
		Prefix:        prefix,
	}
	query, _, _ = dialect.Insert("tbl_agency_transfer_record").Rows(transRecord).ToSQL()
	_, err = db.Exec(query)
	if err != nil {
		return err
	}

	return nil
}

func memberClosureInsert(nodeID, targetID string) string {

	t := "SELECT ancestor, " + nodeID + ",prefix, lvl+1 FROM tbl_members_tree WHERE prefix='" + prefix + "' and descendant = " + targetID + " UNION SELECT " + nodeID + "," + nodeID + "," + "'" + prefix + "'" + ",0"
	query := "INSERT INTO tbl_members_tree (ancestor, descendant,prefix,lvl) (" + t + ")"

	return query
}

func membersByUid(uids []string) (map[string]common.Member, error) {

	data := map[string]common.Member{}

	if len(uids) == 0 {
		return data, errors.New(helper.ParamNull)
	}

	var mbs []common.Member
	query, _, _ := dialect.From("tbl_members").Select(colsMember...).Where(g.Ex{"uid": uids}).ToSQL()
	err := db.Select(&mbs, query)
	if err != nil {
		return data, err
	}

	if len(mbs) > 0 {
		for _, v := range mbs {
			if v.UID != "" {
				data[v.UID] = v
			}
		}
	}

	return data, nil
}
