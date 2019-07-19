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
	Start(stop <-chan struct{})
}

type DataplaneStatusStore interface {
	Upsert(dataplaneId core_model.ResourceKey, subscription *mesh_proto.DiscoverySubscription) error
}

func NewDataplaneStatusSink(
	accessor SubscriptionStatusAccessor,
	newTicker func() *time.Ticker,
	store DataplaneStatusStore) DataplaneStatusSink {
	return &dataplaneStatusSink{newTicker, accessor, store}
}

var _ DataplaneStatusSink = &dataplaneStatusSink{}

type dataplaneStatusSink struct {
	newTicker func() *time.Ticker
	accessor  SubscriptionStatusAccessor
	store     DataplaneStatusStore
}

func (s *dataplaneStatusSink) Start(stop <-chan struct{}) {
	ticker := s.newTicker()
	defer ticker.Stop()

	var lastStoredState *mesh_proto.DiscoverySubscription

	flush := func() {
		dataplaneId, currentState := s.accessor.GetStatus()
		if currentState.Equal(lastStoredState) {
			return
		}
		copy := proto.Clone(currentState).(*mesh_proto.DiscoverySubscription)
		if err := s.store.Upsert(dataplaneId, copy); err != nil {
			xdsServerLog.Error(err, "failed to flush Dataplane status", "dataplaneid", dataplaneId)
		} else {
			xdsServerLog.V(1).Info("saved Dataplane status", "dataplaneid", dataplaneId, "subscription", currentState)
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

func NewDataplaneStatusStore(rs core_store.ResourceStore) DataplaneStatusStore {
	return &dataplaneStatusStore{rs}
}

var _ DataplaneStatusStore = &dataplaneStatusStore{}

type dataplaneStatusStore struct {
	rs core_store.ResourceStore
}

func (s *dataplaneStatusStore) Upsert(dataplaneId core_model.ResourceKey, subscription *mesh_proto.DiscoverySubscription) error {
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
	if create {
		return s.rs.Create(context.Background(), dataplaneStatus, core_store.CreateBy(dataplaneId))
	} else {
		return s.rs.Update(context.Background(), dataplaneStatus)
	}
}
