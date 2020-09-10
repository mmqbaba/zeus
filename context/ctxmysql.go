package zcontext

import (
	"context"
	"errors"
	"github.com/jinzhu/gorm"
)

type ctxMysqlMarker struct{}

type ctxMysql struct {
	cli *gorm.DB
}

var (
	ctxMysqlKey = &ctxMysqlMarker{}
)

// ExtractMysql takes the mysql from ctx.
func ExtractMysql(ctx context.Context) (c *gorm.DB, err error) {
	r, ok := ctx.Value(ctxMysqlKey).(*ctxMysql)
	if !ok || r == nil {
		return nil, errors.New("ctxMysql was not set or nil")
	}
	if r.cli == nil {
		return nil, errors.New("ctxMysql.cli was not set or nil")
	}
	c = r.cli
	return
}

// MysqlToContext adds the mysql to the context for extraction later.
// Returning the new context that has been created.
func MysqlToContext(ctx context.Context, c *gorm.DB) context.Context {
	r := &ctxMysql{
		cli: c,
	}
	return context.WithValue(ctx, ctxMysqlKey, r)
}
