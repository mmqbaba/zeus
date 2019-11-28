package zredis

import (
	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/config"
	"github.com/go-redis/redis"
)

type Redis interface {
	Reload(cfg *config.Redis)
	GetCli() *redis.Client
}
