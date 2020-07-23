package server

import (
	"context"
	"time"

	"github.com/golang/protobuf/proto"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	mesh_core "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
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
		if proto.Equal(currentState, lastStoredState) {
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

func NewDataplaneInsightStore(resManager manager.ResourceManager) DataplaneInsightStore {
	return &dataplaneInsightStore{resManager}
}

var _ DataplaneInsightStore = &dataplaneInsightStore{}

type dataplaneInsightStore struct {
	resManager manager.ResourceManager
}

func (s *dataplaneInsightStore) Upsert(dataplaneId core_model.ResourceKey, subscription *mesh_proto.DiscoverySubscription) error {
	create := false
	dataplaneInsight := &mesh_core.DataplaneInsightResource{}
	err := s.resManager.Get(context.Background(), dataplaneInsight, core_store.GetBy(dataplaneId))
	if err != nil {
		if core_store.IsResourceNotFound(err) {
			create = true
		} else {
			return err
		}
	}
	dataplaneInsight.Spec.UpdateSubscription(subscription)
	if create {
		return s.resManager.Create(context.Background(), dataplaneInsight, core_store.CreateBy(dataplaneId))
	} else {
		return s.resManager.Update(context.Background(), dataplaneInsight)
	}
}
