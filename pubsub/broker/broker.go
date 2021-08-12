package zbroker

import (
	"fmt"
	"log"

	"github.com/Shopify/sarama"
	"github.com/micro/go-micro/broker"
	"github.com/micro/go-plugins/broker/redis"

	"gitlab.dg.com/BackEnd/deliver/tif/zeus/config"
	"gitlab.dg.com/BackEnd/deliver/tif/zeus/pubsub/broker/kafka"
	"gitlab.dg.com/BackEnd/deliver/tif/zeus/pubsub/broker/rabbitmq"
)

type brokerWrap struct {
	mqType string
	broker.Broker
}

func New(conf *config.Broker) (b broker.Broker, err error) {
	log.Printf("new broker type:%s, conf:%+v\n", conf.Type, conf)
	switch conf.Type {
	case "kafka":
		mb := new(brokerWrap)
		mb.mqType = conf.Type
		kconf := sarama.NewConfig()
		kconf.Version = sarama.V0_10_2_0 // 设置kafka最小支持版本
		kconf.Net.SASL.Enable = conf.NeedAuth
		kconf.Net.SASL.User = conf.User
		kconf.Net.SASL.Password = conf.Pwd
		mb.Broker = kafka.NewBroker(
			broker.Addrs(conf.Hosts...),
			kafka.BrokerConfig(kconf),
			kafka.ClusterConfig(kconf),
			kafka.ZBrokerConfig(conf),
		)
		b = mb
		return
	case "rabbitmq":
		mb := new(brokerWrap)
		mb.mqType = conf.Type
		opts := []broker.Option{broker.Addrs(conf.Hosts...), rabbitmq.Exchange(conf.ExchangeName), rabbitmq.ExchangeKind(conf.ExchangeKind)}
		if conf.ExternalAuth {
			opts = append(opts, rabbitmq.ExternalAuth())
		}
		if conf.ExchangeDurable {
			opts = append(opts, rabbitmq.DurableExchange())
		}
		mb.Broker = rabbitmq.NewBroker(opts...)
		b = mb
		return
	case "redis":
		mb := new(brokerWrap)
		mb.mqType = conf.Type
		mb.Broker = redis.NewBroker(
			broker.Addrs(conf.Hosts...),
			redis.ReadTimeout(0),
			redis.WriteTimeout(0),
		)
		b = mb
		return
	default:
		msg := fmt.Sprintf("不支持的broker类型, conf.Type:%s", conf.Type)
		log.Println(msg)
		err = fmt.Errorf(msg)
		return
	}
}
