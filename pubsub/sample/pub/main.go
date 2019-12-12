package main

import (
	"context"
	"fmt"
	"time"

	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/config"
	brokerpb "gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/pubsub/pb/broker"
	zpub "gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/pubsub/pub"
)

var conf *config.Broker

func main() {
	conf = &config.Broker{}
	pub()
}

func pub() {
	// conf.Type = "redis"
	// conf.TopicPrefix = "dev"

	conf.Type = "kafka"
	conf.TopicPrefix = "dev"
	conf.Hosts = []string{"10.1.8.14:9092"}

	zpub.InitDefault(conf)
	for {
		fmt.Printf("broker type: %T, %s\n", zpub.GetClient().Options().Broker, conf.Type)
		h := new(brokerpb.Header)
		h.Id = fmt.Sprint(time.Now().Unix())
		h.Category = "sample"
		h.Source = "zeus"
		body := new(brokerpb.Sample_Body)
		body.Id = "body-" + h.Id
		body.Name = "test Sample"
		msg := new(brokerpb.Sample)
		msg.Header = h
		msg.Body = body
		err := zpub.Publish(context.Background(), h, msg)
		if err != nil {
			fmt.Println("err==>>", err)
		}
		time.Sleep(time.Second)
	}
}
