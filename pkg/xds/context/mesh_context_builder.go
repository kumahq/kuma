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
	xds_topology "github.com/kumahq/kuma/pkg/xds/topology"
)

var logger = core.Log.WithName("xds").WithName("context")

type meshContextBuilder struct {
	rm              manager.ReadOnlyResourceManager
	types           []core_model.ResourceType
	ipFunc          lookup.LookupIPFunc
	zone            string
	vipsPersistence *vips.Persistence
	topLevelDomain  string
}

type MeshContextBuilder interface {
	Build(ctx context.Context, meshName string) (MeshContext, error)

	// BuildIfChanged builds MeshContext only if hash of MeshContext is different
	// If hash is the same, then the function returns (nil, nil)
	// Hash returned in MeshContext can never be empty
	BuildIfChanged(ctx context.Context, meshName string, hash string) (*MeshContext, error)
}

func NewMeshContextBuilder(
	rm manager.ReadOnlyResourceManager,
	types []core_model.ResourceType, // types that should be taken into account when MeshContext is built.
	ipFunc lookup.LookupIPFunc,
	zone string,
	vipsPersistence *vips.Persistence,
	topLevelDomain string,
) MeshContextBuilder {
	return &meshContextBuilder{
		rm:              rm,
		types:           types,
		ipFunc:          ipFunc,
		zone:            zone,
		vipsPersistence: vipsPersistence,
		topLevelDomain:  topLevelDomain,
	}
}

func (m *meshContextBuilder) Build(ctx context.Context, meshName string) (MeshContext, error) {
	meshCtx, err := m.BuildIfChanged(ctx, meshName, "")
	if err != nil {
		return MeshContext{}, err
	}
	return *meshCtx, nil
}

func (m *meshContextBuilder) BuildIfChanged(ctx context.Context, meshName string, hash string) (*MeshContext, error) {
	mesh := core_mesh.NewMeshResource()
	if err := m.rm.Get(ctx, mesh, core_store.GetByKey(meshName, core_model.NoMesh)); err != nil {
		return nil, err
	}

	resources, err := m.fetchResources(ctx, mesh)
	if err != nil {
		return nil, err
	}
	m.resolveAddresses(resources)

	newHash := m.hash(mesh, resources)
	if newHash == hash {
		return nil, nil
	}

	dataplanesByName := map[string]*core_mesh.DataplaneResource{}

	dataplanes := resources.Dataplanes().Items

	for _, dp := range dataplanes {
		dataplanesByName[dp.Meta.GetName()] = dp
	}

	virtualOutboundView, err := m.vipsPersistence.GetByMesh(mesh.GetMeta().GetName())
	if err != nil {
		return nil, err
	}
	// resolve all the domains
	domains, outbounds := xds_topology.VIPOutbounds(virtualOutboundView, m.topLevelDomain)

	zoneIngresses := resources.ZoneIngresses().Items
	zoneEgresses := resources.ZoneEgresses().Items
	externalServices := resources.ExternalServices().Items
	endpointMap := xds_topology.BuildEdsEndpointMap(mesh, m.zone, dataplanes, zoneIngresses, zoneEgresses, externalServices)

	return &MeshContext{
		Hash:                newHash,
		Resource:            mesh,
		Resources:           resources,
		DataplanesByName:    dataplanesByName,
		EndpointMap:         endpointMap,
		VIPDomains:          domains,
		VIPOutbounds:        outbounds,
		ServiceTLSReadiness: m.resolveTLSReadiness(mesh, resources.ServiceInsights()),
	}, nil
}

func (m *meshContextBuilder) fetchResources(ctx context.Context, mesh *core_mesh.MeshResource) (Resources, error) {
	resources := Resources{}

	for _, typ := range m.types {
		switch typ {
		case core_mesh.ZoneIngressType:
			zoneIngresses := &core_mesh.ZoneIngressResourceList{}
			if err := m.rm.List(ctx, zoneIngresses); err != nil {
				return nil, err
			}
			resources[typ] = zoneIngresses
		case core_mesh.ZoneEgressType:
			zoneEgresses := &core_mesh.ZoneEgressResourceList{}
			if err := m.rm.List(ctx, zoneEgresses); err != nil {
				return nil, err
			}
			resources[typ] = zoneEgresses
		case system.ConfigType:
			configs := &system.ConfigResourceList{}
			var items []*system.ConfigResource
			if err := m.rm.List(ctx, configs); err != nil {
				return nil, err
			}
			for _, config := range configs.Items {
				if configInHash(config.Meta.GetName(), mesh.Meta.GetName()) {
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
			if err := m.rm.List(ctx, insights, core_store.ListByMesh(mesh.Meta.GetName())); err != nil {
				return nil, err
			}

			resources[typ] = insights
		default:
			rlist, err := registry.Global().NewList(typ)
			if err != nil {
				return nil, err
			}
			if err := m.rm.List(ctx, rlist, core_store.ListByMesh(mesh.Meta.GetName())); err != nil {
				return nil, err
			}
			resources[typ] = rlist
		}
	}
	return resources, nil
}

func (m *meshContextBuilder) resolveAddresses(resources Resources) {
	zoneIngresses := xds_topology.ResolveZoneIngressAddresses(logger, m.ipFunc, resources.ZoneIngresses().Items)
	resources.ZoneIngresses().Items = zoneIngresses

	dataplanes := xds_topology.ResolveAddresses(logger, m.ipFunc, resources.Dataplanes().Items)
	resources.Dataplanes().Items = dataplanes
}

func (m *meshContextBuilder) hash(mesh *core_mesh.MeshResource, resources Resources) string {
	allResources := []core_model.Resource{
		mesh,
	}
	for _, rl := range resources {
		allResources = append(allResources, rl.GetItems()...)
	}
	return sha256.Hash(m.hashResources(allResources...))
}

func (m *meshContextBuilder) hashResources(rs ...core_model.Resource) string {
	hashes := []string{}
	for _, r := range rs {
		hashes = append(hashes, m.hashResource(r))
	}
	sort.Strings(hashes)
	return strings.Join(hashes, ",")
}

func (m *meshContextBuilder) hashResource(r core_model.Resource) string {
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
				v.Spec.GetNetworking().GetAddress(),
				v.Spec.GetNetworking().GetAdvertisedAddress(),
			}, ":")
	case *core_mesh.ZoneIngressResource:
		return strings.Join(
			[]string{string(v.Descriptor().Name),
				v.GetMeta().GetMesh(),
				v.GetMeta().GetName(),
				v.GetMeta().GetVersion(),
				v.Spec.GetNetworking().GetAddress(),
				v.Spec.GetNetworking().GetAdvertisedAddress(),
			}, ":")
	case *core_mesh.ZoneEgressResource:
		return strings.Join(
			[]string{string(v.Descriptor().Name),
				v.GetMeta().GetMesh(),
				v.GetMeta().GetName(),
				v.GetMeta().GetVersion(),
				v.Spec.GetNetworking().GetAddress(),
			}, ":")
	default:
		return strings.Join(
			[]string{string(v.Descriptor().Name),
				v.GetMeta().GetMesh(),
				v.GetMeta().GetName(),
				v.GetMeta().GetVersion()}, ":")
	}
}

func configInHash(configName string, meshName string) bool {
	return configName == vips.ConfigKey(meshName)
}

func (m *meshContextBuilder) resolveTLSReadiness(mesh *core_mesh.MeshResource, serviceInsights *core_mesh.ServiceInsightResourceList) map[string]bool {
	tlsReady := map[string]bool{}

	backend := mesh.GetEnabledCertificateAuthorityBackend()
	// TLS readiness is irrelevant unless we are using PERMISSIVE TLS, so skip
	// checking ServiceInsights if we aren't.
	if backend == nil || backend.Mode != mesh_proto.CertificateAuthorityBackend_PERMISSIVE {
		return tlsReady
	}

	if len(serviceInsights.Items) == 0 {
		// Nothing about the TLS readiness has been reported yet
		logger.Info("could not determine service TLS readiness, ServiceInsight is not yet present")
		return tlsReady
	}

	for svc, insight := range serviceInsights.Items[0].Spec.GetServices() {
		tlsReady[svc] = insight.IssuedBackends[backend.Name] == insight.Dataplanes.Total
	}
	return tlsReady
}
