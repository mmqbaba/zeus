package zgomicro

import (
	"context"
	"fmt"
	"log"
	"time"

	// grpcserver "github.com/micro/go-grpc/server"
	"github.com/micro/go-micro"
	"github.com/micro/go-micro/registry"
	"github.com/micro/go-micro/server"
	"github.com/micro/go-micro/server/grpc"
	"github.com/micro/go-plugins/registry/etcdv3"

	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/config"
)

func NewService(ctx context.Context, conf config.GoMicro, opts ...micro.Option) micro.Service {
	// discovery/registry
	var reg registry.Registry
	switch conf.RegistryPluginType {
	case "etcd":
		reg = etcdv3.NewRegistry(
			registry.Addrs(conf.RegistryAddrs...),
			etcdv3.Auth(conf.RegistryAuthUser, conf.RegistryAuthPwd),
		)
	default:
		reg = registry.DefaultRegistry
	}

	// grpcS := grpcserver.NewServer(
	// 	server.Advertise(conf.Advertise),
    // )
    grpcS := grpc.NewServer(
        server.Advertise(conf.Advertise),
    )

	o := []micro.Option{
		micro.Server(grpcS),
		micro.Registry(reg),
		micro.Name(conf.ServiceName),
		micro.Address(fmt.Sprintf(":%d", conf.ServerPort)),
		micro.RegisterTTL(15 * time.Second),
		micro.RegisterInterval(10 * time.Second),
		micro.AfterStop(func() error {
			log.Println("[gomicro] afterstop")
			return nil
		}),
		// micro.Flags(
		// 	cli.StringFlag{
		// 		Name:  "string_flag",
		// 		Usage: "This is a string flag",
		// 		Value: "test_string_flag",
		// 	},
		// ),
		// micro.Action(func(c *cli.Context) {
		// 	log.Printf("[micro.Action] called when s.Init(), cli.Context flag\n")
		// 	log.Printf("[micro.Action] The string flag is: %s\n", c.String("string_flag"))
		// }),
	}
	o = append(o, opts...)
	// new micro service
	// s := grpc.NewService(o...)
	s := micro.NewService(o...)
	// // parse command line flags.
	// s.Init() // 禁用掉，不parse
	return s
}
