package xds

import (
	"context"
	"sync"

	envoy_api "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_xds "github.com/envoyproxy/go-control-plane/pkg/server/v2"
	"github.com/prometheus/client_golang/prometheus"
)

type StatsCallbacks struct {
	ResponsesSentMetric    *prometheus.CounterVec
	RequestsReceivedMetric *prometheus.CounterVec
	StreamsActive          int
	sync.RWMutex
}

func NewStatsCallbacks(metrics prometheus.Registerer, dsType string) (envoy_xds.Callbacks, error) {
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

func (s *StatsCallbacks) OnStreamRequest(stream int64, request *envoy_api.DiscoveryRequest) error {
	if request.ResponseNonce != "" {
		if request.ErrorDetail != nil {
			s.RequestsReceivedMetric.WithLabelValues(request.TypeUrl, "NACK").Inc()
		} else {
			s.RequestsReceivedMetric.WithLabelValues(request.TypeUrl, "ACK").Inc()
		}
	}
	return nil
}

func (s *StatsCallbacks) OnStreamResponse(stream int64, request *envoy_api.DiscoveryRequest, response *envoy_api.DiscoveryResponse) {
	s.ResponsesSentMetric.WithLabelValues(response.TypeUrl).Inc()
}

func (s *StatsCallbacks) OnFetchRequest(ctx context.Context, request *envoy_api.DiscoveryRequest) error {
	return nil
}

func (s *StatsCallbacks) OnFetchResponse(request *envoy_api.DiscoveryRequest, response *envoy_api.DiscoveryResponse) {
}

var _ envoy_xds.Callbacks = &StatsCallbacks{}
