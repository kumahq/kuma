package insights

import (
	"context"
	"fmt"
	"sync"
	"time"

	"golang.org/x/time/rate"
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
)

var (
	log = core.Log.WithName("mesh-insight-resyncer")
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
	Registry           registry.TypeRegistry
	ResourceManager    manager.ResourceManager
	EventReaderFactory events.ListenerFactory
	MinResyncTimeout   time.Duration
	MaxResyncTimeout   time.Duration
	Tick               func(d time.Duration) <-chan time.Time
	RateLimiterFactory func() *rate.Limiter
}

type resyncer struct {
	rm                 manager.ResourceManager
	eventFactory       events.ListenerFactory
	minResyncTimeout   time.Duration
	maxResyncTimeout   time.Duration
	tick               func(d time.Duration) <-chan time.Time
	rateLimiterFactory func() *rate.Limiter
	meshInsightMux     sync.Mutex
	serviceInsightMux  sync.Mutex
	rateLimiters       map[string]*rate.Limiter
	// info provides an information about the last sync for both ServiceInsight and MeshInsight
	// Previously this data was stored in the Resource itself, but since resyncer runs only on leader
	// we can just save this in memory and save requests to the DB.
	infos    map[model.ResourceKey]syncInfo
	infosMux sync.RWMutex

	registry registry.TypeRegistry
}

type syncInfo struct {
	resource proto.Message
	lastSync time.Time
}

// NewResyncer creates a new Component that periodically updates insights
// for various policies (right now only for Mesh and services).
//
// It operates with 2 timeouts: MinResyncTimeout and MaxResyncTimeout. Component
// guarantees resync won't happen more often than MinResyncTimeout. It also guarantees
// during MaxResyncTimeout at least one resync will happen. MinResyncTimeout is provided
// by RateLimiter. MaxResyncTimeout is provided by goroutine with Ticker, it runs
// resync every t = MaxResyncTimeout - MinResyncTimeout.
func NewResyncer(config *Config) component.Component {
	r := &resyncer{
		minResyncTimeout:   config.MinResyncTimeout,
		maxResyncTimeout:   config.MaxResyncTimeout,
		eventFactory:       config.EventReaderFactory,
		rm:                 config.ResourceManager,
		rateLimiterFactory: config.RateLimiterFactory,
		rateLimiters:       map[string]*rate.Limiter{},
		infos:              map[model.ResourceKey]syncInfo{},
		registry:           config.Registry,
	}

	r.tick = config.Tick
	if config.Tick == nil {
		r.tick = time.Tick
	}

	return r
}

func (r *resyncer) Start(stop <-chan struct{}) error {
	r.clearInfos() // clear the map if component is restarted
	go func(stop <-chan struct{}) {
		ticker := r.tick(r.maxResyncTimeout - r.minResyncTimeout)
		for {
			select {
			case <-ticker:
				if err := r.createOrUpdateMeshInsights(); err != nil {
					log.Error(err, "unable to resync MeshInsight")
				}
				if err := r.createOrUpdateServiceInsights(); err != nil {
					log.Error(err, "unable to resync ServiceInsight")
				}
			case <-stop:
				log.Info("stop")
				return
			}
		}
	}(stop)

	eventReader := r.eventFactory.New()
	for {
		event, err := eventReader.Recv(stop)
		if err == events.ListenerStoppedErr {
			return nil
		}
		if err != nil {
			return err
		}
		resourceChanged, ok := event.(events.ResourceChangedEvent)
		if !ok {
			continue
		}
		desc, err := r.registry.DescriptorFor(resourceChanged.Type)
		if err != nil {
			log.Error(err, "Resource is not registered in the registry, ignoring it", "resource", resourceChanged.Type)
		}
		if resourceChanged.Type == core_mesh.MeshType && resourceChanged.Operation == events.Delete {
			r.deleteRateLimiter(resourceChanged.Key.Name)
		}
		if !r.getRateLimiter(resourceChanged.Key.Mesh).Allow() {
			continue
		}
		if resourceChanged.Type == core_mesh.DataplaneType || resourceChanged.Type == core_mesh.DataplaneInsightType {
			if err := r.createOrUpdateServiceInsight(resourceChanged.Key.Mesh); err != nil {
				log.Error(err, "unable to resync ServiceInsight", "mesh", resourceChanged.Key.Mesh)
			}
		}
		if desc.Scope == model.ScopeGlobal {
			continue
		}
		if resourceChanged.Operation == events.Update && resourceChanged.Type != core_mesh.DataplaneInsightType {
			// 'Update' events doesn't affect MeshInsight except for DataplaneInsight,
			// because that's how we find online/offline Dataplane's status
			continue
		}
		if err := r.createOrUpdateMeshInsight(resourceChanged.Key.Mesh); err != nil {
			log.Error(err, "unable to resync MeshInsight", "mesh", resourceChanged.Key.Mesh)
			continue
		}
	}
}

func (r *resyncer) getRateLimiter(mesh string) *rate.Limiter {
	if _, ok := r.rateLimiters[mesh]; !ok {
		r.rateLimiters[mesh] = r.rateLimiterFactory()
	}
	return r.rateLimiters[mesh]
}

func (r *resyncer) deleteRateLimiter(mesh string) {
	if _, ok := r.rateLimiters[mesh]; !ok {
		return
	}
	delete(r.rateLimiters, mesh)
}

func (r *resyncer) createOrUpdateServiceInsights() error {
	meshes := &core_mesh.MeshResourceList{}
	if err := r.rm.List(context.Background(), meshes); err != nil {
		return err
	}
	for _, mesh := range meshes.Items {
		if need := r.needResyncServiceInsight(mesh.GetMeta().GetName()); !need {
			continue
		}
		err := r.createOrUpdateServiceInsight(mesh.GetMeta().GetName())
		if err != nil {
			log.Error(err, "unable to resync resources", "mesh", mesh.GetMeta().GetName())
			continue
		}
	}
	return nil
}

func addDpStatusToInsight(insight *mesh_proto.ServiceInsight, svcName string, status core_mesh.Status) {
	if _, ok := insight.Services[svcName]; !ok {
		insight.Services[svcName] = &mesh_proto.ServiceInsight_Service{
			IssuedBackends: map[string]uint32{},
			Dataplanes:     &mesh_proto.ServiceInsight_Service_DataplaneStat{},
		}
	}

	dataplanes := insight.Services[svcName].Dataplanes

	dataplanes.Total++

	switch status {
	case core_mesh.Online:
		dataplanes.Online++
	case core_mesh.Offline:
		dataplanes.Offline++
	case core_mesh.PartiallyDegraded:
		dataplanes.Offline++
	}
}

func addDpOverviewToInsight(insight *mesh_proto.ServiceInsight, dpOverview *core_mesh.DataplaneOverviewResource) {
	status, _ := dpOverview.GetStatus()
	networking := dpOverview.Spec.GetDataplane().GetNetworking()
	backend := dpOverview.Spec.GetDataplaneInsight().GetMTLS().GetIssuedBackend()

	if svc := networking.GetGateway().GetTags()[mesh_proto.ServiceTag]; svc != "" {
		addDpStatusToInsight(insight, svc, status)
		if backend != "" {
			insight.Services[svc].IssuedBackends[backend]++
		}
	}

	for _, inbound := range networking.GetInbound() {
		addDpStatusToInsight(insight, inbound.GetService(), status)
		if backend != "" {
			insight.Services[inbound.GetService()].IssuedBackends[backend]++
		}
	}
}

func (r *resyncer) createOrUpdateServiceInsight(mesh string) error {
	ctx := context.TODO()

	r.serviceInsightMux.Lock()
	defer r.serviceInsightMux.Unlock()

	insight := &mesh_proto.ServiceInsight{
		Services: map[string]*mesh_proto.ServiceInsight_Service{},
	}
	dp := &core_mesh.DataplaneResourceList{}
	if err := r.rm.List(context.Background(), dp, store.ListByMesh(mesh)); err != nil {
		return err
	}
	dpInsights := &core_mesh.DataplaneInsightResourceList{}
	if err := r.rm.List(context.Background(), dpInsights, store.ListByMesh(mesh)); err != nil {
		return err
	}
	dpOverviews := core_mesh.NewDataplaneOverviews(*dp, *dpInsights)

	for _, dpOverview := range dpOverviews.Items {
		addDpOverviewToInsight(insight, dpOverview)
	}

	for _, svc := range insight.Services {
		online, total := svc.Dataplanes.Online, svc.Dataplanes.Total

		switch {
		case online == 0:
			svc.Status = mesh_proto.ServiceInsight_Service_offline
		case online == total:
			svc.Status = mesh_proto.ServiceInsight_Service_online
		case online < total:
			svc.Status = mesh_proto.ServiceInsight_Service_partially_degraded
		}
	}

	key := ServiceInsightKey(mesh)
	info, _ := r.getInfo(key)
	if proto.Equal(info.resource, insight) {
		log.V(1).Info("no need to update ServiceInsight because the resource is the same")
		info.lastSync = core.Now()
		r.setInfo(key, info)
		return nil
	}

	err := manager.Upsert(ctx, r.rm, key, core_mesh.NewServiceInsightResource(), func(resource model.Resource) error {
		return resource.SetSpec(insight)
	})
	if err != nil {
		if manager.IsMeshNotFound(err) {
			log.V(1).Info("ServiceInsight is not updated because mesh no longer exist. This can happen when Mesh is being deleted.")
			// handle the situation when the mesh is deleted and then all the resources connected with the Mesh all deleted.
			// Mesh no longer exist so we cannot upsert the insight for it.
			return nil
		}
		if store.IsResourceConflict(err) {
			log.V(1).Info("ServiceInsight was updated in other place. Retrying")
			return nil
		}
		return err
	}
	log.V(1).Info("ServiceInsights updated")
	r.setInfo(key, syncInfo{
		resource: insight,
		lastSync: core.Now(),
	})
	return nil
}

func (r *resyncer) createOrUpdateMeshInsights() error {
	meshes := &core_mesh.MeshResourceList{}
	if err := r.rm.List(context.Background(), meshes); err != nil {
		return err
	}
	for _, mesh := range meshes.Items {
		if need := r.needResyncMeshInsight(mesh.GetMeta().GetName()); !need {
			continue
		}
		err := r.createOrUpdateMeshInsight(mesh.GetMeta().GetName())
		if err != nil {
			log.Error(err, "unable to resync resources", "mesh", mesh.GetMeta().GetName())
			continue
		}
	}
	return nil
}

func (r *resyncer) createOrUpdateMeshInsight(mesh string) error {
	ctx := context.TODO()

	r.meshInsightMux.Lock()
	defer r.meshInsightMux.Unlock()

	insight := &mesh_proto.MeshInsight{
		Dataplanes: &mesh_proto.MeshInsight_DataplaneStat{},
		DataplanesByType: &mesh_proto.MeshInsight_DataplanesByType{
			Standard: &mesh_proto.MeshInsight_DataplaneStat{},
			Gateway:  &mesh_proto.MeshInsight_DataplaneStat{},
		},
		Policies: map[string]*mesh_proto.MeshInsight_PolicyStat{},
		DpVersions: &mesh_proto.MeshInsight_DpVersions{
			KumaDp: map[string]*mesh_proto.MeshInsight_DataplaneStat{},
			Envoy:  map[string]*mesh_proto.MeshInsight_DataplaneStat{},
		},
		MTLS: &mesh_proto.MeshInsight_MTLS{
			IssuedBackends:    map[string]*mesh_proto.MeshInsight_DataplaneStat{},
			SupportedBackends: map[string]*mesh_proto.MeshInsight_DataplaneStat{},
		},
	}

	dataplanes := &core_mesh.DataplaneResourceList{}
	if err := r.rm.List(context.Background(), dataplanes, store.ListByMesh(mesh)); err != nil {
		return err
	}

	insight.Dataplanes.Total = uint32(len(dataplanes.GetItems()))

	dpInsights := &core_mesh.DataplaneInsightResourceList{}
	if err := r.rm.List(context.Background(), dpInsights, store.ListByMesh(mesh)); err != nil {
		return err
	}

	internalServices := map[string]struct{}{}

	dpOverviews := core_mesh.NewDataplaneOverviews(*dataplanes, *dpInsights)

	for _, dpOverview := range dpOverviews.Items {
		dpInsight := dpOverview.Spec.DataplaneInsight
		dpSubscription := dpInsight.GetLastSubscription().(*mesh_proto.DiscoverySubscription)
		kumaDpVersion := getOrDefault(dpSubscription.GetVersion().GetKumaDp().GetVersion())
		envoyVersion := getOrDefault(dpSubscription.GetVersion().GetEnvoy().GetVersion())
		networking := dpOverview.Spec.GetDataplane().GetNetworking()

		ensureVersionExists(kumaDpVersion, insight.DpVersions.KumaDp)
		ensureVersionExists(envoyVersion, insight.DpVersions.Envoy)

		status, _ := dpOverview.GetStatus()

		statByType := insight.GetDataplanesByType().GetStandard()
		if networking.GetGateway() != nil {
			statByType = insight.GetDataplanesByType().GetGateway()
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
			internalServices[inbound.GetService()] = struct{}{}
		}
	}

	externalServices := &core_mesh.ExternalServiceResourceList{}
	if err := r.rm.List(context.Background(), externalServices, store.ListByMesh(mesh)); err != nil {
		return err
	}

	insight.Services = &mesh_proto.MeshInsight_ServiceStat{
		Total:    uint32(len(internalServices) + len(externalServices.Items)),
		Internal: uint32(len(internalServices)),
		External: uint32(len(externalServices.Items)),
	}

	for _, resDesc := range r.registry.ObjectDescriptors(model.HasScope(model.ScopeMesh), model.Not(model.Named(core_mesh.DataplaneType, core_mesh.DataplaneInsightType))) {
		list := resDesc.NewList()

		if err := r.rm.List(context.Background(), list, store.ListByMesh(mesh)); err != nil {
			return err
		}

		if len(list.GetItems()) != 0 {
			insight.Policies[string(resDesc.Name)] = &mesh_proto.MeshInsight_PolicyStat{
				Total: uint32(len(list.GetItems())),
			}
		}
	}

	key := MeshInsightKey(mesh)
	info, _ := r.getInfo(key)
	if proto.Equal(info.resource, insight) {
		log.V(1).Info("no need to update MeshInsight because the resource is the same")
		info.lastSync = core.Now()
		r.setInfo(key, info)
		return nil
	}

	err := manager.Upsert(ctx, r.rm, model.ResourceKey{Mesh: model.NoMesh, Name: mesh}, core_mesh.NewMeshInsightResource(), func(resource model.Resource) error {
		return resource.SetSpec(insight)
	})
	if err != nil {
		if manager.IsMeshNotFound(err) {
			log.V(1).Info("MeshInsight is not updated because mesh no longer exist. This can happen when Mesh is being deleted.")
			// handle the situation when the mesh is deleted and then all the resources connected with the Mesh all deleted.
			// Mesh no longer exist so we cannot upsert the insight for it.
			return nil
		}
		if store.IsResourceConflict(err) {
			log.V(1).Info("MeshInsight was updated in other place. Retrying")
			return nil
		}
		return err
	}
	log.V(1).Info("MeshInsight updated")
	r.setInfo(key, syncInfo{
		resource: insight,
		lastSync: core.Now(),
	})
	return nil
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

func (r *resyncer) needResyncServiceInsight(mesh string) bool {
	info, exist := r.getInfo(ServiceInsightKey(mesh))
	return !exist || core.Now().Sub(info.lastSync) > r.minResyncTimeout
}

func (r *resyncer) needResyncMeshInsight(mesh string) bool {
	info, exist := r.getInfo(MeshInsightKey(mesh))
	return !exist || core.Now().Sub(info.lastSync) > r.minResyncTimeout
}

func (r *resyncer) getInfo(key model.ResourceKey) (syncInfo, bool) {
	r.infosMux.RLock()
	defer r.infosMux.RUnlock()
	info, ok := r.infos[key]
	return info, ok
}

func (r *resyncer) setInfo(key model.ResourceKey, info syncInfo) {
	r.infosMux.Lock()
	r.infos[key] = info
	r.infosMux.Unlock()
}

func (r *resyncer) clearInfos() {
	r.infosMux.Lock()
	r.infos = map[model.ResourceKey]syncInfo{}
	r.infosMux.Unlock()
}

func (r *resyncer) NeedLeaderElection() bool {
	return true
}
