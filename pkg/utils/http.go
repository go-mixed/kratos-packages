package utils

import (
	"context"
	"io"
	"net"
	"net/url"
	"strings"

	"github.com/go-kratos/kratos/v2/transport"
	trHttp "github.com/go-kratos/kratos/v2/transport/http"
	"github.com/samber/lo"

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

// GetClientIp 获取客户端IP，如果传入了trustProxy，则会尝试从X-Forwarded-For中获取
func GetClientIp(ctx context.Context, trustProxy []string) string {
	httpRequest := GetKratosHttpRequest(ctx)
	if httpRequest != nil {
		if len(trustProxy) > 0 {
			cidrs := lo.Map(trustProxy, func(s string, _ int) *net.IPNet {
				_, ipNet, _ := net.ParseCIDR(s)
				return ipNet
			})
			if forwardedFor := httpRequest.Header.Get("X-Forwarded-For"); forwardedFor != "" {
				segment := strings.Split(forwardedFor, ",")
				for i := len(segment) - 1; i > 0; i-- {
					ip := strings.TrimSpace(segment[i])
					if CIDRContains(cidrs, ip) {
						return strings.TrimSpace(segment[0])
					}
				}
			}
		}

		ip, _, _ := net.SplitHostPort(strings.TrimSpace(httpRequest.RemoteAddr))
		if ip != "" {
			return ip
		}
	}
	return ""
}

type SimpleRequest struct {
	Method   string      `json:"method"`
	Protocol string      `json:"protocol"`
	Host     string      `json:"host"`
	Headers  http.Header `json:"headers"`
	Url      *url.URL    `json:"url"`
	Body     []byte      `json:"body"`
}

// GetSimpleRequestFromKratosHttpContext 从Kratos HTTP上下文获取简单请求
func GetSimpleRequestFromKratosHttpContext(httpCtx trHttp.Context) (simpleRequest *SimpleRequest, err error) {
	request := httpCtx.Request()
	var bodyBytes []byte
	if request.Body != nil {
		defer request.Body.Close()
		bodyBytes, err = io.ReadAll(request.Body)
		if err != nil {
			return nil, err
		}
	}

	return &SimpleRequest{
		Method:   request.Method,
		Protocol: request.Proto,
		Host:     request.Host,
		Url:      request.URL,
		Headers:  request.Header,
		Body:     bodyBytes,
	}, nil
}

// OutputResponseToKratosContext httpCtx响应根据 http.Response 输出
func OutputResponseToKratosContext(httpCtx trHttp.Context, response *http.Response) error {
	return OutputReaderToKratosContext(httpCtx, response.StatusCode, response.Header, response.Body)
}

// OutputBytesToKratosContext httpCtx响应根据statusCode, header, content输出
func OutputBytesToKratosContext(httpCtx trHttp.Context, statusCode int, header http.Header, content []byte) error {
	for k := range header {
		httpCtx.Header().Set(k, header.Get(k))
	}

	contentType := header.Get("Content-Type")
	return httpCtx.Blob(statusCode, lo.Ternary(contentType != "", contentType, "text/plain"), content)
}

// OutputReaderToKratosContext httpCtx响应根据statusCode, header, reader输出
func OutputReaderToKratosContext(httpCtx trHttp.Context, statusCode int, header http.Header, reader io.Reader) error {
	for k := range header {
		httpCtx.Header().Set(k, header.Get(k))
	}

	contentType := header.Get("Content-Type")
	return httpCtx.Stream(statusCode, lo.Ternary(contentType != "", contentType, "text/plain"), reader)
}
