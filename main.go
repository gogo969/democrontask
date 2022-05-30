package main

import (
	"cronTask/modules/captcha"
	"cronTask/modules/deposit"
	"cronTask/modules/event"
	"cronTask/modules/promo"
	"cronTask/modules/reports"
	"cronTask/modules/transfer"
	"cronTask/modules/upgrade"
	"cronTask/modules/withdraw"
	"fmt"
	"os"
	"strings"

	_ "go.uber.org/automaxprocs"
)

var (
	gitReversion   = ""
	buildTime      = ""
	buildGoVersion = ""
)

type fnP func([]string, string, string)

var cbP = map[string]fnP{
	"event":   event.Parse,   // 热门赛事 抓取
	"upgrade": upgrade.Parse, //会员升/保/降级
	"captcha": captcha.Parse, //验证码
	"report":  reports.Parse, //报表计算
	"promo":   promo.Parse,   //活动流水更新
}

type fn func([]string, string)

var cb = map[string]fn{
	"deposit":  deposit.Parse,  //删除未付款订单
	"withdraw": withdraw.Parse, //更新提现（代付）失败订单的状态
	"transfer": transfer.Parse, //团队转代处理脚本
}

func main() {

	argc := len(os.Args)
	if argc != 5 {
		fmt.Printf("%s <etcds> <cfgPath> [upgrade][transferConfirm][dividend][birthDividend][monthlyDividend][rebate][message][report] <proxy>\n", os.Args[0])
		return
	}

	endpoints := strings.Split(os.Args[1], ",")

	fmt.Printf("gitReversion = %s\r\nbuildGoVersion = %s\r\nbuildTime = %s\r\n", gitReversion, buildGoVersion, buildTime)

	if val, ok := cb[os.Args[3]]; ok {
		val(endpoints, os.Args[2])
	}

	// 带参数的脚本
	if val, ok := cbP[os.Args[3]]; ok {
		val(endpoints, os.Args[2], os.Args[4])
	}

	fmt.Println(os.Args[3], "done")
}
