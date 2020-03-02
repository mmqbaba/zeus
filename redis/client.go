package redisclient

import (
	"log"
	"strings"
	"sync"
	"time"

	"github.com/go-redis/redis"

	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/config"
)

type Client struct {
	client *redis.Client
	rw     sync.RWMutex
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
