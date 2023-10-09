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
	kuma_log "github.com/kumahq/kuma/pkg/log"
	"github.com/kumahq/kuma/pkg/multitenant"
)

type zoneTenant struct {
	zone     string
	tenantID string
}

type ZoneWatch struct {
	log        logr.Logger
	poll       time.Duration
	timeout    time.Duration
	bus        events.EventBus
	extensions context.Context
	rm         manager.ReadOnlyResourceManager
	zones      map[zoneTenant]time.Time
}

func NewZoneWatch(
	log logr.Logger,
	cfg multizone.KdsServerConfig,
	bus events.EventBus,
	rm manager.ReadOnlyResourceManager,
	extensions context.Context,
) *ZoneWatch {
	return &ZoneWatch{
		log:        log,
		poll:       cfg.ZoneHealthCheck.PollInterval.Duration,
		timeout:    cfg.ZoneHealthCheck.Timeout.Duration,
		bus:        bus,
		extensions: extensions,
		rm:         rm,
		zones:      map[zoneTenant]time.Time{},
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

				log := kuma_log.AddFieldsFromCtx(zw.log, ctx, zw.extensions)
				if err := zw.rm.Get(ctx, zoneInsight, store.GetByKey(zone.zone, model.NoMesh)); err != nil {
					log.Info("error getting ZoneInsight", "zone", zone.zone, "error", err)
					continue
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

			log := kuma_log.AddFieldsFromCtx(zw.log, ctx, zw.extensions)
			if err := zw.rm.Get(ctx, zoneInsight, store.GetByKey(newStream.Zone, model.NoMesh)); err != nil {
				log.Info("error getting ZoneInsight", "zone", newStream.Zone)
				continue
			}

			// We keep a record of the time we see in the ZoneInsight when the
			// stream is opened.
			// This is to prevent the zone from timing out on a poll
			// where the last health check is still from a previous connect, so:
			// a long time ago: zone CP disconnects, no more health checks are sent
			// now:
			//  zone CP opens streams
			//  global CP gets ZoneOpenedStream
			//  global CP runs poll and see the last health check from "a long time ago"
			//  BAD: global CP kills streams
			//  zone CP health check arrives
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
