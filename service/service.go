package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path"
	"runtime"
	"strings"
	"sync"
	"time"

	assetfs "github.com/elazarl/go-bindata-assetfs"
	"github.com/gorilla/mux"
	gruntime "github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/micro/go-micro"
	"github.com/micro/go-micro/client"

	// "github.com/urfave/negroni"

	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/config"
	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/engine"
	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/engine/etcd"
	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/microsrv/gomicro"
	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/plugin"
	swagger "gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/swagger/ui"
	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/utils"
)

const (
	retryPeriod       = 5 * time.Second
	changesBufferSize = 10
)

var confEntry *config.Entry
var confEntryPath string
var engineProvidors map[string]engine.NewEngineFn
var swaggerDir string

func init() {
	log.SetPrefix("[zeus] ")
	log.SetFlags(3)

	os := runtime.GOOS
	if os == "linux" || os == "darwin" {
		confEntryPath = "/etc/tif/zeus.json"
	} else {
		confEntryPath = "C:\\tif\\zeus.json"
	}

	engineProvidors = map[string]engine.NewEngineFn{
		"etcd": newEtcdEngine,
		"file": newFileEngine,
	}
}

func newEtcdEngine(cnt *plugin.Container) (engine.Engine, error) {
	return etcd.New(confEntry, cnt)
}

func newFileEngine(cnt *plugin.Container) (engine.Engine, error) {
	return nil, nil
}

func Run(cnt *plugin.Container, opts ...Option) (err error) {
	opt, err := ParseCommandLine()
	if err != nil {
		log.Printf("[zeus] [service.Run] err: %s\n", err)
		return
	}
	s := NewService(opt, cnt, opts...)

	s.initConfEntry()

	if fn, ok := engineProvidors[confEntry.EngineType]; ok && fn != nil {
		if s.ng, err = fn(cnt); err != nil {
			log.Printf("[zeus] [service.Run] err: %s\n", err)
		}
	} else {
		err = fmt.Errorf("no newEnginFn providor for engineType: %s", confEntry.EngineType)
		log.Printf("[zeus] [service.Run] err: %s\n", err)
		return
	}

	// 启动engine，监听状态变化，处理更新，各种组件状态变化
	if err = s.loadNG(); err != nil {
		log.Printf("[zeus] [service.Run] err: %s\n", err)
		return
	}

	// TODO: 启动服务，服务发现注册，http/grpc
	if err = s.startServer(); err != nil {
		log.Printf("[zeus] [service.Run] err: %s\n", err)
		s.watcherCancelC <- struct{}{} // 服务启动失败，通知停止engine的监听
		return
	}
	return
}

type Service struct {
	options        *Options
	container      *plugin.Container
	ng             engine.Engine
	watcherCancelC chan struct{}
	watcherErrorC  chan struct{}
	watcherWg      sync.WaitGroup
}

func NewService(options Options, container *plugin.Container, opts ...Option) *Service {
	o := options
	for _, opt := range opts {
		if opt != nil {
			opt(&o)
		}
	}
	s := &Service{
		options:        &o,
		container:      container,
		watcherErrorC:  make(chan struct{}),
		watcherCancelC: make(chan struct{}),
	}
	if !utils.IsEmptyString(options.ConfEntryPath) {
		confEntryPath = options.ConfEntryPath
	}

	return s
}

func (s *Service) initConfEntry() {
	b, err := ioutil.ReadFile(confEntryPath)
	if err != nil {
		panic(err)
	}
	confEntry = &config.Entry{}
	if err := json.Unmarshal(b, confEntry); err != nil {
		panic(err)
	}
	log.Printf("[zeus] [service.initConfEntry] confEntry: %+v", confEntry)
	return
}

// loadNG 初始化engine，开启监听
func (s *Service) loadNG() (err error) {
	if err = s.ng.Init(); err != nil {
		log.Println("[zeus] [service.loadNG] s.ng.Init err:", err)
		return
	}

	changesC := make(chan interface{}, changesBufferSize)
	// 监听配置变化
	go func() {
		defer close(changesC)
		defer close(s.watcherCancelC)
		if err := s.ng.Subscribe(changesC, s.watcherCancelC); err != nil {
			log.Println("[zeus] [s.n.subscribe] err:", err)
			s.watcherErrorC <- struct{}{}
			return
		}
	}()

	// 处理事件
	go func() {
		defer close(s.watcherErrorC)
		for {
			select {
			case change := <-changesC:
				if err := s.processChange(change); err != nil {
					log.Printf("[zeus] failed to processChange, change=%#v, err=%s\n", change, err)
				}
			case <-s.watcherErrorC:
				log.Println("[zeus] watcher error, change processor shutdown")
			}
		}
	}()

	if s.options.LoadEngineFn != nil {
		s.options.LoadEngineFn(s.ng)
	}
	return
}

func (s *Service) processChange(ev interface{}) (err error) {
	switch ev.(type) {
	case config.Configer:
		log.Printf("[zeus] config change\n")
	default:
		log.Printf("[zeus] unsupported event change\n")
	}
	if s.options.ProcessChangeFn != nil {
		utils.AsyncFuncSafe(context.Background(), func(args ...interface{}) {
			s.options.ProcessChangeFn(ev)
		}, nil)
	}
	return
}

type gwOption struct {
	grpcEndpoint string
}

func (s *Service) startServer() (err error) {
	// gomicro-grpc and gw-http
	log.Println("[zeus] start server ...")
	configer, err := s.ng.GetConfiger()
	if err != nil {
		log.Println("[zeus] [s.startServer] err:", err)
		return
	}
	microConf := configer.Get().GoMicro
	serverPort := microConf.ServerPort
	if s.options.Port > 0 {
		serverPort = uint32(s.options.Port)
	}
	if serverPort == 0 {
		serverPort = 9090
	}
	microConf.ServerPort = serverPort
	gomicroservice, err := s.newGomicroSrv(microConf)
	if err != nil {
		return
	}

	gw, err := s.newHttpGateway(gwOption{grpcEndpoint: fmt.Sprintf("localhost:%d", serverPort)})
	if err != nil {
		return
	}

	go func() {
		addr := fmt.Sprintf("%s:%d", s.options.ApiInterface, s.options.ApiPort)
		log.Printf("http server listen on %s", addr)
		// Start HTTP server (and proxy calls to gRPC server endpoint, serve http, serve swagger)
		if err := http.ListenAndServe(addr, gw); err != nil {
			log.Fatal(err)
		}
	}()

	// Run the service
	if err = gomicroservice.Run(); err != nil {
		log.Println("[zeus] err:", err)
		return
	}
	return
}

func (s *Service) newGomicroSrv(conf config.GoMicro) (gms micro.Service, err error) {
	opts := []micro.Option{}
	s.options.GoMicroServerWrapGenerateFn = append(s.options.GoMicroServerWrapGenerateFn, gomicro.GenerateServerLogWrap)
	if len(s.options.GoMicroServerWrapGenerateFn) != 0 {
		for _, fn := range s.options.GoMicroServerWrapGenerateFn {
			sw := fn(s.ng)
			if sw != nil {
				opts = append(opts, micro.WrapHandler(sw))
			}
		}
	}
	// new micro client
	s.options.GoMicroClientWrapGenerateFn = append(s.options.GoMicroClientWrapGenerateFn, gomicro.GenerateClientLogWrap)
	cliOpts := []client.Option{}
	if len(s.options.GoMicroClientWrapGenerateFn) != 0 {
		for _, fn := range s.options.GoMicroClientWrapGenerateFn {
			cw := fn(s.ng)
			if cw != nil {
				cliOpts = append(cliOpts, client.Wrap(cw))
			}
		}
	}
	cli, err := gomicro.NewClient(context.Background(), conf, cliOpts...)
	if err != nil {
		log.Println("[zeus] [s.newGomicroSrv] gomicro.NewClient err:", err)
		return
	}
	// 把client设置到container
	s.ng.GetContainer().SetGoMicroClient(cli)
	opts = append(opts, micro.Client(cli))
	// new micro service
	gomicroservice := gomicro.NewService(context.Background(), conf, opts...)
	if s.options.GoMicroHandlerRegisterFn != nil {
		if err = s.options.GoMicroHandlerRegisterFn(gomicroservice.Server()); err != nil {
			log.Println("[zeus] [s.newGomicroSrv] GoMicroHandlerRegister err:", err)
			return
		}
		log.Println("[zeus] [s.newGomicroSrv] GoMicroHandlerRegister success.")
	}
	gms = gomicroservice
	return
}

func (s *Service) newHttpGateway(opt gwOption) (h http.Handler, err error) {
	// 必须注意r.PathPrefix的顺序问题
	r := mux.NewRouter()

	// swagger handler
	r.PathPrefix("/swagger/").HandlerFunc(serveSwaggerFile)
	serveSwaggerUI("/swagger-ui/", r)
	log.Println("[zeus] [s.newHttpGateway] swaggerRegister success.")

	// http handler
	if s.options.HttpHandlerRegisterFn != nil {
		var handler http.Handler
		handlerPrefix := ""
		configer, e := s.ng.GetConfiger()
		if e != nil {
			log.Println("[zeus] [s.newHttpGateway] s.ng.GetConfiger err:", e)
			return nil, e
		}
		conf := configer.Get()
		if conf != nil {
			if v, ok := conf.Ext["httphandler_pathprefix"]; ok {
				handlerPrefix = fmt.Sprint(v)
			}
		}
		if utils.IsEmptyString(handlerPrefix) {
			handlerPrefix = "/zeus/"
		}
		if handler, err = s.options.HttpHandlerRegisterFn(context.Background(), handlerPrefix, s.ng); err != nil {
			log.Println("[zeus] [s.newHttpGateway] HttpHandlerRegister err:", err)
			return
		}
		if handler != nil {
			r.PathPrefix(handlerPrefix).Handler(handler)
			log.Println("[zeus] [s.newHttpGateway] HttpHandlerRegister success.")
		}
	}

	// gateway handler
	if s.options.HttpGWHandlerRegisterFn != nil {
		var gwmux *gruntime.ServeMux
		if gwmux, err = s.options.HttpGWHandlerRegisterFn(context.Background(), opt.grpcEndpoint, nil); err != nil {
			log.Println("[zeus] [s.newHttpGateway] HttpGWHandlerRegister err:", err)
			return
		}
		if gwmux != nil {
			gwPrefix := "/"
			r.PathPrefix(gwPrefix).HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
				rr := r.WithContext(r.Context())
				rr.URL.Path = strings.Replace(r.URL.Path, gwPrefix, "/", 1)
				gwmux.ServeHTTP(rw, rr)
			})
			log.Println("[zeus] [s.newHttpGateway] HttpGWHandlerRegister success.")
		}
	}

	// r := http.NewServeMux()
	// r.HandleFunc("/swagger/", serveSwaggerFile)
	// serveGin(r)
	// r.Handle("/", gwmux)

	// mux + negroni 可实现完整的前后置处理和路由
	// r.WithContext + r.Context() 实现上下文传递
	// n := negroni.New()
	// n.UseFunc(func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc){
	// 	c := r.Context()
	// 	rr := r.WithContext(context.WithValue(c, "a", "b")) // 上下文传递
	// 	log.Println("========================1 begin")
	// 	next(rw, rr)
	// 	log.Println("========================1 end")
	// })
	// n.UseFunc(func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc){
	// 	_ := r.Context().Value("a") // 读取
	// 	log.Println("========================2 begin")
	// 	next(rw, r)
	// 	log.Println("========================2 end")
	// })
	// n.UseHandlerFunc(func(rw http.ResponseWriter, r *http.Request){
	// 	log.Println("========================process begin")
	// 	rw.Write([]byte("hello, abc user."))
	// 	log.Println("========================process end")
	// })
	// r.Path("/abc/user").Methods("GET").HandlerFunc(n.ServeHTTP)

	h = r
	return
}

func serveSwaggerFile(w http.ResponseWriter, r *http.Request) {
	if !strings.HasSuffix(r.URL.Path, "swagger.json") {
		log.Printf("Not Found: %s", r.URL.Path)
		http.NotFound(w, r)
		return
	}

	swaggerDir = "proto"
	p := strings.TrimPrefix(r.URL.Path, "/swagger/")
	p = path.Join(swaggerDir, p)

	log.Printf("Serving swagger-file: %s", p)

	http.ServeFile(w, r, p)
}

func serveSwaggerUI(prefix string, mux *mux.Router) {
	fileServer := http.FileServer(&assetfs.AssetFS{
		Asset:    swagger.Asset,
		AssetDir: swagger.AssetDir,
		Prefix:   "third_party/swagger-ui",
	})
	mux.PathPrefix(prefix).Handler(http.StripPrefix(prefix, fileServer))
}
