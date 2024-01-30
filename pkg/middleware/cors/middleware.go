package cors

import (
	"net/http"
	"strings"
)

type corsMiddleware struct {
	nextHandler http.Handler
}

func newCorsMiddleware(nextHandler http.Handler) *corsMiddleware {
	return &corsMiddleware{
		nextHandler: nextHandler,
	}
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
	header.Set("Access-Control-Allow-Origin", "*")
	header.Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, HEAD")
	header.Set("Access-Control-Allow-Credentials", "true")
	header.Set("Access-Control-Allow-Headers", "Accept, Accept-Language, Content-Language, Accept-Encoding, "+
		"Content-Type, Content-Length,  X-CSRF-Token, Authorization, Content-Disposition, "+
		"X-Forwarded-For, X-Real-IP, X-Requested-With, X-Request-Id, "+
		"Host, Connection, Origin, User-Agent, Referer, Cache-Control, Vary, "+
		"Access-Control-Request-Headers, Access-Control-Request-Method, Access-Control-Allow-Origin, Access-Control-Allow-Credentials")
	header.Set("Access-Control-Expose-Headers", "ETag, Content-Length, Access-Control-Allow-Origin, "+
		"Access-Control-Allow-Credentials")
	header.Set("Access-Control-Max-Age", "86400")
}

// EnableCORS 启用跨域请求，所有http请求会回复Access-Control-Allow-Origin: *等跨域请求头
func EnableCORS() func(http.Handler) http.Handler {
	return func(nextHandler http.Handler) http.Handler {
		return newCorsMiddleware(nextHandler)
	}
}
