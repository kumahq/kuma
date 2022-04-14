package xds

import (
	"context"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/kumahq/kuma/pkg/core"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
)

var statsLogger = core.Log.WithName("stats-callbacks")

const ConfigInFlightThreshold = 100_000

type StatsCallbacks interface {
	// ConfigReadyForDelivery marks a configuration as a ready to be delivered.
	// This means that any config (EDS/CDS/KDS policies etc.) with specified version was set to a Snapshot
	// and it's scheduled to be delivered.
	ConfigReadyForDelivery(configVersion string)
	// DiscardConfig removes a configuration from being delivered.
	// This should be called when the client of xDS/KDS server disconnects.
	DiscardConfig(configVersion string)
	Callbacks
}

type statsCallbacks struct {
	NoopCallbacks
	responsesSentMetric    *prometheus.CounterVec
	requestsReceivedMetric *prometheus.CounterVec
	deliveryMetric         prometheus.Summary
	deliveryMetricName     string
	streamsActive          int
	configsQueue           map[string]time.Time
	sync.RWMutex
}

func (s *statsCallbacks) ConfigReadyForDelivery(configVersion string) {
	s.Lock()
	if len(s.configsQueue) > ConfigInFlightThreshold {
		// We clean up times of ready for delivery configs when config is delivered or client is disconnected.
		// However, there is always a potential case that may have missed.
		// When we get to the point of ConfigInFlightThreshold elements in the map we want to wipe the map
		// instead of grow it to the point that CP runs out of memory.
		// The statistic is not critical for CP to work, and we will still get data points of configs that are constantly being delivered.
		statsLogger.Info("cleaning up config ready for delivery times to avoid potential memory leak. This operation may cause problems with metric for a short period of time", "metric", s.deliveryMetricName)
		s.configsQueue = map[string]time.Time{}
	}
	s.configsQueue[configVersion] = core.Now()
	s.Unlock()
}

func (s *statsCallbacks) DiscardConfig(configVersion string) {
	s.Lock()
	delete(s.configsQueue, configVersion)
	s.Unlock()
}

var _ StatsCallbacks = &statsCallbacks{}

func NewStatsCallbacks(metrics prometheus.Registerer, dsType string) (StatsCallbacks, error) {
	stats := &statsCallbacks{
		configsQueue: map[string]time.Time{},
	}

	stats.responsesSentMetric = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: dsType + "_responses_sent",
		Help: "Number of responses sent by the server to a client",
	}, []string{"type_url"})
	if err := metrics.Register(stats.responsesSentMetric); err != nil {
		return nil, err
	}

	stats.requestsReceivedMetric = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: dsType + "_requests_received",
		Help: "Number of confirmations requests from a client",
	}, []string{"type_url", "confirmation"})
	if err := metrics.Register(stats.requestsReceivedMetric); err != nil {
		return nil, err
	}

	streamsActive := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: dsType + "_streams_active",
		Help: "Number of active connections between a server and a client",
	}, func() float64 {
		stats.RLock()
		defer stats.RUnlock()
		return float64(stats.streamsActive)
	})
	if err := metrics.Register(streamsActive); err != nil {
		return nil, err
	}

	stats.deliveryMetricName = dsType + "_delivery"
	stats.deliveryMetric = prometheus.NewSummary(prometheus.SummaryOpts{
		Name:       stats.deliveryMetricName,
		Help:       "Summary of config delivery including a response (ACK/NACK) from the client",
		Objectives: core_metrics.DefaultObjectives,
	})
	if err := metrics.Register(stats.deliveryMetric); err != nil {
		return nil, err
	}

	return stats, nil
}

func (s *statsCallbacks) OnStreamOpen(context.Context, int64, string) error {
	s.Lock()
	defer s.Unlock()
	s.streamsActive++
	return nil
}

func (s *statsCallbacks) OnStreamClosed(int64) {
	s.Lock()
	defer s.Unlock()
	s.streamsActive--
}

func (s *statsCallbacks) OnStreamRequest(_ int64, request DiscoveryRequest) error {
	if request.VersionInfo() == "" {
		return nil // It's initial DiscoveryRequest to ask for resources. It's neither ACK nor NACK.
	}

	if request.HasErrors() {
		s.requestsReceivedMetric.WithLabelValues(request.GetTypeUrl(), "NACK").Inc()
	} else {
		s.requestsReceivedMetric.WithLabelValues(request.GetTypeUrl(), "ACK").Inc()
	}

	if configTime, exists := s.takeConfigTimeFromQueue(request.VersionInfo()); exists {
		s.deliveryMetric.Observe(float64(core.Now().Sub(configTime).Milliseconds()))
	}
	return nil
}

func (s *statsCallbacks) takeConfigTimeFromQueue(configVersion string) (time.Time, bool) {
	s.Lock()
	generatedTime, ok := s.configsQueue[configVersion]
	delete(s.configsQueue, configVersion)
	s.Unlock()
	return generatedTime, ok
}

func (s *statsCallbacks) OnStreamResponse(_ int64, _ DiscoveryRequest, response DiscoveryResponse) {
	s.responsesSentMetric.WithLabelValues(response.GetTypeUrl()).Inc()
}
