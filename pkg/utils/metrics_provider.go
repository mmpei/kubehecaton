package utils

import (
	"k8s.io/client-go/util/workqueue"
	"github.com/prometheus/client_golang/prometheus"
)

type PrometheusMetricsProvider struct {}

func NewPrometheusMetricsProvider() *PrometheusMetricsProvider {
	return &PrometheusMetricsProvider{}
}

func (pmp PrometheusMetricsProvider) NewDepthMetric(name string) workqueue.GaugeMetric {
	metric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: name + "_depth",
			Help: "current depth of a workqueue",
		},
	)
	prometheus.Register(metric)
	return metric
}

func (pmp PrometheusMetricsProvider) NewAddsMetric(name string) workqueue.CounterMetric {
	metric := prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: name + "_adds",
			Help: "total number of adds handled by a workqueue",
		},
	)
	prometheus.Register(metric)
	return metric
}

func (pmp PrometheusMetricsProvider) NewLatencyMetric(name string) workqueue.HistogramMetric {
	metric := prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name: name + "_latency",
			Help: "how long an item stays in a workqueue",
		},
	)
	prometheus.Register(metric)
	return metric
}

func (pmp PrometheusMetricsProvider) NewWorkDurationMetric(name string) workqueue.HistogramMetric {
	metric := prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name: name + "_work_duration",
			Help: "how long processing an item from a workqueue takes",
		},
	)
	prometheus.Register(metric)
	return metric
}

func (pmp PrometheusMetricsProvider) NewUnfinishedWorkSecondsMetric(name string) workqueue.SettableGaugeMetric {
	metric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: name + "_unfinished_work_seconds",
			Help: "how long have current threads been working",
		},
	)
	prometheus.Register(metric)
	return metric
}

func (pmp PrometheusMetricsProvider) NewLongestRunningProcessorSecondsMetric(name string) workqueue.SettableGaugeMetric {
	metric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: name + "_longest_running_processor_seconds",
			Help: "how long has the longest running processor working",
		},
	)
	prometheus.Register(metric)
	return metric
}

func (pmp PrometheusMetricsProvider) NewRetriesMetric(name string) workqueue.CounterMetric {
	metric := prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: name + "_retries",
			Help: "how many reties have all items done",
		},
	)
	prometheus.Register(metric)
	return metric
}

func (pmp PrometheusMetricsProvider) NewDeprecatedDepthMetric(name string) workqueue.GaugeMetric {
	return &FakeMetric{}
}

func (pmp PrometheusMetricsProvider) NewDeprecatedAddsMetric(name string) workqueue.CounterMetric {
	return &FakeMetric{}
}

func (pmp PrometheusMetricsProvider) NewDeprecatedLatencyMetric(name string) workqueue.SummaryMetric {
	return &FakeMetric{}
}

func (pmp PrometheusMetricsProvider) NewDeprecatedWorkDurationMetric(name string) workqueue.SummaryMetric {
	return &FakeMetric{}
}

func (pmp PrometheusMetricsProvider) NewDeprecatedUnfinishedWorkSecondsMetric(name string) workqueue.SettableGaugeMetric {
	return &FakeMetric{}
}

func (pmp PrometheusMetricsProvider) NewDeprecatedLongestRunningProcessorMicrosecondsMetric(name string) workqueue.SettableGaugeMetric {
	return &FakeMetric{}
}

func (pmp PrometheusMetricsProvider) NewDeprecatedRetriesMetric(name string) workqueue.CounterMetric {
	return &FakeMetric{}
}

// fake Metric for deprecated metrics
type FakeMetric struct {}

func (_ *FakeMetric) Inc() {}

func (_ *FakeMetric) Dec() {}

func (_ *FakeMetric) Set(float64) {}

func (_ *FakeMetric) Observe(float64) {}
