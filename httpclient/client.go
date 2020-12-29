package httpclient

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/config"
	zeusctx "gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/context"
	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/errors"
	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/httpclient/zhttpclient"
	tracing "gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/trace"
	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/utils"
)

const (
	defaultRetryCount  = 0
	defaultHTTPTimeout = 30 * 1000 * time.Millisecond
)

var httpclientInstance = make(map[string]*Client)

type httpClientSettings struct {
	Transport       http.Transport
	RetryCount      uint32
	Timeout         time.Duration
	Hosts           []string
	TraceOnlyLogErr bool
}

type Client struct {
	client   *http.Client
	settings httpClientSettings
	retrier  Retriable
}

func ReloadHttpClientConf(conf map[string]config.HttpClientConf) error {
	var tmpInstanceMap = make(map[string]*Client)
	for instanceName, httpClientConf := range conf {
		if v, ok := httpclientInstance[instanceName]; ok {
			v.client.CloseIdleConnections()
		}
		tmpInstanceMap[instanceName] = newClient(&httpClientConf)
	}
	httpclientInstance = tmpInstanceMap
	return nil
}

func GetClient(instance string) (*Client, error) {
	v, ok := httpclientInstance[instance]
	if !ok {
		log.Printf("unknown instance: " + instance)
		return nil, errors.ECodeHttpClient.ParseErr("unknown instance: " + instance)
	}
	return v, nil
}

func (c *Client) GetHttpClient(instance string) (zhttpclient.Client, error) {
	clent, err := GetClient(instance)
	return clent, err
}

func DefaultClient() *Client {
	settings := httpClientSettings{
		Transport:       http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: false}},
		Timeout:         defaultHTTPTimeout,
		RetryCount:      defaultRetryCount,
		TraceOnlyLogErr: true,
	}

	client := Client{
		client: &http.Client{
			Transport: &settings.Transport,
			Timeout:   settings.Timeout,
		},
		settings: settings,
		retrier:  NewNoRetrier(),
	}
	return &client
}

func newClient(cfg *config.HttpClientConf) *Client {
	settings := httpClientSettings{
		Transport:       http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: false}},
		Timeout:         defaultHTTPTimeout,
		RetryCount:      defaultRetryCount,
		TraceOnlyLogErr: true,
	}

	settings.TraceOnlyLogErr = cfg.TraceOnlyLogErr

	if len(cfg.HostName) == 0 {
		panic("host_name不能为空...")
	}
	settings.Hosts = cfg.HostName

	if cfg.RetryCount != 0 {
		settings.RetryCount = cfg.RetryCount
	}

	if cfg.TimeOut != 0 {
		settings.Timeout = cfg.TimeOut * time.Millisecond
	}
	transport := http.Transport{}
	if cfg.IdleConnTimeout != 0 {
		transport.IdleConnTimeout = cfg.IdleConnTimeout * time.Millisecond
	}
	if cfg.MaxConnsPerHost != 0 {
		transport.MaxConnsPerHost = cfg.MaxConnsPerHost
	}
	if cfg.MaxIdleConns != 0 {
		transport.MaxIdleConns = cfg.MaxIdleConns
	}
	if cfg.MaxIdleConnsPerHost != 0 {
		transport.MaxIdleConnsPerHost = cfg.MaxIdleConnsPerHost
	}

	if !cfg.InsecureSkipVerify && !utils.IsEmptyString(cfg.CaCertPath) {
		caCrt, err := ioutil.ReadFile(cfg.CaCertPath)
		if err != nil {
			panic(fmt.Sprintf("%v:读取证书文件错误:%v!", cfg.HostName, err.Error()))
		}
		pool := x509.NewCertPool()
		pool.AppendCertsFromPEM(caCrt)
		transport.TLSClientConfig = &tls.Config{RootCAs: pool}
	}
	transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: cfg.InsecureSkipVerify}

	settings.Transport = http.Transport{
		TLSClientConfig:     transport.TLSClientConfig,
		DisableKeepAlives:   cfg.DisableKeepAlives,
		MaxIdleConns:        transport.MaxIdleConns,
		MaxIdleConnsPerHost: transport.MaxIdleConnsPerHost,
		MaxConnsPerHost:     transport.MaxConnsPerHost,
		IdleConnTimeout:     transport.IdleConnTimeout,
	}

	retrier := NewNoRetrier()
	if cfg.RetryCount > 0 {
		retrier = NewRetrier(NewConstantBackoff(cfg.BackoffInterval*time.Millisecond, cfg.MaximumJitterInterval*time.Millisecond))
	}

	client := Client{
		client: &http.Client{
			Transport: &settings.Transport,
			Timeout:   settings.Timeout,
		},
		settings: settings,
		retrier:  retrier,
	}
	return &client
}

func (c *Client) Get(ctx context.Context, url string, headers map[string]string) ([]byte, error) {
	request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%v%v", c.getRandomHost(), url), nil)
	if err != nil {
		return nil, errors.ECodeHttpClient.ParseErr("GET - request creation failed", "err: "+err.Error())
	}

	rsp, _, err := c.do(ctx, request, headers)
	return rsp, err
}

func (c *Client) Post(ctx context.Context, url string, body io.Reader, headers map[string]string) ([]byte, error) {
	request, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%v%v", c.getRandomHost(), url), body)
	if err != nil {
		return nil, errors.ECodeHttpClient.ParseErr("POST - request creation failed", "err: "+err.Error())
	}

	rsp, _, err := c.do(ctx, request, headers)
	return rsp, err
}

func (c *Client) PostWithStatusCode(ctx context.Context, url string, body io.Reader, headers map[string]string) (rsp []byte, s int, err error) {
	request, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%v%v", c.getRandomHost(), url), body)
	if err != nil {
		return nil, 0, errors.ECodeHttpClient.ParseErr("POST WithStatusCode - request creation failed", "err: "+err.Error())
	}

	rsp, s, err = c.do(ctx, request, headers)
	return
}

func (c *Client) Put(ctx context.Context, url string, body io.Reader, headers map[string]string) ([]byte, error) {
	request, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%v%v", c.getRandomHost(), url), body)
	if err != nil {
		return nil, errors.ECodeHttpClient.ParseErr("PUT - request creation failed", "err: "+err.Error())
	}

	rsp, _, err := c.do(ctx, request, headers)
	return rsp, err
}

func (c *Client) Patch(ctx context.Context, url string, body io.Reader, headers map[string]string) ([]byte, error) {
	request, err := http.NewRequest(http.MethodPatch, fmt.Sprintf("%v%v", c.getRandomHost(), url), body)
	if err != nil {
		return nil, errors.ECodeHttpClient.ParseErr("PATCH - request creation failed", "err: "+err.Error())
	}

	rsp, _, err := c.do(ctx, request, headers)
	return rsp, err
}

func (c *Client) Delete(ctx context.Context, url string, headers map[string]string) ([]byte, error) {
	request, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%v%v", c.getRandomHost(), url), nil)
	if err != nil {
		return nil, errors.ECodeHttpClient.ParseErr("DELETE - request creation failed", "err: "+err.Error())
	}

	rsp, _, err := c.do(ctx, request, headers)
	return rsp, err
}

func (c *Client) getRandomHost() string {
	return c.settings.Hosts[rand.Intn(len(c.settings.Hosts))]
}

func (c *Client) do(ctx context.Context, request *http.Request, headers map[string]string) (rsp []byte, s int, err error) {

	if len(headers) > 0 {
		for k, v := range headers {
			request.Header.Add(k, v)
		}
	}

	//request.Close = true
	loger := zeusctx.ExtractLogger(ctx)
	tracer := tracing.NewTracerWrap(opentracing.GlobalTracer())
	name := request.URL.RawPath
	ctx, span, _ := tracer.StartSpanFromContext(ctx, name)
	ext.SpanKindConsumer.Set(span)
	span.SetTag("httpclient request.method", request.Method)
	defer func() {
		if c.settings.TraceOnlyLogErr && err == nil {
			return
		}
		span.Finish()
	}()

	var bodyReader *bytes.Reader

	if request.Body != nil {
		reqData, err := ioutil.ReadAll(request.Body)
		if err != nil {
			return nil, 0, err
		}
		span.SetTag("httpclient request.body", string(reqData))
		bodyReader = bytes.NewReader(reqData)
		request.Body = ioutil.NopCloser(bodyReader) // prevents closing the body between retries
	}

	var response *http.Response

	for i := 0; i <= int(c.settings.RetryCount); i++ {
		if response != nil {
			response.Body.Close()
		}
		response, err = c.client.Do(request)
		if bodyReader != nil {
			// Reset the body reader after the request since at this point it's already read
			// Note that it's safe to ignore the error here since the 0,0 position is always valid
			_, _ = bodyReader.Seek(0, 0)
		}

		if err != nil {
			backoffTime := c.retrier.NextInterval(i)
			time.Sleep(backoffTime)
			continue
		}

		if response.StatusCode >= http.StatusInternalServerError {
			backoffTime := c.retrier.NextInterval(i)
			time.Sleep(backoffTime)
			continue
		}
		break
	}

	if response == nil {
		return nil, 0, err
	}

	defer func() {
		if response.Body != nil {
			response.Body.Close()
		}
	}()

	rspBody, err := ioutil.ReadAll(response.Body)
	span.SetTag("httpclient response.status", response.StatusCode)
	span.SetTag("httpclient response.body", string(rspBody))
	span.SetTag("httpclient response.error", err)
	if err != nil {
		loger.Errorf("ReadAll error:%+v", err)
		return nil, 0, err
	}

	return rspBody, response.StatusCode, err
}
