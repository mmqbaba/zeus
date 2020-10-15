package zcontext

import (
	"context"
	"errors"
	zeusredis "gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/redis/zredis"

	"github.com/go-redis/redis"
)

type ctxRedisMarker struct{}
type ctxRedisWithPromMarker struct{}

type ctxRedis struct {
	cli *redis.Client
}

var (
	ctxRedisKey         = &ctxRedisMarker{}
	ctxRedisWithPromKey = &ctxRedisWithPromMarker{}
)

// ExtractRedis takes the rediscli from ctx.
func ExtractRedis(ctx context.Context) (rdc *redis.Client, err error) {
	r, ok := ctx.Value(ctxRedisKey).(*ctxRedis)
	if !ok || r == nil {
		return nil, errors.New("ctxRedis was not set or nil")
	}
	if r.cli == nil {
		return nil, errors.New("ctxRedis.cli was not set or nil")
	}

	rdc = r.cli
	return
}

// RedisToContext adds the rediscli to the context for extraction later.
// Returning the new context that has been created.
func RedisToContext(ctx context.Context, rdc *redis.Client) context.Context {
	r := &ctxRedis{
		cli: rdc,
	}
	return context.WithValue(ctx, ctxRedisKey, r)
}

// ExtractRedisWithProm takes the zeusrediscli from ctx.
func ExtractRedisWithProm(ctx context.Context) (zrdc zeusredis.Redis, err error) {
	r, ok := ctx.Value(ctxRedisWithPromKey).(zeusredis.Redis)
	if !ok || r == nil {
		return nil, errors.New("ctxRedisProm was not set or nil")
	}
	zrdc = r
	return
}

// RedisWithPromToContext adds the rediscli to the context for extraction later.
// Returning the new context that has been created.
func RedisWithPromToContext(ctx context.Context, zrdc zeusredis.Redis) context.Context {
	return context.WithValue(ctx, ctxRedisWithPromKey, zrdc)
}
