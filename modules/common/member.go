package common

import (
	"cronTask/contrib/helper"
	"database/sql"
	"errors"
	"fmt"
	g "github.com/doug-martin/goqu/v9"
	"github.com/go-redis/redis/v8"
	"github.com/jmoiron/sqlx"
	"github.com/shopspring/decimal"
)

var (
	dialect    = g.Dialect("mysql")
	colsMember = helper.EnumFields(Member{})
)

// 通过用户名获取用户在redis中的数据
func MemberFindOne(db *sqlx.DB, name, prefix string) (Member, error) {

	m := Member{}
	if name == "" {
		return m, errors.New(helper.UsernameErr)
	}

	t := dialect.From("tbl_members")
	query, _, _ := t.Select(colsMember...).Where(g.Ex{"username": name, "prefix": prefix}).ToSQL()
	err := db.Get(&m, query)
	if err != nil && err != sql.ErrNoRows {
		return m, err
	}

	if err == sql.ErrNoRows {
		return m, errors.New(helper.UsernameErr)
	}

	return m, nil
}

func MemberMCache(db *sqlx.DB, names []string, prefix string) (map[string]Member, string, error) {

	data := map[string]Member{}

	if len(names) == 0 {
		return data, "", errors.New(helper.ParamNull)
	}

	var mbs []Member
	t := dialect.From("tbl_members")
	query, _, _ := t.Select(colsMember...).Where(g.Ex{"username": names, "prefix": prefix}).ToSQL()
	err := db.Select(&mbs, query)
	if err != nil {
		return data, "db", err
	}

	if len(mbs) > 0 {
		for _, v := range mbs {
			if v.Username != "" {
				data[v.Username] = v
			}
		}
	}

	return data, "", nil
}

func MembersPageNames(db *sqlx.DB, page, pageSize int, ex g.Ex) ([]string, error) {

	var v []string
	offset := (page - 1) * pageSize
	query, _, _ := dialect.From("tbl_members").Select("username").
		Where(ex).Offset(uint(offset)).Limit(uint(pageSize)).Order(g.C("created_at").Asc()).ToSQL()
	fmt.Println(query)
	err := db.Select(&v, query)

	return v, err
}

func MembersNames(db *sqlx.DB, ex g.Ex) ([]string, error) {

	var v []string
	query, _, _ := dialect.From("tbl_members").Select("username").Where(ex).ToSQL()
	fmt.Println(query)
	err := db.Select(&v, query)

	return v, err
}

func MembersCount(db *sqlx.DB, ex g.Ex) (int, error) {

	var count int
	query, _, _ := dialect.From("tbl_members").Select(g.COUNT("uid")).Where(ex).ToSQL()
	fmt.Println(query)
	err := db.Get(&count, query)

	return count, err
}

func MemberBalance(db *sqlx.DB, uid, prefix string) (decimal.Decimal, error) {

	var b string
	ex := g.Ex{
		"uid":    uid,
		"prefix": prefix,
	}
	query, _, _ := dialect.From("tbl_members").Select("balance").Where(ex).ToSQL()
	fmt.Println(query)
	err := db.Get(&b, query)
	if err != nil {
		return decimal.Zero, err
	}

	balance, err := decimal.NewFromString(b)
	if err != nil {
		return decimal.Zero, err
	}

	return balance, nil
}

func MemberUpdateCache(db *sqlx.DB, cli *redis.ClusterClient, prefix, username string) error {

	var (
		err error
		dst Member
	)

	dst, err = MemberFindOne(db, prefix, username)
	if err != nil {
		return err
	}

	key := prefix + ":member:" + dst.Username
	fields := []interface{}{"uid", dst.UID, "username", dst.Username, "tester", dst.Tester, "top_uid", dst.TopUid, "top_name", dst.TopName, "parent_uid", dst.ParentUid, "parent_name", dst.ParentName, "level", dst.Level, "balance", dst.Balance, "lock_amount", dst.LockAmount, "commission", dst.Commission, "group_name", dst.GroupName, "agency_type", dst.AgencyType, "address", dst.Address, "avatar", dst.Avatar}

	pipe := cli.TxPipeline()
	pipe.Del(ctx, key)
	pipe.HMSet(ctx, key, fields...)
	pipe.Persist(ctx, key)
	pipe.Exec(ctx)
	pipe.Close()
	return nil
}
