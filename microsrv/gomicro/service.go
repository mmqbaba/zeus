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
		micro.AfterStart(func() error {
			log.Println("[gomicro] afterstart")
			return nil
		}),
		micro.AfterStop(func() error {
			log.Println("[gomicro] afterstop")
			return nil
		}),
	}
	o = append(o, opts...)
	// new micro service
	s := grpc.NewService(o...)
	// parse command line flags.
	s.Init()
	return s
}
