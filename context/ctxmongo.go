package context

import (
	"context"
	"errors"

	zeusmongo "gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/mongo"
)

type ctxMongoMarker struct{}

type ctxMongo struct {
	cli *zeusmongo.Client
}

var (
	ctxMongoKey = &ctxMongoMarker{}
)

// ExtractMongo takes the mongo from ctx.
func ExtractMongo(ctx context.Context) (c *zeusmongo.Client, err error) {
	r, ok := ctx.Value(ctxMongoKey).(*ctxMongo)
	if !ok || r == nil {
		return nil, errors.New("ctxMongo was not set or nil")
	}

	c = r.cli
	return
}

// MongoToContext adds the mongo to the context for extraction later.
// Returning the new context that has been created.
func MongoToContext(ctx context.Context, c *zeusmongo.Client) context.Context {
	r := &ctxMongo{
		cli: c,
	}
	return context.WithValue(ctx, ctxMongoKey, r)
}
