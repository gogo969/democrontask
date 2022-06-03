package captcha

import (
	"bytes"
	"context"
	"cronTask/contrib/conn"
	"cronTask/modules/common"
	"fmt"
	"github.com/afocus/captcha"
	"github.com/go-redis/redis/v8"
	"image/png"
	"io/ioutil"
	"log"
	"strings"
	"time"
)

var (
	cli    *redis.Client
	c      *captcha.Captcha
	prefix string
	ctx    = context.Background()
)

func Parse(endpoints []string, path, fpath string) {

	conf := common.ConfParse(endpoints, path)
	prefix = conf.Prefix
	// 初始化redis
	cli = conn.InitRedisSentinel(conf.Redis.Addr, conf.Redis.Password, conf.Redis.Sentinel, conf.Redis.Db)

	fmt.Println("fpath = ", fpath)
	c = captcha.New()
	fp, err := ioutil.ReadDir(fpath + "/fonts/")
	if err != nil {
		log.Fatal(err)
	}
	for _, f := range fp {
		// 设置字体
		_ = c.SetFont(fpath + "/fonts/" + f.Name())
	}

	c.SetSize(120, 40)
	c.SetDisturbance(captcha.MEDIUM)
	handle()
}

func handle() {

	pipe := cli.TxPipeline()
	defer pipe.Close()

	key := fmt.Sprintf("%s:captcha", prefix)
	pipe.Unlink(ctx, key)
	for i := 0; i < 3000; i++ {

		img, str := c.Create(4, captcha.CLEAR)
		str = strings.ToLower(str)
		/*
			fp, err := os.Create(str+".png")
			if err != nil {
				fmt.Println("os.Create = ", err)
				continue
			}

			png.Encode(fp, img)
			fp.Close()
		*/
		fmt.Println("str = ", str)
		buf := new(bytes.Buffer)
		if err := png.Encode(buf, *img); err != nil {
			fmt.Println("png.Encode = ", err)
			continue
		}

		pipe.Set(ctx, str, buf.Bytes(), time.Duration(48)*time.Hour)
		pipe.LPush(ctx, key, str)
		buf.Reset()
		buf = nil
	}

	_, err := pipe.Exec(ctx)
	fmt.Println("pipe.Exec = ", err)
	// bonus, err := cli.Get(ctx, key).Result()
}
