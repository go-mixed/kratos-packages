package http

import (
	"net/http"
	"strings"
	"time"

	"github.com/samber/lo"
)

// DefaultHttpClient 返回默认httpClient
//   - timeout: 等待Header响应头，以及总超时时间
func DefaultHttpClient(timeout time.Duration) *http.Client {
	options := DefaultTransportOptions()
	options.ResponseHeaderTimeout = timeout
	return &http.Client{
		Timeout:   timeout,
		Transport: NewHttpTransport(options),
	}
}

// BuildNginxProxyPassUrl 构建nginx的proxy_pass url
// 如果proxy_pass以/结尾，则最终的url为 proxy_pass + URI相对Location的路径；否则，proxy_pass + URI请求路径
// eg: location is '/api/v1/misc/', URI is '/api/v1/misc/image/operate'
// pass_pass is 'https://domain.com/third/' -> 'https://domain.com/third/image/operate'
// pass_pass is 'https://domain.com/third' -> 'https://domain.com/third/api/v1/misc/image/operate'
func BuildNginxProxyPassUrl(proxyPass string, location string, uri string) string {
	// 如果proxy_pass以/结尾，则最终的url为 proxy_pass + 相对路径
	// 否则，proxy_pass + 请求路径
	path := lo.IfF(proxyPass[len(proxyPass)-1] == '/', func() string {
		relativePath := strings.TrimPrefix(uri, location)
		// 为了避免有//，去除掉第一个/
		if relativePath != "" && relativePath[0] == '/' {
			return relativePath[1:]
		}
		return relativePath
	}).Else(uri)

	return proxyPass + path
}
