package xds

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/kumahq/kuma/v3/pkg/core"
)

var statsLogger = core.Log.WithName("stats-callbacks")

const (
	ConfigInFlightThreshold = 100_000
	failedCallingWebhook    = "failed calling webhook"
	userErrorType           = "user"
	otherErrorType          = "other"
	noErrorType             = "no_error"
	// closureWindow is the maximum delay between sending a large xDS
	// response and observing stream closure that we treat as evidence of
	// gRPC C-Core receive flow-control window depletion (see
	// kumahq/kuma#16355). The reported real-world failure cancels the
	// stream within a few hundred milliseconds; 5s leaves a comfortable
	// margin without inflating false positives.
	closureWindow = 5 * time.Second
	// sizeThreshold is the minimum response size (in bytes) that we
	// consider "large" for the purposes of suspecting receive-window
	// depletion. The default gRPC max_receive_message_length is 4 MiB;
	// 1 MiB triggers the heuristic well below that ceiling so we also
	// catch configurations near the limit.
	sizeThreshold = 1 << 20
)

type VersionExtractor = func(metadata *structpb.Struct) string

var NoopVersionExtractor = func(metadata *structpb.Struct) string {
	return ""
}

type StatsCallbacks interface {
	// ConfigReadyForDelivery marks a configuration as a ready to be delivered.
	// This means that any config (EDS/CDS/KDS policies etc.) with specified version was set to a Snapshot
	// and it's scheduled to be delivered.
	ConfigReadyForDelivery(configVersion string)
	// DiscardConfig removes a configuration from being delivered.
	// This should be called when the client of xDS/KDS server disconnects.
	DiscardConfig(configVersion string)
	Callbacks
	DeltaCallbacks
}

type lastSend struct {
	size    int
	sentAt  time.Time
	typeURL string
}

type statsCallbacks struct {
	NoopCallbacks
	responsesSentMetric         *prometheus.CounterVec
	responseBytesMetric         *prometheus.HistogramVec
	requestsReceivedMetric      *prometheus.CounterVec
	versionsMetric              *prometheus.GaugeVec
	likelyWindowDepletionMetric *prometheus.CounterVec
	deliveryMetric              prometheus.Histogram
	deliveryMetricName          string
	streamsActive               int
	configsQueue                map[string]time.Time
	versionsForStream           map[int64]string
	lastSendByStream            map[int64]lastSend
	versionExtractor            VersionExtractor
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

func NewStatsCallbacks(metrics prometheus.Registerer, dsType string, versionExtractor VersionExtractor) (StatsCallbacks, error) {
	stats := &statsCallbacks{
		configsQueue:      map[string]time.Time{},
		versionsForStream: map[int64]string{},
		lastSendByStream:  map[int64]lastSend{},
		versionExtractor:  versionExtractor,
	}

	stats.responsesSentMetric = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: dsType + "_responses_sent",
		Help: "Number of responses sent by the server to a client",
	}, []string{"type_url"})
	if err := metrics.Register(stats.responsesSentMetric); err != nil {
		return nil, err
	}

	stats.responseBytesMetric = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    dsType + "_responses_sent_bytes",
		Help:    "Size in bytes of DiscoveryResponse messages sent by the server",
		Buckets: prometheus.ExponentialBuckets(1024, 2, 16),
	}, []string{"type_url"})
	if err := metrics.Register(stats.responseBytesMetric); err != nil {
		return nil, err
	}

	stats.requestsReceivedMetric = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: dsType + "_requests_received",
		Help: "Number of confirmations requests from a client",
	}, []string{"type_url", "confirmation", "error_type"})
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
	stats.deliveryMetric = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name: stats.deliveryMetricName,
		Help: "Summary of config delivery including a response (ACK/NACK) from the client",
	})
	if err := metrics.Register(stats.deliveryMetric); err != nil {
		return nil, err
	}

	stats.versionsMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: dsType + "_client_versions",
		Help: "Number of clients for each version. It only counts connections where they sent at least one request",
	}, []string{"client_version"})
	if err := metrics.Register(stats.versionsMetric); err != nil {
		return nil, err
	}

	stats.likelyWindowDepletionMetric = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: dsType + "_stream_likely_window_depletion_total",
		Help: "xDS streams that closed within 5s of sending a >=1 MiB response - likely gRPC C-Core receive flow-control window depletion (kumahq/kuma#16355).",
	}, []string{"type_url"})
	if err := metrics.Register(stats.likelyWindowDepletionMetric); err != nil {
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

func (s *statsCallbacks) OnStreamClosed(streamID int64) {
	s.Lock()
	s.streamsActive--
	if ver, ok := s.versionsForStream[streamID]; ok {
		s.versionsMetric.WithLabelValues(ver).Dec()
		delete(s.versionsForStream, streamID)
	}
	last, hadSend := s.lastSendByStream[streamID]
	delete(s.lastSendByStream, streamID)
	s.Unlock()

	s.maybeReportWindowDepletion(streamID, last, hadSend)
}

func (s *statsCallbacks) OnStreamRequest(_ int64, request DiscoveryRequest) error {
	if request.VersionInfo() == "" {
		return nil // It's initial DiscoveryRequest to ask for resources. It's neither ACK nor NACK.
	}

	if request.HasErrors() {
		s.requestsReceivedMetric.WithLabelValues(request.GetTypeUrl(), "NACK", classifyError(request.ErrorMsg())).Inc()
	} else {
		s.requestsReceivedMetric.WithLabelValues(request.GetTypeUrl(), "ACK", noErrorType).Inc()
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

func (s *statsCallbacks) OnStreamResponse(streamID int64, request DiscoveryRequest, response DiscoveryResponse) {
	if ver := s.versionExtractor(request.Metadata()); ver != "" {
		s.Lock()
		if _, ok := s.versionsForStream[streamID]; !ok {
			s.versionsForStream[streamID] = ver
			s.versionsMetric.WithLabelValues(ver).Inc()
		}
		s.Unlock()
	}

	size := response.ByteSize()
	s.responsesSentMetric.WithLabelValues(response.GetTypeUrl()).Inc()
	s.responseBytesMetric.WithLabelValues(response.GetTypeUrl()).Observe(float64(size))
	s.recordLastSend(streamID, size, response.GetTypeUrl())
}

func (s *statsCallbacks) OnDeltaStreamOpen(context.Context, int64, string) error {
	s.Lock()
	defer s.Unlock()
	s.streamsActive++
	return nil
}

func (s *statsCallbacks) OnDeltaStreamClosed(streamID int64) {
	s.Lock()
	s.streamsActive--
	if ver, ok := s.versionsForStream[streamID]; ok {
		s.versionsMetric.WithLabelValues(ver).Dec()
		delete(s.versionsForStream, streamID)
	}
	last, hadSend := s.lastSendByStream[streamID]
	delete(s.lastSendByStream, streamID)
	s.Unlock()

	s.maybeReportWindowDepletion(streamID, last, hadSend)
}

func (s *statsCallbacks) OnStreamDeltaRequest(_ int64, request DeltaDiscoveryRequest) error {
	if request.GetResponseNonce() == "" {
		return nil // It's initial DiscoveryRequest to ask for resources. It's neither ACK nor NACK.
	}

	if request.HasErrors() {
		s.requestsReceivedMetric.WithLabelValues(request.GetTypeUrl(), "NACK", classifyError(request.ErrorMsg())).Inc()
	} else {
		s.requestsReceivedMetric.WithLabelValues(request.GetTypeUrl(), "ACK", noErrorType).Inc()
	}

	// Delta only has an initial version, therefore we need to change the key to nodeID and typeURL.
	if configTime, exists := s.takeConfigTimeFromQueue(request.NodeId() + request.GetTypeUrl()); exists {
		s.deliveryMetric.Observe(float64(core.Now().Sub(configTime).Milliseconds()))
	}
	return nil
}

func (s *statsCallbacks) OnStreamDeltaResponse(streamID int64, request DeltaDiscoveryRequest, response DeltaDiscoveryResponse) {
	if ver := s.versionExtractor(request.Metadata()); ver != "" {
		s.Lock()
		if _, ok := s.versionsForStream[streamID]; !ok {
			s.versionsForStream[streamID] = ver
			s.versionsMetric.WithLabelValues(ver).Inc()
		}
		s.Unlock()
	}

	size := response.ByteSize()
	s.responsesSentMetric.WithLabelValues(response.GetTypeUrl()).Inc()
	s.responseBytesMetric.WithLabelValues(response.GetTypeUrl()).Observe(float64(size))
	s.recordLastSend(streamID, size, response.GetTypeUrl())
}

func (s *statsCallbacks) recordLastSend(streamID int64, size int, typeURL string) {
	s.Lock()
	s.lastSendByStream[streamID] = lastSend{
		size:    size,
		sentAt:  core.Now(),
		typeURL: typeURL,
	}
	s.Unlock()
}

// maybeReportWindowDepletion emits the warning log and increments the
// counter when the most recent send on a now-closed stream matches the
// receive-window-depletion signature: a large response immediately
// followed by stream closure. See kumahq/kuma#16355.
func (s *statsCallbacks) maybeReportWindowDepletion(streamID int64, last lastSend, hadSend bool) {
	if !hadSend {
		return
	}
	elapsed := core.Now().Sub(last.sentAt)
	if elapsed >= closureWindow || last.size < sizeThreshold {
		return
	}
	s.likelyWindowDepletionMetric.WithLabelValues(last.typeURL).Inc()
	statsLogger.Info(
		"xds stream closed within window of sending a large response; likely caller-side gRPC receive-window depletion (kumahq/kuma#16355). "+
			"If using google_grpc xDS transport, raise KUMA_BOOTSTRAP_SERVER_PARAMS_XDS_GRPC_MAX_RECEIVE_MESSAGE_BYTES.",
		"streamID", streamID,
		"elapsed", elapsed,
		"bytes", last.size,
		"typeURL", last.typeURL,
	)
}

func classifyError(err string) string {
	if strings.Contains(err, failedCallingWebhook) {
		return userErrorType
	} else {
		return otherErrorType
	}
}
