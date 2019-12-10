package zcontext

import (
	"context"
	"errors"

	"github.com/micro/go-micro/client"
)

type ctxGMClientMarker struct{}

// gomicro client
type ctxGMClient struct {
	cli client.Client
}

var (
	ctxGMClientKey = &ctxGMClientMarker{}
)

func ExtractGMClient(ctx context.Context) (cli client.Client, err error) {
	c, ok := ctx.Value(ctxGMClientKey).(*ctxGMClient)
	if !ok || c == nil {
		return nil, errors.New("ctxGMClient was not set or nil")
	}
	if c.cli == nil {
		return nil, errors.New("ctxGMClient.cli was not set or nil")
	}

	cli = c.cli
	return
}

func GMClientToContext(ctx context.Context, cli client.Client) context.Context {
	c := &ctxGMClient{
		cli: cli,
	}
	return context.WithValue(ctx, ctxGMClientKey, c)
}
