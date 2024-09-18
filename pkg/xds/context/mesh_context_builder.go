package context

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"hash/fnv"
	"slices"

	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/datasource"
	"github.com/kumahq/kuma/pkg/core/dns/lookup"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	meshextenralservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshexternalservice/api/v1alpha1"
	meshmzservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshmultizoneservice/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/meshservice/api/v1alpha1"
	meshservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshservice/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/xds"
	xds_types "github.com/kumahq/kuma/pkg/core/xds/types"
	"github.com/kumahq/kuma/pkg/dns/vips"
	"github.com/kumahq/kuma/pkg/log"
	"github.com/kumahq/kuma/pkg/util/maps"
	util_protocol "github.com/kumahq/kuma/pkg/util/protocol"
	xds_topology "github.com/kumahq/kuma/pkg/xds/topology"
)

type meshContextBuilder struct {
	rm              manager.ReadOnlyResourceManager
	typeSet         map[core_model.ResourceType]struct{}
	ipFunc          lookup.LookupIPFunc
	zone            string
	vipsPersistence *vips.Persistence
	topLevelDomain  string
	vipPort         uint32
	rsGraphBuilder  ReachableServicesGraphBuilder
}

// MeshContextBuilder
type MeshContextBuilder interface {
	Build(ctx context.Context, meshName string) (MeshContext, error)

	// BuildGlobalContextIfChanged builds GlobalContext only if `latest` is nil or hash is different
	// If hash is the same, the return `latest`
	BuildGlobalContextIfChanged(ctx context.Context, latest *GlobalContext) (*GlobalContext, error)

	// BuildBaseMeshContextIfChanged builds BaseMeshContext only if `latest` is nil or hash is different
	// If hash is the same, the return `latest`
	BuildBaseMeshContextIfChanged(ctx context.Context, meshName string, latest *BaseMeshContext) (*BaseMeshContext, error)

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
	rsGraphBuilder ReachableServicesGraphBuilder,
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
		rsGraphBuilder:  rsGraphBuilder,
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
	globalContext, err := m.BuildGlobalContextIfChanged(ctx, nil)
	if err != nil {
		return nil, err
	}
	baseMeshContext, err := m.BuildBaseMeshContextIfChanged(ctx, meshName, nil)
	if err != nil {
		return nil, err
	}

	var managedTypes []core_model.ResourceType // The types not managed by global nor baseMeshContext
	resources := NewResources()
	// Build all the local entities from the parent contexts
	for resType := range m.typeSet {
		rl, ok := globalContext.ResourceMap[resType]
		if ok { // Exists in global context take it from there
			switch resType {
			case core_mesh.MeshType: // Remove our own mesh from the list
				otherMeshes := rl.NewItem().Descriptor().NewList()
				for _, rentry := range rl.GetItems() {
					if rentry.GetMeta().GetName() != meshName {
						err := otherMeshes.AddItem(rentry)
						if err != nil {
							return nil, err
						}
					}
				}
				otherMeshes.GetPagination().SetTotal(uint32(len(otherMeshes.GetItems())))
				rl = otherMeshes
			}
			resources.MeshLocalResources[resType] = rl
		} else {
			rl, ok = baseMeshContext.ResourceMap[resType]
			if ok { // Exist in the baseMeshContext take it from there
				resources.MeshLocalResources[resType] = rl
			} else { // absent from all parent contexts get it now
				managedTypes = append(managedTypes, resType)
				rl, err = m.fetchResourceList(ctx, resType, baseMeshContext.Mesh, nil)
				if err != nil {
					return nil, errors.Wrap(err, fmt.Sprintf("could not fetch resources of type:%s", resType))
				}
				resources.MeshLocalResources[resType] = rl
			}
		}
	}

	if err := m.decorateWithCrossMeshResources(ctx, resources); err != nil {
		return nil, errors.Wrap(err, "failed to retrieve cross mesh resources")
	}

	// This base64 encoding seems superfluous but keeping it for backward compatibility
	newHash := base64.StdEncoding.EncodeToString(m.hash(globalContext, baseMeshContext, managedTypes, resources))
	if latestMeshCtx != nil && newHash == latestMeshCtx.Hash {
		return latestMeshCtx, nil
	}

	dataplanes := resources.Dataplanes().Items
	dataplanesByName := make(map[string]*core_mesh.DataplaneResource, len(dataplanes))
	for _, dp := range dataplanes {
		dataplanesByName[dp.Meta.GetName()] = dp
	}
	meshServices := resources.MeshServices().Items
	meshServicesByName := make(map[core_model.ResourceIdentifier]*v1alpha1.MeshServiceResource, len(meshServices))
	meshServicesByLabelByValue := LabelsToValuesToResourceIdentifier{}
	for _, ms := range meshServices {
		ri := core_model.NewResourceIdentifier(ms)
		meshServicesByName[ri] = ms
		buildLabelValueToServiceNames(ri, meshServicesByLabelByValue, ms.Meta.GetLabels())
	}

	meshExternalServices := resources.MeshExternalServices().Items
	meshExternalServicesByName := make(map[core_model.ResourceIdentifier]*meshextenralservice_api.MeshExternalServiceResource, len(meshExternalServices))
	meshExternalServicesByLabelByValue := LabelsToValuesToResourceIdentifier{}
	for _, mes := range meshExternalServices {
		ri := core_model.NewResourceIdentifier(mes)
		meshExternalServicesByName[ri] = mes
		buildLabelValueToServiceNames(ri, meshExternalServicesByLabelByValue, mes.Meta.GetLabels())
	}
	meshMultiZoneServices := resources.MeshMultiZoneServices().Items
	meshMultiZoneServicesByName := make(map[core_model.ResourceIdentifier]*meshmzservice_api.MeshMultiZoneServiceResource, len(meshMultiZoneServices))
	meshMultiZoneServiceNameByLabelByValue := LabelsToValuesToResourceIdentifier{}
	for _, svc := range meshMultiZoneServices {
		ri := core_model.NewResourceIdentifier(svc)
		meshMultiZoneServicesByName[ri] = svc
		buildLabelValueToServiceNames(ri, meshMultiZoneServiceNameByLabelByValue, svc.Meta.GetLabels())
	}

	var domains []xds_types.VIPDomains
	var outbounds []*xds_types.Outbound
	if baseMeshContext.Mesh.Spec.MeshServicesEnabled() != mesh_proto.Mesh_MeshServices_Exclusive {
		virtualOutboundView, err := m.vipsPersistence.GetByMesh(ctx, meshName)
		if err != nil {
			return nil, errors.Wrap(err, "could not fetch vips")
		}
		// resolve all the domains
		vipDomains, vipOutbounds := xds_topology.VIPOutbounds(virtualOutboundView, m.topLevelDomain, m.vipPort)
		outbounds = append(outbounds, vipOutbounds...)
		domains = append(domains, vipDomains...)
	}

	outbounds = append(outbounds, xds_topology.Outbounds(meshServices)...)
	outbounds = append(outbounds, xds_topology.Outbounds(meshExternalServices)...)
	outbounds = append(outbounds, xds_topology.Outbounds(meshMultiZoneServices)...)

	domains = append(domains, xds_topology.Domains(meshServices)...)
	domains = append(domains, xds_topology.Domains(meshExternalServices)...)
	domains = append(domains, xds_topology.Domains(meshMultiZoneServices)...)

	loader := datasource.NewStaticLoader(resources.Secrets().Items)

	mesh := baseMeshContext.Mesh
	zoneIngresses := resources.ZoneIngresses().Items
	zoneEgresses := resources.ZoneEgresses().Items
	externalServices := resources.ExternalServices().Items
	endpointMap := xds_topology.BuildEdsEndpointMap(mesh, m.zone, meshServicesByName, meshMultiZoneServices, meshExternalServices, dataplanes, zoneIngresses, zoneEgresses, externalServices)
	esEndpointMap := xds_topology.BuildExternalServicesEndpointMap(ctx, mesh, externalServices, loader, m.zone)
	ingressEndpointMap := xds_topology.BuildIngressEndpointMap(mesh, m.zone, meshServicesByName, meshMultiZoneServices, meshExternalServices, dataplanes, externalServices, resources.Gateways().Items, zoneEgresses)

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
		Hash:                                newHash,
		Resource:                            mesh,
		Resources:                           resources,
		DataplanesByName:                    dataplanesByName,
		MeshServiceByIdentifier:             meshServicesByName,
		MeshServicesByLabelByValue:          meshServicesByLabelByValue,
		MeshExternalServiceByIdentifier:     meshExternalServicesByName,
		MeshExternalServicesByLabelByValue:  meshExternalServicesByLabelByValue,
		MeshMultiZoneServiceByIdentifier:    meshMultiZoneServicesByName,
		MeshMultiZoneServicesByLabelByValue: meshMultiZoneServiceNameByLabelByValue,
		EndpointMap:                         endpointMap,
		ExternalServicesEndpointMap:         esEndpointMap,
		IngressEndpointMap:                  ingressEndpointMap,
		CrossMeshEndpoints:                  crossMeshEndpointMap,
		VIPDomains:                          domains,
		VIPOutbounds:                        outbounds,
		ServicesInformation:                 m.generateServicesInformation(mesh, resources.ServiceInsights(), endpointMap, esEndpointMap),
		DataSourceLoader:                    loader,
		ReachableServicesGraph:              m.rsGraphBuilder(meshName, resources),
	}, nil
}

func (m *meshContextBuilder) BuildGlobalContextIfChanged(ctx context.Context, latest *GlobalContext) (*GlobalContext, error) {
	rmap := ResourceMap{}
	// Only pick the global stuff
	for t := range m.typeSet {
		desc, err := registry.Global().DescriptorFor(t)
		if err != nil {
			return nil, err
		}
		if desc.Scope == core_model.ScopeGlobal && desc.Name != system.ConfigType { // For config we ignore them atm and prefer to rely on more specific filters.
			rmap[t], err = m.fetchResourceList(ctx, t, nil, nil)
			if err != nil {
				return nil, errors.Wrap(err, "failed to build global context")
			}
		}
	}

	newHash := rmap.Hash()
	if latest != nil && bytes.Equal(newHash, latest.hash) {
		return latest, nil
	}
	return &GlobalContext{
		hash:        newHash,
		ResourceMap: rmap,
	}, nil
}

func (m *meshContextBuilder) BuildBaseMeshContextIfChanged(ctx context.Context, meshName string, latest *BaseMeshContext) (*BaseMeshContext, error) {
	mesh := core_mesh.NewMeshResource()
	if err := m.rm.Get(ctx, mesh, core_store.GetByKey(meshName, core_model.NoMesh)); err != nil {
		return nil, err
	}
	rmap := ResourceMap{}
	// Add the mesh to the resourceMap
	rmap[core_mesh.MeshType] = mesh.Descriptor().NewList()
	_ = rmap[core_mesh.MeshType].AddItem(mesh)
	rmap[core_mesh.MeshType].GetPagination().SetTotal(1)
	for t := range m.typeSet {
		desc, err := registry.Global().DescriptorFor(t)
		if err != nil {
			return nil, err
		}
		// Only pick the policies, gateways, external services and the vip config map
		switch {
		case desc.IsPolicy || desc.Name == core_mesh.MeshGatewayType || desc.Name == core_mesh.ExternalServiceType || desc.Name == meshservice_api.MeshServiceType:
			rmap[t], err = m.fetchResourceList(ctx, t, mesh, nil)
		case desc.Name == system.ConfigType:
			rmap[t], err = m.fetchResourceList(ctx, t, mesh, func(rs core_model.Resource) bool {
				return rs.GetMeta().GetName() == vips.ConfigKey(meshName)
			})
		default:
			// DO nothing we're not interested in this type
		}
		if err != nil {
			return nil, errors.Wrap(err, "failed to build base mesh context")
		}
	}
	newHash := rmap.Hash()
	if latest != nil && bytes.Equal(newHash, latest.hash) {
		return latest, nil
	}
	return &BaseMeshContext{
		hash:        newHash,
		Mesh:        mesh,
		ResourceMap: rmap,
	}, nil
}

type filterFn = func(rs core_model.Resource) bool

// fetch all resources of a type with potential filters etc
func (m *meshContextBuilder) fetchResourceList(ctx context.Context, resType core_model.ResourceType, mesh *core_mesh.MeshResource, filterFn filterFn) (core_model.ResourceList, error) {
	l := log.AddFieldsFromCtx(logger, ctx, context.Background())
	var listOptsFunc []core_store.ListOptionsFunc
	desc, err := registry.Global().DescriptorFor(resType)
	if err != nil {
		return nil, err
	}
	switch desc.Scope {
	case core_model.ScopeGlobal:
	case core_model.ScopeMesh:
		if mesh != nil {
			listOptsFunc = append(listOptsFunc, core_store.ListByMesh(mesh.GetMeta().GetName()))
		}
	default:
		return nil, fmt.Errorf("unknown resource scope:%s", desc.Scope)
	}
	// For some resources we apply extra filters
	switch resType {
	case core_mesh.ServiceInsightType:
		if mesh == nil {
			return desc.NewList(), nil
		}
		// ServiceInsights in XDS generation are only used to check whether the destination is ready to receive mTLS traffic.
		// This information is only useful when mTLS is enabled with PERMISSIVE mode.
		// Not including this into mesh hash for other cases saves us unnecessary XDS config generations.
		if backend := mesh.GetEnabledCertificateAuthorityBackend(); backend == nil || backend.Mode == mesh_proto.CertificateAuthorityBackend_STRICT {
			return desc.NewList(), nil
		}
	}
	listOptsFunc = append(listOptsFunc, core_store.ListOrdered())
	list := desc.NewList()
	if err := m.rm.List(ctx, list, listOptsFunc...); err != nil {
		return nil, err
	}
	if resType != core_mesh.ZoneIngressType && resType != core_mesh.DataplaneType && filterFn == nil {
		// No post processing stuff so return the list as is
		return list, nil
	}
	list, err = modifyAllEntries(list, func(resource core_model.Resource) (core_model.Resource, error) {
		// Because we're not using the pagination store we need to do the filtering ourselves outside of the store
		// I believe this is to maximize cachability of the store
		if filterFn != nil && !filterFn(resource) {
			return nil, nil
		}
		switch resType {
		case core_mesh.ZoneIngressType:
			zi, ok := resource.(*core_mesh.ZoneIngressResource)
			if !ok {
				return nil, errors.New("entry is not a zoneIngress this shouldn't happen")
			}

			resolvedZoneIngress, err := xds_topology.ResolveZoneIngressPublicAddress(m.ipFunc, zi)
			if err != nil {
				l.Error(err, "failed to resolve zoneIngress's domain name, ignoring zoneIngress", "name", zi.GetMeta().GetName())
				return nil, nil
			}
			return resolvedZoneIngress, nil
		case core_mesh.DataplaneType:
			list, err = modifyAllEntries(list, func(resource core_model.Resource) (core_model.Resource, error) {
				dp, ok := resource.(*core_mesh.DataplaneResource)
				if !ok {
					return nil, errors.New("entry is not a dataplane this shouldn't happen")
				}
				zi, err := xds_topology.ResolveDataplaneAddress(m.ipFunc, dp)
				if err != nil {
					l.Error(err, "failed to resolve dataplane's domain name, ignoring dataplane", "mesh", dp.GetMeta().GetMesh(), "name", dp.GetMeta().GetName())
					return nil, nil
				}
				return zi, nil
			})
		}
		return resource, nil
	})
	if err != nil {
		return nil, err
	}
	return list, nil
}

// takes a resourceList and modify it as needed
func modifyAllEntries(list core_model.ResourceList, fn func(resource core_model.Resource) (core_model.Resource, error)) (core_model.ResourceList, error) {
	newList := list.NewItem().Descriptor().NewList()
	for _, v := range list.GetItems() {
		ni, err := fn(v)
		if err != nil {
			return nil, err
		}
		if ni != nil {
			err := newList.AddItem(ni)
			if err != nil {
				return nil, err
			}
		}
	}
	newList.GetPagination().SetTotal(uint32(len(newList.GetItems())))
	return newList, nil
}

func (m *meshContextBuilder) generateServicesInformation(
	mesh *core_mesh.MeshResource,
	serviceInsights *core_mesh.ServiceInsightResourceList,
	endpointMap xds.EndpointMap,
	esEndpointMap xds.EndpointMap,
) map[string]*ServiceInformation {
	servicesInformation := map[string]*ServiceInformation{}
	m.resolveProtocol(mesh, endpointMap, esEndpointMap, servicesInformation)
	m.resolveTLSReadiness(mesh, serviceInsights, servicesInformation)
	return servicesInformation
}

func (m *meshContextBuilder) resolveProtocol(
	mesh *core_mesh.MeshResource,
	endpointMap xds.EndpointMap,
	esEndpointMap xds.EndpointMap,
	servicesInformation map[string]*ServiceInformation,
) {
	// endpointMap has only informations about externalServices when egress is enabled
	// that's why we have to iterate over second map with external services
	if !mesh.ZoneEgressEnabled() {
		for svc, endpoints := range esEndpointMap {
			serviceInfo := getServiceInformation(servicesInformation, svc)
			serviceInfo.Protocol = inferServiceProtocol(endpoints)
			serviceInfo.IsExternalService = true
			servicesInformation[svc] = serviceInfo
		}
	}
	for svc, endpoints := range endpointMap {
		serviceInfo := getServiceInformation(servicesInformation, svc)
		serviceInfo.Protocol = inferServiceProtocol(endpoints)
		serviceInfo.IsExternalService = isExternalService(endpoints)
		servicesInformation[svc] = serviceInfo
	}
}

func (m *meshContextBuilder) resolveTLSReadiness(
	mesh *core_mesh.MeshResource,
	serviceInsights *core_mesh.ServiceInsightResourceList,
	servicesInformation map[string]*ServiceInformation,
) {
	backend := mesh.GetEnabledCertificateAuthorityBackend()
	// TLS readiness is irrelevant unless we are using PERMISSIVE TLS, so skip
	// checking ServiceInsights if we aren't.
	if backend == nil || backend.Mode != mesh_proto.CertificateAuthorityBackend_PERMISSIVE {
		return
	}

	if len(serviceInsights.Items) == 0 {
		// Nothing about the TLS readiness has been reported yet
		logger.Info("could not determine service TLS readiness, ServiceInsight is not yet present")
		return
	}
	for svc, insight := range serviceInsights.Items[0].Spec.GetServices() {
		serviceInfo := getServiceInformation(servicesInformation, svc)
		if insight.ServiceType == mesh_proto.ServiceInsight_Service_external {
			serviceInfo.TLSReadiness = true
		} else {
			serviceInfo.TLSReadiness = insight.IssuedBackends[backend.Name] == (insight.Dataplanes.Offline + insight.Dataplanes.Online)
		}
		servicesInformation[svc] = serviceInfo
	}
}

func (m *meshContextBuilder) decorateWithCrossMeshResources(ctx context.Context, resources Resources) error {
	// Expand with crossMesh info
	otherMeshesByName := map[string]*core_mesh.MeshResource{}
	for _, m := range resources.OtherMeshes().GetItems() {
		otherMeshesByName[m.GetMeta().GetName()] = m.(*core_mesh.MeshResource)
		resources.CrossMeshResources[m.GetMeta().GetName()] = map[core_model.ResourceType]core_model.ResourceList{
			core_mesh.DataplaneType:   &core_mesh.DataplaneResourceList{},
			core_mesh.MeshGatewayType: &core_mesh.MeshGatewayResourceList{},
		}
	}
	var gatewaysByMesh map[string]core_model.ResourceList
	if _, ok := m.typeSet[core_mesh.MeshGatewayType]; ok {
		// For all meshes, get all cross mesh gateways
		rl, err := m.fetchResourceList(ctx, core_mesh.MeshGatewayType, nil, func(rs core_model.Resource) bool {
			_, exists := otherMeshesByName[rs.GetMeta().GetMesh()]
			return exists && rs.(*core_mesh.MeshGatewayResource).Spec.IsCrossMesh()
		})
		if err != nil {
			return errors.Wrap(err, "could not fetch cross mesh meshGateway resources")
		}
		gatewaysByMesh, err = core_model.ResourceListByMesh(rl)
		if err != nil {
			return errors.Wrap(err, "failed building cross mesh meshGateway resources")
		}
	}
	if _, ok := m.typeSet[core_mesh.DataplaneType]; ok {
		for otherMeshName, gws := range gatewaysByMesh { // Only iterate over meshes that have crossMesh gateways
			otherMesh := otherMeshesByName[otherMeshName]
			var gwResources []*core_mesh.MeshGatewayResource
			for _, gw := range gws.GetItems() {
				gwResources = append(gwResources, gw.(*core_mesh.MeshGatewayResource))
			}
			rl, err := m.fetchResourceList(ctx, core_mesh.DataplaneType, otherMesh, func(rs core_model.Resource) bool {
				dp := rs.(*core_mesh.DataplaneResource)
				if !dp.Spec.IsBuiltinGateway() {
					return false
				}
				return xds_topology.SelectGateway(gwResources, dp.Spec.Matches) != nil
			})
			if err != nil {
				return errors.Wrap(err, "could not fetch cross mesh meshGateway resources")
			}
			resources.CrossMeshResources[otherMeshName][core_mesh.DataplaneType] = rl
			resources.CrossMeshResources[otherMeshName][core_mesh.MeshGatewayType] = gws
		}
	}
	return nil
}

func buildLabelValueToServiceNames(ri core_model.ResourceIdentifier, resourceNamesByLabels LabelsToValuesToResourceIdentifier, labels map[string]string) {
	for label, value := range labels {
		key := LabelValue{
			Label: label,
			Value: value,
		}
		if _, ok := resourceNamesByLabels[key]; ok {
			resourceNamesByLabels[key][ri] = true
		} else {
			resourceNamesByLabels[key] = map[core_model.ResourceIdentifier]bool{
				ri: true,
			}
		}
	}
}

func (m *meshContextBuilder) hash(globalContext *GlobalContext, baseMeshContext *BaseMeshContext, managedTypes []core_model.ResourceType, resources Resources) []byte {
	slices.Sort(managedTypes)
	hasher := fnv.New128a()
	_, _ = hasher.Write(globalContext.hash)
	_, _ = hasher.Write(baseMeshContext.hash)
	for _, resType := range managedTypes {
		_, _ = hasher.Write(core_model.ResourceListHash(resources.MeshLocalResources[resType]))
	}

	for _, m := range maps.SortedKeys(resources.CrossMeshResources) {
		_, _ = hasher.Write([]byte(m))
		_, _ = hasher.Write(resources.CrossMeshResources[m].Hash())
	}
	return hasher.Sum(nil)
}

func getServiceInformation(servicesInformation map[string]*ServiceInformation, serviceName string) *ServiceInformation {
	if info, found := servicesInformation[serviceName]; found {
		return info
	}
	return &ServiceInformation{
		Protocol: core_mesh.ProtocolUnknown,
	}
}

func isExternalService(endpoints []xds.Endpoint) bool {
	for _, endpoint := range endpoints {
		return endpoint.IsExternalService()
	}
	return false
}

func inferServiceProtocol(endpoints []xds.Endpoint) core_mesh.Protocol {
	if len(endpoints) == 0 {
		return core_mesh.ProtocolUnknown
	}
	serviceProtocol := core_mesh.ParseProtocol(endpoints[0].Tags[mesh_proto.ProtocolTag])
	if endpoints[0].ExternalService != nil && endpoints[0].ExternalService.Protocol != "" {
		serviceProtocol = endpoints[0].ExternalService.Protocol
	}
	for _, endpoint := range endpoints[1:] {
		endpointProtocol := core_mesh.ParseProtocol(endpoint.Tags[mesh_proto.ProtocolTag])
		if endpoint.ExternalService != nil && endpoint.ExternalService.Protocol != "" {
			endpointProtocol = endpoint.ExternalService.Protocol
		}
		serviceProtocol = util_protocol.GetCommonProtocol(serviceProtocol, endpointProtocol)
	}
	return serviceProtocol
}
