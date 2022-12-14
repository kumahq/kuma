package catalog

import (
	"context"
	"time"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/api/system/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	"github.com/kumahq/kuma/pkg/core/user"
)

var heartbeatLog = core.Log.WithName("intercp").WithName("catalog").WithName("heartbeat")

type heartbeatComponent struct {
	catalog     Catalog
	newClientFn NewClientFn
	request     *v1alpha1.PingRequest
	interval    time.Duration

	leader *Instance
	client Client
}

var _ component.Component = &heartbeatComponent{}

type NewClientFn = func(url string) (Client, error)

func NewHeartbeatComponent(
	catalog Catalog,
	instance Instance,
	interval time.Duration,
	newClientFn NewClientFn,
) component.Component {
	return &heartbeatComponent{
		catalog: catalog,
		request: &v1alpha1.PingRequest{
			InstanceId:  instance.Id,
			Address:     instance.Address,
			InterCpPort: uint32(instance.InterCpPort),
		},
		newClientFn: newClientFn,
		interval:    interval,
	}
}

func (h *heartbeatComponent) Start(stop <-chan struct{}) error {
	heartbeatLog.Info("starting heartbeats to a leader")
	ctx := user.Ctx(context.Background(), user.ControlPlane)
	ticker := time.NewTicker(h.interval)

	for {
		select {
		case <-ticker.C:
			if err := h.heartbeat(ctx, true); err != nil {
				h.leader = nil
				heartbeatLog.Error(err, "could not heartbeat the leader")
			}
		case <-stop:
			// send final heartbeat to gracefully signal that the instance is going down
			if err := h.heartbeat(ctx, false); err != nil {
				h.leader = nil
				heartbeatLog.Error(err, "could not heartbeat the leader")
			}
			return nil
		}
	}
}

func (h *heartbeatComponent) heartbeat(ctx context.Context, ready bool) error {
	if h.leader == nil {
		if err := h.connectToLeader(ctx); err != nil {
			return err
		}
	}
	if h.leader.Id == h.request.InstanceId {
		heartbeatLog.V(1).Info("this instance is a leader. No need to send a heartbeat.")
		return nil
	}
	heartbeatLog.Info("sending a heartbeat to a leader",
		"instanceId", h.request.InstanceId,
		"leaderAddress", h.leader.Address,
		"ready", ready,
	)
	h.request.Ready = ready
	resp, err := h.client.Ping(ctx, h.request)
	if err != nil {
		return errors.Wrap(err, "could not send a heartbeat to a leader")
	}
	if !resp.Leader {
		heartbeatLog.Info("instance responded that it is no longer a leader")
		h.leader = nil
	}
	return nil
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
	heartbeatLog.Info("leader has changed. Closing connection to the old leader. Creating connection to the new leader.",
		"previousLeaderAddress", h.leader.Address,
		"newLeaderAddress", newLeader.Leader,
	)
	if h.client != nil {
		if err := h.client.Close(); err != nil {
			heartbeatLog.Error(err, "could not close a client to a previous leader")
			// continue anyway. This is a better fallback of this error, because the component would restart and create new client anyways.
		}
	}
	h.client, err = h.newClientFn(h.leader.InterCpURL())
	if err != nil {
		return errors.Wrap(err, "could not create a client to a leader")
	}
	return nil
}

func (h *heartbeatComponent) NeedLeaderElection() bool {
	return false
}
