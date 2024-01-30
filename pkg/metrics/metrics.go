package metrics

import (
	"errors"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/samber/lo"
	"net/http"
	"regexp"
)

type Metrics struct {
	registry           *prometheus.Registry
	options            *Opts
	metricsRouteServer http.Handler
}

var _ http.Handler = (*Metrics)(nil)

func NewMetrics(
	namespace string,
) *Metrics {
	registry := prometheus.NewRegistry()

	reg := Metrics{
		registry: registry,
		options: &Opts{
			ConstLabels: map[string]string{},
		},
		metricsRouteServer: promhttp.HandlerFor(registry, promhttp.HandlerOpts{}),
	}

	return reg.WithNamespace(namespace)
}

func (reg *Metrics) clone() *Metrics {
	m := &Metrics{
		registry:           reg.registry,
		options:            reg.options,
		metricsRouteServer: reg.metricsRouteServer,
	}

	// 深拷贝map
	m.options.ConstLabels = lo.MapEntries(reg.options.ConstLabels, func(key string, value string) (string, string) {
		return key, value
	})

	return m
}

var prometheusNameReplacer = regexp.MustCompile(`[^a-zA-Z0-9_]`)

// formatPrometheusName 格式化prometheus指标名称，将非法字符替换为下划线
func (reg *Metrics) formatPrometheusName(name string) string {
	return prometheusNameReplacer.ReplaceAllString(name, "_")
}

// WithSubsystem 设置subsystem，并返回新的Metrics（registry仍然是同一个）
func (reg *Metrics) WithSubsystem(subsystem string) *Metrics {
	_reg := reg.clone()
	_reg.options.Subsystem = reg.formatPrometheusName(subsystem)
	return _reg
}

// WithNamespace 设置命名空间，并返回新的Metrics（registry仍然是同一个）
func (reg *Metrics) WithNamespace(namespace string) *Metrics {
	_reg := reg.clone()
	_reg.options.Namespace = reg.formatPrometheusName(namespace)
	return _reg
}

// WithHelp 设置帮助信息，并返回新的Metrics（registry仍然是同一个）
func (reg *Metrics) WithHelp(format string, args ...any) *Metrics {
	_reg := reg.clone()
	_reg.options.Help = fmt.Sprintf(format, args...)
	return _reg
}

// WithConstLabels 设置常量标签，并返回新的Metrics（registry仍然是同一个）
func (reg *Metrics) WithConstLabels(kv ...string) *Metrics {
	if len(kv)%2 != 0 {
		panic("kv must be key-value pairs")
	}
	_reg := reg.clone()
	for i := 0; i < len(kv); i += 2 {
		_reg.options.ConstLabels[kv[i]] = kv[i+1]
	}
	return _reg
}

// ServeHTTP 显示所有metrics指标，可用于prometheus采集
func (reg *Metrics) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	reg.metricsRouteServer.ServeHTTP(w, r)
}

// MustRegister implements Registerer.
func (reg *Metrics) MustRegister(cs ...Collector) {
	for _, c := range cs {
		if err := reg.Register(c); err != nil {
			panic(err)
		}
	}
}

// Register 注册一个指标
func (reg *Metrics) Register(collector Collector) Collector {
	if err := reg.registry.Register(collector); err != nil {
		var are prometheus.AlreadyRegisteredError
		if errors.As(err, &are) {
			collector = are.ExistingCollector
		} else {
			panic(err)
		}
	}

	return collector
}

func (reg *Metrics) parseLabelKV(lablelKVs []string) ([]string, []string) {
	if len(lablelKVs)%2 != 0 {
		panic("labelKV must be key-value pairs")
	}

	chunks := lo.Chunk(lablelKVs, 2)
	labels := make([]string, 0, len(chunks))
	labelValues := make([]string, 0, len(chunks))
	for _, chunk := range chunks {
		labels = append(labels, chunk[0])
		labelValues = append(labelValues, chunk[1])
	}

	return labels, labelValues
}

// RegisterCounterVec 注册一个累加器指标, 可用于请求数/网络流量总数等
func (reg *Metrics) RegisterCounterVec(name string, label ...string) *CounterVec {
	counter := prometheus.NewCounterVec(
		CounterOpts{
			Namespace:   reg.options.Namespace,
			Subsystem:   reg.options.Subsystem,
			Name:        name,
			Help:        reg.options.Help,
			ConstLabels: reg.options.ConstLabels,
		},
		label,
	)
	counter = reg.Register(counter).(*prometheus.CounterVec)
	return counter
}

// Counter 快捷注册一个累加器指标, 返回Counter对象
func (reg *Metrics) Counter(name string, labelKVs ...string) Counter {
	labels, values := reg.parseLabelKV(labelKVs)
	return reg.RegisterCounterVec(name, labels...).WithLabelValues(values...)
}

// RegisterGaugeVec 注册一个仪表盘 瞬时指标, 可增可减。 比如并发数
func (reg *Metrics) RegisterGaugeVec(name string, labels ...string) *GaugeVec {
	gauge := prometheus.NewGaugeVec(
		GaugeOpts{
			Namespace:   reg.options.Namespace,
			Subsystem:   reg.options.Subsystem,
			Name:        name,
			Help:        reg.options.Help,
			ConstLabels: reg.options.ConstLabels,
		},
		labels,
	)
	gauge = reg.Register(gauge).(*prometheus.GaugeVec)
	return gauge
}

// Gauge 快捷注册一个仪表盘 瞬时指标, 返回Gauge对象
func (reg *Metrics) Gauge(name string, labelKVs ...string) Gauge {
	labels, values := reg.parseLabelKV(labelKVs)
	return reg.RegisterGaugeVec(name, labels...).WithLabelValues(values...)
}

// RegisterHistogramVec 注册一个累积直方图
func (reg *Metrics) RegisterHistogramVec(name string, buckets []float64, labels ...string) *HistogramVec {
	histogram := prometheus.NewHistogramVec(
		HistogramOpts{
			Namespace:   reg.options.Namespace,
			Subsystem:   reg.options.Subsystem,
			Name:        name,
			Help:        reg.options.Help,
			Buckets:     buckets,
			ConstLabels: reg.options.ConstLabels,
		},
		labels,
	)
	histogram = reg.Register(histogram).(*prometheus.HistogramVec)
	return histogram
}

// Histogram 快捷注册一个累积直方图, 返回Observer对象
func (reg *Metrics) Histogram(name string, buckets []float64, labelKVs ...string) Observer {
	labels, values := reg.parseLabelKV(labelKVs)
	return reg.RegisterHistogramVec(name, buckets, labels...).WithLabelValues(values...)
}

// RegisterSummaryVec 注册一个摘要
func (reg *Metrics) RegisterSummaryVec(name string, objectives map[float64]float64, labels ...string) *SummaryVec {
	summary := prometheus.NewSummaryVec(
		SummaryOpts{
			Namespace:   reg.options.Namespace,
			Subsystem:   reg.options.Subsystem,
			Name:        name,
			Help:        reg.options.Help,
			Objectives:  objectives,
			ConstLabels: reg.options.ConstLabels,
		},
		labels,
	)
	summary = reg.Register(summary).(*prometheus.SummaryVec)
	return summary
}

// Summary 快捷注册一个摘要, 返回Observer对象
func (reg *Metrics) Summary(name string, objectives map[float64]float64, labelKVs ...string) Observer {
	labels, values := reg.parseLabelKV(labelKVs)
	return reg.RegisterSummaryVec(name, objectives, labels...).WithLabelValues(values...)
}

func (reg *Metrics) GetRegistry() *prometheus.Registry {
	return reg.registry
}
