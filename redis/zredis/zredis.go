package zredis

import (
	"github.com/go-redis/redis"
	"github.com/mmqbaba/zeus/config"
	"github.com/mmqbaba/zeus/utils"
)

type Redis interface {
	utils.Releaser
	Reload(cfg *config.Redis)
	GetCli() *redis.Client
}
