package cache

import (
	"github.com/prometheus/client_golang/prometheus"
)

const ResultLabel = "result"

func NewMetric(name, help string) *prometheus.CounterVec {
	return prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: name,
		Help: help,
	}, []string{ResultLabel})
}
