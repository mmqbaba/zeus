package zsub

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/micro/go-micro"
	"github.com/micro/go-micro/server"

	gmbroker "github.com/micro/go-micro/broker"
	"github.com/mmqbaba/zeus/config"
	zbroker "github.com/mmqbaba/zeus/pubsub/broker"
	"github.com/mmqbaba/zeus/utils"
)

type subServer struct {
	topicPrefix string
	topics      []*config.TopicInfo
	srv         server.Server
	// subscribers map[string]server.Subscriber
	wr       sync.RWMutex
	handlers map[string]interface{}
}

func (ss *subServer) subscribe(ctx context.Context, handlers map[string]interface{}) (err error) {
	for _, t := range ss.topics {
		key := t.Topic
		if utils.IsEmptyString(key) {
			key = fmt.Sprintf("%s.%s", t.Category, t.Source)
		}
		h, ok := handlers[key]
		if h == nil || !ok {
			log.Printf("%s topic handler was nil\n", key)
			continue
		}
		actualTopic := t.Topic
		if utils.IsEmptyString(actualTopic) {
			actualTopic = fmt.Sprintf("%s.%s.%s", ss.topicPrefix, t.Category, t.Source)
		}
		log.Printf("micro.RegisterSubscriber topic:%s, queue:%s", actualTopic, t.Queue)
		// subscriber := server.NewSubscriber(actualTopic, h)
		// err = ss.srv.Subscribe(subscriber)
		err = micro.RegisterSubscriber(actualTopic, ss.srv, h, server.SubscriberQueue(t.Queue))
		if err != nil {
			log.Println(err)
			return
		}
	}
	return
}

func (ss *subServer) start(ctx context.Context) (err error) {
	err = ss.srv.Start()
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("=========================[subscribe] start srv=====================")
	return
}

func (ss *subServer) stop(ctx context.Context) (err error) {
	err = ss.srv.Stop()
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("=========================[unsubscribe] stop srv=====================")
	return
}

func (ss *subServer) run(ctx context.Context) (err error) {
	if err = ss.start(ctx); err != nil {
		return
	}
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)
	log.Println("=========================[Run subscribe] waiting syscall=====================")
	log.Printf("SubServer Received signal %s\n", <-ch)
	return ss.stop(ctx)
}

func newS(conf *config.Broker, srv server.Server) (s *subServer, err error) {
	topicPrefix := conf.TopicPrefix
	if strings.TrimSpace(conf.TopicPrefix) == "" {
		topicPrefix = "broker"
	}
	s = new(subServer)
	s.topicPrefix = topicPrefix
	if srv == nil {
		var b gmbroker.Broker
		b, err = zbroker.New(conf)
		if err != nil {
			log.Println(err)
			return
		}
		// 初始化并连接broker
		if err = b.Init(); err != nil {
			log.Println("newS b.Init err:", err)
			return
		}
		if err = b.Connect(); err != nil {
			log.Println("newS b.Connect err:", err)
			return
		}
		// log.Printf("newS b.Options().Addrs========%+v\n", b.Options().Addrs)
		// log.Printf("newS b.Address()========%+v\n", b.Address())
		srv = server.NewServer(
			server.Broker(b),
			// server.Codec("application/json", server.DefaultCodecs["application/json"]),
		)
		// 初始化
		if err = srv.Init(); err != nil {
			log.Println(err)
			return
		}
	}
	s.srv = srv
	log.Printf("newS srv.Options().Options().Addrs========%s\n", srv.Options().Broker.Options().Addrs)
	s.topics = conf.SubscribeTopics
	return
}

// Subscribe 订阅
func Subscribe(ctx context.Context, handlers map[string]interface{}) (err error) {
	if defaultSubServer == nil {
		log.Println("defaultSubServer未初始化")
		return errors.New("defaultSubServer未初始化")
	}
	if err = defaultSubServer.subscribe(ctx, handlers); err != nil {
		return
	}
	return
}

// Run 开启订阅
func Run(ctx context.Context) (err error) {
	if defaultSubServer == nil {
		log.Println("defaultSubServer未初始化")
		return errors.New("defaultSubServer未初始化")
	}
	return defaultSubServer.run(ctx)
}

func GetServer() server.Server {
	if defaultSubServer == nil {
		log.Println("defaultSubServer未初始化")
		return nil
	}
	return defaultSubServer.srv
}

var defaultSubServer *subServer

var onceDefaultInit sync.Once

// InitDefault 初始化
func InitDefault(conf *config.Broker) {
	var err error
	onceDefaultInit.Do(func() {
		defaultSubServer, err = newS(conf, nil)
		if err != nil {
			log.Println(err)
			panic(err)
		}
	})
	log.Println("init default subServer")
}

// TODO: 完善重载功能
// // ReloadDefault 重载
// func ReloadDefault(conf *config.Broker) (err error) {
// 	if defaultSubServer == nil {
// 		log.Println("defaultSubServer未初始化")
// 		return errors.New("defaultSubServer未初始化")
// 	}

// 	err = defaultSubServer.srv.Stop()
// 	if err != nil {
// 		log.Println(err)
// 		return
// 	}

// 	var tmp *subServer
// 	tmp, err = newS(conf, nil)
// 	if err != nil {
// 		log.Println(err)
// 		return
// 	}

// 	defaultSubServer = tmp
// 	log.Println("reload default subServer")
// 	return
// }
