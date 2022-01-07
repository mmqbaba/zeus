package zcontext

import (
	"context"
	"errors"

	"github.com/mmqbaba/zeus/plugin/zcontainer"
)

type ctxContainerMarker struct{}

type ctxContainer struct {
	cnt zcontainer.Container
}

var (
	ctxContainerKey = &ctxContainerMarker{}
)

// ExtractContainer takes the container from ctx.
func ExtractContainer(ctx context.Context) (cnt zcontainer.Container, err error) {
	r, ok := ctx.Value(ctxContainerKey).(*ctxContainer)
	if !ok || r == nil {
		return nil, errors.New("ctxContainer was not set or nil")
	}
	if r.cnt == nil {
		return nil, errors.New("ctxContainer.cnt was not set or nil")
	}

	cnt = r.cnt
	return
}

// ContainerToContext adds the container to the context for extraction later.
// Returning the new context that has been created.
func ContainerToContext(ctx context.Context, cnt zcontainer.Container) context.Context {
	r := &ctxContainer{
		cnt: cnt,
	}
	return context.WithValue(ctx, ctxContainerKey, r)
}
