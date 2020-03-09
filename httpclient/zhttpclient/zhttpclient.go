package zhttpclient

import (
	"context"
	"io"
	"net/http"
)

type HttpClient interface {
	GetHttpClient(instance string) (Client, error)
}

type Client interface {
	Get(ctx context.Context, url string, headers http.Header) ([]byte, error)
	Post(ctx context.Context, url string, body io.Reader, headers http.Header) ([]byte, error)
	Put(ctx context.Context, url string, body io.Reader, headers http.Header) ([]byte, error)
	Patch(ctx context.Context, url string, body io.Reader, headers http.Header) ([]byte, error)
	Delete(ctx context.Context, url string, headers http.Header) ([]byte, error)
}
