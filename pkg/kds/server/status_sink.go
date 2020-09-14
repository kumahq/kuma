package server

import (
	"time"

	core_model "github.com/kumahq/kuma/pkg/core/resources/model"

	"github.com/go-logr/logr"

	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"

	"github.com/golang/protobuf/proto"

	"github.com/kumahq/kuma/pkg/core/resources/manager"
)

type ZoneInsightSink interface {
	Start(stop <-chan struct{})
}

type ZoneInsightStore interface {
	Upsert(zone string, subscription *system_proto.KDSSubscription) error
}

func NewZoneInsightSink(
	accessor StatusAccessor,
	newTicker func() *time.Ticker,
	store ZoneInsightStore,
	log logr.Logger) ZoneInsightSink {
	return &zoneInsightSink{newTicker, accessor, store, log}
}

var _ ZoneInsightSink = &zoneInsightSink{}

type zoneInsightSink struct {
	newTicker func() *time.Ticker
	accessor  StatusAccessor
	store     ZoneInsightStore
	log       logr.Logger
}

func (s *zoneInsightSink) Start(stop <-chan struct{}) {
	ticker := s.newTicker()
	defer ticker.Stop()

	var lastStoredState *system_proto.KDSSubscription

	flush := func() {
		zone, currentState := s.accessor.GetStatus()
		if proto.Equal(currentState, lastStoredState) {
			return
		}
		copy := proto.Clone(currentState).(*system_proto.KDSSubscription)
		if err := s.store.Upsert(zone, copy); err != nil {
			s.log.Error(err, "failed to flush zone status", "zone", zone)
		} else {
			s.log.V(1).Info("saved zone status", "zone", zone, "subscription", currentState)
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

func NewDataplaneInsightStore(resManager manager.ResourceManager) ZoneInsightStore {
	return &zoneInsightStore{resManager}
}

var _ ZoneInsightStore = &zoneInsightStore{}

type zoneInsightStore struct {
	resManager manager.ResourceManager
}

func (s *zoneInsightStore) Upsert(zone string, subscription *system_proto.KDSSubscription) error {
	key := core_model.ResourceKey{
		Mesh: core_model.DefaultMesh,
		Name: zone,
	}
	zoneInsight := &system.ZoneInsightResource{}
	return manager.Upsert(s.resManager, key, zoneInsight, func(resource core_model.Resource) {
		zoneInsight.Spec.UpdateSubscription(subscription)
	})
}
