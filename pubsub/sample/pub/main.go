package main

import (
	"context"
	"fmt"
	"log"
	"time"

	structpb "github.com/golang/protobuf/ptypes/struct"
	"github.com/micro/go-micro/metadata"
	"github.com/mmqbaba/zeus/config"
	brokerpb "github.com/mmqbaba/zeus/pubsub/proto"
	zpub "github.com/mmqbaba/zeus/pubsub/pub"
)

func main() {
	// pub()
	pubManager()
}

type Request struct {
	Message  string `json:"message"`
	Count    int32  `json:"count"`
	Finished bool   `json:"finished"`
}

func pub() {
	conf := &config.Broker{}

	// conf.Type = "redis"
	// conf.TopicPrefix = "dev"

	conf.Type = "kafka"
	conf.TopicPrefix = "dev"
	conf.Hosts = []string{"10.1.8.14:9094"}
	// conf.PubWithOriginalData = true

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
		log.Printf("Publish Sample message: %+v\n", msg)
		err := zpub.Publish(context.Background(), h, msg)
		if err != nil {
			fmt.Println("err==>>", err)
		}

		sreq := &brokerpb.RequestSample{Message: "hello world!", Count: 100, Finished: true}
		log.Printf("Publish SampleRequest message: %+v\n", sreq)
		err = zpub.Publish(context.Background(), &brokerpb.Header{Id: fmt.Sprint(time.Now().Unix()), Category: "samplerequest", Source: "zeus"}, sreq)
		if err != nil {
			fmt.Println("err==>>", err)
		}

		sMsg := &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"message": &structpb.Value{
					Kind: &structpb.Value_StringValue{StringValue: "hello world!"},
				},
				"count": &structpb.Value{
					Kind: &structpb.Value_NumberValue{NumberValue: 100},
				},
				"finished": &structpb.Value{
					Kind: &structpb.Value_BoolValue{BoolValue: true},
				},
			},
		}
		log.Printf("Publish pbstruct message: %+v\n", sMsg)
		err = zpub.Publish(context.Background(), &brokerpb.Header{Id: fmt.Sprint(time.Now().Unix()), Category: "pbstruct", Source: "zeus"}, sMsg)
		if err != nil {
			fmt.Println("err==>>", err)
		}

		req := &Request{Message: "hello world!", Count: 100, Finished: true}
		log.Printf("Publish JSON Request message: %+v\n", req)
		err = zpub.Publish(metadata.NewContext(context.Background(), metadata.Metadata{"for-handler": "json-request"}), &brokerpb.Header{Id: fmt.Sprint(time.Now().Unix()), Category: "jsonrequest", Source: "zeus"}, req)
		if err != nil {
			fmt.Println("err==>>", err)
		}

		time.Sleep(time.Second)
	}
}

// pubManager 多个数据源
func pubManager() {
	brokerSource := map[string]config.Broker{
		"zeus": config.Broker{
			Type:                "kafka",
			TopicPrefix:         "dev",
			Hosts:               []string{"10.1.8.14:9094"},
			EnablePub:           true,
			PubWithOriginalData: false,
		},
		"dzqz": config.Broker{
			Type:                "kafka",
			TopicPrefix:         "dev",
			Hosts:               []string{"10.1.8.14:9094"},
			EnablePub:           true,
			PubWithOriginalData: true,
		},
		"zeus-rb": config.Broker{
			Type:  "rabbitmq",
			Hosts: []string{"amqp://guest:guest@127.0.0.1:5672"},
			// ExchangeName: "zeus",
			// ExchangeKind: "direct",
			EnablePub: false,
		},
	}
	confList := make(map[string]*zpub.PubConfig)
	for s, b := range brokerSource {
		if !b.EnablePub {
			continue
		}
		// 此处需要注意：b为结构体类型，不能直接指针&b赋值，需要使用中间变量复制b结构体值
		tmp := b
		pc := &zpub.PubConfig{
			BrokerConf: &tmp,
		}
		confList[s] = pc
	}
	mc := &zpub.ManagerConfig{
		Conf: confList,
		// JSONCodecFn: NewCodec, // 可使用自定义实现的json Codec覆盖
	}
	m, err := zpub.NewManager(mc)
	if err != nil {
		log.Println(err)
		return
	}

	for {
		// h := new(brokerpb.Header)
		// h.Id = fmt.Sprint(time.Now().Unix())
		// h.Category = "sample"
		// h.Source = "zeus"
		// body := new(brokerpb.Sample_Body)
		// body.Id = "body-" + h.Id
		// body.Name = "test Sample"
		// msg := new(brokerpb.Sample)
		// msg.Header = h
		// msg.Body = body
		// log.Printf("Publish Sample message: %+v\n", msg)
		// err := m.Publish(context.Background(), "zeus", h, msg)
		// if err != nil {
		// 	fmt.Println("err==>>", err)
		// }

		sreq := &brokerpb.RequestSample{Message: "hello world!", Count: 100, Finished: true}
		log.Printf("Publish SampleRequest message: %+v\n", sreq)
		err = m.Publish(context.Background(), "zeus", &brokerpb.Header{Category: "samplerequest", Source: "zeus"}, sreq)
		if err != nil {
			fmt.Println("err==>>", err)
		}

		// sMsg := &structpb.Struct{
		// 	Fields: map[string]*structpb.Value{
		// 		"message": &structpb.Value{
		// 			Kind: &structpb.Value_StringValue{StringValue: "hello world!"},
		// 		},
		// 		"count": &structpb.Value{
		// 			Kind: &structpb.Value_NumberValue{NumberValue: 100},
		// 		},
		// 		"finished": &structpb.Value{
		// 			Kind: &structpb.Value_BoolValue{BoolValue: true},
		// 		},
		// 	},
		// }
		// log.Printf("Publish pbstruct message: %+v\n", sMsg)
		// err = m.Publish(context.Background(), "zeus", &brokerpb.Header{Id: fmt.Sprint(time.Now().Unix()), Category: "pbstruct", Source: "zeus"}, sMsg)
		// if err != nil {
		// 	fmt.Println("err==>>", err)
		// }

		req := &Request{Message: "hello world!", Count: 100, Finished: true}
		log.Printf("Publish JSON Request message: %+v\n", req)
		err = m.Publish(metadata.NewContext(context.Background(), metadata.Metadata{"for-handler": "json-request"}), "dzqz", &brokerpb.Header{Id: fmt.Sprint(time.Now().Unix()), Category: "jsonrequest", Source: "zeus"}, req)
		if err != nil {
			fmt.Println("err==>>", err)
		}

		time.Sleep(time.Second)
	}
}
