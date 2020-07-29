package server

import (
	"context"
	"time"

	core_model "github.com/kumahq/kuma/pkg/core/resources/model"

	"github.com/go-logr/logr"

	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"

	"github.com/golang/protobuf/proto"

	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
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
	create := false
	zoneInsight := &system.ZoneInsightResource{}
	err := s.resManager.Get(context.Background(), zoneInsight, core_store.GetByKey(zone, core_model.DefaultMesh))
	if err != nil {
		if core_store.IsResourceNotFound(err) {
			create = true
		} else {
			return err
		}
	}
	zoneInsight.Spec.UpdateSubscription(subscription)
	if create {
		return s.resManager.Create(context.Background(), zoneInsight, core_store.CreateByKey(zone, core_model.DefaultMesh))
	} else {
		return s.resManager.Update(context.Background(), zoneInsight)
	}
}
