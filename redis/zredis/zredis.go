package zredis

import (
	"github.com/go-redis/redis"
	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/config"
	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/utils"
)

type Redis interface {
	utils.Releaser
	Reload(cfg *config.Redis)
	GetCli() *redis.Client
}
