package common

import (
	"fmt"
	"github.com/jmoiron/sqlx"
)

// 会员等级信息
type LevelInfo struct {
	Level             int    `db:"level"`
	LevelName         string `db:"level_name"`
	UpgradeDeposit    string `db:"upgrade_deposit"`
	UpgradeRecord     string `db:"upgrade_record"`
	RelegationFlowing string `db:"relegation_flowing"`
	UpgradeGift       string `db:"upgrade_gift"`
	EarlyMonthPacket  string `db:"early_month_packet"`
	BirthGift         string `db:"birth_gift"`
}

// 最近降级会员记录
type DowngradeInfo struct {
	Username  string `db:"username"`   //会员名
	CreatedAt uint64 `db:"created_at"` //降级时间
}

func MemberLevelList(db *sqlx.DB) (map[int]LevelInfo, error) {

	var info []LevelInfo
	levels := map[int]LevelInfo{}
	query, _, _ := dialect.From("tbl_member_level").Select("level", "level_name", "upgrade_deposit",
		"upgrade_record", "relegation_flowing", "upgrade_gift", "early_month_packet", "birth_gift").ToSQL()
	fmt.Println(query)
	err := db.Select(&info, query)
	if err != nil {
		return levels, err
	}

	for _, v := range info {
		levels[v.Level] = v
		fmt.Println(v)
	}

	return levels, nil
}

func MemberLevelDowngradeList(db *sqlx.DB) (map[string]uint64, error) {

	var info []DowngradeInfo
	downgrade := map[string]uint64{}
	query, _, _ := dialect.From("tbl_member_level_downgrade").Select("username", "created_at").ToSQL()
	err := db.Select(&info, query)
	if err != nil {
		return downgrade, err
	}

	for _, v := range info {
		downgrade[v.Username] = v.CreatedAt
	}

	return downgrade, nil
}
