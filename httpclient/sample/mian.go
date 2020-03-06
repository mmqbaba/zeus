package main

import (
	"context"
	"fmt"
	conf "gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/config"
	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/httpclient"
	"net/http"
	"strings"
)

func main() {
	cfg := map[string]conf.HttpClientConf{
		"test0": conf.HttpClientConf{
			InstanceName:          "test0",
			HostName:              "http://ip-api.com",
			RetryCount:            2,
			BackoffInterval:       0,
			MaximumJitterInterval: 2000,
			TimeOut:               10000,
			CaCertPath:            "",
			UserAgent:             "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.88 Safari/537.36",
			InsecureSkipVerify:    false,
			DisableKeepAlives:     false,
			//MaxIdleConns:          0,
			//MaxIdleConnsPerHost:   0,
			//MaxConnsPerHost:       0,
			//IdleConnTimeout:       0,
			Enable: true,
		},
	}

	err := httpclient.ReloadHttpClientConf(cfg)
	ctx := context.Background()
	client, err := httpclient.GetClient(ctx, "test0")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	body, err := client.Get(ctx, "/json/?lang=zh-CN", nil)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(string(body))
	headers := http.Header{}
	headers.Set("Content-Type", "application/json")
	body2, err := client.Post(ctx, "/json/demo", strings.NewReader("xxxxx=1&eeeee=2"), headers)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(string(body2))
}
