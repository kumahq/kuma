package server

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	"google.golang.org/protobuf/proto"

	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	config_store "github.com/kumahq/kuma/pkg/config/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/user"
	kuma_log "github.com/kumahq/kuma/pkg/log"
	"github.com/kumahq/kuma/pkg/multitenant"
)

type ZoneInsightSink interface {
	Start(ctx context.Context, stop <-chan struct{})
}

type ZoneInsightStore interface {
	Upsert(ctx context.Context, zone string, subscription *system_proto.KDSSubscription) error
}

func NewZoneInsightSink(
	accessor StatusAccessor,
	flushTicker func() *time.Ticker,
	generationTicker func() *time.Ticker,
	flushBackoff time.Duration,
	store ZoneInsightStore,
	log logr.Logger,
	extensions context.Context,
) ZoneInsightSink {
	return &zoneInsightSink{
		flushTicker:      flushTicker,
		generationTicker: generationTicker,
		flushBackoff:     flushBackoff,
		accessor:         accessor,
		store:            store,
		log:              log,
	}
}

var _ ZoneInsightSink = &zoneInsightSink{}

type zoneInsightSink struct {
	flushTicker      func() *time.Ticker
	generationTicker func() *time.Ticker
	flushBackoff     time.Duration
	accessor         StatusAccessor
	store            ZoneInsightStore
	log              logr.Logger
	extensions       context.Context
}

func (s *zoneInsightSink) Start(ctx context.Context, stop <-chan struct{}) {
	flushTicker := s.flushTicker()
	defer flushTicker.Stop()

	generationTicker := s.generationTicker()
	defer generationTicker.Stop()

	var lastStoredState *system_proto.KDSSubscription
	var generation uint32

	gracefulCtx, cancel := context.WithCancel(multitenant.CopyIntoCtx(ctx, context.Background()))
	defer cancel()

	log := kuma_log.AddFieldsFromCtx(s.log, ctx, s.extensions)

	flush := func() {
		zone, currentState := s.accessor.GetStatus()
		select {
		case <-generationTicker.C:
			generation++
		default:
		}
		currentState.Generation = generation
		if proto.Equal(currentState, lastStoredState) {
			return
		}

		if err := s.store.Upsert(gracefulCtx, zone, currentState); err != nil {
			if store.IsResourceConflict(err) || store.IsResourceAlreadyExists(err) {
				log.V(1).Info("failed to flush ZoneInsight because it was updated in other place. Will retry in the next tick", "zone", zone)
			} else {
				log.Error(err, "failed to flush zone status", "zone", zone)
			}
		} else {
			log.V(1).Info("ZoneInsight saved", "zone", zone, "subscription", currentState)
			lastStoredState = currentState
		}
	}

	for {
		select {
		case <-flushTicker.C:
			flush()
		case <-stop:
			flush()
			return
		}
	}
}

func NewZonesInsightStore(resManager manager.ResourceManager, upsertCfg config_store.UpsertConfig) ZoneInsightStore {
	return &zoneInsightStore{
		resManager: resManager,
		upsertCfg:  upsertCfg,
	}
}

var _ ZoneInsightStore = &zoneInsightStore{}

type zoneInsightStore struct {
	resManager manager.ResourceManager
	upsertCfg  config_store.UpsertConfig
}

func (s *zoneInsightStore) Upsert(ctx context.Context, zone string, subscription *system_proto.KDSSubscription) error {
	ctx = user.Ctx(ctx, user.ControlPlane)

	key := core_model.ResourceKey{
		Name: zone,
	}
	zoneInsight := system.NewZoneInsightResource()
	return manager.Upsert(ctx, s.resManager, key, zoneInsight, func(resource core_model.Resource) error {
		return zoneInsight.Spec.UpdateSubscription(subscription)
	})
}
