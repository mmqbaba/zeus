package leakylimiter

import (
	"go.uber.org/ratelimit"
)

type Limiter interface {
	ratelimit.Limiter
}

// New returns a Limiter
// rate 每秒请求数
// withoutslack 是否禁用松弛量
func New(rate int, withOutSlack bool) Limiter {
	if withOutSlack {
		return ratelimit.New(rate, ratelimit.WithoutSlack)
	}
	return ratelimit.New(rate)
}
