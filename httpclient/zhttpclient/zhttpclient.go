package zhttpclient

import (
	"context"
	"io"
)

type HttpClient interface {
	GetHttpClient(instance string) (Client, error)
}

type Client interface {
	Get(ctx context.Context, url string, headers map[string]string) ([]byte, error)
	Post(ctx context.Context, url string, body io.Reader, headers map[string]string) ([]byte, error)
	Put(ctx context.Context, url string, body io.Reader, headers map[string]string) ([]byte, error)
	Patch(ctx context.Context, url string, body io.Reader, headers map[string]string) ([]byte, error)
	Delete(ctx context.Context, url string, headers map[string]string) ([]byte, error)
}
