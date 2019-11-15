package generator

import "fmt"

func GenerateHttp(PD *Generator, rootdir string) (err error) {
	err = genHttpInit(PD, rootdir)
	if err != nil {
		return
	}
	return genHttp(PD, rootdir)
}

func genHttpInit(PD *Generator, rootdir string) error {
	header := _defaultHeader
	tmpContext := `package http

import (
	"context"
	"log"

	gruntime "github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"

	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/service"

	"%s/global"
	gw "%s/proto/gw"
)

func init() {
	// grpc gateway
	global.ServiceOpts = append(global.ServiceOpts, service.WithHttpGWhandlerRegisterFnOption(gwHandlerRegister))
	// http handler
	global.ServiceOpts = append(global.ServiceOpts, service.WithHttpHandlerRegisterFnOption(getHandlerRegisterFn()))
}

func gwHandlerRegister(ctx context.Context, endpoint string, opts []grpc.DialOption) (m *gruntime.ServeMux, err error) {
	optsTmp := opts
	mux := gruntime.NewServeMux()
	if len(opts) == 0 {
		optsTmp = []grpc.DialOption{grpc.WithInsecure()}
	}
	if err = gw.Register%sHandlerFromEndpoint(ctx, mux, endpoint, optsTmp); err != nil {
		log.Println("gw.Register%sHandlerFromEndpoint err:", err)
		return
	}
	m = mux
	return
}

func getHandlerRegisterFn() service.HttpHandlerRegisterFn {
	return serveHttpHandler
}

`
	context := fmt.Sprintf(tmpContext, _defaultPkgPrefix+PD.PackageName, _defaultPkgPrefix+PD.PackageName,
		PD.SvrName, PD.SvrName)
	fn := GetTargetFileName(PD, "http.init", rootdir)
	return writeContext(fn, header, context, false)
}

func genHttp(PD *Generator, rootdir string) error {
	header := ``
	tmpContext := `package http

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/engine"
	zeusmwhttp "gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/middleware/http"
)

func serveHttpHandler(ctx context.Context, pathPrefix string, ng engine.Engine) (http.Handler, error) {
	log.Println("serveHttpHandler pathPrefix:", pathPrefix)
	g := gin.New()
	g.NoRoute(zeusmwhttp.NotFound(ng))
	g.Use(zeusmwhttp.Access(ng))

	prefixGroup := g.Group(pathPrefix)

	apiGroup := prefixGroup.Group("api")
	apiGroup.GET("/echo", getEcho)

	return g, nil
}

func getEcho(c *gin.Context) {
	zeusmwhttp.ExtractLogger(c).Debug("echo")
	zeusmwhttp.SuccessResponse(c, gin.H{"message": "hello, zeus enginego."})
}

`
	context := tmpContext
	fn := GetTargetFileName(PD, "http", rootdir)
	return writeContext(fn, header, context, false)
}
