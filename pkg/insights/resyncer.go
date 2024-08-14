package insights

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/protobuf/proto"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	"github.com/kumahq/kuma/pkg/events"
	kuma_log "github.com/kumahq/kuma/pkg/log"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
	"github.com/kumahq/kuma/pkg/multitenant"
	util_maps "github.com/kumahq/kuma/pkg/util/maps"
)

var log = core.Log.WithName("mesh-insight-resyncer")

const (
	ReasonResync    = "resync"
	ReasonForce     = "force"
	ReasonEvent     = "event"
	ResultChanged   = "changed"
	ResultNoChanges = "no_changes"
)

func ServiceInsightKey(mesh string) model.ResourceKey {
	return model.ResourceKey{
		Name: fmt.Sprintf("all-services-%s", mesh),
		Mesh: mesh,
	}
}

func MeshInsightKey(mesh string) model.ResourceKey {
	return model.ResourceKey{
		Name: mesh,
		Mesh: model.NoMesh,
	}
}

type Config struct {
	Registry            registry.TypeRegistry
	ResourceManager     manager.ResourceManager
	EventReaderFactory  events.ListenerFactory
	MinResyncInterval   time.Duration
	FullResyncInterval  time.Duration
	Tick                func(d time.Duration) <-chan time.Time
	TenantFn            multitenant.Tenants
	EventBufferCapacity int
	EventProcessors     int
	Metrics             core_metrics.Metrics
	Now                 func() time.Time
	Extensions          context.Context
}

type resyncer struct {
	rm                    manager.ResourceManager
	eventFactory          events.ListenerFactory
	minResyncInterval     time.Duration
	stepsBeforeFullResync int
	tick                  func(d time.Duration) <-chan time.Time

	registry            registry.TypeRegistry
	tenantFn            multitenant.Tenants
	eventBufferCapacity int
	eventProcessors     int
	metrics             core_metrics.Metrics
	now                 func() time.Time

	idleTime           prometheus.Summary
	timeToProcessItem  prometheus.Summary
	itemProcessingTime *prometheus.SummaryVec

	allResourceTypes []model.ResourceType
	extensions       context.Context
}

// NewResyncer creates a new Component that periodically updates insights
// for various policies (right now only for Mesh and services).
//
// It operates with 2 timeouts: MinResyncInterval and FullResyncInterval. Component
// guarantees resync won't happen more often than MinResyncInterval. It also guarantees
// during FullResyncInterval at least one resync will happen. MinResyncInterval is provided
// by RateLimiter. FullResyncInterval is provided by goroutine with Ticker, it runs
// resync every t = FullResyncInterval - MinResyncInterval.
func NewResyncer(config *Config) component.Component {
	idleTime := prometheus.NewSummary(prometheus.SummaryOpts{
		Name:       "insights_resyncer_processor_idle_time",
		Help:       "Summary of the time that the processor loop sits idle, the closer this gets to 0 the more the processing loop is at capacity",
		Objectives: core_metrics.DefaultObjectives,
	})
	timeToProcessItem := prometheus.NewSummary(prometheus.SummaryOpts{
		Name:       "insights_resyncer_event_time_to_process",
		Help:       "Summary of the time between an event being added to a batch and it being processed, in a well behaving system this should be less of equal to the MinResyncInterval",
		Objectives: core_metrics.DefaultObjectives,
	})
	itemProcessingTime := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name:       "insights_resyncer_event_time_processing",
		Help:       "Summary of the time spent to process an event",
		Objectives: core_metrics.DefaultObjectives,
	}, []string{"reason", "result"})
	config.Metrics.MustRegister(idleTime, timeToProcessItem, itemProcessingTime)

	var allResourceTypes []model.ResourceType
	for _, desc := range config.Registry.ObjectDescriptors(model.HasScope(model.ScopeMesh), model.Not(model.IsInsight())) {
		allResourceTypes = append(allResourceTypes, desc.Name)
	}

	r := &resyncer{
		rm:                    config.ResourceManager,
		eventFactory:          config.EventReaderFactory,
		minResyncInterval:     config.MinResyncInterval,
		stepsBeforeFullResync: int(config.FullResyncInterval.Round(config.MinResyncInterval).Seconds() / config.MinResyncInterval.Seconds()),
		tick:                  config.Tick,
		registry:              config.Registry,
		tenantFn:              config.TenantFn,
		eventBufferCapacity:   config.EventBufferCapacity,
		metrics:               config.Metrics,
		now:                   config.Now,
		idleTime:              idleTime,
		timeToProcessItem:     timeToProcessItem,
		itemProcessingTime:    itemProcessingTime,
		allResourceTypes:      allResourceTypes,
		eventProcessors:       config.EventProcessors,
		extensions:            config.Extensions,
	}
	if config.Now == nil {
		r.now = time.Now
	}

	if config.Tick == nil {
		r.tick = time.Tick
	}

	return r
}

type resyncEvent struct {
	mesh     string
	tenantId string
	time     time.Time
	flag     actionFlag
	types    map[model.ResourceType]struct{}
	reasons  map[string]struct{}
}

type actionFlag uint8

const (
	FlagMesh = 1 << iota
	FlagService
)

// eventBatch keeps all the outstanding changes. The idea is that we linger an entire batch for some amount of time and we flush the batch all at once
type eventBatch struct {
	events map[string]*resyncEvent
	sync.Mutex
}

var flushAll = func(*resyncEvent) bool {
	return true
}

// flush sends the current batch to the resyncEvents chan, if the context is cancelled we interrupt the sending but keep the items in the batch.
// if an item is successfully put in the chanel we remove it.
func (e *eventBatch) flush(ctx context.Context, resyncEvents chan resyncEvent, predicate func(*resyncEvent) bool) error {
	e.Lock()
	defer e.Unlock()
	for k, event := range e.events {
		if !predicate(event) {
			continue
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("context done and the batch wasn't complete, update will be delayed, outstanding events: %d", len(e.events))
		case resyncEvents <- *event:
			// Once an event is sent we remove it from the batch
			delete(e.events, k)
		}
	}
	return nil
}

// add adds an item to the batch, if an item with a similar key exists we simply merge the actionFlags.
func (e *eventBatch) add(
	now time.Time,
	tenantId string,
	mesh string,
	actionFlag actionFlag,
	types []model.ResourceType,
	reason string,
) {
	if actionFlag == 0x00 { // No action so no need to persist
		return
	}
	e.Lock()
	defer e.Unlock()
	key := tenantId + ":" + mesh
	if elt := e.events[key]; elt != nil {
		// If the item is already present just merge the actionFlag
		elt.flag |= actionFlag
		for _, typ := range types {
			elt.types[typ] = struct{}{}
		}
		elt.reasons[reason] = struct{}{}
	} else {
		event := &resyncEvent{
			time:     now,
			tenantId: tenantId,
			mesh:     mesh,
			flag:     actionFlag,
			types:    map[model.ResourceType]struct{}{},
			reasons: map[string]struct{}{
				reason: {},
			},
		}
		for _, typ := range types {
			event.types[typ] = struct{}{}
		}
		e.events[key] = event
	}
}

func (r *resyncer) Start(stop <-chan struct{}) error {
	resyncEvents := make(chan resyncEvent, r.eventBufferCapacity)
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
		close(resyncEvents)
	}()
	for i := 0; i < r.eventProcessors; i++ {
		go func() {
			// We dequeue from the resyncEvents channel and actually do the insight update we want.
			for {
				start := r.now()
				select {
				case <-ctx.Done():
					return
				case event, more := <-resyncEvents:
					if !more {
						// Usually this shouldn't close if there's no closed context
						continue
					}
					tenantCtx := multitenant.WithTenant(ctx, event.tenantId)
					log := kuma_log.AddFieldsFromCtx(log, tenantCtx, r.extensions).WithValues("event", event)
					if err := r.processEvent(tenantCtx, start, event); err != nil && !errors.Is(err, context.Canceled) {
						log.Error(err, "could not process an event")
					}
				}
			}
		}()
	}
	eventReader := r.eventFactory.Subscribe(func(event events.Event) bool {
		if _, ok := event.(events.TriggerInsightsComputationEvent); ok {
			return true
		}
		if resourceChanged, ok := event.(events.ResourceChangedEvent); ok {
			desc, err := r.registry.DescriptorFor(resourceChanged.Type)
			if err != nil {
				log.Error(err, "Resource is not registered in the registry, ignoring it", "resource", resourceChanged.Type)
				return false
			}
			if desc.Scope == model.ScopeGlobal && desc.Name != core_mesh.MeshType {
				return false
			}
		}
		return true
	})
	defer eventReader.Close()
	batch := &eventBatch{events: map[string]*resyncEvent{}}
	ticker := r.tick(r.minResyncInterval)
	steps := 0
	for {
		select {
		// We tick every minResyncInterval and flush the batch so we process updates
		case now := <-ticker:
			steps += 1
			// Every fullResyncInterval we also add to the batch an update for each existing entities so we refresh all of them
			tickCtx, cancelTimeout := context.WithDeadline(ctx, now.Add(r.minResyncInterval))
			if steps == r.stepsBeforeFullResync {
				steps = 0
				tenantIds, err := r.tenantFn.GetIDs(ctx)
				if err != nil && !errors.Is(err, context.Canceled) {
					log.Error(err, "could not get tenants")
				}
				wg := sync.WaitGroup{}
				wg.Add(len(tenantIds))
				for _, tenantId := range tenantIds {
					go func(tenantId string) {
						r.addMeshesToBatch(tickCtx, batch, tenantId, ReasonResync)
						wg.Done()
					}(tenantId)
				}
				wg.Wait()
			}
			// We flush the batch. We want to check if the original context is cancelled to avoid logging on CP shutdown.
			if err := batch.flush(tickCtx, resyncEvents, flushAll); err != nil && !errors.Is(ctx.Err(), context.Canceled) {
				log.Error(err, "Flush of batch didn't complete, some insights won't be refreshed until next tick")
			}
			cancelTimeout()
		case event, ok := <-eventReader.Recv():
			if !ok {
				return errors.New("end of events channel")
			}
			if triggerEvent, ok := event.(events.TriggerInsightsComputationEvent); ok {
				ctx := context.Background()
				r.addMeshesToBatch(ctx, batch, triggerEvent.TenantID, ReasonForce)
				if err := batch.flush(ctx, resyncEvents, func(event *resyncEvent) bool {
					return event.tenantId == triggerEvent.TenantID
				}); err != nil {
					log.Error(err, "Flush of batch didn't complete, some insights won't be refreshed until next tick")
				}
			}
			if resourceChanged, ok := event.(events.ResourceChangedEvent); ok {
				supported, err := r.tenantFn.IDSupported(ctx, resourceChanged.TenantID)
				if err != nil && !errors.Is(err, context.Canceled) {
					log.Error(err, "could not determine if tenant ID is supported", "tenantID", resourceChanged.TenantID)
					continue
				}
				if !supported {
					continue
				}
				meshName := resourceChanged.Key.Mesh
				if resourceChanged.Type == core_mesh.MeshType {
					meshName = resourceChanged.Key.Name
				}
				var f actionFlag
				// 'Update' events doesn't affect MeshInsight except for DataplaneInsight, because that's how we find online/offline Dataplane's status
				if resourceChanged.Operation != events.Update || resourceChanged.Type == core_mesh.DataplaneInsightType {
					f |= FlagMesh
				}
				// Only a subset of types influence service insights
				if resourceChanged.Type == core_mesh.DataplaneType || resourceChanged.Type == core_mesh.DataplaneInsightType || resourceChanged.Type == core_mesh.ExternalServiceType {
					f |= FlagService
				}
				if f != 0 {
					batch.add(r.now(), resourceChanged.TenantID, meshName, f, []model.ResourceType{resourceChanged.Type}, ReasonEvent)
				}
			}
		case <-stop:
			log.Info("stop")
			return nil
		}
	}
}

func (r *resyncer) processEvent(ctx context.Context, start time.Time, event resyncEvent) error {
	if event.flag == 0 {
		return nil
	}
	startProcessingTime := r.now()
	r.idleTime.Observe(float64(startProcessingTime.Sub(start).Milliseconds()))
	r.timeToProcessItem.Observe(float64(startProcessingTime.Sub(event.time).Milliseconds()))
	dpOverviews, err := r.dpOverviews(ctx, event.mesh)
	if err != nil {
		return errors.Wrap(err, "unable to get DataplaneOverviews to recompute insights")
	}

	externalServices := &core_mesh.ExternalServiceResourceList{}
	if err := r.rm.List(ctx, externalServices, store.ListByMesh(event.mesh)); err != nil {
		return errors.Wrap(err, "unable to get ExternalServices to recompute insights")
	}

	anyChanged := false
	if event.flag&FlagService == FlagService {
		err, changed := r.createOrUpdateServiceInsight(ctx, event.mesh, dpOverviews, externalServices.Items)
		if err != nil {
			return errors.Wrap(err, "unable to resync ServiceInsight")
		}
		if changed {
			anyChanged = true
		}
	}
	if event.flag&FlagMesh == FlagMesh {
		err, changed := r.createOrUpdateMeshInsight(ctx, event.mesh, dpOverviews, externalServices.Items, event.types)
		if err != nil {
			return errors.Wrap(err, "unable to resync MeshInsight")
		}
		if changed {
			anyChanged = true
		}
	}
	reason := strings.Join(util_maps.SortedKeys(event.reasons), "_and_")
	result := ResultNoChanges
	if anyChanged {
		result = ResultChanged
	}
	r.itemProcessingTime.WithLabelValues(reason, result).Observe(float64(time.Since(startProcessingTime).Milliseconds()))
	return nil
}

func (r *resyncer) dpOverviews(ctx context.Context, mesh string) ([]*core_mesh.DataplaneOverviewResource, error) {
	dataplanes := &core_mesh.DataplaneResourceList{}
	if err := r.rm.List(ctx, dataplanes, store.ListByMesh(mesh)); err != nil {
		return nil, err
	}

	dpInsights := &core_mesh.DataplaneInsightResourceList{}
	if err := r.rm.List(ctx, dpInsights, store.ListByMesh(mesh)); err != nil {
		return nil, err
	}
	overviews := core_mesh.NewDataplaneOverviews(*dataplanes, *dpInsights)
	return overviews.Items, nil
}

func (r *resyncer) addMeshesToBatch(ctx context.Context, batch *eventBatch, tenantID string, reason string) {
	meshList := &core_mesh.MeshResourceList{}
	tenantCtx := multitenant.WithTenant(ctx, tenantID)
	if err := r.rm.List(tenantCtx, meshList); err != nil && !errors.Is(err, context.Canceled) {
		log := kuma_log.AddFieldsFromCtx(log, tenantCtx, r.extensions)
		log.Error(err, "failed to get list of meshes")
		return
	}
	for _, mesh := range meshList.Items {
		batch.add(time.Now(), tenantID, mesh.GetMeta().GetName(), FlagMesh|FlagService, r.allResourceTypes, reason)
	}
}

func populateInsight(serviceType mesh_proto.ServiceInsight_Service_Type, insight *mesh_proto.ServiceInsight, svcName string, status core_mesh.Status, backend string, addressPort string) {
	if _, ok := insight.Services[svcName]; !ok {
		insight.Services[svcName] = &mesh_proto.ServiceInsight_Service{
			IssuedBackends: map[string]uint32{},
			Dataplanes:     &mesh_proto.ServiceInsight_Service_DataplaneStat{},
			ServiceType:    serviceType,
			AddressPort:    addressPort,
		}
	}
	if backend != "" {
		insight.Services[svcName].IssuedBackends[backend]++
	}

	dataplanes := insight.Services[svcName].Dataplanes

	switch status {
	case core_mesh.Online:
		dataplanes.Online++
	case core_mesh.Offline:
		dataplanes.Offline++
	case core_mesh.PartiallyDegraded:
		dataplanes.Offline++
	}
}

func (r *resyncer) createOrUpdateServiceInsight(
	ctx context.Context,
	mesh string,
	dpOverviews []*core_mesh.DataplaneOverviewResource,
	externalServices []*core_mesh.ExternalServiceResource,
) (error, bool) {
	log := kuma_log.AddFieldsFromCtx(log, ctx, r.extensions).WithValues("mesh", mesh) // Add info
	insight := &mesh_proto.ServiceInsight{
		Services: map[string]*mesh_proto.ServiceInsight_Service{},
	}

	zonesMap := map[string]map[string]struct{}{}
	addSvcToZones := func(svc, zone string) {
		if zone == "" {
			return
		}
		if _, ok := zonesMap[svc]; !ok {
			zonesMap[svc] = map[string]struct{}{}
		}
		zonesMap[svc][zone] = struct{}{}
	}
	for _, dpOverview := range dpOverviews {
		status, _ := dpOverview.Status()
		networking := dpOverview.Spec.GetDataplane().GetNetworking()
		backend := dpOverview.Spec.GetDataplaneInsight().GetMTLS().GetIssuedBackend()

		if gw := networking.GetGateway(); gw != nil {
			var svcType mesh_proto.ServiceInsight_Service_Type
			switch gw.Type {
			case mesh_proto.Dataplane_Networking_Gateway_BUILTIN:
				svcType = mesh_proto.ServiceInsight_Service_gateway_builtin
			case mesh_proto.Dataplane_Networking_Gateway_DELEGATED:
				svcType = mesh_proto.ServiceInsight_Service_gateway_delegated
			}
			populateInsight(svcType, insight, gw.GetTags()[mesh_proto.ServiceTag], status, backend, "")
			addSvcToZones(gw.GetTags()[mesh_proto.ServiceTag], gw.GetTags()[mesh_proto.ZoneTag])
		}

		for _, inbound := range networking.GetInbound() {
			if inbound.State == mesh_proto.Dataplane_Networking_Inbound_Ignored {
				continue
			}
			// address port is empty to save space in the resource. It will be filled by the server on API response
			populateInsight(mesh_proto.ServiceInsight_Service_internal, insight, inbound.GetService(), status, backend, "")
			addSvcToZones(inbound.GetService(), inbound.GetTags()[mesh_proto.ZoneTag])
		}
	}

	for _, es := range externalServices {
		populateInsight(mesh_proto.ServiceInsight_Service_external, insight, es.Spec.GetService(), "", "", es.Spec.Networking.GetAddress())
		addSvcToZones(es.Spec.GetService(), es.Spec.GetTags()[mesh_proto.ZoneTag])
	}

	for svcName, svc := range insight.Services {
		online := svc.Dataplanes.Online
		total := svc.Dataplanes.Online + svc.Dataplanes.Offline

		switch {
		case svc.ServiceType == mesh_proto.ServiceInsight_Service_external:
			svc.Status = mesh_proto.ServiceInsight_Service_not_available
		case online == 0:
			svc.Status = mesh_proto.ServiceInsight_Service_offline
		case online == total:
			svc.Status = mesh_proto.ServiceInsight_Service_online
		case online < total:
			svc.Status = mesh_proto.ServiceInsight_Service_partially_degraded
		}
		if zones, ok := zonesMap[svcName]; ok {
			svc.Zones = util_maps.SortedKeys(zones)
		}
	}

	key := ServiceInsightKey(mesh)
	changed := false
	err := manager.Upsert(ctx, r.rm, key, core_mesh.NewServiceInsightResource(), func(resource model.Resource) error {
		if resource.GetSpec() != nil && proto.Equal(resource.GetSpec().(proto.Message), insight) {
			log.V(1).Info("no need to update ServiceInsight because the resource is the same")
			return manager.ErrSkipUpsert
		}
		changed = true
		return resource.SetSpec(insight)
	})
	if err != nil {
		if manager.IsMeshNotFound(err) {
			log.V(1).Info("ServiceInsight is not updated because mesh no longer exist. This can happen when Mesh is being deleted.")
			// handle the situation when the mesh is deleted and then allByType the resources connected with the Mesh allByType deleted.
			// Mesh no longer exist so we cannot upsert the insight for it.
			return nil, false
		}
		if errors.Is(err, &store.ResourceConflictError{}) {
			log.V(1).Info(err.Error())
			return nil, false
		}
		return err, false
	}
	log.V(1).Info("ServiceInsights updated")
	return nil, changed
}

func (r *resyncer) createOrUpdateMeshInsight(
	ctx context.Context,
	mesh string,
	dpOverviews []*core_mesh.DataplaneOverviewResource,
	externalServices []*core_mesh.ExternalServiceResource,
	types map[model.ResourceType]struct{},
) (error, bool) {
	log := kuma_log.AddFieldsFromCtx(log, ctx, r.extensions).WithValues("mesh", mesh) // Add info
	insight := &mesh_proto.MeshInsight{
		Dataplanes: &mesh_proto.MeshInsight_DataplaneStat{},
		DataplanesByType: &mesh_proto.MeshInsight_DataplanesByType{
			Standard:         &mesh_proto.MeshInsight_DataplaneStat{},
			Gateway:          &mesh_proto.MeshInsight_DataplaneStat{},
			GatewayBuiltin:   &mesh_proto.MeshInsight_DataplaneStat{},
			GatewayDelegated: &mesh_proto.MeshInsight_DataplaneStat{},
		},
		Policies:  map[string]*mesh_proto.MeshInsight_PolicyStat{},
		Resources: map[string]*mesh_proto.MeshInsight_ResourceStat{},
		DpVersions: &mesh_proto.MeshInsight_DpVersions{
			KumaDp: map[string]*mesh_proto.MeshInsight_DataplaneStat{},
			Envoy:  map[string]*mesh_proto.MeshInsight_DataplaneStat{},
		},
		MTLS: &mesh_proto.MeshInsight_MTLS{
			IssuedBackends:    map[string]*mesh_proto.MeshInsight_DataplaneStat{},
			SupportedBackends: map[string]*mesh_proto.MeshInsight_DataplaneStat{},
		},
	}

	insight.Dataplanes.Total = uint32(len(dpOverviews))
	internalServices := map[string]struct{}{}

	for _, dpOverview := range dpOverviews {
		dpInsight := dpOverview.Spec.DataplaneInsight
		dpSubscription := dpInsight.GetLastSubscription().(*mesh_proto.DiscoverySubscription)
		kumaDpVersion := getOrDefault(dpSubscription.GetVersion().GetKumaDp().GetVersion())
		envoyVersion := getOrDefault(dpSubscription.GetVersion().GetEnvoy().GetVersion())
		networking := dpOverview.Spec.GetDataplane().GetNetworking()

		ensureVersionExists(kumaDpVersion, insight.DpVersions.KumaDp)
		ensureVersionExists(envoyVersion, insight.DpVersions.Envoy)

		status, _ := dpOverview.Status()

		statByType := insight.GetDataplanesByType().GetStandard()
		if networking.GetGateway() != nil {
			switch networking.GetGateway().GetType() {
			case mesh_proto.Dataplane_Networking_Gateway_BUILTIN:
				statByType = insight.GetDataplanesByType().GetGatewayBuiltin()
			case mesh_proto.Dataplane_Networking_Gateway_DELEGATED:
				statByType = insight.GetDataplanesByType().GetGatewayDelegated()
			}
		}

		statByType.Total++

		switch status {
		case core_mesh.Online:
			insight.Dataplanes.Online++
			statByType.Online++
			insight.DpVersions.KumaDp[kumaDpVersion].Online++
			insight.DpVersions.Envoy[envoyVersion].Online++
		case core_mesh.PartiallyDegraded:
			insight.Dataplanes.PartiallyDegraded++
			statByType.PartiallyDegraded++
			insight.DpVersions.KumaDp[kumaDpVersion].PartiallyDegraded++
			insight.DpVersions.Envoy[envoyVersion].PartiallyDegraded++
		case core_mesh.Offline:
			insight.Dataplanes.Offline++
			statByType.Offline++
			insight.DpVersions.KumaDp[kumaDpVersion].Offline++
			insight.DpVersions.Envoy[envoyVersion].Offline++
		}

		updateTotal(kumaDpVersion, insight.DpVersions.KumaDp)
		updateTotal(envoyVersion, insight.DpVersions.Envoy)
		updateMTLS(dpInsight.GetMTLS(), status, insight.MTLS)

		if svc := networking.GetGateway().GetTags()[mesh_proto.ServiceTag]; svc != "" {
			internalServices[svc] = struct{}{}
		}

		for _, inbound := range networking.GetInbound() {
			if inbound.State == mesh_proto.Dataplane_Networking_Inbound_Ignored {
				continue
			}
			internalServices[inbound.GetService()] = struct{}{}
		}
	}

	insight.DataplanesByType.Gateway.Online = insight.GetDataplanesByType().GetGatewayBuiltin().GetOnline() + insight.GetDataplanesByType().GetGatewayDelegated().GetOnline()
	insight.DataplanesByType.Gateway.Offline = insight.GetDataplanesByType().GetGatewayBuiltin().GetOffline() + insight.GetDataplanesByType().GetGatewayDelegated().GetOffline()
	insight.DataplanesByType.Gateway.PartiallyDegraded = insight.GetDataplanesByType().GetGatewayBuiltin().GetPartiallyDegraded() + insight.GetDataplanesByType().GetGatewayDelegated().GetPartiallyDegraded()
	insight.DataplanesByType.Gateway.Total = insight.GetDataplanesByType().GetGatewayBuiltin().GetTotal() + insight.GetDataplanesByType().GetGatewayDelegated().GetTotal()

	insight.Services = &mesh_proto.MeshInsight_ServiceStat{
		Total:    uint32(len(internalServices) + len(externalServices)),
		Internal: uint32(len(internalServices)),
		External: uint32(len(externalServices)),
	}

	key := MeshInsightKey(mesh)
	changed := false
	err := manager.Upsert(ctx, r.rm, key, core_mesh.NewMeshInsightResource(), func(resource model.Resource) error {
		oldInsight := resource.GetSpec().(*mesh_proto.MeshInsight)
		for k, v := range oldInsight.Resources {
			insight.Resources[k] = proto.Clone(v).(*mesh_proto.MeshInsight_ResourceStat)
		}
		for k, v := range oldInsight.Policies {
			insight.Policies[k] = proto.Clone(v).(*mesh_proto.MeshInsight_PolicyStat)
		}
		if proto.Equal(oldInsight, &mesh_proto.MeshInsight{}) {
			// insight was not yet computed, need to update all
			for _, typ := range r.allResourceTypes {
				types[typ] = struct{}{}
			}
		}

		for typ := range types {
			desc, err := r.registry.DescriptorFor(typ)
			if err != nil {
				return err
			}

			if desc.IsInsight() {
				// It's expensive to retrieve insights and the counter is not useful without the parent object.
				continue
			}

			var count int
			// Reuse counter of resources that we already have
			switch typ {
			case core_mesh.ExternalServiceType:
				count = len(externalServices)
			case core_mesh.DataplaneType:
				count = len(dpOverviews)
			default:
				list := desc.NewList()
				if err := r.rm.List(ctx, list, store.ListByMesh(mesh)); err != nil {
					return err
				}
				count = len(list.GetItems())
			}

			if count != 0 {
				insight.Resources[string(typ)] = &mesh_proto.MeshInsight_ResourceStat{
					Total: uint32(count),
				}
				if desc.IsPolicy {
					// backwards compatibility
					insight.Policies[string(typ)] = &mesh_proto.MeshInsight_PolicyStat{
						Total: uint32(count),
					}
				}
			}
			if count == 0 {
				delete(insight.Resources, string(typ))
				delete(insight.Policies, string(typ))
			}
		}

		if proto.Equal(resource.GetSpec().(proto.Message), insight) {
			log.V(1).Info("no need to update MeshInsight because the resource is the same")
			return manager.ErrSkipUpsert
		}
		changed = true
		return resource.SetSpec(insight)
	})
	if err != nil {
		if manager.IsMeshNotFound(err) {
			log.V(1).Info("MeshInsight is not updated because mesh no longer exist. This can happen when Mesh is being deleted.")
			// handle the situation when the mesh is deleted and then allByType the resources connected with the Mesh allByType deleted.
			// Mesh no longer exist so we cannot upsert the insight for it.
			return nil, false
		}
		if errors.Is(err, &store.ResourceConflictError{}) {
			log.V(1).Info(err.Error())
			return nil, false
		}
		return err, false
	}
	log.V(1).Info("MeshInsight updated")
	return nil, changed
}

func updateMTLS(mtlsInsight *mesh_proto.DataplaneInsight_MTLS, status core_mesh.Status, stats *mesh_proto.MeshInsight_MTLS) {
	if mtlsInsight == nil {
		return
	}

	backend := mtlsInsight.GetIssuedBackend()
	if stat := stats.IssuedBackends[backend]; stat == nil {
		stats.IssuedBackends[backend] = &mesh_proto.MeshInsight_DataplaneStat{}
	}

	switch status {
	case core_mesh.Online:
		stats.IssuedBackends[backend].Online++
	case core_mesh.PartiallyDegraded:
		stats.IssuedBackends[backend].PartiallyDegraded++
	case core_mesh.Offline:
		stats.IssuedBackends[backend].Offline++
	}
	stats.IssuedBackends[backend].Total++

	for _, backend := range mtlsInsight.GetSupportedBackends() {
		if stat := stats.SupportedBackends[backend]; stat == nil {
			stats.SupportedBackends[backend] = &mesh_proto.MeshInsight_DataplaneStat{}
		}

		switch status {
		case core_mesh.Online:
			stats.SupportedBackends[backend].Online++
		case core_mesh.PartiallyDegraded:
			stats.SupportedBackends[backend].PartiallyDegraded++
		case core_mesh.Offline:
			stats.SupportedBackends[backend].Offline++
		}
		stats.SupportedBackends[backend].Total++
	}
}

func updateTotal(version string, dpStats map[string]*mesh_proto.MeshInsight_DataplaneStat) {
	dpStats[version].Total = dpStats[version].Online + dpStats[version].Offline
}

func ensureVersionExists(version string, m map[string]*mesh_proto.MeshInsight_DataplaneStat) {
	if _, versionExists := m[version]; !versionExists {
		m[version] = &mesh_proto.MeshInsight_DataplaneStat{}
	}
}

func getOrDefault(version string) string {
	if version == "" {
		return "unknown"
	}
	return version
}

func (r *resyncer) NeedLeaderElection() bool {
	return !r.tenantFn.SupportsSharding()
}
