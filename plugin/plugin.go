package plugin

import (
	"log"
	"net/http"

	"github.com/micro/go-micro"
	"github.com/micro/go-micro/client"
	"github.com/opentracing/opentracing-go"
	"github.com/sirupsen/logrus"

	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/config"
	zeuslog "gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/log"
	zeusmongo "gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/mongo"
	zeusredis "gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/redis"
	tracing "gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/trace"
	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/trace/zipkin"
)

// Container contain comm obj
type Container struct {
	serviceID     string
	appcfg        config.AppConf
	redis         *zeusredis.Client
	mongo         *zeusmongo.Client
	gomicroClient client.Client
	logger        *logrus.Logger
	tracer        *tracing.TracerWrap
	// http
	httpHandler http.Handler
	// gomicro grpc
	gomicroService micro.Service

	// dbPool          *sql.DB
	// transport       *http.Transport
	// svc             XUtil
	// mqProducer      *mq.MqProducer
}

func NewContainer() *Container {
	return &Container{}
}

func (c *Container) Init(appcfg *config.AppConf) {
	log.Println("[Container.Init] start")
	c.initRedis(&appcfg.Redis)
	c.initLogger(&appcfg.LogConf)
	c.initTracer(&appcfg.Trace)
	c.initMongo(&appcfg.MongoDB)
	log.Println("[Container.Init] finish")
	c.appcfg = *appcfg
}

func (c *Container) Reload(appcfg *config.AppConf) {
	log.Println("[Container.Reload] start")
	if c.appcfg.Redis != appcfg.Redis {
		c.reloadRedis(&appcfg.Redis)
	}
	if c.appcfg.LogConf != appcfg.LogConf {
		c.reloadLogger(&appcfg.LogConf)
	}
	if c.appcfg.Trace != appcfg.Trace {
		c.reloadTracer(&appcfg.Trace)
	}
	if c.appcfg.MongoDB != appcfg.MongoDB {
		c.reloadMongo(&appcfg.MongoDB)
	}
	log.Println("[Container.Reload] finish")
	c.appcfg = *appcfg
}

// Redis
func (c *Container) initRedis(cfg *config.Redis) {
	if cfg.Enable {
		c.redis = zeusredis.InitClient(cfg)
	}
}

func (c *Container) reloadRedis(cfg *config.Redis) {
	if cfg.Enable {
		if c.redis != nil {
			c.redis.Reload(cfg)
		} else {
			c.redis = zeusredis.InitClient(cfg)
		}
	} else if c.redis != nil {
		// 释放
		// c.redis.Release()
		c.redis = nil
	}
}

func (c *Container) GetRedisCli() *zeusredis.Client {
	return c.redis
}

// GoMicroClient
func (c *Container) SetGoMicroClient(cli client.Client) {
	c.gomicroClient = cli
}

func (c *Container) GetGoMicroClient() client.Client {
	return c.gomicroClient
}

// Logger
func (c *Container) initLogger(cfg *config.LogConf) {
	l, err := zeuslog.New(cfg)
	if err != nil {
		log.Println("initLogger err:", err)
		return
	}
	c.logger = l.Logger
}

func (c *Container) reloadLogger(cfg *config.LogConf) {
	c.initLogger(cfg)
}

func (c *Container) GetLogger() *logrus.Logger {
	return c.logger
}

// func (r *Container) SetDBPool(p *sql.DB) {
// 	r.dbPool = p
// }

// func (r *Container) GetDBPool() *sql.DB {
// 	return r.dbPool
// }

// func (r *Container) SetTransport(tr *http.Transport) {
// 	r.transport = tr
// }

// func (r *Container) GetTransport() *http.Transport {
// 	return r.transport
// }

// func (r *Container) SetSvcOptions(opt interface{}) {
// 	r.serviceOptions = opt
// }

// func (r *Container) GetSvcOptions() interface{} {
// 	return r.serviceOptions
// }

// func (r *Container) SetSvc(svc XUtil) {
// 	r.svc = svc
// }

// func (r *Container) GetSvc() XUtil {
// 	return r.svc
// }

// func (r *Container) SetMQProducer(p *mq.MqProducer) {
// 	r.mqProducer = p
// }

// func (r *Container) GetMQProducer() *mq.MqProducer {
// 	return r.mqProducer
// }

// func (r *Container) Release() {
// if r.redisPool != nil {
// 	r.redisPool.Close()
// }

// if r.dbPool != nil {
// 	r.dbPool.Close()
// }
// }

// Tracer
func (c *Container) initTracer(cfg *config.Trace) (err error) {
	err = zipkin.InitTracer(cfg)
	if err != nil {
		log.Println("initTracer err:", err)
		return
	}
	c.tracer = tracing.NewTracerWrap(opentracing.GlobalTracer())
	return
}

func (c *Container) reloadTracer(cfg *config.Trace) (err error) {
	return c.initTracer(cfg)
}

func (c *Container) GetTracer() *tracing.TracerWrap {
	return c.tracer
}

func (c *Container) SetServiceID(id string) {
	c.serviceID = id
}

func (c *Container) GetServiceID() string {
	return c.serviceID
}

func (c *Container) SetHTTPHandler(h http.Handler) {
	c.httpHandler = h
}

func (c *Container) GetHTTPHandler() http.Handler {
	return c.httpHandler
}

func (c *Container) SetGoMicroService(s micro.Service) {
	c.gomicroService = s
}

func (c *Container) GetGoMicroService() micro.Service {
	return c.gomicroService
}

// Mongo
func (c *Container) initMongo(cfg *config.MongoDB) {
	var err error
	if cfg.Enable {
		zeusmongo.InitDefalut(cfg)
		c.mongo, err = zeusmongo.DefaultClient()
		if err != nil {
			log.Println("mgoc.DefaultClient err: ", err)
			return
		}
	}
}

func (c *Container) reloadMongo(cfg *config.MongoDB) {
	var err error
	if cfg.Enable {
		if c.mongo != nil {
			zeusmongo.ReloadDefault(cfg)
			c.mongo, err = zeusmongo.DefaultClient()
			if err != nil {
				log.Println("mgoc.DefaultClient err: ", err)
				return
			}
		} else {
			c.initMongo(cfg)
		}
	} else if c.mongo != nil {
		// 释放
		zeusmongo.DefaultClientRelease()
		c.mongo = nil
	}
}

func (c *Container) GetMongo() *zeusmongo.Client {
	return c.mongo
}
