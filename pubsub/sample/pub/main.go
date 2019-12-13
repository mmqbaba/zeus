package main

import (
	"context"
	"fmt"
	"log"
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
	conf.Hosts = []string{"10.1.8.14:9094"}

	zpub.InitDefault(conf)
	for {
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
		err := zpub.Publish(context.Background(), h, "hello")
		log.Println("hello")
		if err != nil {
			fmt.Println("err==>>", err)
		}
		// err = zpub.Publish(context.Background(), h, msg)
		// log.Println(msg)
		// if err != nil {
		// 	fmt.Println("err==>>", err)
		// }
		// rMsg := &brokerpb.RequestSample{Message: "hello world!", Count: 100, Finished: true}
		// err = zpub.Publish(context.Background(), h, rMsg)
		// log.Println(rMsg)
		// if err != nil {
		// 	fmt.Println("err==>>", err)
		// }
		time.Sleep(time.Second)
	}
}
