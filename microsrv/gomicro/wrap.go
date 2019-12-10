package zgomicro

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/micro/go-micro/client"
	gmerrors "github.com/micro/go-micro/errors"
	"github.com/micro/go-micro/server"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/sirupsen/logrus"

	zeusctx "gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/context"
	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/engine"
	zeuserrors "gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/errors"
	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/utils"
)

type validator interface {
	Validate() error
}

func GenerateServerLogWrap(ng engine.Engine) func(fn server.HandlerFunc) server.HandlerFunc {
	return func(fn server.HandlerFunc) server.HandlerFunc {
		return func(ctx context.Context, req server.Request, rsp interface{}) (err error) {
			logger := ng.GetContainer().GetLogger()
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
			l = l.WithFields(logrus.Fields{"tracerid": tracer.GetTraceID(spnctx)})
			c = zeusctx.LoggerToContext(spnctx, l)
			l.Debug("serverLogWrap")
			if v, ok := req.Body().(validator); ok && v != nil {
				if err = v.Validate(); err != nil {
					zeusErr := zeuserrors.New(zeuserrors.ECodeInvalidParams, err.Error(), "validator.Validate")
					err = &gmerrors.Error{Id: ng.GetContainer().GetServiceID(), Code: int32(zeusErr.ErrCode), Detail: zeusErr.ErrMsg, Status: zeusErr.Cause}
					return
				}
			}
			if ng.GetContainer().GetRedisCli() != nil {
				c = zeusctx.RedisToContext(c, ng.GetContainer().GetRedisCli().GetCli())
			}
			if ng.GetContainer().GetMongo() != nil {
				c = zeusctx.MongoToContext(c, ng.GetContainer().GetMongo())
			}
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
					err = &gmerrors.Error{Id: serviceID, Code: int32(zeusErr.ErrCode), Detail: zeusErr.ErrMsg, Status: zeusErr.Cause}
					return
				}
				if errors.As(err, &gmErr) {
					err = gmErr
					return
				}

				err = &gmerrors.Error{Id: ng.GetContainer().GetServiceID(), Code: int32(zeuserrors.ECodeSystem), Detail: err.Error(), Status: err.Error()}
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
	logger := zeusctx.ExtractLogger(ctx)
	logger.Debug("clientLogWrap")

	ng := l.ng
	///////// tracer begin
	name := fmt.Sprintf("%s.%s", req.Service(), req.Endpoint())
	cfg, err := ng.GetConfiger()
	if err != nil {
		logger.Error(err)
		err = &gmerrors.Error{Id: ng.GetContainer().GetServiceID(), Code: int32(zeuserrors.ECodeSystem), Detail: err.Error(), Status: "ng.GetConfiger"}
		return
	}

	tracer := ng.GetContainer().GetTracer()
	if tracer == nil {
		err = fmt.Errorf("tracer is nil")
		logger.Error(err)
		err = &gmerrors.Error{Id: ng.GetContainer().GetServiceID(), Code: int32(zeuserrors.ECodeSystem), Detail: err.Error(), Status: ""}
		return
	}
	spnctx, span, err := tracer.StartSpanFromContext(ctx, name)
	if err != nil {
		logger.Error(err)
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

	err = l.Client.Call(ctx, req, rsp, opts...)
	if err != nil {
		// gomicro错误解包为zeus错误
		var gmErr *gmerrors.Error
		if errors.As(err, &gmErr) {
			if gmErr != nil {
				span.SetTag("grpc client receive error", gmErr)

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
