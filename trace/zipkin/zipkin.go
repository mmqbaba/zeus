package zipkin

import (
	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/config"
	"log"
	"os"
	//"github.com/micro/go-micro/metadata"
	"github.com/opentracing/opentracing-go"
	zipkin "github.com/openzipkin/zipkin-go-opentracing"
	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/utils"
)

func InitTracer(cfg *config.Trace) error {
	zipkinURL := cfg.TraceUrl
	hostPort, _ := os.Hostname()
	serviceName := cfg.ServerName
	rate := cfg.Rate
	sampler := cfg.Sampler
	mod := cfg.Mod
	collector, err := zipkin.NewHTTPCollector(zipkinURL)
	if err != nil {
		log.Printf("unable to create Zipkin HTTP collector: %v", err)
		return err
	}
	// 为0默认完全开启采样
	// 为负值则关闭采样
	// 大于1则完全开启采样
	if rate == 0 {
		rate = 1.0
	}
	if utils.IsEmptyString(sampler) {
		sampler = "boundary"
	}
	log.Printf("final sampler: %s, rate: %f, mod: %d", sampler, rate, mod)
	samplerOpt := zipkin.WithSampler(zipkin.NewBoundarySampler(rate, 0))
	switch sampler {
	case "boundary":
		samplerOpt = zipkin.WithSampler(zipkin.NewBoundarySampler(rate, 0))
	case "counting":
		samplerOpt = zipkin.WithSampler(zipkin.NewCountingSampler(rate))
	case "mod":
		samplerOpt = zipkin.WithSampler(zipkin.ModuloSampler(mod))
	}
	tracer, err := zipkin.NewTracer(
		zipkin.NewRecorder(collector, false, hostPort, serviceName),
		//zipkin.ClientServerSameSpan(false),
		zipkin.TraceID128Bit(true),
		samplerOpt,
	)
	if err != nil {
		log.Printf("unable to create Zipkin tracer: %v", err)
		return err
	}
	log.Printf("reload tracer")
	opentracing.SetGlobalTracer(tracer)
	return nil
}
