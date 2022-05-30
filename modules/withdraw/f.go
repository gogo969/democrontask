package withdraw

import (
	"crypto/hmac"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"cronTask/contrib/helper"
	"cronTask/modules/common"

	"github.com/valyala/fasthttp"
	"github.com/valyala/fastjson"
)

type F struct {
	AppID string
	WKey  string
	Host  string
}

func newF() F {
	return F{
		AppID: "MC210429163124",
		WKey:  "SIUAY7GQLR5KSWVQ",
		Host:  "https://api.zzz.ph",
	}
}

var _ Payment = &F{}

// 查询f pay提款(三方代付)订单
func (f F) query(order withdraw) (state int, err error) {

	var p fastjson.Parser

	recs := map[string]string{
		"orderNum": order.ID, // 订单号
		"merchant": f.AppID,
	}
	recs["sign"] = f.sign(fmt.Sprintf("%s/%s", f.AppID, order.ID))

	body, err := helper.JsonMarshal(recs)
	if err != nil {
		return state, errors.New(helper.FormatErr)
	}

	header := map[string]string{
		"Content-Type": "application/json",
	}
	dst := fmt.Sprintf("%s/payout/queryByOrderNum", f.Host)

	resp, statusCode, err := helper.HttpDoTimeout(body, "POST", dst, header, time.Second*8)
	common.Log("withdraw", "f query withdrawal order response: [%s] recs: [%v] code: [%d] error: [%v]",
		string(resp), recs, statusCode, err)
	if err != nil || statusCode != fasthttp.StatusOK {
		return state, errors.New(helper.RequestFail)
	}

	v, err := p.ParseBytes(resp)
	if err != nil {
		return state, errors.New(helper.FormatErr)
	}

	if string(v.GetStringBytes("code")) != "success" {
		return state, errors.New(string(v.GetStringBytes("msg")))
	}

	switch string(v.Get("data").GetStringBytes("status")) {
	case "success":
		state = common.WithdrawSuccess
	case "reject":
		state = common.WithdrawAutoPayFailed
	case "withdrawing":
		state = common.WithdrawDealing
	}

	return state, nil
}

func (f F) sign(str string) string {

	key := []byte(f.WKey)
	mac := hmac.New(md5.New, key)
	mac.Write([]byte(str))

	return hex.EncodeToString(mac.Sum(nil))
}
