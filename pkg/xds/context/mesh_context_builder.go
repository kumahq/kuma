package context

import (
	"context"
	"sort"
	"strings"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/datasource"
	"github.com/kumahq/kuma/pkg/core/dns/lookup"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/dns/vips"
	"github.com/kumahq/kuma/pkg/xds/cache/sha256"
	xds_topology "github.com/kumahq/kuma/pkg/xds/topology"
)

var logger = core.Log.WithName("xds").WithName("context")

type meshContextBuilder struct {
	rm              manager.ReadOnlyResourceManager
	typeSet         map[core_model.ResourceType]struct{}
	ipFunc          lookup.LookupIPFunc
	zone            string
	vipsPersistence *vips.Persistence
	topLevelDomain  string
	vipPort         uint32
}

type MeshContextBuilder interface {
	Build(ctx context.Context, meshName string) (MeshContext, error)

	// BuildIfChanged builds MeshContext only if latestMeshCtx is nil or hash of
	// latestMeshCtx is different.
	// If hash is the same, then the function returns the passed latestMeshCtx.
	// Hash returned in MeshContext can never be empty.
	BuildIfChanged(ctx context.Context, meshName string, latestMeshCtx *MeshContext) (*MeshContext, error)
}

func NewMeshContextBuilder(
	rm manager.ReadOnlyResourceManager,
	types []core_model.ResourceType, // types that should be taken into account when MeshContext is built.
	ipFunc lookup.LookupIPFunc,
	zone string,
	vipsPersistence *vips.Persistence,
	topLevelDomain string,
	vipPort uint32,
) MeshContextBuilder {
	typeSet := map[core_model.ResourceType]struct{}{}
	for _, typ := range types {
		typeSet[typ] = struct{}{}
	}

	return &meshContextBuilder{
		rm:              rm,
		typeSet:         typeSet,
		ipFunc:          ipFunc,
		zone:            zone,
		vipsPersistence: vipsPersistence,
		topLevelDomain:  topLevelDomain,
		vipPort:         vipPort,
	}
}

func (m *meshContextBuilder) Build(ctx context.Context, meshName string) (MeshContext, error) {
	meshCtx, err := m.BuildIfChanged(ctx, meshName, nil)
	if err != nil {
		return MeshContext{}, err
	}
	return *meshCtx, nil
}

func (m *meshContextBuilder) BuildIfChanged(ctx context.Context, meshName string, latestMeshCtx *MeshContext) (*MeshContext, error) {
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
	if latestMeshCtx != nil && newHash == latestMeshCtx.Hash {
		return latestMeshCtx, nil
	}

	dataplanesByName := map[string]*core_mesh.DataplaneResource{}

	dataplanes := resources.Dataplanes().Items

	for _, dp := range dataplanes {
		dataplanesByName[dp.Meta.GetName()] = dp
	}

	virtualOutboundView, err := m.vipsPersistence.GetByMesh(ctx, mesh.GetMeta().GetName())
	if err != nil {
		return nil, err
	}
	// resolve all the domains
	domains, outbounds := xds_topology.VIPOutbounds(virtualOutboundView, m.topLevelDomain, m.vipPort)

	zoneIngresses := resources.ZoneIngresses().Items
	zoneEgresses := resources.ZoneEgresses().Items
	externalServices := resources.ExternalServices().Items
	endpointMap := xds_topology.BuildEdsEndpointMap(mesh, m.zone, dataplanes, zoneIngresses, zoneEgresses, externalServices)

	crossMeshEndpointMap := map[string]xds.EndpointMap{}
	for otherMeshName, gateways := range resources.gatewaysAndDataplanesForMesh(mesh) {
		crossMeshEndpointMap[otherMeshName] = xds_topology.BuildCrossMeshEndpointMap(
			mesh,
			gateways.Mesh,
			m.zone,
			gateways.Gateways,
			gateways.Dataplanes,
			zoneIngresses,
			zoneEgresses,
		)
	}

	return &MeshContext{
		Hash:                newHash,
		Resource:            mesh,
		Resources:           resources,
		DataplanesByName:    dataplanesByName,
		EndpointMap:         endpointMap,
		CrossMeshEndpoints:  crossMeshEndpointMap,
		VIPDomains:          domains,
		VIPOutbounds:        outbounds,
		ServiceTLSReadiness: m.resolveTLSReadiness(mesh, resources.ServiceInsights()),
		DataSourceLoader:    datasource.NewStaticLoader(resources.Secrets().Items),
	}, nil
}

func (m *meshContextBuilder) fetchCrossMesh(
	ctx context.Context,
	typ core_model.ResourceType,
	localMesh string,
	otherMeshes []string,
	resources *Resources,
	crossMeshPredicate func(xds.MeshName) bool,
	crossMeshResourcePredicate func(core_model.Resource) bool,
) error {
	local, err := registry.Global().NewList(typ)
	if err != nil {
		return err
	}

	if err := m.rm.List(ctx, local, core_store.ListByMesh(localMesh), core_store.ListOrdered()); err != nil {
		return err
	}

	resources.MeshLocalResources[typ] = local

	for _, otherMesh := range otherMeshes {
		if crossMeshPredicate != nil && !crossMeshPredicate(otherMesh) {
			continue
		}

		allOtherMeshItems, err := registry.Global().NewList(typ)
		if err != nil {
			return err
		}
		if err := m.rm.List(ctx, allOtherMeshItems, core_store.ListByMesh(otherMesh), core_store.ListOrdered()); err != nil {
			return err
		}

		other, found := resources.CrossMeshResources[otherMesh]
		if !found {
			other = map[core_model.ResourceType]core_model.ResourceList{}
		}

		otherMeshItems, err := registry.Global().NewList(typ)
		if err != nil {
			return err
		}
		for _, item := range allOtherMeshItems.GetItems() {
			if crossMeshResourcePredicate != nil && !crossMeshResourcePredicate(item) {
				continue
			}
			if err := otherMeshItems.AddItem(item); err != nil {
				return err
			}
		}

		other[typ] = otherMeshItems
		resources.CrossMeshResources[otherMesh] = other
	}

	return nil
}

func (m *meshContextBuilder) fetchResources(ctx context.Context, mesh *core_mesh.MeshResource) (Resources, error) {
	// fetchResources fetches in stages, first getting all resources that only
	// depend on which mesh is in context. It uses those results to iterate over
	// cross-mesh resources.
	resources := NewResources()
	meshName := mesh.GetMeta().GetName()

	for typ := range m.typeSet {
		switch typ {
		case core_mesh.MeshType:
			meshes := &core_mesh.MeshResourceList{}
			if err := m.rm.List(ctx, meshes, core_store.ListOrdered()); err != nil {
				return Resources{}, err
			}
			otherMeshes := &core_mesh.MeshResourceList{}
			for _, someMesh := range meshes.Items {
				if someMesh.GetMeta().GetName() != mesh.GetMeta().GetName() {
					if err := otherMeshes.AddItem(someMesh); err != nil {
						return Resources{}, err
					}
				}
			}
			resources.MeshLocalResources[typ] = otherMeshes
		case core_mesh.ZoneIngressType:
			zoneIngresses := &core_mesh.ZoneIngressResourceList{}
			if err := m.rm.List(ctx, zoneIngresses, core_store.ListOrdered()); err != nil {
				return Resources{}, err
			}
			resources.MeshLocalResources[typ] = zoneIngresses
		case core_mesh.ZoneEgressType:
			zoneEgresses := &core_mesh.ZoneEgressResourceList{}
			if err := m.rm.List(ctx, zoneEgresses, core_store.ListOrdered()); err != nil {
				return Resources{}, err
			}
			resources.MeshLocalResources[typ] = zoneEgresses
		case system.ConfigType:
			configs := &system.ConfigResourceList{}
			var items []*system.ConfigResource
			if err := m.rm.List(ctx, configs, core_store.ListOrdered()); err != nil {
				return Resources{}, err
			}
			for _, config := range configs.Items {
				if configInHash(config.Meta.GetName(), mesh.Meta.GetName()) {
					items = append(items, config)
				}
			}
			configs.Items = items
			resources.MeshLocalResources[typ] = configs
		case core_mesh.ServiceInsightType:
			// ServiceInsights in XDS generation are only used to check whether the destination is ready to receive mTLS traffic.
			// This information is only useful when mTLS is enabled with PERMISSIVE mode.
			// Not including this into mesh hash for other cases saves us unnecessary XDS config generations.
			if backend := mesh.GetEnabledCertificateAuthorityBackend(); backend == nil || backend.Mode == mesh_proto.CertificateAuthorityBackend_STRICT {
				break
			}

			insights := &core_mesh.ServiceInsightResourceList{}
			if err := m.rm.List(ctx, insights, core_store.ListByMesh(mesh.Meta.GetName()), core_store.ListOrdered()); err != nil {
				return Resources{}, err
			}

			resources.MeshLocalResources[typ] = insights
		default:
			rlist, err := registry.Global().NewList(typ)
			if err != nil {
				return Resources{}, err
			}
			if err := m.rm.List(ctx, rlist, core_store.ListByMesh(mesh.Meta.GetName()), core_store.ListOrdered()); err != nil {
				return Resources{}, err
			}
			resources.MeshLocalResources[typ] = rlist
		}
	}

	var otherMeshNames []string
	for _, mesh := range resources.MeshLocalResources.listOrEmpty(core_mesh.MeshType).GetItems() {
		otherMeshNames = append(otherMeshNames, mesh.GetMeta().GetName())
	}

	if _, ok := m.typeSet[core_mesh.MeshGatewayType]; ok {
		// For all meshes, get all cross mesh gateways
		if err := m.fetchCrossMesh(
			ctx, core_mesh.MeshGatewayType, meshName, otherMeshNames, &resources,
			nil,
			func(gateway core_model.Resource) bool {
				return gateway.(*core_mesh.MeshGatewayResource).Spec.IsCrossMesh()
			},
		); err != nil {
			return Resources{}, err
		}
	}
	if _, ok := m.typeSet[core_mesh.DataplaneType]; ok {
		// for all meshes with a cross mesh gateway, get all builtin gateway
		// dataplanes
		if err := m.fetchCrossMesh(
			ctx, core_mesh.DataplaneType, meshName, otherMeshNames, &resources,
			func(mesh string) bool {
				meshGateways := resources.CrossMeshResources[mesh].listOrEmpty(core_mesh.MeshGatewayType)
				return len(meshGateways.GetItems()) > 0
			},
			func(dataplane core_model.Resource) bool {
				return dataplane.(*core_mesh.DataplaneResource).Spec.IsBuiltinGateway()
			},
		); err != nil {
			return Resources{}, err
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
	for _, rl := range resources.MeshLocalResources {
		allResources = append(allResources, rl.GetItems()...)
	}
	for _, ml := range resources.CrossMeshResources {
		for _, rl := range ml {
			allResources = append(allResources, rl.GetItems()...)
		}
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
<<<<<<< HEAD
		tlsReady[svc] = insight.IssuedBackends[backend.Name] == insight.Dataplanes.Total
=======
		if insight.ServiceType == mesh_proto.ServiceInsight_Service_external {
			tlsReady[svc] = true
		} else {
			tlsReady[svc] = insight.IssuedBackends[backend.Name] == (insight.Dataplanes.Offline + insight.Dataplanes.Online)
		}
>>>>>>> 6e228b7e5 (fix(kuma-cp): handle external services with permissive mtls (#7179))
	}
	return tlsReady
}
