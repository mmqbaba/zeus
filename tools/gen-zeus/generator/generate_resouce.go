package generator

import (
	"fmt"
	"strings"
)

func GenerateResource(PD *Generator, rootdir string) (err error) {
	err = genResourceDao(PD, rootdir)
	if err != nil {
		return
	}
	err = genResourceCache(PD, rootdir)
	if err != nil {
		return
	}

	err = genResourceRpcClient(PD, rootdir)
	if err != nil {
		return
	}

	err = genResourceHttpClient(PD, rootdir)
	if err != nil {
		return
	}

	return
}

func genResourceDao(PD *Generator, rootdir string) error {
	header := ``
	context := `package dao

`
	fn := GetTargetFileName(PD, "resource.dao", rootdir)
	return writeContext(fn, header, context, false)
}

func genResourceCache(PD *Generator, rootdir string) error {
	header := ``
	context := `package cache

`
	fn := GetTargetFileName(PD, "resource.cache", rootdir)
	return writeContext(fn, header, context, false)
}

func genResourceRpcClient(PD *Generator, rootdir string) error {

	header := ``
	context := `package rpcclient
import (
	"context"
	"sync"

	"github.com/micro/go-micro/client"

	zeusctx "gitlab.dg.com/BackEnd/deliver/tif/zeus/context"

	gomicro "%s{PKG}/proto/{PKG}pb"
)

var {PKG}Srv {SRV}Service

type {SRV}Service struct {
	gomicro.{SRV}Service
	once sync.Once
	name   string
}

func New{SRV}Service(ctx context.Context) (gomicro.{SRV}Service, error) {
	var err error
	{PKG}Srv.once.Do(func() {
		var cli client.Client
		cli, err = zeusctx.ExtractGMClient(ctx)
		if err != nil {
			return
		}
		{PKG}Srv.name = "{PKG}"
		{PKG}Srv.{SRV}Service = gomicro.New{SRV}Service({PKG}Srv.name, cli)
	})
	if err != nil {
		return nil, err
	}
	return &{PKG}Srv, nil
}

`
	camelSrvName := CamelCase(PD.SvrName)

	context = strings.ReplaceAll(context, "{PKG}", PD.PackageName)
	context = strings.ReplaceAll(context, "{SRV}", camelSrvName)
	context = fmt.Sprintf(context, projectBasePrefix)

	fn := GetTargetFileName(PD, "resource.rpcclient", rootdir)
	return writeContext(fn, header, context, false)
}

func genResourceHttpClient(PD *Generator, rootdir string) error {

	header := ``
	tmpContext := `package httpclient


`
	context := tmpContext
	fn := GetTargetFileName(PD, "resource.httpclient", rootdir)
	return writeContext(fn, header, context, false)
}
