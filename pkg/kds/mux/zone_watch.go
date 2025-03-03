package mux

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/kumahq/kuma/pkg/config/multizone"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/events"
	"github.com/kumahq/kuma/pkg/kds/service"
	kuma_log "github.com/kumahq/kuma/pkg/log"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
	"github.com/kumahq/kuma/pkg/multitenant"
)

type zoneTenant struct {
	zone     string
	tenantID string
}

type streamConnStart struct {
	id             int64
	connectionTime time.Time
}

type ZoneWatch struct {
	log        logr.Logger
	poll       time.Duration
	timeout    time.Duration
	bus        events.EventBus
	extensions context.Context
	rm         manager.ReadOnlyResourceManager
	summary    prometheus.Summary
	zones      map[zoneTenant]streamConnStart
}

func NewZoneWatch(
	log logr.Logger,
	cfg multizone.ZoneHealthCheckConfig,
	metrics prometheus.Registerer,
	bus events.EventBus,
	rm manager.ReadOnlyResourceManager,
	extensions context.Context,
) (*ZoneWatch, error) {
	summary := prometheus.NewSummary(prometheus.SummaryOpts{
		Name:       "component_zone_watch",
		Help:       "Summary of ZoneWatch component interval",
		Objectives: core_metrics.DefaultObjectives,
	})
	if err := metrics.Register(summary); err != nil {
		return nil, err
	}

	return &ZoneWatch{
		log:        log,
		poll:       cfg.PollInterval.Duration,
		timeout:    cfg.Timeout.Duration,
		bus:        bus,
		extensions: extensions,
		rm:         rm,
		summary:    summary,
		zones:      map[zoneTenant]streamConnStart{},
	}, nil
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
			start := core.Now()
			for zone, lastStreamOpened := range zw.zones {
				ctx := multitenant.WithTenant(context.TODO(), zone.tenantID)
				zoneInsight := system.NewZoneInsightResource()

				log := kuma_log.AddFieldsFromCtx(zw.log, ctx, zw.extensions)
				if err := zw.rm.Get(ctx, zoneInsight, store.GetByKey(zone.zone, model.NoMesh)); err != nil {
					if store.IsResourceNotFound(err) {
						zw.bus.Send(service.ZoneWentOffline{
							Zone:     zone.zone,
							TenantID: zone.tenantID,
						})
						delete(zw.zones, zone)
					} else {
						log.Info("error getting ZoneInsight", "zone", zone.zone, "error", err)
					}
					continue
				}

				// It may be that we don't have a health check yet so we use the
				// lastSeen time because we know the zone was connected at that
				// point at least
				lastHealthCheck := zoneInsight.Spec.GetHealthCheck().GetTime().AsTime()
				if lastStreamOpened.connectionTime.After(lastHealthCheck) {
					lastHealthCheck = lastStreamOpened.connectionTime
				}
				if time.Since(lastHealthCheck) > zw.timeout {
					zw.bus.Send(service.ZoneWentOffline{
						Zone:     zone.zone,
						TenantID: zone.tenantID,
					})
					delete(zw.zones, zone)
				}
			}
			zw.summary.Observe(float64(core.Now().Sub(start).Milliseconds()))
		case e := <-connectionWatch.Recv():
			newStream := e.(service.ZoneOpenedStream)
			// Disconnect the old stream.
			// There should not be two streams for the same zone, as only the leader can connect to the global control plane.
			// Instead, we generate a unique StreamID for each stream. If we detect a second control plane 
			// connecting to the same zone with a different StreamID, we cancel the previous stream.
			if val, found := zw.zones[zoneTenant{tenantID: newStream.TenantID, zone: newStream.Zone}]; found && val.id != newStream.StreamID {
				zw.bus.Send(service.StreamCancelled{
					Zone:     newStream.Zone,
					TenantID: newStream.TenantID,
					StreamID: val.id,
				})
			}

			// We keep a record of the time we open a stream.
			// This is to prevent the zone from timing out on a poll
			// where the last health check is still from a previous connect, so:
			// a long time ago: zone CP disconnects, no more health checks are sent
			// now:
			//  zone CP opens streams
			//  global CP gets ZoneOpenedStream (but we don't stash the time as below)
			//  global CP runs poll and see the last health check from "a long time ago"
			//  BAD: global CP kills streams
			//  zone CP health check arrives
			zw.zones[zoneTenant{
				tenantID: newStream.TenantID,
				zone:     newStream.Zone,
			}] = streamConnStart{
				id:             newStream.StreamID,
				connectionTime: time.Now(),
			}
		case <-stop:
			return nil
		}
	}
}

func (zw *ZoneWatch) NeedLeaderElection() bool {
	return false
}
