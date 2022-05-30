package withdraw

import (
	"errors"
	"fmt"
	"net/url"
	"sort"
	"time"

	"cronTask/contrib/helper"
	"cronTask/modules/common"

	"github.com/valyala/fasthttp"
	"github.com/valyala/fastjson"
)

type PN struct {
	AppID string
	Key   string
	Host  string
}

func newPN() PN {
	return PN{
		AppID: "101",
		Key:   "03522909d0f0d484df50ddcf3f04f502",
		Host:  "https://app.vndpay.today",
	}
}

var _ Payment = &PN{}

// 查询pn pay提款(三方代付)订单
func (pn PN) query(order withdraw) (state int, err error) {

	recs := map[string]string{
		"app_id":       pn.AppID, // 商户ID 平台提供
		"app_order_no": order.ID, // 商户侧订单号
	}
	recs["sign"] = pn.sign(recs)

	formData := url.Values{}
	for k, v := range recs {
		formData.Set(k, v)
	}
	dst := fmt.Sprintf("%s/order/withdrawal/query", pn.Host)
	header := map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	}

	resp, statusCode, err := helper.HttpDoTimeout([]byte(formData.Encode()), "POST", dst, header, time.Second*8)
	common.Log("withdraw", "pn query withdrawal order response: [%s] recs: [%v] code: [%d] error: [%v]",
		string(resp), recs, statusCode, err)
	if err != nil || statusCode != fasthttp.StatusOK {
		return state, errors.New(helper.FormatErr)
	}

	var p fastjson.Parser
	v, err := p.ParseBytes(resp)
	if err != nil {
		return state, errors.New(helper.FormatErr)
	}

	if v.GetInt("success") != 1 {
		return state, errors.New(string(v.GetStringBytes("msg")))
	}

	switch v.GetInt("status") {
	case 1:
		state = common.WithdrawSuccess
	case 2:
		state = common.WithdrawAutoPayFailed
	case 0:
		state = common.WithdrawDealing
	}

	return state, nil
}

func (pn PN) sign(recs map[string]string) string {

	i := 0
	qs := ""
	keys := make([]string, len(recs))

	for k := range recs {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	for _, v := range keys {
		qs += fmt.Sprintf("%s=%s&", v, recs[v])
	}
	qs = qs[:len(qs)-1] + pn.Key

	return helper.GetMD5Hash(qs)
}
