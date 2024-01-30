package metrics

import (
	prom "github.com/go-kratos/kratos/contrib/metrics/prometheus/v2"
	kratosMiddleware "github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/metrics"
	kratosHttp "github.com/go-kratos/kratos/v2/transport/http"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

// Collect 采集全局经过http.Handler的Request、Response的指标，比如请求耗时、请求次数、请求状态码等。建议使用promhttp中的各种方法
// 注意：本方法不会recover panic，请勿在此方法中调用可能会panic的方法
func (reg *Metrics) Collect() kratosHttp.FilterFunc {
	return func(next http.Handler) http.Handler {
		handler := promhttp.InstrumentMetricHandler(reg.registry, next)
		return handler
	}
}

// Middleware 适用于kratos的metrics中间件，用于收集路由层的请求的指标
func (reg *Metrics) Middleware() kratosMiddleware.Middleware {
	_metricSeconds := reg.WithSubsystem("requests").
		WithHelp("server requests duration(sec).").
		RegisterHistogramVec("duration_sec", []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.250, 0.5, 1}, "kind", "operation")

	_metricRequests := reg.WithSubsystem("requests").
		WithHelp("The total number of processed requests").
		RegisterCounterVec("code_total", "kind", "operation", "code", "reason")

	return metrics.Server(
		metrics.WithSeconds(prom.NewHistogram(_metricSeconds)),
		metrics.WithRequests(prom.NewCounter(_metricRequests)),
	)
}
