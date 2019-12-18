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
	brokerSource := map[string]config.Broker{
		"zeus": config.Broker{
			Type:        "kafka",
			TopicPrefix: "dev",
			Hosts:       []string{"10.1.8.14:9094"},
			SubscribeTopics: []*config.TopicInfo{
				&config.TopicInfo{Category: "sample", Source: "zeus", Queue: "cache-zeus"},
				&config.TopicInfo{Category: "pbstruct", Source: "zeus", Queue: "cache-zeus"},
			},
		},
		"dzqz": config.Broker{
			Type:        "kafka",
			TopicPrefix: "dev",
			Hosts:       []string{"10.1.8.14:9094"},
			SubscribeTopics: []*config.TopicInfo{
				&config.TopicInfo{Category: "samplerequest", Source: "zeus", Queue: "cache-dzqz"},
				&config.TopicInfo{Category: "jsonrequest", Source: "zeus", Queue: "cache-dzqz"},
			},
		},
		"zeus-rb": config.Broker{
			Type:  "rabbitmq",
			Hosts: []string{"amqp://guest:guest@127.0.0.1:5672"},
			// ExchangeName: "zeus",
			// ExchangeKind: "direct",
			SubscribeTopics: []*config.TopicInfo{
				&config.TopicInfo{Topic: "pbstruct.zeus", Queue: "cache-zeus-rb"}, // 这里的Topic 是作为 routing-key
			},
		},
	}
	handlers := map[string]map[string]interface{}{
		"zeus": map[string]interface{}{
			"sample.zeus":   SampleHandler,
			"pbstruct.zeus": PBStructHandler,
		},
		"dzqz": map[string]interface{}{
			"samplerequest.zeus": SampleRequestHandler,
			"jsonrequest.zeus":   JSONRequestHandler,
		},
		"zeus-rb": map[string]interface{}{
			"pbstruct.zeus": RawDataHandler,
		},
	}
	confList := make(map[string]*zsub.SubConfig)
	for s, b := range brokerSource {
		// 此处需要注意：b为结构体类型，不能直接指针&b赋值，需要使用中间变量复制b结构体值
		tmp := b
		sc := &zsub.SubConfig{
			BrokerConf: &tmp,
			Handlers:   handlers[s],
		}
		confList[s] = sc
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

func RawDataHandler(ctx context.Context, msg *[]byte) error {
	m, _ := metadata.FromContext(ctx)
	log.Println("metadata", m)
	log.Printf("RawDataHandler %s\n", string(*msg))
	return nil
}
