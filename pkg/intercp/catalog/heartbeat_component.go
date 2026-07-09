package catalog

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"

	system_proto "github.com/kumahq/kuma/v2/api/system/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/core"
	"github.com/kumahq/kuma/v2/pkg/core/runtime/component"
	"github.com/kumahq/kuma/v2/pkg/core/user"
	core_metrics "github.com/kumahq/kuma/v2/pkg/metrics"
	"github.com/kumahq/kuma/v2/pkg/multitenant"
)

var heartbeatLog = core.Log.WithName("intercp").WithName("catalog").WithName("heartbeat")

// maxHeartbeatBackoff caps the exponential backoff applied while the leader
// cannot be reached, so a persistent connectivity problem does not produce a
// high-volume stream of failure logs.
const maxHeartbeatBackoff = 1 * time.Minute

type heartbeatMetrics struct {
	interval prometheus.Histogram
	failures prometheus.Counter
}

type heartbeatComponent struct {
	catalog     Catalog
	getClientFn GetClientFn
	request     *system_proto.PingRequest
	interval    time.Duration

	leader   *Instance
	failures int
	metrics  *heartbeatMetrics
}

var _ component.Component = &heartbeatComponent{}

type GetClientFn = func(url string) (system_proto.InterCpPingServiceClient, error)

func NewHeartbeatComponent(
	catalog Catalog,
	instance Instance,
	interval time.Duration,
	newClientFn GetClientFn,
	metrics core_metrics.Metrics,
) (component.Component, error) {
	m := &heartbeatMetrics{
		interval: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name: "component_heartbeat",
			Help: "Summary of Inter CP Heartbeat component interval",
		}),
		failures: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "component_heartbeat_failures",
			Help: "Counter of failed Inter CP heartbeats to the leader",
		}),
	}
	if err := metrics.BulkRegister(m.interval, m.failures); err != nil {
		return nil, err
	}

	return &heartbeatComponent{
		catalog: catalog,
		request: &system_proto.PingRequest{
			InstanceId:  instance.Id,
			Address:     instance.Address,
			InterCpPort: uint32(instance.InterCpPort),
			Version:     instance.Version,
		},
		getClientFn: newClientFn,
		interval:    interval,
		metrics:     m,
	}, nil
}

func (h *heartbeatComponent) Start(stop <-chan struct{}) error {
	heartbeatLog.Info("starting heartbeats to a leader")
	ctx := user.Ctx(context.Background(), user.ControlPlane)
	ctx = multitenant.WithTenant(ctx, multitenant.GlobalTenantID)
	timer := time.NewTimer(h.interval)
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			start := core.Now()
			if h.heartbeat(ctx, true) {
				h.metrics.interval.Observe(float64(core.Now().Sub(start).Milliseconds()))
			}
			timer.Reset(h.backoff())
		case <-stop:
			// send final heartbeat to gracefully signal that the instance is going down
			_ = h.heartbeat(ctx, false)
			return nil
		}
	}
}

// backoff returns the delay until the next heartbeat: the configured interval
// on success, or an exponentially growing delay after consecutive failures.
// The cap never reduces the delay below the configured interval.
func (h *heartbeatComponent) backoff() time.Duration {
	if h.failures == 0 {
		return h.interval
	}
	backoff := h.interval << min(h.failures, 16)
	maxBackoff := max(maxHeartbeatBackoff, h.interval)
	if backoff <= 0 || backoff > maxBackoff {
		return maxBackoff
	}
	return backoff
}

func (h *heartbeatComponent) heartbeat(ctx context.Context, ready bool) bool {
	log := heartbeatLog.WithValues(
		"instanceId", h.request.InstanceId,
		"ready", ready,
	)

	// The catalog is only consulted when the leader is unknown or heartbeats
	// are currently failing - in the calm state we keep talking to the
	// already-known leader instead of reading the catalog on every tick.
	if (h.leader == nil || h.failures > 0) && !h.ensureLeader(ctx, log) {
		return false
	}

	if h.leader.Id == h.request.InstanceId {
		log.V(1).Info("this instance is a leader. No need to send a heartbeat.")
		return true
	}
	log = log.WithValues("leaderAddress", h.leader.Address)
	log.V(1).Info("sending a heartbeat to a leader")

	h.request.Ready = ready
	client, err := h.getClientFn(h.leader.InterCpURL())
	if err != nil {
		log.Error(err, "could not get or create a client to a leader")
		h.recordFailure()
		return false
	}
	resp, err := client.Ping(ctx, h.request)
	if err != nil {
		h.recordFailure()
		log.Info(
			"could not send a heartbeat to a leader, will retry with backoff. This is a connectivity problem, not a leader change",
			"cause", err,
			"consecutiveFailures", h.failures,
		)
		return false
	}
	h.failures = 0
	if !resp.Leader {
		// the instance no longer considers itself the leader; clear it so the
		// leader is re-resolved from the catalog on the next tick.
		log.V(1).Info("instance responded that it is no longer a leader")
		h.leader = nil
	}
	return true
}

func (h *heartbeatComponent) ensureLeader(ctx context.Context, log logr.Logger) bool {
	newLeader, err := Leader(ctx, h.catalog)
	if err != nil {
		if errors.Is(err, ErrNoLeader) {
			log.Info("leader is not yet present in the cluster. No heartbeat to send")
		} else {
			log.Error(err, "could not resolve the leader from the catalog")
			h.recordFailure()
		}
		return false
	}

	if h.leader != nil && h.leader.Id == newLeader.Id {
		// same leader as before, nothing to update
		return true
	}

	previousLeaderAddress := ""
	if h.leader != nil {
		previousLeaderAddress = h.leader.Address
	}
	h.failures = 0
	h.leader = &newLeader

	if newLeader.Id != h.request.InstanceId {
		log.Info(
			"leader has changed. Creating connection to the new leader.",
			"previousLeaderAddress", previousLeaderAddress,
			"newLeaderAddress", newLeader.Address,
		)
	}
	return true
}

func (h *heartbeatComponent) recordFailure() {
	h.failures++
	h.metrics.failures.Inc()
}

func (h *heartbeatComponent) NeedLeaderElection() bool {
	return false
}
