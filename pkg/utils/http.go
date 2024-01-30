package utils

import (
	"context"
	"github.com/go-kratos/kratos/v2/transport"
	trHttp "github.com/go-kratos/kratos/v2/transport/http"
	"io"

	"net/http"
)

func GetHttpRequest(ctx context.Context) *http.Request {
	if t, ok := transport.FromServerContext(ctx); ok {
		if t.Kind() == transport.KindHTTP {
			if info, ok := t.(*trHttp.Transport); ok {
				return info.Request()
			}
		}
	}
	return nil
}

func GetHttpHeader(ctx context.Context) transport.Header {
	if t, ok := transport.FromServerContext(ctx); ok {
		return t.RequestHeader()
	}
	return nil
}

func GetHttpBody(ctx context.Context) []byte {
	httpRequest := GetHttpRequest(ctx)
	if httpRequest != nil {
		body, _ := io.ReadAll(httpRequest.Body)
		return body
	}
	return nil
}
