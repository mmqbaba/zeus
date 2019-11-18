package gomicro

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/micro/go-grpc"
	"github.com/micro/go-micro"
	"github.com/micro/go-micro/registry"
	"github.com/micro/go-plugins/registry/etcdv3"

	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/config"
)

func NewService(ctx context.Context, conf config.GoMicro, opts ...micro.Option) micro.Service {
	// discovery/registry
	reg := etcdv3.NewRegistry(
		registry.Addrs(conf.RegistryAddrs...),
		etcdv3.Auth(conf.RegistryAuthUser, conf.RegistryAutPwd),
	)

	o := []micro.Option{
		micro.Name(conf.ServerName),
		micro.Address(fmt.Sprintf(":%d", conf.ServerPort)),
		micro.RegisterTTL(30 * time.Second),
		micro.RegisterInterval(20 * time.Second),
		micro.Registry(reg),
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
	s := grpc.NewService(o...)
	// // parse command line flags.
	// s.Init() // 禁用掉，不parse
	return s
}
