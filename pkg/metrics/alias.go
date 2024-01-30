package metrics

import "github.com/prometheus/client_golang/prometheus"

type (
	Opts      = prometheus.Opts
	Collector = prometheus.Collector

	Counter     = prometheus.Counter
	CounterVec  = prometheus.CounterVec
	CounterOpts = prometheus.CounterOpts

	Gauge     = prometheus.Gauge
	GaugeVec  = prometheus.GaugeVec
	GaugeOpts = prometheus.GaugeOpts

	Histogram     = prometheus.Histogram
	HistogramVec  = prometheus.HistogramVec
	HistogramOpts = prometheus.HistogramOpts

	Observer     = prometheus.Observer
	ObserverFunc = prometheus.ObserverFunc
	ObserverVec  = prometheus.ObserverVec

	Summary     = prometheus.Summary
	SummaryOpts = prometheus.SummaryOpts
	SummaryVec  = prometheus.SummaryVec
)
