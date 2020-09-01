package server

import (
	"time"

	"github.com/golang/protobuf/proto"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	mesh_core "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
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

	flush := func(closing bool) {
		dataplaneId, currentState := s.accessor.GetStatus()
		if proto.Equal(currentState, lastStoredState) {
			return
		}
		copy := proto.Clone(currentState).(*mesh_proto.DiscoverySubscription)
		if err := s.store.Upsert(dataplaneId, copy); err != nil {
			if closing {
				// When XDS stream is closed, Dataplane Status Tracker executes OnStreamClose which closes stop channel
				// The problem is that close() does not wait for this sink to do it's final work
				// In the meantime Dataplane Lifecycle executes OnStreamClose which can remove Dataplane entity (and Insights due to ownership). Therefore both scenarios can happen:
				// 1) upsert fail because it successfully retrieved DataplaneInsight but cannot Update because by this time, Insight is gone (ResourceConflict error)
				// 2) upsert fail because it tries to create a new insight, but there is no Dataplane so ownership returns an error
				// We could build a synchronous mechanism that waits for Sink to be stopped before moving on to next Callbacks, but this is potentially dangerous
				// that we could block waiting for storage instead of executing next callbacks.
				xdsServerLog.V(1).Info("failed to flush Dataplane status on stream close. It can happen when Dataplane is deleted at the same time", "dataplaneid", dataplaneId, "err", err)
			} else {
				xdsServerLog.Error(err, "failed to flush Dataplane status", "dataplaneid", dataplaneId)
			}
		} else {
			xdsServerLog.V(1).Info("saved Dataplane status", "dataplaneid", dataplaneId, "subscription", currentState)
			lastStoredState = currentState
		}
	}

	for {
		select {
		case <-ticker.C:
			flush(false)
		case <-stop:
			flush(true)
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
	insight := &mesh_core.DataplaneInsightResource{}
	return manager.Upsert(s.resManager, dataplaneId, insight, func(resource core_model.Resource) {
		insight.Spec.UpdateSubscription(subscription)
	})
}
