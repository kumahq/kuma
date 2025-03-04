package mux

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	"github.com/prometheus/client_golang/prometheus"

	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
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
	"github.com/kumahq/kuma/pkg/util/proto"
)

type zoneTenant struct {
	zone     string
	tenantID string
}

type ZoneWatch struct {
	log         logr.Logger
	poll        time.Duration
	timeout     time.Duration
	bus         events.EventBus
	extensions  context.Context
	rm          manager.ReadOnlyResourceManager
	summary     prometheus.Summary
	zones       map[zoneTenant]time.Time
	zoneStreams map[zoneTenant]map[service.StreamType]time.Time
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
		log:         log,
		poll:        cfg.PollInterval.Duration,
		timeout:     cfg.Timeout.Duration,
		bus:         bus,
		extensions:  extensions,
		rm:          rm,
		summary:     summary,
		zones:       map[zoneTenant]time.Time{},
		zoneStreams: map[zoneTenant]map[service.StreamType]time.Time{},
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
						delete(zw.zoneStreams, zone)
					} else {
						log.Info("error getting ZoneInsight", "zone", zone.zone, "error", err)
					}
					continue
				}

				// It may be that we don't have a health check yet so we use the
				// lastSeen time because we know the zone was connected at that
				// point at least
				lastHealthCheck := zoneInsight.Spec.GetHealthCheck().GetTime().AsTime()
				if lastStreamOpened.After(lastHealthCheck) {
					lastHealthCheck = lastStreamOpened
				}
				if time.Since(lastHealthCheck) > zw.timeout {
					zw.bus.Send(service.ZoneWentOffline{
						Zone:     zone.zone,
						TenantID: zone.tenantID,
					})
					delete(zw.zones, zone)
					delete(zw.zoneStreams, zone)
				}
				// Now we want to check individual stream
				for stream, connOpenTime := range zw.zoneStreams[zone] {
					var conf *system_proto.KDSStream
					switch stream {
					case service.Clusters:
						conf = zoneInsight.Spec.GetKdsStreams().GetClusters()
					case service.ConfigDump:
						conf = zoneInsight.Spec.GetKdsStreams().GetConfigDump()
					case service.Stats:
						conf = zoneInsight.Spec.GetKdsStreams().GetStats()
					case service.GlobalToZone:
						conf = zoneInsight.Spec.GetKdsStreams().GetGlobalToZone()
					case service.ZoneToGlobal:
						conf = zoneInsight.Spec.GetKdsStreams().GetZoneToGlobal()
					}
					if conf == nil {
						continue
					}
					// If we have a connection that started before the one from insight, cancel the stream.
					// There's no need to check globalId since the connection exists in the map, meaning it is local.
					activeStreamConnTime := proto.MustTimestampFromProto(conf.GetConnectTime())
					if connOpenTime.Before(*activeStreamConnTime) {
						log.Info("the same zone has connected but the previous connection wasn't closed, closing",
							"zone", zone.zone, "streamType", stream, "previouslyConnected", connOpenTime, "currentlyConnected", activeStreamConnTime)
						zw.bus.Send(service.StreamCancelled{
							Zone:     zone.zone,
							TenantID: zone.tenantID,
							Type:     stream,
							ConnTime: connOpenTime,
						})
						delete(zw.zoneStreams, zone)
					}
				}
			}
			zw.summary.Observe(float64(core.Now().Sub(start).Milliseconds()))
		case e := <-connectionWatch.Recv():
			newStream := e.(service.ZoneOpenedStream)
			// Disconnect the old stream.
			// There should not be two streams for the same zone, as only the leader can connect to the global control plane.
			// Instead, we generate a unique StreamID for each stream. If we detect a second control plane
			// connecting to the same zone with a different StreamID, we cancel the previous stream.
			streams, found := zw.zoneStreams[zoneTenant{tenantID: newStream.TenantID, zone: newStream.Zone}]
			if found {
				if prevConnTime, exist := streams[newStream.Type]; exist && prevConnTime.Before(newStream.ConnTime) {
					ctx := multitenant.WithTenant(context.TODO(), newStream.TenantID)
					log := kuma_log.AddFieldsFromCtx(zw.log, ctx, zw.extensions)
					log.Info("the same zone has connected but the previous connection wasn't closed, closing",
						"zone", newStream.Zone, "streamType", newStream.Type, "previouslyConnected", prevConnTime, "currentlyConnected", newStream.ConnTime)
					zw.bus.Send(service.StreamCancelled{
						Zone:     newStream.Zone,
						TenantID: newStream.TenantID,
						Type:     newStream.Type,
						ConnTime: prevConnTime,
					})
					zw.zoneStreams[zoneTenant{tenantID: newStream.TenantID, zone: newStream.Zone}][newStream.Type] = newStream.ConnTime
				}
			} else {
				zw.zoneStreams[zoneTenant{tenantID: newStream.TenantID, zone: newStream.Zone}] = map[service.StreamType]time.Time{
					newStream.Type: newStream.ConnTime,
				}
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
			}] = core.Now()
		case <-stop:
			return nil
		}
	}
}

func (zw *ZoneWatch) NeedLeaderElection() bool {
	return false
}
