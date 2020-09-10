package zcontext

import (
	"context"
	"errors"
	prom "gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/prometheus"
)

type ctxPromMarker struct{}

type ctxPrometheus struct {
	cli *prom.PubClient
}

var (
	ctxPromKey = &ctxPromMarker{}
)

// ExtractPrometheus takes the prometheus cli from ctx.
func ExtractPrometheus(ctx context.Context) (promc *prom.PubClient, err error) {
	p, ok := ctx.Value(ctxPromKey).(*ctxPrometheus)
	if !ok || p == nil {
		return nil, errors.New("ctxProm was not set or nil")
	}
	if p.cli == nil {
		return nil, errors.New("ctxProm.cli was not set or nil")
	}
	promc = p.cli
	return
}

// PrometheusToContext adds the prometheus cli to the context for extraction later.
// Returning the new context that has been created.
func PrometheusToContext(ctx context.Context, promc *prom.PubClient) context.Context {
	r := &ctxPrometheus{
		cli: promc,
	}
	return context.WithValue(ctx, ctxPromKey, r)
}
