package utils

import (
	"context"
	"github.com/go-kratos/kratos/v2/transport"
	trHttp "github.com/go-kratos/kratos/v2/transport/http"
	"github.com/samber/lo"
	"io"
	"net"
	"strings"

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
