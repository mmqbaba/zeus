package plugin

import (
	"log"

	"github.com/micro/go-micro/client"
	"github.com/sirupsen/logrus"

	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/config"
	zeuslog "gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/log"
	redisclient "gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/redis"
)

// Container contain comm obj
type Container struct {
	appcfg config.AppConf

	redis *redisclient.Client
	// gomicro
	gomicroClient client.Client
	logger        *logrus.Logger
	// http
	// grpc
	// middlewareSpecs map[string]*MiddlewareSpec
	// redisPool       *redis.Pool
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
	log.Println("[Container.Reload] finish")
	c.appcfg = *appcfg
}

// Redis
func (c *Container) initRedis(cfg *config.Redis) {
	if cfg.Enable {
		c.redis = redisclient.InitClient(cfg)
	}
}

func (c *Container) reloadRedis(cfg *config.Redis) {
	if cfg.Enable {
		if c.redis != nil {
			c.redis.Reload(cfg)
		} else {
			c.redis = redisclient.InitClient(cfg)
		}
	} else if c.redis != nil {
		// 释放
		// c.redis.Release()
		c.redis = nil
	}
}

func (c *Container) GetRedisCli() *redisclient.Client {
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

// func (r *Container) SetRedisPool(p *redis.Pool) {
// 	r.redisPool = p
// }

// func (r *Container) GetRedisPool() *redis.Pool {
// 	return r.redisPool
// }

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
