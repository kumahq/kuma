package xds

import (
	"context"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

type StatsCallbacks struct {
	NoopCallbacks
	ResponsesSentMetric    *prometheus.CounterVec
	RequestsReceivedMetric *prometheus.CounterVec
	StreamsActive          int
	sync.RWMutex
}

var _ Callbacks = &StatsCallbacks{}

func NewStatsCallbacks(metrics prometheus.Registerer, dsType string) (Callbacks, error) {
	stats := &StatsCallbacks{}

	stats.ResponsesSentMetric = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: dsType + "_responses_sent",
		Help: "Number of responses sent by the server to a client",
	}, []string{"type_url"})
	if err := metrics.Register(stats.ResponsesSentMetric); err != nil {
		return nil, err
	}

	stats.RequestsReceivedMetric = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: dsType + "_requests_received",
		Help: "Number of confirmations requests from a client",
	}, []string{"type_url", "confirmation"})
	if err := metrics.Register(stats.RequestsReceivedMetric); err != nil {
		return nil, err
	}

	streamsActive := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: dsType + "_streams_active",
		Help: "Number of active connections between a server and a client",
	}, func() float64 {
		stats.RLock()
		defer stats.RUnlock()
		return float64(stats.StreamsActive)
	})
	if err := metrics.Register(streamsActive); err != nil {
		return nil, err
	}

	return stats, nil
}

func (s *StatsCallbacks) OnStreamOpen(ctx context.Context, stream int64, typ string) error {
	s.Lock()
	defer s.Unlock()
	s.StreamsActive++
	return nil
}

func (s *StatsCallbacks) OnStreamClosed(stream int64) {
	s.Lock()
	defer s.Unlock()
	s.StreamsActive--
}

func (s *StatsCallbacks) OnStreamRequest(stream int64, request DiscoveryRequest) error {
	if request.GetResponseNonce() != "" {
		if request.HasErrors() {
			s.RequestsReceivedMetric.WithLabelValues(request.GetTypeUrl(), "NACK").Inc()
		} else {
			s.RequestsReceivedMetric.WithLabelValues(request.GetTypeUrl(), "ACK").Inc()
		}
	}
	return nil
}

func (s *StatsCallbacks) OnStreamResponse(_ int64, _ DiscoveryRequest, response DiscoveryResponse) {
	s.ResponsesSentMetric.WithLabelValues(response.GetTypeUrl()).Inc()
}
