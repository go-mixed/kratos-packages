package metrics

import (
	kratosMiddleware "github.com/go-kratos/kratos/v2/middleware"
	kratosHttp "github.com/go-kratos/kratos/v2/transport/http"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"

	"github.com/go-kratos/kratos/v2/middleware/metrics"
	"go.opentelemetry.io/otel/exporters/prometheus"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
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
	// 将prometheus.exporter注册到registry中
	exporter, err := prometheus.New(prometheus.WithRegisterer(reg.registry))
	if err != nil {
		panic(err)
	}
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(exporter))
	meter := provider.Meter("server")

	_metricSeconds, err := metrics.DefaultSecondsHistogram(meter, metrics.DefaultServerSecondsHistogramName)
	if err != nil {
		panic(err)
	}

	_metricRequests, err := metrics.DefaultRequestsCounter(meter, metrics.DefaultServerRequestsCounterName)
	if err != nil {
		panic(err)
	}

	return metrics.Server(
		metrics.WithSeconds(_metricSeconds),
		metrics.WithRequests(_metricRequests),
	)
}
