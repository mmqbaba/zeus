package tifclient

import (
	"context"
	"crypto/md5"
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"

	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/config"
	zeusctx "gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/context"
	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/errors"
	tracing "gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/trace"
	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/utils"
)

var appconf *config.AppConf
var hclient *http.Client

func InitClient(conf *config.AppConf) {
	appconf = conf
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	if conf.EBus.IdleConnTimeout != 0 {
		transport.IdleConnTimeout = conf.EBus.IdleConnTimeout * time.Second
	}
	if conf.EBus.MaxConnsPerHost != 0 {
		transport.MaxConnsPerHost = conf.EBus.MaxConnsPerHost
	}
	if conf.EBus.MaxIdleConns != 0 {
		transport.MaxIdleConns = conf.EBus.MaxIdleConns
	}
	if conf.EBus.MaxIdleConnsPerHost != 0 {
		transport.MaxIdleConnsPerHost = conf.EBus.MaxIdleConnsPerHost
	}
	tc := &http.Client{
		Transport: transport,
	}
	if hclient != nil {
		hclient.CloseIdleConnections()
	}
	hclient = tc
}

type IdentificationInfo struct {
	Uid   string `json:"uid,omitempty"`
	Uinfo string `json:"uinfo,omitempty"`
	Ext   string `json:"ext,omitempty"`
}

type AccessToken struct {
	AccessToken string `json:"access_token,omitempty"`
	ExpiresIn   uint32 `json:"expires_in,omitempty"`
}

type AccessTokenTif struct {
	ErrCode int         `json:"errcode,omitempty"`
	ErrMsg  string      `json:"errmsg,omitempty"`
	Data    AccessToken `json:"data,omitempty"`
}

func GetAccessTokenByPassId(ctx context.Context, paasId string) (*AccessTokenTif, error) {
	logger := zeusctx.ExtractLogger(ctx)
	url := fmt.Sprintf(appconf.EBus.PathMap["tifapi_gettoken"], paasId)
	accessInfo := new(AccessTokenTif)
	rspBody, _, err := TifRequest(ctx, "GET", url, "", nil)
	if err != nil {
		logger.Errorf("TifRequest error, errMsg:%+v", err)
		return nil, errors.ECodeTifClientRequest.ParseErr("GetAccessTokenByPassId err")
	}
	err = utils.Unmarshal(rspBody, accessInfo)
	if err != nil {
		logger.Errorf("Unmarshal error, errMsg:%+v", err)
		return nil, errors.ECodeJsonUnmarshal.ParseErr("")
	}
	if accessInfo.ErrCode != 0 {
		logger.Errorf("getAccessToken error,rspBody:%s", string(rspBody))
		return nil, errors.ECodeTifClientRequest.ParseErr("getAccessToken fail")
	}

	return accessInfo, nil
}

func Request(ctx context.Context, method, url, postData string, headers map[string]string) (rspBody []byte, status int, err error) {
	logger := zeusctx.ExtractLogger(ctx)
	tracer := tracing.NewTracerWrap(opentracing.GlobalTracer())
	name := url
	ctx, span, _ := tracer.StartSpanFromContext(ctx, name)
	ext.SpanKindConsumer.Set(span)
	span.SetTag("tif request.method", method)
	span.SetTag("tif request.body", postData)
	defer func() {
		if appconf.Trace.OnlyLogErr && err == nil {
			return
		}
		span.Finish()
	}()
	var postBody io.Reader
	if len(postData) > 0 {
		postBody = strings.NewReader(postData)
	}
	req, err := SetUpTifSignature(ctx, method, url, postBody, nil)
	if err != nil {
		return nil, 0, err
	}

	if len(headers) > 0 {
		for k, v := range headers {
			req.Header.Add(k, v)
		}
	}

	contentType := req.Header.Get("Content-Type")
	if len(contentType) == 0 {
		req.Header.Set("Content-Type", "application/json")
	}
	client := hclient
	httpRsp, err := client.Do(req)
	if err != nil {
		logger.Errorf("client.do err:%+v", err)
		return nil, status, errors.ECodeBadRequest.ParseErr("")
	}

	if httpRsp == nil {
		logger.Error("httpRsp is nil")
		return nil, status, errors.ECodeBadRequest.ParseErr("")
	}

	defer func() {
		if httpRsp.Body != nil {
			httpRsp.Body.Close()
		}
	}()

	status = httpRsp.StatusCode
	rspBody, err = ioutil.ReadAll(httpRsp.Body)
	span.SetTag("tif response.status", status)
	span.SetTag("tif response.body", string(rspBody))
	span.SetTag("tif response.error", err)
	if err != nil {
		logger.Errorf("ReadAll error:%+v", err)
		return nil, status, errors.ECodeBadRequest.ParseErr("")
	}
	if status/100 != 2 {
		logger.Errorf("http request fail, status:%d, body:%s", status, rspBody)
		return nil, status, errors.ECodeBadRequest.ParseErr("")
	}

	return rspBody, status, nil
}

func TifRequest(ctx context.Context, method, url, postData string, info *IdentificationInfo) (rspBody []byte, status int, err error) {
	logger := zeusctx.ExtractLogger(ctx)
	tracer := tracing.NewTracerWrap(opentracing.GlobalTracer())
	name := url
	ctx, span, _ := tracer.StartSpanFromContext(ctx, name)
	ext.SpanKindConsumer.Set(span)
	span.SetTag("tif request.method", method)
	span.SetTag("tif request.body", postData)
	defer func() {
		if appconf.Trace.OnlyLogErr && err == nil {
			return
		}
		span.Finish()
	}()
	var postBody io.Reader
	if len(postData) > 0 {
		postBody = strings.NewReader(postData)
	}
	req, err := SetUpTifSignature(ctx, method, url, postBody, info)
	if err != nil {
		return nil, 0, err
	}

	contentType := req.Header.Get("Content-Type")
	if len(contentType) == 0 {
		req.Header.Set("Content-Type", "application/json")
	}
	client := hclient
	httpRsp, err := client.Do(req)
	if err != nil {
		logger.Errorf("client.do err:%+v", err)
		return nil, status, errors.ECodeBadRequest.ParseErr("")
	}

	if httpRsp == nil {
		logger.Error("httpRsp is nil")
		return nil, status, errors.ECodeBadRequest.ParseErr("")
	}

	defer func() {
		if httpRsp.Body != nil {
			httpRsp.Body.Close()
		}
	}()

	status = httpRsp.StatusCode
	rspBody, err = ioutil.ReadAll(httpRsp.Body)
	span.SetTag("tif response.status", status)
	span.SetTag("tif response.body", string(rspBody))
	span.SetTag("tif response.error", err)
	if err != nil {
		logger.Errorf("ReadAll error:%+v", err)
		return nil, status, errors.ECodeBadRequest.ParseErr("")
	}
	if status/100 != 2 {
		logger.Errorf("http request fail, status:%d, body:%s", status, rspBody)
		return nil, status, errors.ECodeBadRequest.ParseErr("")
	}

	return rspBody, status, nil
}

func SetUpTifSignature(ctx context.Context, method, path string, body io.Reader, info *IdentificationInfo) (*http.Request, error) {
	logger := zeusctx.ExtractLogger(ctx)
	host := appconf.EBus.Hosts[rand.Intn(len(appconf.EBus.Hosts))]
	paasId := appconf.EBus.PaasId
	paasToken := appconf.EBus.PaasToken
	url := fmt.Sprintf("%v%v", host, path)
	logger.Info("url==>", url)
	httpReq, err := http.NewRequest(method, url, body)
	if err != nil {
		logger.Error(err)
		return nil, errors.ECodeSystem.ParseErr("new request err")
	}
	now := time.Now()
	nonce := tifNonce(now)
	sign := tifSign(paasToken, now.Unix(), nonce)

	httpReq.Header.Set("x-tif-paasid", paasId)
	httpReq.Header.Set("x-tif-signature", sign)
	httpReq.Header.Set("x-tif-timestamp", fmt.Sprintf("%d", now.Unix()))
	httpReq.Header.Set("x-tif-nonce", nonce)

	return httpReq, nil

}

func tifNonce(now time.Time) string {
	r := rand.New(rand.NewSource(now.Unix()))
	str := fmt.Sprintf("%d_%d_%d", os.Getpid(), now.Unix(), r.Uint32())
	return fmt.Sprintf("%x", md5.Sum([]byte(str)))
}

func tifSign(secret string, now int64, nonce string) string {
	rawStr := fmt.Sprintf("%d%s%s%d", now, secret, nonce, now)
	return strings.ToUpper(fmt.Sprintf("%x", sha256.Sum256([]byte(rawStr))))
}

func tifIdentificationSign(secret string, now int64, nonce string, info *IdentificationInfo) string {
	rawStr := fmt.Sprintf("%d%s%s,%s,%s,%s%d", now, secret, nonce, info.Uid, info.Uinfo, info.Ext, now)
	return strings.ToUpper(fmt.Sprintf("%x", sha256.Sum256([]byte(rawStr))))
}
