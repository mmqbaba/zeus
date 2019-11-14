package gomicro

import (
	"context"
	"errors"

	"github.com/micro/go-micro/client"
	gmerrors "github.com/micro/go-micro/errors"
	"github.com/micro/go-micro/server"
	"github.com/sirupsen/logrus"

	zeusctx "gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/context"
	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/engine"
	zeuserrors "gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/errors"
)

type validator interface {
	Validate() error
}

func GenerateServerLogWrap(ng engine.Engine) func(fn server.HandlerFunc) server.HandlerFunc {
	return func(fn server.HandlerFunc) server.HandlerFunc {
		return func(ctx context.Context, req server.Request, rsp interface{}) (err error) {
			// TODO: ctx添加tracer

			logger := ng.GetContainer().GetLogger()
			c := zeusctx.LoggerToContext(ctx, logger.WithFields(logrus.Fields{"tag": "gomicro-serverlogwrap"}))
			c = zeusctx.EngineToContext(c, ng)
			c = zeusctx.GMClientToContext(c, ng.GetContainer().GetGoMicroClient())
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
}

func (l *clientLogWrap) Call(ctx context.Context, req client.Request, rsp interface{}, opts ...client.CallOption) (err error) {
	zeusctx.ExtractLogger(ctx).Debug("clientLogWrap")
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

type clientWrapTest struct {
	client.Client
}

func (l *clientWrapTest) Call(ctx context.Context, req client.Request, rsp interface{}, opts ...client.CallOption) (err error) {
	zeusctx.ExtractLogger(ctx).Debug("clientWrapTest")
	err = l.Client.Call(ctx, req, rsp, opts...)
	return
}
