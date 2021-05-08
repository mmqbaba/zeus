package httpclient

import (
	"context"
	"fmt"
	conf "gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/config"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestClient_Get(t *testing.T) {

	cfg := map[string]conf.HttpClientConf{
		"test0": conf.HttpClientConf{
			InstanceName:          "test0",
			HostName:              []string{"http://ip-api.com", "http://ip-api2.com", "http://ip-api3.com"},
			RetryCount:            2,
			BackoffInterval:       0,
			MaximumJitterInterval: 2000,
			TimeOut:               10000,
			CaCertPath:            "",
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
		headers map[string]string
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
				map[string]string{"Content-Type": "application/json"},
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
				map[string]string{"Content-Type": "application/json"},
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
				nil,
			},
			[]byte{},
			false,
		},
		{
			"test-get-3",
			fields{
				client.client,
				client.settings,
				client.retrier,
			},
			args{
				ctx,
				"/json/?lang=zh-CN",
				nil,
			},
			[]byte{},
			false,
		},
		{
			"test-get-4",
			fields{
				client.client,
				client.settings,
				client.retrier,
			},
			args{
				ctx,
				"/json/?lang=zh-CN",
				nil,
			},
			[]byte{},
			false,
		},
		{
			"test-get-5",
			fields{
				client.client,
				client.settings,
				client.retrier,
			},
			args{
				ctx,
				"/json/?lang=zh-CN",
				nil,
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

func TestClient_PostForm(t *testing.T) {

	cfg := map[string]conf.HttpClientConf{
		"test0": conf.HttpClientConf{
			InstanceName:          "test0",
			HostName:              []string{"http://portal.dclingcloud.com:60163"},
			RetryCount:            2,
			BackoffInterval:       0,
			MaximumJitterInterval: 2000,
			TimeOut:               10000,
			CaCertPath:            "",
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
		headers map[string]string
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
				"/NPM/api/get-api-token.html",
				nil,
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
			data := url.Values{
				"username":   {"test"},
				"password":   {"OHJBQlcBjspIQXyBaaXC+9KvndE9RHP8m6Zgm6AcC3pfQOGiI8TIUXtMIqXjM760UDKdduyorP60B7nVFKXr3tdTnPI1rit55NCPKhjnygU+DJoo0WLaUevW+mcAPwi6/R5O65eGGgLzyngxeikPItWfeCn4tnnxTNpP9wb/d9A="},
				"expireTime": {"100000000"},
			}
			got, err := c.PostForm(tt.args.ctx, tt.args.url, strings.NewReader(data.Encode()), tt.args.headers)
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
