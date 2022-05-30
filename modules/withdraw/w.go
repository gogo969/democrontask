package withdraw

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"

	"cronTask/contrib/helper"
	"cronTask/modules/common"

	"github.com/valyala/fasthttp"
	"github.com/valyala/fastjson"
)

type W struct {
	AppID  string
	Key    string
	Secret string
	Host   string
}

func newW() W {
	return W{
		AppID:  "10679",
		Key:    "KRYD47E53kWoP4p9yVcSJw",
		Secret: "kwl56qForUC9sMYVVf7zUA",
		Host:   "https://wv.wppas.com",
	}
}

var _ Payment = &W{}

// 查询w pay提款(三方代付)订单
func (w W) query(order withdraw) (state int, err error) {

	args := map[string]string{
		"merchantNo": w.AppID,                              // 商户编号
		"orderNo":    order.ID,                             // 商户订单号
		"appSecret":  w.Secret,                             //
		"tradeNo":    order.OID,                            // 3方订单号
		"time":       fmt.Sprintf("%d", time.Now().Unix()), // 时间戳
	}

	args["sign"] = w.sign(args)
	formData := url.Values{}
	for k, v := range args {
		formData.Set(k, v)
	}
	dst := fmt.Sprintf("%s/payout/status", w.Host)
	header := map[string]string{}

	resp, statusCode, err := helper.HttpDoTimeout([]byte(formData.Encode()), "POST", dst, header, time.Second*8)
	common.Log("withdraw", "w query withdrawal order response: [%s] recs: [%v] code: [%d] error: [%v]",
		string(resp), formData, statusCode, err)
	if err != nil || statusCode != fasthttp.StatusOK {
		return state, errors.New(helper.FormatErr)
	}

	var p fastjson.Parser
	v, err := p.ParseBytes(resp)
	if err != nil {
		return state, errors.New(helper.FormatErr)
	}

	if v.GetInt("code") != 0 {
		return state, errors.New(string(v.GetStringBytes("text")))
	}

	switch string(v.GetStringBytes("status")) {
	case "PAID":
		state = common.WithdrawSuccess
	case "CANCELLED":
		state = common.WithdrawAutoPayFailed
	case "PENDING":
		state = common.WithdrawDealing
	}

	return state, nil
}

func (w W) sign(args map[string]string) string {

	qs := ""
	keys := make([]string, 0)

	for k := range args {
		switch k {
		case "bankBranch", "memo", "appSecret":
			continue
		}

		keys = append(keys, k)
	}

	sort.Strings(keys)
	for _, v := range keys {
		qs += fmt.Sprintf("%s=%s&", v, args[v])
	}
	qs = qs[:len(qs)-1] + w.Key

	s256 := fmt.Sprintf("%x", sha256.Sum256([]byte(qs)))

	return strings.ToUpper(helper.GetMD5Hash(s256))
}
