package redisclient

import (
	zeusprometheus "gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/prometheus"
	"log"
	"sync"
	"time"

	"github.com/go-redis/redis"
	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/config"
)

type ClusterClientClient struct {
	client *redis.ClusterClient
	rw     sync.RWMutex
}

func InitClusterClientWithProm(cfg *config.Redis, promClient *zeusprometheus.Prom) *ClusterClientClient {
	prom = promClient
	rds := new(ClusterClientClient)
	rds.client = newRedisClusterClient(cfg)
	return rds
}

func InitClusterClient(cfg *config.Redis) *ClusterClientClient {
	rds := new(ClusterClientClient)
	rds.client = newRedisClusterClient(cfg)
	return rds
}

func newRedisClusterClient(cfg *config.Redis) *redis.ClusterClient {
	var clusterClient *redis.ClusterClient
	clusterClient = redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:       cfg.ClusterHost,
		Password:    cfg.Pwd,
		PoolSize:    cfg.PoolSize,
		IdleTimeout: time.Duration(cfg.ConnIdleTimeout) * time.Second,
	})
	if err := clusterClient.Ping().Err(); err != nil {
		log.Fatalf("[redis.newRedisClient] redis ping failed: %s\n", err.Error())
		return nil
	}
	log.Printf("[redis.newRedisClient] success \n")
	return clusterClient
}

func (rds *ClusterClientClient) Reload(cfg *config.Redis) {
	rds.rw.Lock()
	defer rds.rw.Unlock()
	if err := rds.client.Close(); err != nil {
		log.Printf("redis close failed: %s\n", err.Error())
		return
	}
	log.Printf("[redis.Reload] redisclient reload with new conf: %+v\n", cfg)
	rds.client = newRedisClusterClient(cfg)
}

func (rds *ClusterClientClient) GetCli() *redis.ClusterClient {
	rds.rw.RLock()
	defer rds.rw.RUnlock()
	return rds.client
}

func (rds *ClusterClientClient) release() {
	rds.rw.Lock()
	defer rds.rw.Unlock()
	if err := rds.client.Close(); err != nil {
		log.Printf("redis close failed: %s\n", err.Error())
		return
	}
}

func (rds *ClusterClientClient) ZGet(key string) *redis.StringCmd {
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

func (rds *ClusterClientClient) ZSet(key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	println("ClusterClientClient zsetting")
	getStartTime := time.Now()
	result := rds.client.Set(key, value, expiration)
	if result.Err() != nil {
		log.Printf("ClusterClientClient Set Error:", result.Err().Error())
		prom.Incr(redisSet, key, result.Err().Error())
	} else {
		prom.Incr(redisSet, key, OPTION_SUC)
	}
	prom.Timing(redisSet, int64(time.Since(getStartTime)/time.Millisecond), key)
	prom.StateIncr(redisSet, key)
	return result
}

func (rds *ClusterClientClient) ZDel(key string) *redis.IntCmd {
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

func (rds *ClusterClientClient) ZIncr(key string) *redis.IntCmd {
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

func (rds *ClusterClientClient) ZTTL(key string) *redis.DurationCmd {
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

func (rds *ClusterClientClient) ZSetRange(key string, offset int64, value string) *redis.IntCmd {
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

func (rds *ClusterClientClient) ZSetNX(key string, value interface{}, expiration time.Duration) *redis.BoolCmd {
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

func (rds *ClusterClientClient) ZExpire(key string, expiration time.Duration) *redis.BoolCmd {
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

func (rds *ClusterClientClient) ZExists(key string) *redis.IntCmd {
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
