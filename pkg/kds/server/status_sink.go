package server

import (
	"time"

	"github.com/go-logr/logr"
	"github.com/golang/protobuf/proto"

	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
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
	flushBackoff time.Duration,
	store ZoneInsightStore,
	log logr.Logger) ZoneInsightSink {
	return &zoneInsightSink{
		newTicker:    newTicker,
		flushBackoff: flushBackoff,
		accessor:     accessor,
		store:        store,
		log:          log,
	}
}

var _ ZoneInsightSink = &zoneInsightSink{}

type zoneInsightSink struct {
	newTicker    func() *time.Ticker
	flushBackoff time.Duration
	accessor     StatusAccessor
	store        ZoneInsightStore
	log          logr.Logger
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
			if store.IsResourceConflict(err) {
				s.log.V(1).Info("failed to flush ZoneInsight because it was updated in other place. Will retry in the next tick", "zone", zone)
			} else {
				s.log.Error(err, "failed to flush zone status", "zone", zone)
			}
		} else {
			s.log.V(1).Info("ZoneInsight saved", "zone", zone, "subscription", currentState)
			lastStoredState = currentState
		}
	}

	for {
		select {
		case <-ticker.C:
			flush()
			time.Sleep(s.flushBackoff)
		case <-stop:
			flush()
			return
		}
	}
}

func NewZonesInsightStore(resManager manager.ResourceManager) ZoneInsightStore {
	return &zoneInsightStore{resManager}
}

var _ ZoneInsightStore = &zoneInsightStore{}

type zoneInsightStore struct {
	resManager manager.ResourceManager
}

func (s *zoneInsightStore) Upsert(zone string, subscription *system_proto.KDSSubscription) error {
	key := core_model.ResourceKey{
		Name: zone,
	}
	zoneInsight := system.NewZoneInsightResource()
	return manager.Upsert(s.resManager, key, zoneInsight, func(resource core_model.Resource) {
		zoneInsight.Spec.UpdateSubscription(subscription)
	})
}
