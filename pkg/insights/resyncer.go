package insights

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-kit/kit/ratelimit"
	"github.com/golang/protobuf/ptypes"
	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	"github.com/kumahq/kuma/pkg/events"
	"github.com/kumahq/kuma/pkg/util/proto"
)

var (
	log = core.Log.WithName("mesh-insight-resyncer")
)

func ServiceInsightName(mesh string) string {
	return fmt.Sprintf("all-services-%s", mesh)
}

type Config struct {
	ResourceManager    manager.ResourceManager
	EventReaderFactory events.ListenerFactory
	MinResyncTimeout   time.Duration
	MaxResyncTimeout   time.Duration
	Tick               func(d time.Duration) <-chan time.Time
	RateLimiterFactory func() ratelimit.Allower
}

type resyncer struct {
	rm                 manager.ResourceManager
	eventFactory       events.ListenerFactory
	minResyncTimeout   time.Duration
	maxResyncTimeout   time.Duration
	tick               func(d time.Duration) <-chan time.Time
	rateLimiterFactory func() ratelimit.Allower
	meshInsightMux     sync.Mutex
	serviceInsightMux  sync.Mutex
	rateLimiters       map[string]ratelimit.Allower
}

// NewResyncer creates a new Component that periodically updates insights
// for various policies (right now only for Mesh).
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
		rateLimiters:       map[string]ratelimit.Allower{},
	}

	r.tick = config.Tick
	if config.Tick == nil {
		r.tick = time.Tick
	}

	return r
}

func (r *resyncer) Start(stop <-chan struct{}) error {
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
		if err != nil {
			return err
		}
		resourceChanged, ok := event.(events.ResourceChangedEvent)
		if !ok {
			continue
		}
		if resourceChanged.Type == core_mesh.MeshType && resourceChanged.Operation == events.Delete {
			r.deleteRateLimiter(resourceChanged.Key.Name)
		}
		if resourceChanged.Type == core_mesh.DataplaneType || resourceChanged.Type == core_mesh.DataplaneInsightType {
			if err := r.createOrUpdateServiceInsight(resourceChanged.Key.Mesh); err != nil {
				log.Error(err, "unable to resync ServiceInsight", "mesh", resourceChanged.Key.Mesh)
			}
		}
		if !meshScoped(resourceChanged.Type) {
			continue
		}
		if resourceChanged.Operation == events.Update && resourceChanged.Type != core_mesh.DataplaneInsightType {
			// 'Update' events doesn't affect MeshInsight expect for DataplaneInsight,
			// because that's how we find online/offline Dataplane's status
			continue
		}
		if !r.getRateLimiter(resourceChanged.Key.Mesh).Allow() {
			continue
		}
		if err := r.createOrUpdateMeshInsight(resourceChanged.Key.Mesh); err != nil {
			log.Error(err, "unable to resync MeshInsight", "mesh", resourceChanged.Key.Mesh)
			continue
		}
	}
}

func (r *resyncer) getRateLimiter(mesh string) ratelimit.Allower {
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

func meshScoped(t model.ResourceType) bool {
	if obj, err := registry.Global().NewObject(t); err != nil || obj.Scope() != model.ScopeMesh {
		return false
	}
	return true
}

func (r *resyncer) createOrUpdateServiceInsights() error {
	meshes := &core_mesh.MeshResourceList{}
	if err := r.rm.List(context.Background(), meshes); err != nil {
		return err
	}
	for _, mesh := range meshes.Items {
		if need, err := r.needResyncServiceInsight(mesh.GetMeta().GetName()); err != nil || !need {
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

func (r *resyncer) createOrUpdateServiceInsight(mesh string) error {
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
		status, _ := dpOverview.GetStatus()

		for _, inbound := range dpOverview.Spec.Dataplane.Networking.Inbound {
			svcName := inbound.GetService()

			if _, ok := insight.Services[svcName]; !ok {
				insight.Services[svcName] = &mesh_proto.ServiceInsight_Service{
					Dataplanes: &mesh_proto.ServiceInsight_Service_DataplaneStat{},
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
				if inbound.Health != nil && !inbound.Health.Ready {
					dataplanes.Offline++
				} else {
					dataplanes.Online++
				}
			}
		}
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

	err := manager.Upsert(r.rm, model.ResourceKey{Mesh: mesh, Name: ServiceInsightName(mesh)}, core_mesh.NewServiceInsightResource(), func(resource model.Resource) {
		insight.LastSync = proto.MustTimestampProto(core.Now())
		_ = resource.SetSpec(insight)
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
	return nil
}

func (r *resyncer) createOrUpdateMeshInsights() error {
	meshes := &core_mesh.MeshResourceList{}
	if err := r.rm.List(context.Background(), meshes); err != nil {
		return err
	}
	for _, mesh := range meshes.Items {
		if need, err := r.needResyncMeshInsight(mesh.GetMeta().GetName()); err != nil || !need {
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
	r.meshInsightMux.Lock()
	defer r.meshInsightMux.Unlock()

	insight := &mesh_proto.MeshInsight{
		Dataplanes: &mesh_proto.MeshInsight_DataplaneStat{},
		Policies:   map[string]*mesh_proto.MeshInsight_PolicyStat{},
		DpVersions: &mesh_proto.MeshInsight_DpVersions{
			KumaDp: map[string]*mesh_proto.MeshInsight_DataplaneStat{},
			Envoy:  map[string]*mesh_proto.MeshInsight_DataplaneStat{},
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

	dpOverviews := core_mesh.NewDataplaneOverviews(*dataplanes, *dpInsights)

	for _, dpOverview := range dpOverviews.Items {
		dpInsight := dpOverview.Spec.DataplaneInsight
		dpSubscription, _ := dpInsight.GetLatestSubscription()
		kumaDpVersion := getOrDefault(dpSubscription.GetVersion().GetKumaDp().GetVersion())
		envoyVersion := getOrDefault(dpSubscription.GetVersion().GetEnvoy().GetVersion())
		ensureVersionExists(kumaDpVersion, insight.DpVersions.KumaDp)
		ensureVersionExists(envoyVersion, insight.DpVersions.Envoy)

		status, _ := dpOverview.GetStatus()

		switch status {
		case core_mesh.Online:
			insight.Dataplanes.Online++
			insight.DpVersions.KumaDp[kumaDpVersion].Online++
			insight.DpVersions.Envoy[envoyVersion].Online++
		case core_mesh.PartiallyDegraded:
			insight.Dataplanes.PartiallyDegraded++
			insight.DpVersions.KumaDp[kumaDpVersion].PartiallyDegraded++
			insight.DpVersions.Envoy[envoyVersion].PartiallyDegraded++
		case core_mesh.Offline:
			insight.Dataplanes.Offline++
			insight.DpVersions.KumaDp[kumaDpVersion].Offline++
			insight.DpVersions.Envoy[envoyVersion].Offline++
		}

		updateTotal(kumaDpVersion, insight.DpVersions.KumaDp)
		updateTotal(envoyVersion, insight.DpVersions.Envoy)
	}

	for _, resType := range registry.Global().ListTypes() {
		if !meshScoped(resType) || resType == core_mesh.DataplaneType || resType == core_mesh.DataplaneInsightType {
			continue
		}
		list, err := registry.Global().NewList(resType)
		if err != nil {
			return err
		}
		if err := r.rm.List(context.Background(), list, store.ListByMesh(mesh)); err != nil {
			return err
		}

		if len(list.GetItems()) != 0 {
			insight.Policies[string(resType)] = &mesh_proto.MeshInsight_PolicyStat{
				Total: uint32(len(list.GetItems())),
			}
		}
	}

	err := manager.Upsert(r.rm, model.ResourceKey{Mesh: model.NoMesh, Name: mesh}, core_mesh.NewMeshInsightResource(), func(resource model.Resource) {
		insight.LastSync = proto.MustTimestampProto(core.Now())
		_ = resource.SetSpec(insight)
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
	return nil
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

func (r *resyncer) needResyncServiceInsight(mesh string) (bool, error) {
	serviceInsight := core_mesh.NewServiceInsightResource()
	if err := r.rm.Get(context.Background(), serviceInsight, store.GetByKey(ServiceInsightName(mesh), mesh)); err != nil {
		if !store.IsResourceNotFound(err) {
			return false, errors.Wrap(err, "failed to get ServiceInsight")
		}
		return true, nil
	}
	lastSync, err := ptypes.Timestamp(serviceInsight.Spec.LastSync)
	if err != nil {
		return false, errors.Wrapf(err, "lastSync has wrong value: %s", serviceInsight.Spec.LastSync)
	}
	if core.Now().Sub(lastSync) < r.minResyncTimeout {
		return false, nil
	}
	return true, nil
}

func (r *resyncer) needResyncMeshInsight(mesh string) (bool, error) {
	meshInsight := core_mesh.NewMeshInsightResource()
	if err := r.rm.Get(context.Background(), meshInsight, store.GetByKey(mesh, model.NoMesh)); err != nil {
		if !store.IsResourceNotFound(err) {
			return false, errors.Wrap(err, "failed to get MeshInsight")
		}
		return true, nil
	}
	lastSync, err := ptypes.Timestamp(meshInsight.Spec.LastSync)
	if err != nil {
		return false, errors.Wrapf(err, "lastSync has wrong value: %s", meshInsight.Spec.LastSync)
	}
	if core.Now().Sub(lastSync) < r.minResyncTimeout {
		return false, nil
	}
	return true, nil
}

func (r *resyncer) NeedLeaderElection() bool {
	return true
}
