package catalog

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"

	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	"github.com/kumahq/kuma/pkg/core/user"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
	"github.com/kumahq/kuma/pkg/multitenant"
)

var heartbeatLog = core.Log.WithName("intercp").WithName("catalog").WithName("heartbeat")

type heartbeatComponent struct {
	catalog     Catalog
	getClientFn GetClientFn
	request     *system_proto.PingRequest
	interval    time.Duration

	leader *Instance
	metric prometheus.Summary
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
	metric := prometheus.NewSummary(prometheus.SummaryOpts{
		Name:       "component_heartbeat",
		Help:       "Summary of Inter CP Heartbeat component interval",
		Objectives: core_metrics.DefaultObjectives,
	})
	if err := metrics.Register(metric); err != nil {
		return nil, err
	}

	return &heartbeatComponent{
		catalog: catalog,
		request: &system_proto.PingRequest{
			InstanceId:  instance.Id,
			Address:     instance.Address,
			InterCpPort: uint32(instance.InterCpPort),
		},
		getClientFn: newClientFn,
		interval:    interval,
		metric:      metric,
	}, nil
}

func (h *heartbeatComponent) Start(stop <-chan struct{}) error {
	heartbeatLog.Info("starting heartbeats to a leader")
	ctx := user.Ctx(context.Background(), user.ControlPlane)
	ctx = multitenant.WithTenant(ctx, multitenant.GlobalTenantID)
	ticker := time.NewTicker(h.interval)

	for {
		select {
		case <-ticker.C:
			start := core.Now()
			if !h.heartbeat(ctx, true) {
				continue
			}
			h.metric.Observe(float64(core.Now().Sub(start).Milliseconds()))
		case <-stop:
			// send final heartbeat to gracefully signal that the instance is going down
			_ = h.heartbeat(ctx, false)
			return nil
		}
	}
}

func (h *heartbeatComponent) heartbeat(ctx context.Context, ready bool) bool {
	heartbeatLog := heartbeatLog.WithValues(
		"instanceId", h.request.InstanceId,
		"ready", ready,
	)
	if h.leader == nil {
		if err := h.connectToLeader(ctx); err != nil {
			if err == ErrNoLeader {
				heartbeatLog.Info("leader is not yet present in the cluster. No heartbeat to send")
			} else {
				heartbeatLog.Error(err, "could not connect to leader")
			}
			return false
		}
	}
	if h.leader.Id == h.request.InstanceId {
		heartbeatLog.V(1).Info("this instance is a leader. No need to send a heartbeat.")
		return true
	}
	heartbeatLog = heartbeatLog.WithValues(
		"leaderAddress", h.leader.Address,
	)
	heartbeatLog.V(1).Info("sending a heartbeat to a leader")
	h.request.Ready = ready
	client, err := h.getClientFn(h.leader.InterCpURL())
	if err != nil {
		heartbeatLog.Error(err, "could not get or create a client to a leader")
		h.leader = nil
		return false
	}
	resp, err := client.Ping(ctx, h.request)
	if err != nil {
		heartbeatLog.Info("could not send a heartbeat to a leader. It's likely due to leader change", "cause", err)
		h.leader = nil
		return false
	}
	if !resp.Leader {
		heartbeatLog.V(1).Info("instance responded that it is no longer a leader")
		h.leader = nil
	}
	return true
}

func (h *heartbeatComponent) connectToLeader(ctx context.Context) error {
	newLeader, err := Leader(ctx, h.catalog)
	if err != nil {
		return err
	}
	h.leader = &newLeader
	if h.leader.Id == h.request.InstanceId {
		return nil
	}
	heartbeatLog.Info("leader has changed. Creating connection to the new leader.",
		"previousLeaderAddress", h.leader.Address,
		"newLeaderAddress", newLeader.Address,
	)
	_, err = h.getClientFn(h.leader.InterCpURL())
	if err != nil {
		return errors.Wrap(err, "could not create a client to a leader")
	}
	return nil
}

func (h *heartbeatComponent) NeedLeaderElection() bool {
	return false
}
