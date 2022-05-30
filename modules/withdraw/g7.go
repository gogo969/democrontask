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

type G7 struct {
	AppID         string
	Key           string
	ApiToken      string
	CallbackToken string
	Host          string
}

func newG7() G7 {
	return G7{
		AppID:         "101",
		Key:           "03522909d0f0d484df50ddcf3f04f502",
		ApiToken:      "8OgLSiqV0nlrYwb8hfa2R3VCAhkdmnMkcguKjELJSDzfWnQhQiNfYl2zxHmW",
		CallbackToken: "EaYma6kKewPc2I1nRf1gPrK1gEt9rh9ReVJr64VFwI5noummPrz8n1tCklqs",
		Host:          "https://tianciv990901.com",
	}
}

var _ Payment = &G7{}

// 查询g7 pay提款(三方代付)订单
func (g7 G7) query(order withdraw) (state int, err error) {

	header := map[string]string{
		"Content-Type":  "application/json",
		"Accept":        "application/json",
		"Authorization": "Bearer " + g7.ApiToken, // 使用直连的token
	}
	dst := fmt.Sprintf("%s/api/payment/%s", g7.Host, order.ID)

	resp, statusCode, err := helper.HttpDoTimeout([]byte{}, "GET", dst, header, time.Second*8)
	common.Log("withdraw", "g7 query withdrawal order response: [%s] code: [%d] error: [%v]",
		string(resp), statusCode, err)

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

	switch string(v.Get("data").GetStringBytes("state")) {
	case "completed":
		state = common.WithdrawSuccess
	case "failed", "reject":
		state = common.WithdrawAutoPayFailed
	case "new", "processing":
		state = common.WithdrawDealing
	}

	return state, nil
}

func (g7 G7) sign(args map[string]string) string {

	i := 0
	qs := ""
	keys := make([]string, len(args))

	for k := range args {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	for _, v := range keys {
		qs += fmt.Sprintf("%s=%s&", v, args[v])
	}
	qs = qs[:len(qs)-1] + g7.Key

	return helper.GetMD5Hash(qs)
}
