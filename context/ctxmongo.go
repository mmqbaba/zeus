package zcontext

import (
	"context"
	"errors"

	"gitlab.dg.com/BackEnd/deliver/tif/zeus/mongo/zmongo"
)

type ctxMongoMarker struct{}

type ctxMongo struct {
	cli zmongo.Mongo
}

var (
	ctxMongoKey = &ctxMongoMarker{}
)

// ExtractMongo takes the mongo from ctx.
func ExtractMongo(ctx context.Context) (c zmongo.Mongo, err error) {
	r, ok := ctx.Value(ctxMongoKey).(*ctxMongo)
	if !ok || r == nil {
		return nil, errors.New("ctxMongo was not set or nil")
	}
	if r.cli == nil {
		return nil, errors.New("ctxMongo.cli was not set or nil")
	}

	c = r.cli
	return
}

// MongoToContext adds the mongo to the context for extraction later.
// Returning the new context that has been created.
func MongoToContext(ctx context.Context, c zmongo.Mongo) context.Context {
	r := &ctxMongo{
		cli: c,
	}
	return context.WithValue(ctx, ctxMongoKey, r)
}
