package insights

import (
	"context"
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
	rm               manager.ResourceManager
	eventFactory     events.ListenerFactory
	minResyncTimeout time.Duration
	maxResyncTimeout time.Duration
	tick             func(d time.Duration) <-chan time.Time
	rateLimiter      ratelimit.Allower
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
				if err := p.resync(); err != nil {
					log.Error(err, "unable to resync resources")
					continue
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
		if !meshScoped(resourceChanged.Type) {
			continue
		}
		if resourceChanged.Operation == events.Update && resourceChanged.Type != core_mesh.DataplaneInsightType {
			// 'Update' events doesn't affect MeshInsight expect for DataplaneInsight,
			// because that's how we find online/offline Dataplane's status
			continue
		}
		if resourceChanged.Type == core_mesh.MeshType || resourceChanged.Type == core_mesh.MeshInsightType {
			// when don't count Mesh itself, we coun't Mesh resources, so there is no need to react on this event
			continue
		}
		if !p.rateLimiter.Allow() {
			continue
		}
		if err := p.resyncMesh(resourceChanged.Key.Mesh); err != nil {
			log.Error(err, "unable to resync resources")
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

func (p *resyncer) resync() error {
	meshes := &core_mesh.MeshResourceList{}
	if err := p.rm.List(context.Background(), meshes); err != nil {
		return err
	}
	for _, mesh := range meshes.Items {
		if need, err := p.needResync(mesh.GetMeta().GetName()); err != nil || !need {
			continue
		}
		err := p.resyncMesh(mesh.GetMeta().GetName())
		if err != nil {
			log.Error(err, "unable to resync resources", "mesh", mesh.GetMeta().GetName())
			continue
		}
	}
	return nil
}

func (p *resyncer) resyncMesh(mesh string) error {
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

	if err := manager.Upsert(p.rm, model.ResourceKey{Mesh: model.NoMesh, Name: mesh}, &core_mesh.MeshInsightResource{}, func(resource model.Resource) {
		insight.LastSync = proto.MustTimestampProto(core.Now())
		_ = resource.SetSpec(insight)
	}); err != nil {
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
