package server

import (
	"context"
	"time"

	mesh_proto "github.com/Kong/konvoy/components/konvoy-control-plane/api/mesh/v1alpha1"
	mesh_core "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/apis/mesh"
	core_model "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/model"
	core_store "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/store"
	"github.com/gogo/protobuf/proto"
)

type DataplaneStatusSink interface {
	Start(accessor SubscriptionStatusAccessor, stop <-chan struct{})
}

type SubscriptionStatusAccessor interface {
	GetStatus() (core_model.ResourceKey, *mesh_proto.DiscoverySubscription)
}

func DefaultDataplaneStatusSink(rs core_store.ResourceStore) DataplaneStatusSink {
	return &dataplaneStatusSink{
		rs: rs,
	}
}

var _ DataplaneStatusSink = &dataplaneStatusSink{}

type dataplaneStatusSink struct {
	rs core_store.ResourceStore
}

func (s *dataplaneStatusSink) Start(accessor SubscriptionStatusAccessor, stop <-chan struct{}) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	var lastStoredState *mesh_proto.DiscoverySubscription

	flush := func() {
		dataplaneId, currentState := accessor.GetStatus()
		if currentState.Equal(lastStoredState) {
			return
		}
		copy := proto.Clone(currentState).(*mesh_proto.DiscoverySubscription)
		if err := s.store(dataplaneId, copy); err != nil {
			xdsServerLog.Error(err, "failed to flush Dataplane status", "dataplaneid", dataplaneId)
		} else {
			lastStoredState = currentState
		}
	}

	for {
		select {
		case <-ticker.C:
			flush()
		case <-stop:
			flush()
			return
		}
	}
}

func (s *dataplaneStatusSink) store(dataplaneId core_model.ResourceKey, subscription *mesh_proto.DiscoverySubscription) error {
	create := false
	dataplaneStatus := &mesh_core.DataplaneStatusResource{}
	err := s.rs.Get(context.Background(), dataplaneStatus, core_store.GetBy(dataplaneId))
	if err != nil {
		if core_store.IsResourceNotFound(err) {
			create = true
		} else {
			return err
		}
	}
	dataplaneStatus.Spec.UpdateSubscription(subscription)
	xdsServerLog.V(1).Info("updating Dataplane status", "dataplaneid", dataplaneId, "subscription", subscription)
	if create {
		return s.rs.Create(context.Background(), dataplaneStatus, core_store.CreateBy(dataplaneId))
	} else {
		return s.rs.Update(context.Background(), dataplaneStatus)
	}
}
