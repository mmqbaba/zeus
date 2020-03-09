package httpclient

import (
	"context"
	"fmt"
	conf "gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/config"
	"net/http"
	"testing"
)

func TestClient_Get(t *testing.T) {

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
			//IdleConnTimeout:       0
		},
	}

	err := ReloadHttpClientConf(cfg)
	ctx := context.Background()
	client, err := GetClient("test0")
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	type fields struct {
		client   *http.Client
		settings httpClientSettings
		retrier  Retriable
	}
	type args struct {
		ctx     context.Context
		url     string
		headers http.Header
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr bool
	}{
		{
			"test-get-0",
			fields{
				client.client,
				client.settings,
				client.retrier,
			},
			args{
				ctx,
				"/json/?lang=zh-CN",
				http.Header{},
			},
			[]byte{},
			false,
		},
		{
			"test-get-1",
			fields{
				client.client,
				client.settings,
				client.retrier,
			},
			args{
				ctx,
				"/json/?lang=zh-CN",
				http.Header{},
			},
			[]byte{},
			false,
		},
		{
			"test-get-2",
			fields{
				client.client,
				client.settings,
				client.retrier,
			},
			args{
				ctx,
				"/json/?lang=zh-CN",
				http.Header{},
			},
			[]byte{},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				client:   tt.fields.client,
				settings: tt.fields.settings,
				retrier:  tt.fields.retrier,
			}
			got, err := c.Get(tt.args.ctx, tt.args.url, tt.args.headers)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			fmt.Println(string(got))
			//if !reflect.DeepEqual(got, tt.want) {
			//    t.Errorf("Get() got = %v, want %v", got, tt.want)
			//}
		})
	}
}
