package insights

import (
	"context"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/pkg/errors"
	"golang.org/x/time/rate"

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
	ResourceManager  manager.ResourceManager
	EventReader      events.Reader
	MinResyncTimeout time.Duration
	MaxResyncTimeout time.Duration
	Tick             func(d time.Duration) <-chan time.Time
}

type resyncer struct {
	rm               manager.ResourceManager
	events           events.Reader
	minResyncTimeout time.Duration
	maxResyncTimeout time.Duration
	tick             func(d time.Duration) <-chan time.Time
}

func NewResyncer(config *Config) component.Component {
	r := &resyncer{
		minResyncTimeout: config.MinResyncTimeout,
		maxResyncTimeout: config.MaxResyncTimeout,
		events:           config.EventReader,
		rm:               config.ResourceManager,
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

	limiter := rate.NewLimiter(rate.Every(p.minResyncTimeout), 50)
	for {
		event, err := p.events.Recv(stop)
		if err != nil {
			return err
		}
		if !meshScoped(event.Type) || event.Type == core_mesh.DataplaneType {
			continue
		}
		if event.Operation == events.Update && event.Type != core_mesh.DataplaneInsightType {
			// 'Update' events doesn't affect MeshInsight expect for DataplaneInsight,
			// because that's how we find online/offline Dataplane's status
			continue
		}
		if !limiter.Allow() {
			continue
		}
		if err := p.resyncMesh(event.Key.Mesh); err != nil {
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
		if !meshScoped(resType) || resType == core_mesh.DataplaneType {
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
		case core_mesh.DataplaneInsightType:
			insight.Dataplanes.Total = uint32(len(list.GetItems()))
			for _, dpInsight := range list.(*core_mesh.DataplaneInsightResourceList).Items {
				if dpInsight.Spec.IsOnline() {
					insight.Dataplanes.Online++
				} else {
					insight.Dataplanes.Offline++
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
