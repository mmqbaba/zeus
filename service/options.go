package service

import (
	"context"
	"flag"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/micro/go-micro/client"
	"github.com/micro/go-micro/server"
	"google.golang.org/grpc"

	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/engine"
)

type Options struct {
	ServiceName  string
	ApiPort      int
	ApiInterface string

	Port int
	// Interface string

	Log       string
	LogFormat string
	LogLevel  string

	SwaggerJSONFileName string

	ConfEntryPath string

	GoMicroHandlerRegisterFn GoMicroHandlerRegisterFn

	HttpGWHandlerRegisterFn HttpGWHandlerRegisterFn

	HttpHandlerRegisterFn HttpHandlerRegisterFn

	ProcessChangeFn ProcessChangeFn

	LoadEngineFn          LoadEngineFn
	InitServiceCompleteFn InitServiceCompleteFn

	GoMicroServerWrapGenerateFn []GoMicroServerWrapGenerateFn
	GoMicroClientWrapGenerateFn []GoMicroClientWrapGenerateFn

	Version bool
}

type Option func(o *Options)

type ProcessChangeFn func(event interface{})

type GoMicroHandlerRegisterFn func(s server.Server, opts ...server.HandlerOption) error

type HttpGWHandlerRegisterFn func(ctx context.Context, endpoint string, opts []grpc.DialOption) (*runtime.ServeMux, error)

type HttpHandlerRegisterFn func(ctx context.Context, prefix string, ng engine.Engine) (http.Handler, error)

type LoadEngineFn func(ng engine.Engine)

type InitServiceCompleteFn func(ng engine.Engine)

type GoMicroServerWrapGenerateFn func(ng engine.Engine) func(fn server.HandlerFunc) server.HandlerFunc
type GoMicroClientWrapGenerateFn func(ng engine.Engine) func(c client.Client) client.Client

func WithProcessChangeFnOption(fn ProcessChangeFn) Option {
	return func(o *Options) {
		o.ProcessChangeFn = fn
	}
}

func WithGoMicrohandlerRegisterFnOption(fn GoMicroHandlerRegisterFn) Option {
	return func(o *Options) {
		o.GoMicroHandlerRegisterFn = fn
	}
}

func WithHttpGWhandlerRegisterFnOption(fn HttpGWHandlerRegisterFn) Option {
	return func(o *Options) {
		o.HttpGWHandlerRegisterFn = fn
	}
}

func WithHttpHandlerRegisterFnOption(fn HttpHandlerRegisterFn) Option {
	return func(o *Options) {
		o.HttpHandlerRegisterFn = fn
	}
}

func WithLoadEngineFnOption(fn LoadEngineFn) Option {
	return func(o *Options) {
		o.LoadEngineFn = fn
	}
}

func WithInitServiceCompleteFnOption(fn InitServiceCompleteFn) Option {
	return func(o *Options) {
		o.InitServiceCompleteFn = fn
	}
}

func WithGoMicroServerWrapGenerateFnOption(fn ...GoMicroServerWrapGenerateFn) Option {
	return func(o *Options) {
		o.GoMicroServerWrapGenerateFn = append(o.GoMicroServerWrapGenerateFn, fn...)
	}
}

func WithGoMicroClientWrapGenerateFnOption(fn ...GoMicroClientWrapGenerateFn) Option {
	return func(o *Options) {
		o.GoMicroClientWrapGenerateFn = append(o.GoMicroClientWrapGenerateFn, fn...)
	}
}

func WithServiceNameOption(s string) Option {
	return func(o *Options) {
		o.ServiceName = s
	}
}

func WithSwaggerJSONFileName(s string) Option {
	return func(o *Options) {
		o.SwaggerJSONFileName = s
	}
}

// ParseCommandLine ...
func ParseCommandLine() (options Options, err error) {
	flag.IntVar(&options.Port, "port", 0, "Port to listen on")               // 0-使用随机端口
	flag.IntVar(&options.ApiPort, "apiPort", 8081, "Port to provide api on") // 0-使用随机端口

	// flag.StringVar(&options.Interface, "interface", "", "Interface to bind to")
	flag.StringVar(&options.ApiInterface, "apiInterface", "", "Interface to for API to bind to")

	flag.StringVar(&options.Log, "log", "", "logging to use (console, file, redis, kafka, syslog or logstash)")
	flag.StringVar(&options.LogFormat, "logFormat", "", "log fromat to use (text, json)")
	flag.StringVar(&options.LogLevel, "logLevel", "", "log at or above(debug, info, warn, error, fatal, panic) this level to the logging output(default >=info)")

	flag.StringVar(&options.ConfEntryPath, "confEntryPath", "./conf/zeus.json", "config entry path")
	flag.BoolVar(&options.Version, "version", false, "show version")

	flag.Parse()

	return options, nil
}
