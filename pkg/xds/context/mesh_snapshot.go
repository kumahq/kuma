package context

import (
	"context"
	"sort"
	"strings"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/dns/lookup"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/dns/vips"
	"github.com/kumahq/kuma/pkg/xds/cache/sha256"
)

var meshCacheLog = core.Log.WithName("mesh-cache")

// MeshSnapshot represents all resources that belong to Mesh and allows to calculate hash.
// Calculating and comparing hashes is much faster than call 'equal' for xDS resources. So
// meshSnapshot reduces costs on reconciling Envoy config when resources in the store are
// not changing
type MeshSnapshot struct {
	mesh      *core_mesh.MeshResource
	resources map[core_model.ResourceType]core_model.ResourceList
	Hash      string
}

type MeshSnapshotBuilder interface {
	Build(ctx context.Context, meshName string) (*MeshSnapshot, error)
}

func NewMeshSnapshotBuilder(rm manager.ReadOnlyResourceManager, types []core_model.ResourceType, ipFunc lookup.LookupIPFunc) MeshSnapshotBuilder {
	return &meshSnapshotBuilder{
		rm:     rm,
		types:  types,
		ipFunc: ipFunc,
	}
}

type meshSnapshotBuilder struct {
	rm     manager.ReadOnlyResourceManager
	types  []core_model.ResourceType
	ipFunc lookup.LookupIPFunc
}

var _ MeshSnapshotBuilder = &meshSnapshotBuilder{}

func (m *meshSnapshotBuilder) Build(ctx context.Context, meshName string) (*MeshSnapshot, error) {
	resources := map[core_model.ResourceType]core_model.ResourceList{}

	mesh := core_mesh.NewMeshResource()
	if err := m.rm.Get(ctx, mesh, core_store.GetByKey(meshName, core_model.NoMesh)); err != nil {
		return nil, err
	}

	for _, typ := range m.types {
		switch typ {
		case core_mesh.ZoneIngressType:
			zoneIngresses := &core_mesh.ZoneIngressResourceList{}
			if err := m.rm.List(ctx, zoneIngresses); err != nil {
				return nil, err
			}
			resources[typ] = zoneIngresses
		case system.ConfigType:
			configs := &system.ConfigResourceList{}
			var items []*system.ConfigResource
			if err := m.rm.List(ctx, configs); err != nil {
				return nil, err
			}
			for _, config := range configs.Items {
				if configInHash(config.Meta.GetName(), meshName) {
					items = append(items, config)
				}
			}
			configs.Items = items
			resources[typ] = configs
		case core_mesh.ServiceInsightType:
			// ServiceInsights in XDS generation are only used to check whether the destination is ready to receive mTLS traffic.
			// This information is only useful when mTLS is enabled with PERMISSIVE mode.
			// Not including this into mesh hash for other cases saves us unnecessary XDS config generations.
			if backend := mesh.GetEnabledCertificateAuthorityBackend(); backend == nil || backend.Mode == mesh_proto.CertificateAuthorityBackend_STRICT {
				break
			}

			insights := &core_mesh.ServiceInsightResourceList{}
			if err := m.rm.List(ctx, insights, core_store.ListByMesh(meshName)); err != nil {
				return nil, err
			}

			resources[typ] = insights
		default:
			rlist, err := registry.Global().NewList(typ)
			if err != nil {
				return nil, err
			}
			if err := m.rm.List(ctx, rlist, core_store.ListByMesh(meshName)); err != nil {
				return nil, err
			}
			resources[typ] = rlist
		}
	}

	return &MeshSnapshot{
		mesh:      mesh,
		resources: resources,
		Hash:      m.hash(mesh, resources),
	}, nil
}

func (m *MeshSnapshot) Resources(resourceType core_model.ResourceType) core_model.ResourceList {
	list, ok := m.resources[resourceType]
	if !ok {
		list, err := registry.Global().NewList(resourceType)
		if err != nil {
			panic(err)
		}
		return list
	}
	return list
}

func configInHash(configName string, meshName string) bool {
	return configName == vips.ConfigKey(meshName)
}

func (m *meshSnapshotBuilder) hash(mesh *core_mesh.MeshResource, resourcesByType map[core_model.ResourceType]core_model.ResourceList) string {
	resources := []core_model.Resource{
		mesh,
	}
	for _, rl := range resourcesByType {
		resources = append(resources, rl.GetItems()...)
	}
	return sha256.Hash(m.hashResources(resources...))
}

func (m *meshSnapshotBuilder) hashResources(rs ...core_model.Resource) string {
	hashes := []string{}
	for _, r := range rs {
		hashes = append(hashes, m.hashResource(r))
	}
	sort.Strings(hashes)
	return strings.Join(hashes, ",")
}

func (m *meshSnapshotBuilder) hashResource(r core_model.Resource) string {
	switch v := r.(type) {
	// In case of hashing Dataplane we are also adding '.Spec.Networking.Address' and `.Spec.Networking.Ingress.PublicAddress` into hash.
	// The address could be a domain name and right now we resolve it right after fetching
	// of Dataplane resource. Since DNS Records might be updated and address could be changed
	// after resolving. That's why it is important to include address into hash.
	case *core_mesh.DataplaneResource:
		return strings.Join(
			[]string{string(v.Descriptor().Name),
				v.GetMeta().GetMesh(),
				v.GetMeta().GetName(),
				v.GetMeta().GetVersion(),
				m.hashResolvedIPs(v.Spec.GetNetworking().GetAddress()),
			}, ":")
	case *core_mesh.ZoneIngressResource:
		return strings.Join(
			[]string{string(v.Descriptor().Name),
				v.GetMeta().GetMesh(),
				v.GetMeta().GetName(),
				v.GetMeta().GetVersion(),
				m.hashResolvedIPs(v.Spec.GetNetworking().GetAddress()),
				m.hashResolvedIPs(v.Spec.GetNetworking().GetAdvertisedAddress()),
			}, ":")
	default:
		return strings.Join(
			[]string{string(v.Descriptor().Name),
				v.GetMeta().GetMesh(),
				v.GetMeta().GetName(),
				v.GetMeta().GetVersion()}, ":")
	}
}

// We need to hash all the resolved IPs, not only the first one, because hostname can be resolved to two IPs (ex. LoadBalancer on AWS)
// If we were to pick only the first one, DNS returns addresses in different order, so we could constantly get different hashes.
func (m *meshSnapshotBuilder) hashResolvedIPs(address string) string {
	if address == "" {
		return ""
	}
	ips, err := m.ipFunc(address)
	if err != nil {
		meshCacheLog.V(1).Info("could not resolve hostname", "err", err)
		// we can ignore an error and assume that address is not yet resolvable for some reason, once it will be resolvable the hash will change
		return ""
	}
	var addresses []string
	for _, ip := range ips {
		addresses = append(addresses, ip.String())
	}
	sort.Strings(addresses)
	return strings.Join(addresses, ":")
}
