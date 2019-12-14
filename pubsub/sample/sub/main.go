package main

import (
	"context"
	"log"

	structpb "github.com/golang/protobuf/ptypes/struct"
	"github.com/micro/go-micro/metadata"
	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/config"
	brokerpb "gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/pubsub/proto"
	zsub "gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/pubsub/sub"
	zprotobuf "gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/utils/protobuf"
)

func main() {
	// sub()
	subManager()
}

func sub() {
	conf := &config.Broker{}

	// conf.Type = "redis"
	// conf.TopicPrefix = "dev"

	conf.Type = "kafka"
	conf.TopicPrefix = "dev"
	conf.Hosts = []string{"10.1.8.14:9094"}

	conf.SubscribeTopics = append(conf.SubscribeTopics, &config.TopicInfo{Category: "sample", Source: "zeus", Queue: "cache"})
	conf.SubscribeTopics = append(conf.SubscribeTopics, &config.TopicInfo{Category: "samplerequest", Source: "zeus", Queue: "cache"})
	conf.SubscribeTopics = append(conf.SubscribeTopics, &config.TopicInfo{Category: "pbstruct", Source: "zeus", Queue: "cache"})
	conf.SubscribeTopics = append(conf.SubscribeTopics, &config.TopicInfo{Category: "jsonrequest", Source: "zeus", Queue: "cache"})
	zsub.InitDefault(conf)
	handlers := make(map[string]interface{})
	handlers["sample.zeus"] = SampleHandler
	handlers["samplerequest.zeus"] = SampleRequestHandler
	handlers["pbstruct.zeus"] = PBStructHandler
	handlers["jsonrequest.zeus"] = JSONRequestHandler
	zsub.Subscribe(context.Background(), handlers)
	if err := zsub.Run(context.Background()); err != nil {
		log.Println(err)
		return
	}
}

func subManager() {
	confList := map[string]*zsub.SubConfig{
		"zeus": &zsub.SubConfig{
			BrokerConf: &config.Broker{
				Type:        "kafka",
				TopicPrefix: "dev",
				Hosts:       []string{"10.1.8.14:9094"},
				SubscribeTopics: []*config.TopicInfo{
					&config.TopicInfo{Category: "sample", Source: "zeus", Queue: "cache-zeus"},
					&config.TopicInfo{Category: "pbstruct", Source: "zeus", Queue: "cache-zeus"},
				},
			},
			Handlers: map[string]interface{}{
				"sample.zeus":   SampleHandler,
				"pbstruct.zeus": PBStructHandler,
			},
		},
		"dzqz": &zsub.SubConfig{
			BrokerConf: &config.Broker{
				Type:        "kafka",
				TopicPrefix: "dev",
				Hosts:       []string{"10.1.8.14:9094"},
				SubscribeTopics: []*config.TopicInfo{
					&config.TopicInfo{Category: "samplerequest", Source: "zeus", Queue: "cache-dzqz"},
					&config.TopicInfo{Category: "jsonrequest", Source: "zeus", Queue: "cache-dzqz"},
				},
			},
			Handlers: map[string]interface{}{
				"samplerequest.zeus": SampleRequestHandler,
				"jsonrequest.zeus":   JSONRequestHandler,
			},
		},
	}
	mc := &zsub.ManagerConfig{
		Conf: confList,
	}
	m, err := zsub.NewManager(mc)
	if err != nil {
		log.Println(err)
		return
	}
	if err = m.Subscribe(context.Background()); err != nil {
		log.Println(err)
		return
	}
	m.Run(context.Background())
}

func SampleHandler(ctx context.Context, msg *brokerpb.Sample) error {
	m, _ := metadata.FromContext(ctx)
	log.Println("metadata", m)
	log.Printf("SampleHandler %+v\n", msg)
	return nil
}

func SampleRequestHandler(ctx context.Context, msg *brokerpb.RequestSample) error {
	m, _ := metadata.FromContext(ctx)
	log.Println("metadata", m)
	log.Printf("SampleRequestHandler %+v\n", msg)
	return nil
}

func PBStructHandler(ctx context.Context, msg *structpb.Struct) error {
	m, _ := metadata.FromContext(ctx)
	log.Println("metadata", m)
	req := zprotobuf.DecodeToMap(msg)
	// log.Printf("PBStructHandler %+v\n", msg)
	log.Printf("PBStructHandler %+v\n", req)
	return nil
}

type Request struct {
	Message  string `json:"message"`
	Count    int32  `json:"count"`
	Finished bool   `json:"finished"`
}

func JSONRequestHandler(ctx context.Context, msg *Request) error {
	m, _ := metadata.FromContext(ctx)
	log.Println("metadata", m)
	log.Printf("JSONRequestHandler %+v\n", msg)
	return nil
}
