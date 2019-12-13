package main

import (
	"context"
	"log"

	structpb "github.com/golang/protobuf/ptypes/struct"
	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/config"
	brokerpb "gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/pubsub/pb/broker"
	zsub "gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/pubsub/sub"
)

var conf *config.Broker
var confList map[string]*config.Broker

func main() {
	sub()
	// subManager()
}

func sub() {
	conf = &config.Broker{}

	// conf.Type = "redis"
	// conf.TopicPrefix = "dev"

	conf.Type = "kafka"
	conf.TopicPrefix = "dev"
	conf.Hosts = []string{"10.1.8.14:9094"}

	conf.SubscribeTopics = append(conf.SubscribeTopics, &config.TopicInfo{Category: "sample", Source: "zeus", Queue: "cache"})
	zsub.InitDefault(conf)
	handlers := make(map[string]interface{})
	handlers["sample.zeus"] = AnyHandler
	zsub.Subscribe(context.Background(), handlers)
	if err := zsub.Run(context.Background()); err != nil {
		log.Println(err)
		return
	}
}

func subManager() {
	confList = map[string]*config.Broker{
		"dzqz": &config.Broker{
			Type:            "kafka",
			TopicPrefix:     "dev",
			Hosts:           []string{"10.1.8.14:9092"},
			SubscribeTopics: []*config.TopicInfo{&config.TopicInfo{Category: "sample", Source: "zeus", Queue: "cache-1"}},
		},
		"zeus": &config.Broker{
			Type:            "kafka",
			TopicPrefix:     "dev",
			Hosts:           []string{"10.1.8.14:9092"},
			SubscribeTopics: []*config.TopicInfo{&config.TopicInfo{Category: "sample", Source: "zeus", Queue: "cache-2"}},
		},
	}
	handlers := map[string]map[string]interface{}{
		"dzqz": map[string]interface{}{
			"sample.zeus": SampleHandler,
		},
		"zeus": map[string]interface{}{
			"sample.zeus": SampleHandler,
		},
	}
	m, err := zsub.NewManager(confList)
	if err != nil {
		log.Println(err)
		return
	}
	if err = m.Subscribe(context.Background(), handlers); err != nil {
		log.Println(err)
		return
	}
	m.Run(context.Background())
}

func SampleHandler(ctx context.Context, msg *brokerpb.RequestSample) error {
	log.Printf("%+v\n", msg)
	return nil
}

func AnyHandler(ctx context.Context, msg *structpb.Struct) error {
	log.Printf("%+v\n", msg)
	return nil
}
