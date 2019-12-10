package tokenlimiter

import (
	"context"

	"golang.org/x/time/rate"
)

type Limiter interface {
	Allow() bool
	// AllowN(now time.Time, n int) bool
	Wait(ctx context.Context) (err error)
	// WaitN(ctx context.Context, n int) (err error)
}

// New returns a Limiter
// r 每秒可产生token最大数
// bucketSize token桶容量
func New(r, bucketSize int) Limiter {
	return rate.NewLimiter(rate.Limit(r), bucketSize)
}
