package cache

import (
	"time"

	"github.com/goburrow/cache"
	"github.com/prometheus/client_golang/prometheus"
)

const ResultLabel = "result"

func NewMetric(name, help string) *prometheus.CounterVec {
	return prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: name,
		Help: help,
	}, []string{ResultLabel})
}

type PrometheusStatsCounter struct {
	Metric *prometheus.CounterVec
}

var _ cache.StatsCounter = &PrometheusStatsCounter{}

func (p *PrometheusStatsCounter) RecordHits(count uint64) {
	p.Metric.WithLabelValues("hit").Add(float64(count))
}

func (p *PrometheusStatsCounter) RecordMisses(count uint64) {
	p.Metric.WithLabelValues("miss").Add(float64(count))
}

func (p *PrometheusStatsCounter) RecordLoadSuccess(loadTime time.Duration) {
}

func (p *PrometheusStatsCounter) RecordLoadError(loadTime time.Duration) {
}

func (p *PrometheusStatsCounter) RecordEviction() {
}

func (p *PrometheusStatsCounter) Snapshot(stats *cache.Stats) {
}
