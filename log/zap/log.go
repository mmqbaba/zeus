package log

import (
	"time"

	"go.uber.org/zap"
)

var logger *zap.Logger
var sugar *zap.SugaredLogger
var url string

func init() {
	logger, _ := zap.NewProduction()
	sugar = logger.Sugar()
	url = "http://gitlab.dg.com"
}

func print() {
	// defer logger.Sync() // flushes buffer, if any
	sugar.Infow("failed to fetch URL",
		// Structured context as loosely typed key-value pairs.
		"url", url,
		"attempt", 3,
		"backoff", time.Second,
	)
}
