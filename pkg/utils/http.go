package utils

import (
	"context"
	"github.com/go-kratos/kratos/v2/transport"
	trHttp "github.com/go-kratos/kratos/v2/transport/http"
	"io"

	"net/http"
)

func GetKratosHttpRequest(ctx context.Context) *http.Request {
	if t, ok := transport.FromServerContext(ctx); ok {
		if t.Kind() == transport.KindHTTP {
			if info, ok := t.(*trHttp.Transport); ok {
				return info.Request()
			}
		}
	}
	return nil
}

func GetKratosHttpHeader(ctx context.Context) transport.Header {
	if t, ok := transport.FromServerContext(ctx); ok {
		return t.RequestHeader()
	}
	return nil
}

func GetKratosHttpBody(ctx context.Context) []byte {
	httpRequest := GetKratosHttpRequest(ctx)
	if httpRequest != nil {
		body, _ := io.ReadAll(httpRequest.Body)
		return body
	}
	return nil
}
