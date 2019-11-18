package gomicro

import (
	"context"
	"reflect"
	"errors"
	"fmt"

	"github.com/micro/go-micro/client"
	gmerrors "github.com/micro/go-micro/errors"
	"github.com/micro/go-micro/server"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	zipkintracer "github.com/openzipkin/zipkin-go-opentracing"
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
			c := zeusctx.EngineToContext(ctx, ng)
			c = zeusctx.GMClientToContext(c, ng.GetContainer().GetGoMicroClient())
			///////// tracer begin
			name := fmt.Sprintf("%s.%s", req.Service(), req.Endpoint())
			cfg, err := ng.GetConfiger()
			if err != nil {
				l.Error(err)
				err = &gmerrors.Error{Id: ng.GetContainer().GetServerID(), Code: int32(zeuserrors.ECodeSystem), Detail: err.Error(), Status: "ng.GetConfiger"}
				return
			}

			tracer := ng.GetContainer().GetTracer()
			if tracer == nil {
				err = fmt.Errorf("tracer is nil")
				l.Error(err)
				err = &gmerrors.Error{Id: ng.GetContainer().GetServerID(), Code: int32(zeuserrors.ECodeSystem), Detail: err.Error(), Status: ""}
				return
			}
			spnctx, span, err := tracer.StartSpanFromContext(c, name)
			if err != nil {
				l.Error(err)
				err = &gmerrors.Error{Id: ng.GetContainer().GetServerID(), Code: int32(zeuserrors.ECodeSystem), Detail: err.Error(), Status: ""}
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
			l = l.WithFields(logrus.Fields{"tracerid": span.Context().(zipkintracer.SpanContext).TraceID.ToHex()})
			c = zeusctx.LoggerToContext(spnctx, l)
			l.Debug("serverLogWrap")
			if v, ok := req.Body().(validator); ok && v != nil {
				if err = v.Validate(); err != nil {
					zeusErr := zeuserrors.New(zeuserrors.ECodeInvalidParams, err.Error(), "validator.Validate")
					err = &gmerrors.Error{Id: ng.GetContainer().GetServerID(), Code: int32(zeusErr.ErrCode), Detail: zeusErr.ErrMsg, Status: zeusErr.Cause}
					return
				}
			}
			if ng.GetContainer().GetRedisCli() != nil {
				c = zeusctx.RedisToContext(c, ng.GetContainer().GetRedisCli().GetCli())
			}
			err = fn(c, req, rsp)
			if err != nil && !reflect.ValueOf(err).IsNil() {
				span.SetTag("grpc client receive error", err)
				// zeus错误包装为gomicro错误
				var zeusErr *zeuserrors.Error
				var gmErr *gmerrors.Error
				if errors.As(err, &zeusErr) {
					serverID := zeusErr.ServerID
					if utils.IsEmptyString(serverID) {
						serverID = ng.GetContainer().GetServerID()
					}
					err = &gmerrors.Error{Id: serverID, Code: int32(zeusErr.ErrCode), Detail: zeusErr.ErrMsg, Status: zeusErr.Cause}
					return
				}
				if errors.As(err, &gmErr) {
					err = gmErr
					return
				}

				err = &gmerrors.Error{Id: ng.GetContainer().GetServerID(), Code: int32(zeuserrors.ECodeSystem), Detail: err.Error(), Status: err.Error()}
				return
			}
			err = nil
			rspRaw, _ := utils.Marshal(rsp)
			span.SetTag("grpc client receive", string(rspRaw))
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
	zeusctx.ExtractLogger(ctx).Debug("clientLogWrap")
	err = l.Client.Call(ctx, req, rsp, opts...)
	if err != nil {
		// gomicro错误解包为zeus错误
		var gmErr *gmerrors.Error
		if errors.As(err, &gmErr) {
			if gmErr != nil {
				zeusErr := zeuserrors.New(zeuserrors.ErrorCode(gmErr.Code), gmErr.Detail, gmErr.Status)
				zeusErr.ServerID = gmErr.Id
				if utils.IsEmptyString(zeusErr.ServerID) && l.ng != nil {
					zeusErr.ServerID = l.ng.GetContainer().GetServerID()
				}
				span := opentracing.SpanFromContext(ctx)
				zeusErr.TracerID = span.Context().(zipkintracer.SpanContext).TraceID.ToHex()
				err = zeusErr
				return
			}
			err = nil
		}
	}
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
