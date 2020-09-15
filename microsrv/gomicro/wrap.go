package zgomicro

import (
	"context"
	"errors"
	"fmt"
	"net"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/micro/go-micro/client"
	gmerrors "github.com/micro/go-micro/errors"
	"github.com/micro/go-micro/server"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/sirupsen/logrus"

	"github.com/micro/go-micro/metadata"
	zeusctx "gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/context"
	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/engine"
	zeuserrors "gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/errors"
	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/utils"
)

var fm sync.Map

type validator interface {
	Validate() error
}

const (
	// log file.
	_source = "source"
	// container ID.
	_instanceID = "instance_id"
	// uniq ID from trace.
	_tid = "traceId"
	// appsName.
	_caller = "caller"
)

func GenerateServerLogWrap(ng engine.Engine) func(fn server.HandlerFunc) server.HandlerFunc {
	return func(fn server.HandlerFunc) server.HandlerFunc {
		return func(ctx context.Context, req server.Request, rsp interface{}) (err error) {
			logger := ng.GetContainer().GetLogger()
			prom := ng.GetContainer().GetPrometheus().GetInnerCli()
			var errcode string
			now := time.Now()
			l := logger.WithFields(logrus.Fields{"tag": "gomicro-serverlogwrap"})
			c := zeusctx.GMClientToContext(ctx, ng.GetContainer().GetGoMicroClient())
			///////// tracer begin
			name := fmt.Sprintf("%s.%s", req.Service(), req.Endpoint())
			cfg, err := ng.GetConfiger()
			if err != nil {
				l.Error(err)
				err = &gmerrors.Error{Id: ng.GetContainer().GetServiceID(), Code: int32(zeuserrors.ECodeSystem), Detail: err.Error(), Status: "ng.GetConfiger"}
				return
			}

			tracer := ng.GetContainer().GetTracer()
			if tracer == nil {
				err = fmt.Errorf("tracer is nil")
				l.Error(err)
				err = &gmerrors.Error{Id: ng.GetContainer().GetServiceID(), Code: int32(zeuserrors.ECodeSystem), Detail: err.Error(), Status: ""}
				return
			}
			spnctx, span, err := tracer.StartSpanFromContext(c, name)
			if err != nil {
				l.Error(err)
				err = &gmerrors.Error{Id: ng.GetContainer().GetServiceID(), Code: int32(zeuserrors.ECodeSystem), Detail: err.Error(), Status: ""}
				return
			}
			defer func() {
				if cfg.Get().Trace.OnlyLogErr && err == nil {
					return
				}
				span.Finish()
			}()
			body, _ := utils.Marshal(req.Body())
			span.SetTag("grpc server receive", string(body))
			///////// tracer finish
			tracerID := tracer.GetTraceID(spnctx)
			l = l.WithFields(logrus.Fields{
				_tid:        tracerID,
				_instanceID: getHostIP(),
				_source:     funcName(2),
				_caller:     ng.GetContainer().GetServiceID(),
			})
			c = zeusctx.LoggerToContext(spnctx, l)

			if v, ok := req.Body().(validator); ok && v != nil {
				if err = v.Validate(); err != nil {
					zeusErr := zeuserrors.New(zeuserrors.ECodeInvalidParams, err.Error(), "validator.Validate")
					status := zeusErr.Cause
					if !strings.HasPrefix(status, tracerID+"@") {
						status = tracerID + "@" + zeusErr.Cause
					}
					err = &gmerrors.Error{Id: ng.GetContainer().GetServiceID(), Code: int32(zeusErr.ErrCode), Detail: zeusErr.ErrMsg, Status: status}
					return
				}
			}
			if ng.GetContainer().GetRedisCli() != nil {
				c = zeusctx.RedisToContext(c, ng.GetContainer().GetRedisCli().GetCli())
			}
			if ng.GetContainer().GetMongo() != nil {
				c = zeusctx.MongoToContext(c, ng.GetContainer().GetMongo())
			}
			if ng.GetContainer().GetHttpClient() != nil {
				c = zeusctx.HttpclientToContext(c, ng.GetContainer().GetHttpClient())
			}
			if ng.GetContainer().GetMysqlCli() != nil {
				c = zeusctx.MysqlToContext(c, ng.GetContainer().GetMysqlCli().GetCli())
			}
			if ng.GetContainer().GetPrometheus() != nil {
				c = zeusctx.PrometheusToContext(c, ng.GetContainer().GetPrometheus().GetPubCli())
			}

			defer func() {
				prom.RPCServer.Timing(name, int64(time.Since(now)/time.Millisecond), ng.GetContainer().GetServiceID())
				if errcode != "" {
					prom.RPCServer.Incr(name, ng.GetContainer().GetServiceID(), errcode)
				}

			}()
			err = fn(c, req, rsp)
			if err != nil && !utils.IsBlank(reflect.ValueOf(err)) {
				span.SetTag("grpc server answer error", err)
				// zeus错误包装为gomicro错误
				var zeusErr *zeuserrors.Error
				var gmErr *gmerrors.Error
				if errors.As(err, &zeusErr) {
					serviceID := zeusErr.ServiceID
					if utils.IsEmptyString(serviceID) {
						serviceID = ng.GetContainer().GetServiceID()
					}
					status := zeusErr.Cause
					if !strings.HasPrefix(status, tracerID+"@") {
						status = tracerID + "@" + zeusErr.Cause
					}
					gmErr = &gmerrors.Error{Id: serviceID, Code: int32(zeusErr.ErrCode), Detail: zeusErr.ErrMsg, Status: status}

					// 防止go-micro grpc 将小于errcode小于0的错误转换成 internal error
					// 对于非zeus grpc调用，不做处理（grpc-gateway调用，直接返回负数，否http访问会得不到正确的错误码）
					if isZeusRpc(ctx) {
						if gmErr.Code < 0 {
							gmErr.Code = -gmErr.Code
							gmErr.Detail = "-@" + gmErr.Detail
						}
					}
					errcode = strconv.Itoa(int(zeusErr.ErrCode))
					err = gmErr

					return
				}
				if errors.As(err, &gmErr) {
					err = gmErr
					return
				}

				err = &gmerrors.Error{Id: ng.GetContainer().GetServiceID(), Code: int32(zeuserrors.ECodeSystem), Detail: err.Error(), Status: tracerID + "@" + err.Error()}
				return
			}
			err = nil
			rspRaw, _ := utils.Marshal(rsp)
			span.SetTag("grpc server answer", string(rspRaw))
			return
		}
	}
}

func GenerateClientLogWrap(ng engine.Engine) func(c client.Client) client.Client {
	return func(c client.Client) client.Client {
		return &clientLogWrap{
			Client: c,
			ng:     ng,
		}
	}
}

func GenerateClientWrapTest(ng engine.Engine) func(c client.Client) client.Client {
	return func(c client.Client) client.Client {
		return &clientWrapTest{
			Client: c,
		}
	}
}

func newClientLogWrap(c client.Client) client.Client {
	return &clientLogWrap{
		Client: c,
	}
}

type clientLogWrap struct {
	client.Client
	ng engine.Engine
}

func (l *clientLogWrap) Call(ctx context.Context, req client.Request, rsp interface{}, opts ...client.CallOption) (err error) {
	var now = time.Now()
	var errcode string
	logger := zeusctx.ExtractLogger(ctx)
	ng := l.ng
	prom := ng.GetContainer().GetPrometheus().GetInnerCli()

	///////// tracer begin
	name := fmt.Sprintf("%s.%s", req.Service(), req.Endpoint())
	defer func() {
		prom.RPCClient.Timing(name, int64(time.Since(now)/time.Millisecond), ng.GetContainer().GetServiceID())
		if errcode != "" {
			prom.RPCClient.Incr(name, ng.GetContainer().GetServiceID(), errcode)
		}

	}()
	cfg, err := ng.GetConfiger()
	if err != nil {
		logger.Error(err)
		errcode = string(zeuserrors.ECodeSystem)
		err = &gmerrors.Error{Id: ng.GetContainer().GetServiceID(), Code: int32(zeuserrors.ECodeSystem), Detail: err.Error(), Status: "ng.GetConfiger"}
		return
	}

	tracer := ng.GetContainer().GetTracer()
	if tracer == nil {
		err = fmt.Errorf("tracer is nil")
		logger.Error(err)
		errcode = string(zeuserrors.ECodeSystem)
		err = &gmerrors.Error{Id: ng.GetContainer().GetServiceID(), Code: int32(zeuserrors.ECodeSystem), Detail: err.Error(), Status: ""}
		return
	}
	spnctx, span, err := tracer.StartSpanFromContext(ctx, name)
	if err != nil {
		logger.Error(err)
		errcode = string(zeuserrors.ECodeSystem)
		err = &gmerrors.Error{Id: ng.GetContainer().GetServiceID(), Code: int32(zeuserrors.ECodeSystem), Detail: err.Error(), Status: ""}
		return
	}
	defer func() {
		if cfg.Get().Trace.OnlyLogErr && err == nil {
			return
		}
		span.Finish()
	}()
	ext.SpanKindRPCClient.Set(span)
	body, _ := utils.Marshal(req.Body())
	span.SetTag("grpc client call", string(body))
	///////// tracer finish

	//zeus rpc 调用标识
	ctx = zeusFlagToContext(ctx)
	err = l.Client.Call(ctx, req, rsp, opts...)
	if err != nil {
		// gomicro错误解包为zeus错误
		var gmErr *gmerrors.Error
		if errors.As(err, &gmErr) {
			if gmErr != nil {
				span.SetTag("grpc client receive error", gmErr)
				errcode = strconv.Itoa(int(gmErr.Code))
				// 根据Detail 判断错误码是否负数，将其还原
				strs := strings.SplitN(gmErr.Detail, "@", 2)
				if len(strs) == 2 && strs[0] == "-" {
					gmErr.Code = -gmErr.Code
					gmErr.Detail = strs[1]
				}
				zeusErr := zeuserrors.New(zeuserrors.ErrorCode(gmErr.Code), gmErr.Detail, gmErr.Status)
				zeusErr.ServiceID = gmErr.Id
				if utils.IsEmptyString(zeusErr.ServiceID) && l.ng != nil {
					zeusErr.ServiceID = l.ng.GetContainer().GetServiceID()
				}

				zeusErr.TracerID = tracer.GetTraceID(spnctx)
				err = zeusErr
				return
			}
			err = nil
		}
	}
	rspRaw, _ := utils.Marshal(rsp)
	span.SetTag("grpc client receive", string(rspRaw))
	return
}

type clientWrapTest struct {
	client.Client
}

func (l *clientWrapTest) Call(ctx context.Context, req client.Request, rsp interface{}, opts ...client.CallOption) (err error) {
	zeusctx.ExtractLogger(ctx).Debug("clientWrapTest")
	err = l.Client.Call(ctx, req, rsp, opts...)
	return
}

func getHostIP() (orghost string) {
	addrs, _ := net.InterfaceAddrs()
	for _, address := range addrs {
		// 检查ip地址判断是否回环地址
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				orghost = ipnet.IP.String()
			}

		}
	}
	return
}

// funcName get func name.
func funcName(skip int) (name string) {
	if pc, _, lineNo, ok := runtime.Caller(skip); ok {
		if v, ok := fm.Load(pc); ok {
			name = v.(string)
		} else {
			name = runtime.FuncForPC(pc).Name() + ":" + strconv.FormatInt(int64(lineNo), 10)
			fm.Store(pc, name)
		}
	}
	return
}

func zeusFlagToContext(ctx context.Context) context.Context {
	if md, ok := metadata.FromContext(ctx); ok {
		md["zeus-rpc-flag"] = ""
		//map修改直接生效，不需要重设
		//ctx = metadata.NewContext(ctx, md)
	} else {
		ctx = metadata.NewContext(ctx, metadata.Metadata{"zeus-rpc-flag": ""})
	}
	return ctx
}

func isZeusRpc(ctx context.Context) bool {
	if md, ok := metadata.FromContext(ctx); ok {
		if _, ok := md["zeus-rpc-flag"]; ok {
			return true
		}
	}
	return false
}
