package main

import (
	"context"
	"log"

	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/config"
	brokerpb "gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/pubsub/pb/broker"
	zsub "gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/pubsub/sub"
)

var conf *config.Broker

func main() {
	conf = &config.Broker{}
	sub()
}

func sub() {
	// conf.Type = "redis"
	// conf.TopicPrefix = "dev"

	conf.Type = "kafka"
	conf.TopicPrefix = "dev"
	conf.Hosts = []string{"10.1.8.14:9092"}

	conf.SubscribeTopics = append(conf.SubscribeTopics, &config.TopicInfo{Category: "sample", Source: "zeus", Queue: "cache"})
	zsub.InitDefault(conf)
	handlers := make(map[string]interface{})
	handlers["sample.zeus"] = SampleHandler
	zsub.Subscribe(context.Background(), handlers)
	if err := zsub.Run(context.Background()); err != nil {
		log.Println(err)
		return
	}
}

func SampleHandler(ctx context.Context, msg *brokerpb.Sample) error {
	log.Printf("%+v\n", msg)
	return nil
}
