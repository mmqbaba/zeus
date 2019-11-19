package generator

import "fmt"

func GenerateRpc(PD *Generator, rootdir string) (err error) {
	err = genRpcInit(PD, rootdir)
	return
}

func genRpcInit(PD *Generator, rootdir string) error {
	header := _defaultHeader
	tmpContext := `package rpc

import (
	"log"

	"github.com/micro/go-micro/server"

	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/service"

	"%s/global"
	"%s/handler"
	gomicro "%s/proto/%spb"
)

func init() {
	// gomicro
	global.ServiceOpts = append(global.ServiceOpts, service.WithGoMicrohandlerRegisterFnOption(gm%sHandlerRegister))
	global.ServiceOpts = append(global.ServiceOpts, service.WithServerNameOption("%s"))
}

func gm%sHandlerRegister(s server.Server, opts ...server.HandlerOption) (err error) {
	if err = gomicro.Register%sHandler(s, handler.New%s(), opts...); err != nil {
		log.Println("gomicro.Register%sHandler err:", err)
		return
	}
	return
}

`
	camelSrvName := CamelCase(PD.SvrName)
	context := fmt.Sprintf(tmpContext, projectBasePrefix+PD.PackageName, projectBasePrefix+PD.PackageName,
		projectBasePrefix+PD.PackageName, PD.PackageName, camelSrvName, PD.PackageName, camelSrvName, camelSrvName, camelSrvName, camelSrvName)
	fn := GetTargetFileName(PD, "rpc.init", rootdir)
	return writeContext(fn, header, context, false)
}
