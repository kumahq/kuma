package mesh

import (
	"context"
	"sort"
	"strings"

	"github.com/kumahq/kuma/pkg/core/dns/lookup"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/xds/topology"
)

type meshSnapshot struct {
	mesh      *core_mesh.MeshResource
	resources map[core_model.ResourceType]core_model.ResourceList
}

func GetMeshSnapshot(ctx context.Context, meshName string, rm manager.ReadOnlyResourceManager, types []core_model.ResourceType, ipFunc lookup.LookupIPFunc) (*meshSnapshot, error) {
	snapshot := &meshSnapshot{
		resources: map[core_model.ResourceType]core_model.ResourceList{},
	}

	mesh := &core_mesh.MeshResource{}
	if err := rm.Get(ctx, mesh, core_store.GetByKey(meshName, meshName)); err != nil {
		return nil, err
	}
	snapshot.mesh = mesh

	for _, typ := range types {
		switch typ {
		case core_mesh.DataplaneType:
			dataplanes := &core_mesh.DataplaneResourceList{}
			if err := rm.List(ctx, dataplanes); err != nil {
				return nil, err
			}
			dataplanes.Items = topology.ResolveAddresses(meshCacheLog, ipFunc, dataplanes.Items)
			meshedDpsAndIngresses := &core_mesh.DataplaneResourceList{}
			for _, d := range dataplanes.Items {
				if d.GetMeta().GetMesh() == meshName || d.Spec.IsIngress() {
					_ = meshedDpsAndIngresses.AddItem(d)
				}
			}
			snapshot.resources[typ] = meshedDpsAndIngresses
		default:
			rlist, err := registry.Global().NewList(typ)
			if err != nil {
				return nil, err
			}
			if err := rm.List(ctx, rlist, core_store.ListByMesh(meshName)); err != nil {
				return nil, err
			}
			snapshot.resources[typ] = rlist
		}
	}
	return snapshot, nil
}

func (m *meshSnapshot) hash() string {
	resources := []core_model.Resource{
		m.mesh,
	}
	for _, rl := range m.resources {
		resources = append(resources, rl.GetItems()...)
	}
	return hashResources(resources...)
}

func hashResources(rs ...core_model.Resource) string {
	hashes := []string{}
	for _, r := range rs {
		hashes = append(hashes, hashResource(r))
	}
	sort.Strings(hashes)
	return strings.Join(hashes, ",")
}

func hashResource(r core_model.Resource) string {
	switch v := r.(type) {
	case *core_mesh.DataplaneResource:
		return strings.Join(
			[]string{string(v.GetType()),
				v.GetMeta().GetMesh(),
				v.GetMeta().GetName(),
				v.GetMeta().GetVersion(),
				v.Spec.Networking.Address}, ":")
	default:
		return strings.Join(
			[]string{string(v.GetType()),
				v.GetMeta().GetMesh(),
				v.GetMeta().GetName(),
				v.GetMeta().GetVersion()}, ":")
	}
}
