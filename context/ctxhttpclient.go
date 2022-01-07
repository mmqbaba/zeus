package zcontext

import (
	"context"
	"errors"

	"github.com/mmqbaba/zeus/httpclient/zhttpclient"
)

type ctxHttpClientMarker struct{}

type ctxHttpClient struct {
	cli zhttpclient.HttpClient
}

var (
	ctxHttpClientKey = &ctxHttpClientMarker{}
)

// ExtractHttpClient takes the httpClient from ctx.
func ExtractHttpClient(ctx context.Context) (c zhttpclient.HttpClient, err error) {
	r, ok := ctx.Value(ctxHttpClientKey).(*ctxHttpClient)
	if !ok || r == nil {
		return nil, errors.New("ctxHttpClient was not set or nil")
	}
	if r.cli == nil {
		return c, errors.New("ctxHttpClient.cli was not set or nil")
	}

	c = r.cli
	return
}

// HttpclientToContext adds the httpclient to the context for extraction later.
// Returning the new context that has been created.
func HttpclientToContext(ctx context.Context, c zhttpclient.HttpClient) context.Context {
	r := &ctxHttpClient{
		cli: c,
	}
	return context.WithValue(ctx, ctxHttpClientKey, r)
}
