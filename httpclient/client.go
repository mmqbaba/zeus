package httpclient

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/pkg/errors"
	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/config"
	zeusctx "gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/context"
	tracing "gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/trace"
	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/utils"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	defaultRetryCount  = 0
	defaultHTTPTimeout = 30 * 1000 * time.Millisecond
	defaultUsageAgent  = "zeus-httpclient v0.0.1"
)

var httpclientInstance = make(map[string]*Client)

var defaultSetting = httpClientSettings{
	Transport:  http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: false}},
	UserAgent:  defaultUsageAgent,
	Timeout:    defaultHTTPTimeout,
	RetryCount: defaultRetryCount,
}

type httpClientSettings struct {
	Transport  http.Transport
	UserAgent  string
	RetryCount uint32
	Timeout    time.Duration
	Host       string
}

//type ClientMgr struct {
//    clients map[string]*Client
//    rw      sync.RWMutex
//}

type Client struct {
	client   *http.Client
	settings httpClientSettings
	retrier  Retriable
}

func ReloadHttpClientConf(conf map[string]config.HttpClientConf) error {
	var tmpInstanceMap = make(map[string]*Client)
	for instanceName, httpClientConf := range conf {
		tmpInstanceMap[instanceName] = newClient(&httpClientConf)
	}
	httpclientInstance = tmpInstanceMap
	return nil
}

func GetClient(ctx context.Context, instance string) (*Client, error) {
	loger := zeusctx.ExtractLogger(ctx)
	v, ok := httpclientInstance[instance]
	if !ok {
		loger.Error("unknown instance: " + instance)
		return nil, errors.New("unknown instance: " + instance)
	}
	return v, nil
}

func newClient(cfg *config.HttpClientConf) *Client {
	settings := defaultSetting

	if !utils.IsEmptyString(cfg.UserAgent) {
		settings.UserAgent = cfg.UserAgent
	}

	if utils.IsEmptyString(cfg.HostName) {
		panic("host_name不能为空...")
	}
	settings.Host = cfg.HostName

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
	settings.Transport = transport

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

func (c *Client) Get(ctx context.Context, url string, headers http.Header) ([]byte, error) {
	request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%v%v", c.settings.Host, url), nil)
	if err != nil {
		return nil, errors.Wrap(err, "GET - request creation failed")
	}
	request.Header = headers

	return c.do(ctx, request)
}

func (c *Client) Post(ctx context.Context, url string, body io.Reader, headers http.Header) ([]byte, error) {
	request, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%v%v", c.settings.Host, url), body)
	if err != nil {
		return nil, errors.Wrap(err, "POST - request creation failed")
	}

	request.Header = headers

	return c.do(ctx, request)
}

func (c *Client) Put(ctx context.Context, url string, body io.Reader, headers http.Header) ([]byte, error) {
	request, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%v%v", c.settings.Host, url), body)
	if err != nil {
		return nil, errors.Wrap(err, "PUT - request creation failed")
	}

	request.Header = headers

	return c.do(ctx, request)
}

func (c *Client) Patch(ctx context.Context, url string, body io.Reader, headers http.Header) ([]byte, error) {
	request, err := http.NewRequest(http.MethodPatch, fmt.Sprintf("%v%v", c.settings.Host, url), body)
	if err != nil {
		return nil, errors.Wrap(err, "PATCH - request creation failed")
	}

	request.Header = headers

	return c.do(ctx, request)
}

func (c *Client) Delete(ctx context.Context, url string, headers http.Header) ([]byte, error) {
	request, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%v%v", c.settings.Host, url), nil)
	if err != nil {
		return nil, errors.Wrap(err, "DELETE - request creation failed")
	}

	request.Header = headers

	return c.do(ctx, request)
}

func (c *Client) do(ctx context.Context, request *http.Request) (rsp []byte, err error) {
	// todo
	//request.Close = true
	loger := zeusctx.ExtractLogger(ctx)
	tracer := tracing.NewTracerWrap(opentracing.GlobalTracer())
	name := request.URL.RawPath
	ctx, span, _ := tracer.StartSpanFromContext(ctx, name)
	ext.SpanKindConsumer.Set(span)
	span.SetTag("httpclient request.method", request.Method)
	defer func() {
		//if err == nil { //todo
		//    return
		//}
		span.Finish()
	}()

	var bodyReader *bytes.Reader

	if request.Body != nil {
		reqData, err := ioutil.ReadAll(request.Body)
		if err != nil {
			return nil, err
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
		return nil, err
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
		return nil, err
	}

	return rspBody, err
}
