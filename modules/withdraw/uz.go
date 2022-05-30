package withdraw

import (
	"errors"
	"fmt"
	"sort"
	"time"

	"cronTask/contrib/helper"
	"cronTask/modules/common"

	"github.com/valyala/fasthttp"
	"github.com/valyala/fastjson"
)

type UZ struct {
	AppID string
	Key   string
	Host  string
}

func newUZ() UZ {
	return UZ{
		AppID: "55834",
		Key:   "4db72bcc4dc5a447051c9a9914d22475",
		Host:  "https://www.uz-pay.com",
	}
}

var _ Payment = &UZ{}

// 查询uz pay提款(三方代付)订单
func (uz UZ) query(order withdraw) (state int, err error) {

	recs := map[string]string{
		"uid":     uz.AppID,
		"orderid": order.ID, //贵司订单编号
	}
	recs["sign"] = uz.sign(recs)

	body, err := helper.JsonMarshal(recs)
	if err != nil {
		return state, err
	}

	header := map[string]string{
		"Content-Type": "application/json",
	}
	dst := fmt.Sprintf("%s/Api/withdraw/query", uz.Host)

	resp, statusCode, err := helper.HttpDoTimeout(body, "POST", dst, header, time.Second*8)
	common.Log("withdraw", "uz query withdrawal order response: [%s] recs: [%v] code: [%d] error: [%v]",
		string(resp), recs, statusCode, err)
	if err != nil || statusCode != fasthttp.StatusOK {
		return state, errors.New(helper.RequestFail)
	}

	var p fastjson.Parser
	v, err := p.ParseBytes(resp)
	if err != nil {
		return state, err
	}

	if !v.GetBool("success") {
		return state, errors.New(string(v.GetStringBytes("msg")))
	}

	switch string(v.Get("order").GetStringBytes("status")) {
	case "verified":
		state = common.WithdrawSuccess
	case "timeout", "revoked":
		state = common.WithdrawAutoPayFailed
	case "processing":
		state = common.WithdrawDealing
	}

	return state, nil
}

func (uz UZ) sign(recs map[string]string) string {

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
	qs += "key=" + uz.Key

	return helper.GetMD5Hash(qs)
}
