package gomicro

import (
	"context"

	"github.com/sirupsen/logrus"
	"github.com/micro/go-micro/client"
	"github.com/micro/go-micro/server"

	zeusctx "gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/context"
	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/engine"
)

type validator interface {
	Validate() error
}

func GenerateServerLogWrap(ng engine.Engine) func(fn server.HandlerFunc) server.HandlerFunc {
	return func(fn server.HandlerFunc) server.HandlerFunc {
		return func(ctx context.Context, req server.Request, rsp interface{}) (err error) {
			if v, ok := req.Body().(validator); ok && v != nil {
				if err = v.Validate(); err != nil {
					return
				}
			}
			logger := ng.GetContainer().GetLogger()
			c := zeusctx.LoggerToContext(ctx, logger.WithFields(logrus.Fields{"tag": "gomicro-serverlogwrap"}))
			c = zeusctx.EngineToContext(c, ng)
			c = zeusctx.GMClientToContext(c, ng.GetContainer().GetGoMicroClient())
			c = zeusctx.RedisToContext(c, ng.GetContainer().GetRedisCli().GetCli())
			err = fn(c, req, rsp)
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

func (l *clientLogWrap) Call(ctx context.Context, req client.Request, rsp interface{}, opts ...client.CallOption) error {
	err := l.Client.Call(ctx, req, rsp, opts...)
	return err
}
