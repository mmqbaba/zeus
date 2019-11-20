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

	zeusctx "gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/context"

	gomicro "%s{PKG}/proto/{PKG}pb"
)

var cli client.Client

type {SRV}Service struct {
	mux    sync.RWMutex
	name   string
	client gomicro.{SRV}Service
}

// {PKG}Srv
var {PKG}Srv {SRV}Service

func New{SRV}Service(ctx context.Context) (gomicro.{SRV}Service, error) {
	{PKG}Srv.mux.RLock()
	if {PKG}Srv.client != nil {
		defer {PKG}Srv.mux.RUnlock()
		return {PKG}Srv.client, nil
	}
	{PKG}Srv.mux.RUnlock()

	{PKG}Srv.mux.Lock()
	defer {PKG}Srv.mux.Unlock()
	cli, err := zeusctx.ExtractGMClient(ctx)
	if err != nil {
		return nil, err
	}
	{PKG}Srv.name = "{PKG}"
	{PKG}Srv.client = gomicro.New{SRV}Service({PKG}Srv.name, cli)
	return {PKG}Srv.client, nil
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
