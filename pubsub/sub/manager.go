package zsub

import (
	"context"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	gmbroker "github.com/micro/go-micro/broker"
	"github.com/micro/go-micro/codec"
	"github.com/micro/go-micro/server"

	"gitlab.dg.com/BackEnd/deliver/tif/zeus/config"
	zjson "gitlab.dg.com/BackEnd/deliver/tif/zeus/microsrv/gomicro/codec/json"
	zbroker "gitlab.dg.com/BackEnd/deliver/tif/zeus/pubsub/broker"
	"gitlab.dg.com/BackEnd/deliver/tif/zeus/utils"
)

type Manager struct {
	subs    map[string]*subServer
	options *ManagerConfig
}

type ManagerConfig struct {
	Conf        map[string]*SubConfig
	JSONCodecFn func(io.ReadWriteCloser) codec.Codec
}

type SubConfig struct {
	BrokerConf *config.Broker
	Handlers   map[string]interface{}
}

func NewManager(mc *ManagerConfig) (m *Manager, err error) {
	tmp := &Manager{
		subs:    make(map[string]*subServer),
		options: mc,
	}
	for k, c := range mc.Conf {
		var ss *subServer
		var b gmbroker.Broker
		b, err = zbroker.New(c.BrokerConf)
		if err != nil {
			log.Println(err)
			return
		}
		// 初始化broker
		if err = b.Init(); err != nil {
			log.Println("newS b.Init err:", err)
			return
		}
		if err = b.Connect(); err != nil {
			log.Println("newS b.Connect err:", err)
			return
		}
		srvOpts := []server.Option{server.Broker(b)}
		jsonCodeFn := zjson.NewCodec
		if mc.JSONCodecFn != nil {
			jsonCodeFn = mc.JSONCodecFn
		}
		srvOpts = append(srvOpts, server.Codec("application/json", jsonCodeFn))
		srv := server.NewServer(srvOpts...)
		// 初始化server
		if err = srv.Init(); err != nil {
			log.Println(err)
			return
		}
		ss, err = newS(c.BrokerConf, srv)
		if err != nil {
			log.Println(err)
			panic(err)
		}
		ss.handlers = c.Handlers
		tmp.subs[k] = ss
	}
	m = tmp
	return
}

func (m *Manager) Subscribe(ctx context.Context) (err error) {
	for _, ss := range m.subs {
		err = ss.subscribe(ctx, ss.handlers)
		if err != nil {
			return
		}
	}
	return
}

func (m *Manager) Run(ctx context.Context) {
	for _, ss := range m.subs {
		tmp := ss
		utils.AsyncFuncSafe(ctx, func(args ...interface{}) {
			err := tmp.run(ctx)
			if err != nil {
				log.Println(err)
				return
			}
		})
	}
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)
	log.Println("=========================[Run subserver manager] waiting syscall=====================")
	log.Printf("Subserver Manager Received signal %s\n", <-ch)
}
