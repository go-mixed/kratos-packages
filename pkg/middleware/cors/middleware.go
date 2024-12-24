package cors

import (
	"fmt"
	"github.com/samber/lo"
	"net/http"
	"strings"
	"time"
)

type corsMiddleware struct {
	nextHandler http.Handler

	origin        string
	methods       []string
	credentials   bool
	headers       []string
	exposeHeaders []string
	maxAge        time.Duration
}

func newCorsMiddleware(nextHandler http.Handler, opts ...corsOption) *corsMiddleware {
	m := &corsMiddleware{
		nextHandler: nextHandler,

		origin:      "*",
		methods:     []string{"POST", "GET", "OPTIONS", "PUT", "DELETE", "HEAD"},
		credentials: true,
		headers: []string{"Accept", "Accept-Language", "Content-Language", "Accept-Encoding",
			"Content-Type", "Content-Length", "X-CSRF-Token", "Authorization", "Content-Disposition",
			"X-Forwarded-For", "X-Real-IP", "X-Requested-With", "X-Request-Id",
			"Host", "Connection", "Origin", "User-Agent", "Referer", "Cache-Control", "Vary",
			"Access-Control-Request-Headers", "Access-Control-Request-Method", "Access-Control-Allow-Origin", "Access-Control-Allow-Credentials"},
		exposeHeaders: []string{
			"ETag", "Content-Length", "Access-Control-Allow-Origin", "Access-Control-Allow-Credentials",
		},
		maxAge: 86400 * time.Second,
	}

	for _, opt := range opts {
		opt(m)
	}

	return m
}

func (m *corsMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.EqualFold(r.Method, "OPTIONS") {
		m.writeCorsHeaders(w.Header())
		w.WriteHeader(http.StatusNoContent)
		return
	} else {
		m.writeCorsHeaders(w.Header())
		m.nextHandler.ServeHTTP(w, r)
	}
}

func (m *corsMiddleware) writeCorsHeaders(header http.Header) {
	if m.origin != "" {
		header.Set("Access-Control-Allow-Origin", m.origin)
	}
	if len(m.methods) > 0 {
		header.Set("Access-Control-Allow-Methods", strings.Join(m.methods, ", "))
	}
	if m.credentials {
		header.Set("Access-Control-Allow-Credentials", "true")
	}
	if len(m.headers) > 0 {
		header.Set("Access-Control-Allow-Headers", strings.Join(m.headers, ", "))
	}

	if len(m.exposeHeaders) > 0 {
		header.Set("Access-Control-Expose-Headers", strings.Join(m.exposeHeaders, ", "))
	}
	if m.maxAge > 0 {
		header.Set("Access-Control-Max-Age", fmt.Sprintf("%d", int(m.maxAge.Seconds())))
	}
}

// EnableCORS 启用跨域请求，所有http请求会回复Access-Control-Allow-Origin: *等跨域请求头
func EnableCORS(opts ...corsOption) func(http.Handler) http.Handler {
	return func(nextHandler http.Handler) http.Handler {
		return newCorsMiddleware(nextHandler, opts...)
	}
}

func trimHeaderStrings(ls ...string) []string {
	var _ls []string
	for _, h := range ls {
		h = strings.TrimSpace(h)
		if h == "" {
			continue
		}
		_ls = append(_ls, lo.Map(strings.Split(h, ","), func(item string, index int) string {
			return strings.TrimSpace(item)
		})...)
	}
	return _ls
}

type corsOption func(m *corsMiddleware)

func WithOrigin(origin string) corsOption {
	return func(m *corsMiddleware) {
		m.origin = origin
	}
}

func WithMethods(methods ...string) corsOption {
	return func(m *corsMiddleware) {
		m.methods = trimHeaderStrings(methods...)
	}
}

func WithCredentials(val bool) corsOption {
	return func(m *corsMiddleware) {
		m.credentials = val
	}
}

// WithHeaders 如果，则按逗号分割，忽略空格
func WithHeaders(headers ...string) corsOption {
	return func(m *corsMiddleware) {

		m.headers = trimHeaderStrings(headers...)
	}
}

func WithExposeHeaders(headers ...string) corsOption {
	return func(m *corsMiddleware) {
		m.exposeHeaders = trimHeaderStrings(headers...)
	}
}

func WithMaxAge(maxAge time.Duration) corsOption {
	return func(m *corsMiddleware) {
		m.maxAge = maxAge
	}
}
