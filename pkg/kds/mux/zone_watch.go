package mux

import (
	"context"
	"time"

	"github.com/go-logr/logr"

	"github.com/kumahq/kuma/pkg/config/multizone"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/events"
	"github.com/kumahq/kuma/pkg/kds/service"
	"github.com/kumahq/kuma/pkg/multitenant"
)

type zoneTenant struct {
	zone     string
	tenantID string
}

type ZoneWatch struct {
	log     logr.Logger
	poll    time.Duration
	timeout time.Duration
	bus     events.EventBus
	rm      manager.ReadOnlyResourceManager
	zones   map[zoneTenant]time.Time
}

func NewZoneWatch(
	log logr.Logger,
	cfg multizone.KdsServerConfig,
	bus events.EventBus,
	rm manager.ReadOnlyResourceManager,
) *ZoneWatch {
	return &ZoneWatch{
		log:     log,
		poll:    cfg.ZoneHealthCheck.PollInterval.Duration,
		timeout: cfg.ZoneHealthCheck.Timeout.Duration,
		bus:     bus,
		rm:      rm,
		zones:   map[zoneTenant]time.Time{},
	}
}

func (zw *ZoneWatch) Start(stop <-chan struct{}) error {
	timer := time.NewTicker(zw.poll)
	defer timer.Stop()

	connectionWatch := zw.bus.Subscribe(func(e events.Event) bool {
		_, ok := e.(service.ZoneOpenedStream)
		return ok
	})
	defer connectionWatch.Close()

	for {
		select {
		case <-timer.C:
			for zone, firstSeen := range zw.zones {
				ctx := multitenant.WithTenant(context.TODO(), zone.tenantID)
				zoneInsight := system.NewZoneInsightResource()
				if err := zw.rm.Get(ctx, zoneInsight, store.GetByKey(zone.zone, model.NoMesh)); err != nil {
					zw.log.Info("error getting ZoneInsight", "zone", zone.zone)
				}
				lastHealthCheck := zoneInsight.Spec.GetHealthCheck().GetTime().AsTime()
				if lastHealthCheck.After(firstSeen) && time.Since(lastHealthCheck) > zw.timeout {
					zw.bus.Send(service.ZoneWentOffline{
						Zone:     zone.zone,
						TenantID: zone.tenantID,
					})
					delete(zw.zones, zone)
				}
			}
		case e := <-connectionWatch.Recv():
			newStream := e.(service.ZoneOpenedStream)

			ctx := multitenant.WithTenant(context.TODO(), newStream.TenantID)
			zoneInsight := system.NewZoneInsightResource()

			if err := zw.rm.Get(ctx, zoneInsight, store.GetByKey(newStream.Zone, model.NoMesh)); err != nil {
				zw.log.Info("error getting ZoneInsight", "zone", newStream.Zone)
			}

			zw.zones[zoneTenant{
				tenantID: newStream.TenantID,
				zone:     newStream.Zone,
			}] = zoneInsight.Spec.GetHealthCheck().GetTime().AsTime()
		case <-stop:
			return nil
		}
	}
}

func (zw *ZoneWatch) NeedLeaderElection() bool {
	return false
}
