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

type DataplaneInsightSink interface {
	Start(stop <-chan struct{})
}

type DataplaneInsightStore interface {
	Upsert(dataplaneId core_model.ResourceKey, subscription *mesh_proto.DiscoverySubscription) error
}

func NewDataplaneInsightSink(
	accessor SubscriptionStatusAccessor,
	newTicker func() *time.Ticker,
	store DataplaneInsightStore) DataplaneInsightSink {
	return &dataplaneInsightSink{newTicker, accessor, store}
}

var _ DataplaneInsightSink = &dataplaneInsightSink{}

type dataplaneInsightSink struct {
	newTicker func() *time.Ticker
	accessor  SubscriptionStatusAccessor
	store     DataplaneInsightStore
}

func (s *dataplaneInsightSink) Start(stop <-chan struct{}) {
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

func NewDataplaneInsightStore(rs core_store.ResourceStore) DataplaneInsightStore {
	return &dataplaneInsightStore{rs}
}

var _ DataplaneInsightStore = &dataplaneInsightStore{}

type dataplaneInsightStore struct {
	rs core_store.ResourceStore
}

func (s *dataplaneInsightStore) Upsert(dataplaneId core_model.ResourceKey, subscription *mesh_proto.DiscoverySubscription) error {
	create := false
	dataplaneInsight := &mesh_core.DataplaneInsightResource{}
	err := s.rs.Get(context.Background(), dataplaneInsight, core_store.GetBy(dataplaneId))
	if err != nil {
		if core_store.IsResourceNotFound(err) {
			create = true
		} else {
			return err
		}
	}
	dataplaneInsight.Spec.UpdateSubscription(subscription)
	if create {
		return s.rs.Create(context.Background(), dataplaneInsight, core_store.CreateBy(dataplaneId))
	} else {
		return s.rs.Update(context.Background(), dataplaneInsight)
	}
}
