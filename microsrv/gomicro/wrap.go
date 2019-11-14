package gomicro

import (
	"context"
	"errors"
	"fmt"

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
			// TODO: ctx添加tracer

			logger := ng.GetContainer().GetLogger()
			l := logger.WithFields(logrus.Fields{"tag": "gomicro-serverlogwrap"})
			c := zeusctx.LoggerToContext(ctx, l)
			c = zeusctx.EngineToContext(c, ng)
			c = zeusctx.GMClientToContext(c, ng.GetContainer().GetGoMicroClient())
			///////// tracer begin
			name := fmt.Sprintf("%s.%s", req.Service(), req.Endpoint())
			cfg, err := ng.GetConfiger()
			if err != nil {
				l.Error(err)
				err = &gmerrors.Error{Id: name, Code: int32(zeuserrors.ECodeSystem), Detail: err.Error(), Status: ""}
				return
			}

			tracer := ng.GetContainer().GetTracer()
			if tracer == nil {
				err = fmt.Errorf("tracer is nil")
				l.Error(err)
				err = &gmerrors.Error{Id: name, Code: int32(zeuserrors.ECodeSystem), Detail: err.Error(), Status: ""}
				return
			}
			spnctx, span, err := tracer.StartSpanFromContext(c, name)
			if err != nil {
				l.Error(err)
				err = &gmerrors.Error{Id: name, Code: int32(zeuserrors.ECodeSystem), Detail: err.Error(), Status: ""}
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
			c = spnctx
			///////// tracer finish

			if v, ok := req.Body().(validator); ok && v != nil {
				if err = v.Validate(); err != nil {
					zeusErr := zeuserrors.New(zeuserrors.ECodeInvalidParams, err.Error(), "validator.Validate")
					err = &gmerrors.Error{Id: zeusErr.ServerID, Code: int32(zeusErr.ErrCode), Detail: zeusErr.ErrMsg, Status: zeusErr.Cause}
					return
				}
			}
			if ng.GetContainer().GetRedisCli() != nil {
				c = zeusctx.RedisToContext(c, ng.GetContainer().GetRedisCli().GetCli())
			}
			err = fn(c, req, rsp)
			if err != nil {
				span.SetTag("grpc client receive error", err)
				// zeus错误包装为gomicro错误
				var zeusErr *zeuserrors.Error
				if errors.As(err, &zeusErr) {
					if zeusErr != nil {
						err = &gmerrors.Error{Id: zeusErr.ServerID, Code: int32(zeusErr.ErrCode), Detail: zeusErr.ErrMsg, Status: zeusErr.Cause}
						return
					}
					err = nil
				}
			}
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
			// ng: ng,
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
}

func (l *clientLogWrap) Call(ctx context.Context, req client.Request, rsp interface{}, opts ...client.CallOption) (err error) {
	err = l.Client.Call(ctx, req, rsp, opts...)
	if err != nil {
		// gomicro错误解包为zeus错误
		var gmErr *gmerrors.Error
		if errors.As(err, &gmErr) {
			if gmErr != nil {
				err = zeuserrors.New(zeuserrors.ErrorCode(gmErr.Code), gmErr.Detail, gmErr.Status)
				return
			}
			err = nil
		}
	}
	return
}
