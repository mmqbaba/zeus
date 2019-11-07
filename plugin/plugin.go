package plugin

import (
	"log"

	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/config"
	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/redis"
)

// Container contain comm obj
type Container struct {
	redis *redisclient.Client
	// discovery/registry
	// http
	// grpc
	// middlewareSpecs map[string]*MiddlewareSpec
	// redisPool       *redis.Pool
	// dbPool          *sql.DB
	// transport       *http.Transport
	// serviceOptions  interface{}
	// svc             XUtil
	// mqProducer      *mq.MqProducer
}

func NewContainer() *Container {
	return &Container{}
}

func (c *Container) Init(appcfg *config.AppConf) {
	log.Println("[Container.Init] start")
	c.initRedis(&appcfg.Redis)
	log.Println("[Container.Init] finish")
}

func (c *Container) Reload(appcfg *config.AppConf) {
	log.Println("[Container.Reload] start")
	c.reloadRedis(&appcfg.Redis)
	log.Println("[Container.Reload] finish")
}

func (c *Container) initRedis(cfg *config.Redis) {
	c.redis = redisclient.InitClient(cfg)
}

func (c *Container) reloadRedis(cfg *config.Redis) {
	c.redis.Reload(cfg)
}

func (c *Container) GetRedisCli() *redisclient.Client {
	return c.redis
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
