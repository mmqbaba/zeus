package redisclient

import (
	zeusprometheus "gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/prometheus"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/go-redis/redis"
	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/config"
)

var prom *zeusprometheus.Prom

const (
	redisGet    = "redis:get"
	redisSet    = "redis:set"
	redisDel    = "redis:del"
	redisTtl    = "redis:ttl"
	redisIncr   = "redis:incr"
	redisSetNx  = "redis:setnx"
	redisExpire = "redis:expire"
	redisExist  = "redis:exist"
	OPTION_SUC  = "success"
)

type Client struct {
	client *redis.Client
	rw     sync.RWMutex
}

func InitClientWithProm(cfg *config.Redis, promClient *zeusprometheus.Prom) *Client {
	prom = promClient
	rds := new(Client)
	rds.client = newRedisClient(cfg)
	return rds
}

func InitClient(cfg *config.Redis) *Client {
	rds := new(Client)
	rds.client = newRedisClient(cfg)
	return rds
}

func newRedisClient(cfg *config.Redis) *redis.Client {
	var client *redis.Client
	if cfg.SentinelHost != "" {
		client = redis.NewFailoverClient(&redis.FailoverOptions{
			MasterName:    cfg.SentinelMastername,
			SentinelAddrs: strings.Split(cfg.SentinelHost, ","),
			Password:      cfg.Pwd,
			PoolSize:      cfg.PoolSize,
			IdleTimeout:   time.Duration(cfg.ConnIdleTimeout) * time.Second,
		})
	} else if len(cfg.ClusterHost) != 0 {
		redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:       cfg.ClusterHost,
			Password:    cfg.Pwd,
			PoolSize:    cfg.PoolSize,
			IdleTimeout: time.Duration(cfg.ConnIdleTimeout) * time.Second,
		})
	} else {
		client = redis.NewClient(&redis.Options{
			Addr:        cfg.Host,
			Password:    cfg.Pwd,
			PoolSize:    cfg.PoolSize,
			IdleTimeout: time.Duration(cfg.ConnIdleTimeout) * time.Second,
		})
	}
	if err := client.Ping().Err(); err != nil {
		log.Fatalf("[redis.newRedisClient] redis ping failed: %s\n", err.Error())
		return nil
	}
	log.Printf("[redis.newRedisClient] success \n")
	return client
}

func (rds *Client) Reload(cfg *config.Redis) {
	rds.rw.Lock()
	defer rds.rw.Unlock()
	if err := rds.client.Close(); err != nil {
		log.Printf("redis close failed: %s\n", err.Error())
		return
	}
	log.Printf("[redis.Reload] redisclient reload with new conf: %+v\n", cfg)
	rds.client = newRedisClient(cfg)
}

func (rds *Client) GetCli() *redis.Client {
	rds.rw.RLock()
	defer rds.rw.RUnlock()
	return rds.client
}

func (rds *Client) release() {
	rds.rw.Lock()
	defer rds.rw.Unlock()
	if err := rds.client.Close(); err != nil {
		log.Printf("redis close failed: %s\n", err.Error())
		return
	}
}

func (rds *Client) ZGet(key string) *redis.StringCmd {
	getStartTime := time.Now()
	result := rds.client.Get(key)
	if result.Err() != nil {
		prom.Incr(redisGet, key, result.Err().Error())
	} else {
		prom.Incr(redisGet, key, OPTION_SUC)
	}
	prom.Timing(redisGet, int64(time.Since(getStartTime)/time.Millisecond), key)
	prom.StateIncr(redisGet, key)
	return result
}

func (rds *Client) ZSet(key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	getStartTime := time.Now()

	result := rds.client.Set(key, value, expiration)
	if result.Err() != nil {
		prom.Incr(redisSet, key, result.Err().Error())
	} else {
		prom.Incr(redisSet, key, OPTION_SUC)
	}
	prom.Timing(redisSet, int64(time.Since(getStartTime)/time.Millisecond), key)
	prom.StateIncr(redisSet, key)
	return result
}

func (rds *Client) ZDel(key string) *redis.IntCmd {
	getStartTime := time.Now()
	result := rds.client.Del(key)
	if result.Err() != nil {
		prom.Incr(redisDel, key, result.Err().Error())
	} else {
		prom.Incr(redisDel, key, OPTION_SUC)
	}
	prom.Timing(redisDel, int64(time.Since(getStartTime)/time.Millisecond), key)
	prom.StateIncr(redisDel, key)
	return result
}

func (rds *Client) ZIncr(key string) *redis.IntCmd {
	getStartTime := time.Now()
	result := rds.client.Incr(key)
	if result.Err() != nil {
		prom.Incr(redisIncr, key, result.Err().Error())
	} else {
		prom.Incr(redisIncr, key, OPTION_SUC)
	}
	prom.Timing(redisIncr, int64(time.Since(getStartTime)/time.Millisecond), key)
	prom.StateIncr(redisIncr, key)
	return result
}

func (rds *Client) ZTTL(key string) *redis.DurationCmd {
	getStartTime := time.Now()
	result := rds.client.TTL(key)
	if result.Err() != nil {
		prom.Incr(redisTtl, key, result.Err().Error())
	} else {
		prom.Incr(redisTtl, key, OPTION_SUC)
	}
	prom.Timing(redisTtl, int64(time.Since(getStartTime)/time.Millisecond), key)
	prom.StateIncr(redisTtl, key)
	return result
}

func (rds *Client) ZSetRange(key string, offset int64, value string) *redis.IntCmd {
	getStartTime := time.Now()
	result := rds.client.SetRange(key, offset, value)
	if result.Err() != nil {
		prom.Incr(redisTtl, key, result.Err().Error())
	} else {
		prom.Incr(redisTtl, key, OPTION_SUC)
	}
	prom.Timing(redisTtl, int64(time.Since(getStartTime)/time.Millisecond), key)
	prom.StateIncr(redisTtl, key)
	return result
}

func (rds *Client) ZSetNX(key string, value interface{}, expiration time.Duration) *redis.BoolCmd {
	getStartTime := time.Now()
	result := rds.client.SetNX(key, value, expiration)
	if result.Err() != nil {
		prom.Incr(redisSetNx, key, result.Err().Error())
	} else {
		prom.Incr(redisSetNx, key, OPTION_SUC)
	}
	prom.Timing(redisSetNx, int64(time.Since(getStartTime)/time.Millisecond), key)
	prom.StateIncr(redisSetNx, key)
	return result
}

func (rds *Client) ZExpire(key string, expiration time.Duration) *redis.BoolCmd {
	getStartTime := time.Now()
	result := rds.client.Expire(key, expiration)
	if result.Err() != nil {
		prom.Incr(redisExpire, key, result.Err().Error())
	} else {
		prom.Incr(redisExpire, key, OPTION_SUC)
	}
	prom.Timing(redisExpire, int64(time.Since(getStartTime)/time.Millisecond), key)
	prom.StateIncr(redisExpire, key)
	return result
}

func (rds *Client) ZExists(key string) *redis.IntCmd {
	getStartTime := time.Now()
	result := rds.client.Exists(key)
	if result.Err() != nil {
		prom.Incr(redisExist, key, result.Err().Error())
	} else {
		prom.Incr(redisExist, key, OPTION_SUC)
	}
	prom.Timing(redisExist, int64(time.Since(getStartTime)/time.Millisecond), key)
	prom.StateIncr(redisExist, key)
	return result
}
