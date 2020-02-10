package zcontext

import (
	"context"
	"errors"

	"github.com/gin-gonic/gin"
)

type ctxGinCtxMarker struct{}

type ctxGinCtx struct {
	ctx *gin.Context
}

var (
	ctxGinCtxKey = &ctxGinCtxMarker{}
)

func ExtractGinCtx(ctx context.Context) (gc *gin.Context, err error) {
	r, ok := ctx.Value(ctxGinCtxKey).(*ctxGinCtx)
	if !ok || r == nil {
		return nil, errors.New("ctxGinCtx was not set or nil")
	}
	if r.ctx == nil {
		return nil, errors.New("ctxGinCtx.ctx was not set or nil")
	}

	gc = r.ctx
	return
}

func GinCtxToContext(ctx context.Context, gc *gin.Context) context.Context {
	r := &ctxGinCtx{
		ctx: gc,
	}
	return context.WithValue(ctx, ctxGinCtxKey, r)
}
