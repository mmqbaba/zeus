package zpub

import (
	"context"
	"fmt"
	"io"
	"log"

	gmbroker "github.com/micro/go-micro/broker"
	"github.com/micro/go-micro/client"
	"github.com/micro/go-micro/codec"

	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/config"
	zjson "gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/microsrv/gomicro/codec/json"
	zbroker "gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/pubsub/broker"
	broker "gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/pubsub/proto"
)

type Manager struct {
	pubs    map[string]*pubClient
	options *ManagerConfig
}

type ManagerConfig struct {
	Conf        map[string]*PubConfig
	JSONCodecFn func(io.ReadWriteCloser) codec.Codec
}

type PubConfig struct {
	BrokerConf *config.Broker
}

func NewManager(mc *ManagerConfig) (m *Manager, err error) {
	tmp := &Manager{
		pubs:    make(map[string]*pubClient),
		options: mc,
	}

	for k, c := range mc.Conf {
		var pubCli *pubClient
		var b gmbroker.Broker
		b, err = zbroker.New(c.BrokerConf)
		if err != nil {
			log.Println(err)
			return
		}
		// 初始化并连接broker
		if err = b.Init(); err != nil {
			log.Println("[pub.NewManager] b.Init err:", err)
			return
		}
		if err = b.Connect(); err != nil {
			log.Println("[pub.NewManager] b.Connect err:", err)
			return
		}
		jsonCodeFn := zjson.NewCodec
		if mc.JSONCodecFn != nil {
			jsonCodeFn = mc.JSONCodecFn
		}
		cli := client.NewClient(
			client.Broker(b),
			client.ContentType("application/json"),
			client.Codec("application/json", jsonCodeFn),
		)
		// 初始化
		if err = cli.Init(); err != nil {
			log.Println(err)
			return
		}
		pubCli, err = newC(c.BrokerConf, nil)
		if err != nil {
			log.Println("[pub.NewManager] newC err:", err)
			return
		}
		tmp.pubs[k] = pubCli
	}

	m = tmp
	return
}

// Publish 发布消息
//
// targetInstance 目标数据源
//
// header 消息topic等相关信息
//
// msg 消息体(struct)
func (m *Manager) Publish(ctx context.Context, targetInstance string, header *broker.Header, msg interface{}) (err error) {
	pubCli, ok := m.pubs[targetInstance]
	if !ok || pubCli == nil {
		err = fmt.Errorf("[Manager.Publish] the pubClient was nil, targetInstance: %s", targetInstance)
		log.Println(err)
		return
	}
	if err = pubCli.publish(ctx, header, msg); err != nil {
		log.Println("[Manager.Publish] err:", err)
		return
	}
	return
}

// Release 释放broker连接资源
func (m *Manager) Release() (err error) {
	for k, pc := range m.pubs {
		if err = pc.cli.Options().Broker.Disconnect(); err != nil {
			log.Printf("[Manager.Release] broker: %s, pc.cli.Options().Broker.Disconnect() err: %s\n", k, err)
			return
		}
	}
	return
}
