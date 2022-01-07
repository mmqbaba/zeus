package zcontext

import (
	"context"
	"errors"

	"github.com/mmqbaba/zeus/engine"
)

type ctxEngineMarker struct{}

type ctxEngine struct {
	n engine.Engine
}

var (
	ctxEngineKey = &ctxEngineMarker{}
)

// ExtractEngine takes the engine from ctx.
func ExtractEngine(ctx context.Context) (n engine.Engine, err error) {
	r, ok := ctx.Value(ctxEngineKey).(*ctxEngine)
	if !ok || r == nil {
		return nil, errors.New("ctxEngine was not set or nil")
	}
	if r.n == nil {
		return nil, errors.New("ctxEngine.n was not set or nil")
	}

	n = r.n
	return
}

// EngineToContext adds the engine to the context for extraction later.
// Returning the new context that has been created.
func EngineToContext(ctx context.Context, n engine.Engine) context.Context {
	r := &ctxEngine{
		n: n,
	}
	return context.WithValue(ctx, ctxEngineKey, r)
}
