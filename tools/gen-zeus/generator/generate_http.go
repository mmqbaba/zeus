package generator

import (
	"fmt"
	"strings"
)

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
	"net/http"

	"github.com/gin-gonic/gin"
	gruntime "github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"

	zeusmwhttp "github.com/mmqbaba/zeus/middleware/http"
	"github.com/mmqbaba/zeus/service"

	"%s/global"
	"%s/handler"
	gw "%s/proto/{PKG}pb"
)

const (
%s
)

var {PKG}Hdlr = handler.New{SRV}()
var {PKG}HdlrRoutes = map[zeusmwhttp.RouteLink]*zeusmwhttp.Route{
%s
}

func init() {
	// grpc gateway
	global.ServiceOpts = append(global.ServiceOpts, service.WithHttpGWhandlerRegisterFnOption(gwHandlerRegister))
	// http handler
	global.ServiceOpts = append(global.ServiceOpts, service.WithHttpHandlerRegisterFnOption(getHandlerRegisterFn()))
	global.ServiceOpts = append(global.ServiceOpts, service.WithSwaggerJSONFileName("{PKG}"))
}

func gwHandlerRegister(ctx context.Context, endpoint string, opts []grpc.DialOption) (m *gruntime.ServeMux, err error) {
	optsTmp := opts
	mux := gruntime.NewServeMux()
	if len(opts) == 0 {
		optsTmp = []grpc.DialOption{grpc.WithInsecure()}
	}
	if err = gw.Register{SRV}HandlerFromEndpoint(ctx, mux, endpoint, optsTmp); err != nil {
		log.Println("gw.Register{SRV}HandlerFromEndpoint err:", err)
		return
	}
	m = mux
	return
}

func getHandlerRegisterFn() service.HttpHandlerRegisterFn {
	return serveHTTPHandler
}

func registerRoutesFor{SRV}Handler(groups map[string]*gin.RouterGroup, customFn ...zeusmwhttp.CustomRouteFn) {
	for _, f := range customFn {
		f({PKG}HdlrRoutes)
	}
	for _, r := range {PKG}HdlrRoutes {
		zeusmwhttp.Method(groups, r)
	}
}
`
	constVarBlock := ""
	mapValBlock := ""

	camelSrv := CamelCase(PD.SvrName)
	for _, v := range PD.Rpcapi {
		if v.ApiPath == "" {
			continue
		}

		constVarBlock += fmt.Sprintf(
			`	Route_%sHdlr_%s zeusmwhttp.RouteLink = "Route_%sHdlr_%s"
`, camelSrv, v.Name, camelSrv, v.Name)

		mapValBlock += fmt.Sprintf(`	Route_%sHdlr_%s: &zeusmwhttp.Route{
		RLink:  Route_%sHdlr_%s,
		Method: %s,
		Path:   "%s",
		Handle: zeusmwhttp.GenerateGinHandle(%sHdlr.%s),
	},
`, camelSrv, v.Name, camelSrv, v.Name, v.Method, v.ApiPath, PD.PackageName, v.Name)
	}

	imPkg := projectBasePrefix + PD.PackageName
	context := fmt.Sprintf(tmpContext, imPkg, imPkg, imPkg, constVarBlock, mapValBlock)
	context = strings.ReplaceAll(context, "{PKG}", PD.PackageName)
	context = strings.ReplaceAll(context, "{SRV}", camelSrv)

	fn := GetTargetFileName(PD, "http.init", rootdir)
	return writeContext(fn, header, context, true)
}

func genHttp(PD *Generator, rootdir string) error {
	header := ``
	tmpContext := `package http

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/mmqbaba/zeus/engine"
	zeusmwhttp "github.com/mmqbaba/zeus/middleware/http"
)

func init() {
	// zeusmwhttp.SuccessResponse = customSsuccessResponse // ??????????????????????????????
	// zeusmwhttp.ErrorResponse = customErrorResponse
}

func serveHTTPHandler(ctx context.Context, pathPrefix string, ng engine.Engine) (http.Handler, error) {
	log.Println("serveHTTPHandler pathPrefix:", pathPrefix)
	g := gin.New()

	// TODO: ????????????
	// ???????????????????????????????????????handlerfunc
	g.NoRoute(zeusmwhttp.NotFound(ng))
	g.Use(zeusmwhttp.Access(ng))

	prefixGroup := g.Group(pathPrefix)
	prefixGroup.GET("/ping", func(c *gin.Context) {
		zeusmwhttp.ExtractLogger(c).Debug("ping")
		zeusmwhttp.SuccessResponse(c, gin.H{"message": "hello, zeus enginego."})
		// zeusmwhttp.ErrorResponse(c, nil)
	})

	// TODO: ????????????
	// ????????????????????????????????????grouprouter
	////
	{PKG}Group := g.Group(pathPrefix, func(c *gin.Context) {
		zeusmwhttp.ExtractLogger(c).Debug("{PKG} group")
		c.Next()
	})
	groups := map[string]*gin.RouterGroup{
		"default": prefixGroup,
		"{PKG}":   {PKG}Group,
	}
	////

	// TODO: ????????????
	// ???????????????????????????????????????????????????handlerfunc??????????????????
	////
	customRoute{SRV}Hdlr := zeusmwhttp.CustomRouteFn(func(routes map[zeusmwhttp.RouteLink]*zeusmwhttp.Route) {
		//Route_{SRV}Hdlr_Demo.AddMW(routes, func(c *gin.Context) {
		//	zeusmwhttp.ExtractLogger(c).Debug("customRoute{SRV}Hdlr: ", Route_{SRV}Hdlr_PingPong)
		//	c.Next()
		//})
		//Route_{SRV}Hdlr_Demo.SetGroup(routes, "{PKG}")
	})
	////

	// register routes for {SRV}handler
	registerRoutesFor{SRV}Handler(groups, customRoute{SRV}Hdlr)
	return g, nil
}

`
	context := strings.ReplaceAll(tmpContext, "{PKG}", PD.PackageName)
	context = strings.ReplaceAll(context, "{SRV}", CamelCase(PD.SvrName))
	fn := GetTargetFileName(PD, "http", rootdir)
	return writeContext(fn, header, context, false)
}
