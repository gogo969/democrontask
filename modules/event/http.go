package event

import (
	"crypto/tls"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpproxy"
	"time"
)

const (
	apiTimeOut = time.Second * 12
)

func HttpPostJsonProxy(requestBody []byte, requestURI string, headers map[string]string) (int, []byte, error) {

	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()

	defer func() {
		fasthttp.ReleaseResponse(resp)
		fasthttp.ReleaseRequest(req)
	}()

	req.SetBody(requestBody)
	req.SetRequestURI(requestURI)
	req.Header.SetMethod("POST")

	if headers != nil {
		for k, v := range headers {
			req.Header.Set(k, v)
		}
	}

	client := &fasthttp.Client{
		MaxConnsPerHost: 100,
		TLSConfig:       &tls.Config{InsecureSkipVerify: true},
		ReadTimeout:     apiTimeOut,
		WriteTimeout:    apiTimeOut,
	}

	if httpProxy != "" && httpProxy != "0.0.0.0" {
		client.Dial = fasthttpproxy.FasthttpSocksDialer(httpProxy)
	}

	err := client.DoTimeout(req, resp, apiTimeOut)
	//fmt.Println("======request==================")
	//fmt.Println(req.Header.String(), string(req.Body()))
	//fmt.Println("======response==================")
	//fmt.Println(resp.Header.String(), string(resp.Body()))
	return resp.StatusCode(), resp.Body(), err
}
