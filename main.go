package main

import (
	"cronTask/modules/deposit"
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

type fn func([]string, string)

var cb = map[string]fn{
	"deposit": deposit.Parse, //删除未付款订单
}

func main() {

	argc := len(os.Args)
	if argc != 5 {
		fmt.Printf("%s <etcds> <cfgPath> [deposit]\n", os.Args[0])
		return
	}

	endpoints := strings.Split(os.Args[1], ",")
	fmt.Printf("gitReversion = %s\r\nbuildGoVersion = %s\r\nbuildTime = %s\r\n", gitReversion, buildGoVersion, buildTime)
	if val, ok := cb[os.Args[3]]; ok {
		val(endpoints, os.Args[2])
	}

	fmt.Println(os.Args[3], "done")
}
