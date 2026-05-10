package metrics

import "github.com/prometheus/client_golang/prometheus"

var transportBuckets = []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10}
var storageBuckets = []float64{0.001, 0.0025, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5}
var sagaBuckets = []float64{0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10, 30}
var sizeBuckets = []float64{128, 512, 1024, 4096, 16384, 65536, 262144, 1048576, 4194304, 16777216}

func counter(reg *prometheus.Registry, name, help string, constLabels prometheus.Labels, variableLabels []string) *prometheus.CounterVec {
	collector := prometheus.NewCounterVec(prometheus.CounterOpts{Name: name, Help: help, ConstLabels: constLabels}, variableLabels)
	reg.MustRegister(collector)
	return collector
}

func histogram(reg *prometheus.Registry, name, help string, constLabels prometheus.Labels, variableLabels []string, buckets []float64) *prometheus.HistogramVec {
	collector := prometheus.NewHistogramVec(prometheus.HistogramOpts{Name: name, Help: help, ConstLabels: constLabels, Buckets: buckets}, variableLabels)
	reg.MustRegister(collector)
	return collector
}

func gauge(reg *prometheus.Registry, name, help string, constLabels prometheus.Labels) prometheus.Gauge {
	collector := prometheus.NewGauge(prometheus.GaugeOpts{Name: name, Help: help, ConstLabels: constLabels})
	reg.MustRegister(collector)
	return collector
}

func gaugeVec(reg *prometheus.Registry, name, help string, constLabels prometheus.Labels, variableLabels []string) *prometheus.GaugeVec {
	collector := prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: name, Help: help, ConstLabels: constLabels}, variableLabels)
	reg.MustRegister(collector)
	return collector
}
