package generator

import (
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

	gomicro "zeus_app/{PKGNAME}/proto/gomicro"
)

var cli client.Client

type helloService struct {
	mux    sync.RWMutex
	name   string
	client gomicro.{SRVNAME}Service
}

// {PKGNAME}Srv
var {PKGNAME}Srv {PKGNAME}Service

func New{SRVNAME}Service(ctx context.Context) (gomicro.{SRVNAME}Service, error) {
	{PKGNAME}Srv.mux.RLock()
	if {PKGNAME}Srv.client != nil {
		defer {PKGNAME}Srv.mux.RUnlock()
		return {PKGNAME}Srv.client, nil
	}
	{PKGNAME}Srv.mux.RUnlock()

	{PKGNAME}Srv.mux.Lock()
	defer {PKGNAME}Srv.mux.Unlock()
	cli, err := zeusctx.ExtractGMClient(ctx)
	if err != nil {
		return nil, err
	}
	{PKGNAME}Srv.name = "zeus"
	{PKGNAME}Srv.client = gomicro.New{SRVNAME}Service({PKGNAME}Srv.name, cli)
	return {PKGNAME}Srv.client, nil
}

`
	camelSrvName := CamelCase(PD.SvrName)

	context = strings.ReplaceAll(context, "{PKGNAME}", PD.PackageName)
	context = strings.ReplaceAll(context, "{SRVNAME}", camelSrvName)

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
