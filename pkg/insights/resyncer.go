package insights

import (
	"context"
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

type Config struct {
	ResourceManager    manager.ResourceManager
	EventReaderFactory events.ListenerFactory
	MinResyncTimeout   time.Duration
	MaxResyncTimeout   time.Duration
	Tick               func(d time.Duration) <-chan time.Time
	RateLimiter        ratelimit.Allower
}

type resyncer struct {
	rm                manager.ResourceManager
	eventFactory      events.ListenerFactory
	minResyncTimeout  time.Duration
	maxResyncTimeout  time.Duration
	tick              func(d time.Duration) <-chan time.Time
	rateLimiter       ratelimit.Allower
	meshInsightMux    sync.Mutex
	serviceInsightMux sync.Mutex
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
		minResyncTimeout: config.MinResyncTimeout,
		maxResyncTimeout: config.MaxResyncTimeout,
		eventFactory:     config.EventReaderFactory,
		rm:               config.ResourceManager,
		rateLimiter:      config.RateLimiter,
	}

	r.tick = config.Tick
	if config.Tick == nil {
		r.tick = time.Tick
	}

	return r
}

func (p *resyncer) Start(stop <-chan struct{}) error {
	go func(stop <-chan struct{}) {
		ticker := p.tick(p.maxResyncTimeout - p.minResyncTimeout)
		for {
			select {
			case <-ticker:
				if err := p.createOrUpdateMeshInsights(); err != nil {
					log.Error(err, "unable to resync MeshInsight")
				}
				if err := p.createOrUpdateServiceInsights(); err != nil {
					log.Error(err, "unable to resync ServiceInsight")
				}
			case <-stop:
				log.Info("stop")
				return
			}
		}
	}(stop)

	eventReader := p.eventFactory.New()
	for {
		event, err := eventReader.Recv(stop)
		if err != nil {
			return err
		}
		resourceChanged, ok := event.(events.ResourceChangedEvent)
		if !ok {
			continue
		}
		if resourceChanged.Type == core_mesh.DataplaneType || resourceChanged.Type == core_mesh.DataplaneInsightType {
			if err := p.createOrUpdateServiceInsight(resourceChanged.Key.Mesh); err != nil {
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
		if !p.rateLimiter.Allow() {
			continue
		}
		if err := p.createOrUpdateMeshInsight(resourceChanged.Key.Mesh); err != nil {
			log.Error(err, "unable to resync MeshInsight", "mesh", resourceChanged.Key.Mesh)
			continue
		}
	}
}

func meshScoped(t model.ResourceType) bool {
	if obj, err := registry.Global().NewObject(t); err != nil || obj.Scope() != model.ScopeMesh {
		return false
	}
	return true
}

func (p *resyncer) createOrUpdateServiceInsights() error {
	p.serviceInsightMux.Lock()
	defer p.serviceInsightMux.Unlock()

	meshes := &core_mesh.MeshResourceList{}
	if err := p.rm.List(context.Background(), meshes); err != nil {
		return err
	}
	for _, mesh := range meshes.Items {
		if need, err := p.needResync(mesh.GetMeta().GetName()); err != nil || !need {
			continue
		}
		err := p.createOrUpdateMeshInsight(mesh.GetMeta().GetName())
		if err != nil {
			log.Error(err, "unable to resync resources", "mesh", mesh.GetMeta().GetName())
			continue
		}
	}
	return nil
}

func (p *resyncer) createOrUpdateServiceInsight(mesh string) error {
	insight := &mesh_proto.ServiceInsight{
		Services: map[string]*mesh_proto.ServiceInsight_DataplaneStat{},
	}
	dpList := &core_mesh.DataplaneResourceList{}
	if err := p.rm.List(context.Background(), dpList, store.ListByMesh(mesh)); err != nil {
		return err
	}
	dpMap := map[model.ResourceKey][]string{}
	for _, dp := range dpList.Items {
		dpKey := model.MetaToResourceKey(dp.Meta)
		for _, inbound := range dp.Spec.Networking.Inbound {
			svc := inbound.GetService()
			dpMap[dpKey] = append(dpMap[dpKey], svc)
			if _, ok := insight.Services[svc]; !ok {
				insight.Services[svc] = &mesh_proto.ServiceInsight_DataplaneStat{}
			}
			insight.Services[svc].Total++
		}
	}
	dpInsights := &core_mesh.DataplaneInsightResourceList{}
	if err := p.rm.List(context.Background(), dpInsights, store.ListByMesh(mesh)); err != nil {
		return err
	}
	for _, dpInsight := range dpInsights.Items {
		if dpInsight.Spec.IsOnline() {
			for _, svc := range dpMap[model.MetaToResourceKey(dpInsight.Meta)] {
				insight.Services[svc].Online++
			}
		}
	}
	for _, stat := range insight.Services {
		stat.Offline = stat.Total - stat.Online
	}
	err := manager.Upsert(p.rm, model.ResourceKey{Mesh: model.NoMesh, Name: mesh}, &core_mesh.ServiceInsightResource{}, func(resource model.Resource) {
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
		return err
	}
	return nil
}

func (p *resyncer) createOrUpdateMeshInsights() error {
	p.meshInsightMux.Lock()
	defer p.meshInsightMux.Unlock()

	meshes := &core_mesh.MeshResourceList{}
	if err := p.rm.List(context.Background(), meshes); err != nil {
		return err
	}
	for _, mesh := range meshes.Items {
		if need, err := p.needResync(mesh.GetMeta().GetName()); err != nil || !need {
			continue
		}
		err := p.createOrUpdateMeshInsight(mesh.GetMeta().GetName())
		if err != nil {
			log.Error(err, "unable to resync resources", "mesh", mesh.GetMeta().GetName())
			continue
		}
	}
	return nil
}

func (p *resyncer) createOrUpdateMeshInsight(mesh string) error {
	insight := &mesh_proto.MeshInsight{
		Dataplanes: &mesh_proto.MeshInsight_DataplaneStat{},
		Policies:   map[string]*mesh_proto.MeshInsight_PolicyStat{},
	}
	for _, resType := range registry.Global().ListTypes() {
		if !meshScoped(resType) {
			continue
		}
		list, err := registry.Global().NewList(resType)
		if err != nil {
			return err
		}
		if err := p.rm.List(context.Background(), list, store.ListByMesh(mesh)); err != nil {
			return err
		}
		switch resType {
		case core_mesh.DataplaneType:
			insight.Dataplanes.Total = uint32(len(list.GetItems()))
		case core_mesh.DataplaneInsightType:
			for _, dpInsight := range list.(*core_mesh.DataplaneInsightResourceList).Items {
				if dpInsight.Spec.IsOnline() {
					insight.Dataplanes.Online++
				}
			}
		default:
			if len(list.GetItems()) != 0 {
				insight.Policies[string(resType)] = &mesh_proto.MeshInsight_PolicyStat{
					Total: uint32(len(list.GetItems())),
				}
			}
		}
	}
	insight.Dataplanes.Offline = insight.Dataplanes.Total - insight.Dataplanes.Online

	err := manager.Upsert(p.rm, model.ResourceKey{Mesh: model.NoMesh, Name: mesh}, &core_mesh.MeshInsightResource{}, func(resource model.Resource) {
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
		return err
	}
	return nil
}

func (p *resyncer) needResync(mesh string) (bool, error) {
	meshInsight := &core_mesh.MeshInsightResource{}
	if err := p.rm.Get(context.Background(), meshInsight, store.GetByKey(mesh, model.NoMesh)); err != nil {
		if !store.IsResourceNotFound(err) {
			return false, errors.Wrap(err, "failed to get MeshInsight")
		}
		return true, nil
	}
	lastSync, err := ptypes.Timestamp(meshInsight.Spec.LastSync)
	if err != nil {
		return false, errors.Wrapf(err, "lastSync has wrong value: %s", meshInsight.Spec.LastSync)
	}
	if core.Now().Sub(lastSync) < p.minResyncTimeout {
		return false, nil
	}
	return true, nil
}

func (p *resyncer) NeedLeaderElection() bool {
	return true
}
