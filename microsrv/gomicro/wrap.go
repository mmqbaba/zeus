package gomicro

import (
	"context"
	"log"

	"github.com/micro/go-micro/client"
	"github.com/micro/go-micro/server"

	zeusctx "gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/context"
	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/engine"
)

func GenerateServerLogWrap(ng engine.Engine) func(fn server.HandlerFunc) server.HandlerFunc {
	return func(fn server.HandlerFunc) server.HandlerFunc {
		return func(ctx context.Context, req server.Request, rsp interface{}) (err error) {
			log.Printf("server wrap log begin, Endpoint: %s, %T, %T\n", req.Endpoint(), req.Body(), rsp)
			c := context.WithValue(ctx, "ZEUS_NG", ng)
			c = zeusctx.RedisToContext(c, ng.GetContainer().GetRedisCli().GetCli())
			err = fn(c, req, rsp)
			log.Println("server wrap log end")
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
	log.Printf("客户端请求服务开始：%s，方法：%s\n", req.Service(), req.Endpoint())
	ng := ctx.Value("ZEUS_NG")
	log.Printf("%T\n", ng)
	err := l.Client.Call(ctx, req, rsp, opts...)
	log.Printf("客户端请求服务结束：%s，方法：%s", req.Service(), req.Endpoint())
	return err
}
