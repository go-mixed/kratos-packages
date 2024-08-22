package http

import (
	"net/http"
	"time"
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
