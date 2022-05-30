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

type YFB struct {
	AppID  string
	Key    string
	Secret string
	Host   string
}

func newYFB() YFB {
	return YFB{
		AppID:  "9016",
		Key:    "ItPUAxr2Bk2uP4lO27Gp8Q",
		Secret: "Mv58pOzJgk2romn15EFbFw",
		Host:   "https://vn.pasvn.com",
	}
}

var _ Payment = &YFB{}

// 查询优付宝提款(三方代付)订单
func (yfb YFB) query(order withdraw) (state int, err error) {

	recs := map[string]string{
		"merchantNo": yfb.AppID,                            // 商户编号
		"orderNo":    order.ID,                             // 商户订单号
		"appSecret":  yfb.Secret,                           //
		"tradeNo":    order.OID,                            // 3方订单号
		"time":       fmt.Sprintf("%d", time.Now().Unix()), // 时间戳
	}

	recs["sign"] = yfb.sign(recs)
	formData := url.Values{}
	for k, v := range recs {
		formData.Set(k, v)
	}
	dst := fmt.Sprintf("%s/payout/status", yfb.Host)
	header := map[string]string{}

	resp, statusCode, err := helper.HttpDoTimeout([]byte(formData.Encode()), "POST", dst, header, time.Second*8)
	common.Log("withdraw", "yfb query withdrawal order response: [%s] recs: [%v] code: [%d] error: [%v]",
		string(resp), formData, statusCode, err)
	if err != nil || statusCode != fasthttp.StatusOK {
		return state, errors.New(helper.RequestFail)
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

func (yfb YFB) sign(args map[string]string) string {

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
	qs = qs[:len(qs)-1] + yfb.Key

	s256 := fmt.Sprintf("%x", sha256.Sum256([]byte(qs)))

	return strings.ToUpper(helper.GetMD5Hash(s256))
}
