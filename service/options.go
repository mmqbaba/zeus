package service

import "flag"

type Options struct {
	ApiPort      int
	ApiInterface string

	Port      int
	Interface string

	Log      string
	LogLevel string

	ConfEntryPath string

	ProcessChangeFn CustomerEventFn
}

type CustomerEventFn func(event interface{})

type Option func(o *Options)

func WithCustomerEventFnOption(fn CustomerEventFn) Option {
	return func(o *Options) {
		o.ProcessChangeFn = fn
	}
}

// ParseCommandLine ...
func ParseCommandLine() (options Options, err error) {
	flag.IntVar(&options.Port, "port", 9799, "Port to listen on")
	flag.IntVar(&options.ApiPort, "apiPort", 9788, "Port to provide api on")

	flag.StringVar(&options.Interface, "interface", "", "Interface to bind to")
	flag.StringVar(&options.ApiInterface, "apiInterface", "127.0.0.1", "Interface to for API to bind to")

	flag.StringVar(&options.Log, "log", "console", "Logging to use (console, json, redis, kafka, syslog or logstash)")
	flag.StringVar(&options.LogLevel, "logLevel", "info", "log at or above(debug,info,warn,error,fatal,panic) this level to the logging output(default >=error)")

	flag.StringVar(&options.ConfEntryPath, "confEntryPath", "", "config entry path")

	flag.Parse()

	return options, nil
}
