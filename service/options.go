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
	ApiPort      int
	ApiInterface string

	Port int
	// Interface string

	Log       string
	LogFormat string
	LogLevel  string

	ConfEntryPath string

	GoMicroHandlerRegisterFn GoMicroHandlerRegisterFn

	HttpGWHandlerRegisterFn HttpGWHandlerRegisterFn

	HttpHandlerRegisterFn HttpHandlerRegisterFn

	ProcessChangeFn ProcessChangeFn

	LoadEngineFn LoadEngineFn

	GoMicroServerWrapGenerateFn []GoMicroServerWrapGenerateFn
	GoMicroClientWrapGenerateFn []GoMicroClientWrapGenerateFn
}

type Option func(o *Options)

type ProcessChangeFn func(event interface{})

type GoMicroHandlerRegisterFn func(s server.Server, opts ...server.HandlerOption) error

type HttpGWHandlerRegisterFn func(ctx context.Context, endpoint string, opts []grpc.DialOption) (*runtime.ServeMux, error)

type HttpHandlerRegisterFn func(ctx context.Context, prefix string, ng engine.Engine) (http.Handler, error)

type LoadEngineFn func(ng engine.Engine)

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

// ParseCommandLine ...
func ParseCommandLine() (options Options, err error) {
	flag.IntVar(&options.Port, "port", 9090, "Port to listen on")
	flag.IntVar(&options.ApiPort, "apiPort", 8081, "Port to provide api on")

	// flag.StringVar(&options.Interface, "interface", "", "Interface to bind to")
	flag.StringVar(&options.ApiInterface, "apiInterface", "127.0.0.1", "Interface to for API to bind to")

	flag.StringVar(&options.Log, "log", "", "logging to use (console, file, redis, kafka, syslog or logstash)")
	flag.StringVar(&options.LogFormat, "logFormat", "", "log fromat to use (text, json)")
	flag.StringVar(&options.LogLevel, "logLevel", "", "log at or above(debug, info, warn, error, fatal, panic) this level to the logging output(default >=info)")

	flag.StringVar(&options.ConfEntryPath, "confEntryPath", "", "config entry path")

	flag.Parse()

	return options, nil
}
