package zpub

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/micro/go-micro"
	gmbroker "github.com/micro/go-micro/broker"
	"github.com/micro/go-micro/client"

	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/config"
	zbroker "gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/pubsub/broker"
	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/pubsub/pb/broker"
)

type pubClient struct {
	topicPrefix  string
	cli          client.Client
	publishers   map[string]micro.Publisher
	wrPublishers sync.RWMutex
}

func (pc *pubClient) publish(ctx context.Context, header *broker.Header, msg interface{}) (err error) {
	topic := fmt.Sprintf("%s.%s.%s", pc.topicPrefix, header.Category, header.Source)
	if p, ok := pc.publishers[topic]; ok && p != nil {
		return p.Publish(ctx, msg)
	}
	pc.wrPublishers.Lock()
	defer pc.wrPublishers.Unlock()
	p := micro.NewPublisher(topic, pc.cli)
	pc.publishers[topic] = p
	return p.Publish(ctx, msg)
}

// Publish 发布消息
func Publish(ctx context.Context, header *broker.Header, msg interface{}) error {
	if defaultPubClient == nil {
		return errors.New("DefaultPubClient未初始化")
	}
	return defaultPubClient.publish(ctx, header, msg)
}

func GetClient() client.Client {
	if defaultPubClient == nil {
		log.Println("DefaultPubClient未初始化")
		return nil
	}
	return defaultPubClient.cli
}

var defaultPubClient *pubClient

func newC(conf *config.Broker, cli client.Client) (c *pubClient, err error) {
	topicPrefix := conf.TopicPrefix
	if strings.TrimSpace(conf.TopicPrefix) == "" {
		topicPrefix = "broker"
	}
	c = new(pubClient)
	c.topicPrefix = topicPrefix
	if cli == nil {
		var b gmbroker.Broker
		b, err = zbroker.New(conf)
		if err != nil {
			log.Println(err)
			return
		}
		// 初始化并连接broker
		if err = b.Init(); err != nil {
			log.Println("newC b.Init err:", err)
			return
		}
		if err = b.Connect(); err != nil {
			log.Println("newC b.Connect err:", err)
			return
		}
		cli = client.NewClient(
			client.Broker(b),
		)
		// 初始化
		if err = cli.Init(); err != nil {
			log.Println(err)
			return
		}
	}
	c.cli = cli
	log.Printf("newC c.cli.Options().Broker.Options().Addrs========%s\n", c.cli.Options().Broker.Options().Addrs)
	// log.Printf("newC c.cli.Options().Broker.Address()========%s\n", c.cli.Options().Broker.Address())
	c.publishers = make(map[string]micro.Publisher)
	return
}

var onceDefaultInit sync.Once

// InitDefault 初始化
func InitDefault(conf *config.Broker) {
	var err error
	onceDefaultInit.Do(func() {
		defaultPubClient, err = newC(conf, nil)
		if err != nil {
			log.Println(err)
			panic(err)
		}
	})
	log.Println("init default pubClient")
}

// ReloadDefault 重载
func ReloadDefault(conf *config.Broker) (err error) {
	if defaultPubClient == nil {
		log.Println("DefaultPubClient未初始化")
		return errors.New("DefaultPubClient未初始化")
	}

	if err = defaultPubClient.cli.Options().Broker.Disconnect(); err != nil {
		log.Println(err)
		return
	}

	var tmp *pubClient
	tmp, err = newC(conf, nil)
	if err != nil {
		log.Println(err)
		return
	}

	defaultPubClient = tmp
	log.Println("reload default pubClient")
	return
}
