package context

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"hash/fnv"
	"os"
	"slices"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"

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
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	"github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/dns/vips"
	"github.com/kumahq/kuma/pkg/events"
	"github.com/kumahq/kuma/pkg/log"
	"github.com/kumahq/kuma/pkg/util/maps"
	util_protocol "github.com/kumahq/kuma/pkg/util/protocol"
	xds_topology "github.com/kumahq/kuma/pkg/xds/topology"
)

var logger = core.Log.WithName("xds").WithName("context")

type meshContextBuilder struct {
	rm                       manager.ReadOnlyResourceManager
	typeSet                  map[core_model.ResourceType]struct{}
	ipFunc                   lookup.LookupIPFunc
	zone                     string
	vipsPersistence          *vips.Persistence
	topLevelDomain           string
	vipPort                  uint32
	rsGraphBuilder           ReachableServicesGraphBuilder
	changedTypesByMesh       map[string]map[core_model.ResourceType]struct{}
	eventBus                 events.EventBus
	hashCacheBaseMeshContext *cache.Cache
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
	BuildBaseMeshContextIfChangedV2(ctx context.Context, meshName string, latest *BaseMeshContext) (*BaseMeshContext, error)

	// BuildIfChanged builds MeshContext only if latestMeshCtx is nil or hash of
	// latestMeshCtx is different.
	// If hash is the same, then the function returns the passed latestMeshCtx.
	// Hash returned in MeshContext can never be empty.
	BuildIfChanged(ctx context.Context, meshName string, latestMeshCtx *MeshContext) (*MeshContext, error)

	Start(stop <-chan struct{}) error
	NeedLeaderElection() bool
}

// cleanupTime is the time after which the mesh context is removed from
// the longer TTL cache.
// It exists to ensure contexts of deleted Meshes are eventually cleaned up.
const cleanupTime = 10 * time.Minute

type MeshContextBuilderComponent interface {
	MeshContextBuilder
	component.Component
}

func NewMeshContextBuilderComponent(
	rm manager.ReadOnlyResourceManager,
	types []core_model.ResourceType, // types that should be taken into account when MeshContext is built.
	ipFunc lookup.LookupIPFunc,
	zone string,
	vipsPersistence *vips.Persistence,
	topLevelDomain string,
	vipPort uint32,
	rsGraphBuilder ReachableServicesGraphBuilder,
	eventBus events.EventBus,
) MeshContextBuilderComponent {
	typeSet := map[core_model.ResourceType]struct{}{}
	for _, typ := range types {
		typeSet[typ] = struct{}{}
	}

	return &meshContextBuilder{
		rm:                       rm,
		typeSet:                  typeSet,
		ipFunc:                   ipFunc,
		zone:                     zone,
		vipsPersistence:          vipsPersistence,
		topLevelDomain:           topLevelDomain,
		vipPort:                  vipPort,
		rsGraphBuilder:           rsGraphBuilder,
		changedTypesByMesh:       map[string]map[core_model.ResourceType]struct{}{},
		eventBus:                 eventBus,
		hashCacheBaseMeshContext: cache.New(cleanupTime, time.Duration(int64(float64(cleanupTime)*0.9))),
	}
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
		rm:                       rm,
		typeSet:                  typeSet,
		ipFunc:                   ipFunc,
		zone:                     zone,
		vipsPersistence:          vipsPersistence,
		topLevelDomain:           topLevelDomain,
		vipPort:                  vipPort,
		rsGraphBuilder:           rsGraphBuilder,
		changedTypesByMesh:       nil,
		hashCacheBaseMeshContext: nil,
	}
}

func useReactiveBuildBaseMeshContext() bool {
	return os.Getenv("EXPERIMENTAL_REACTIVE_BASE_MESH_CONTEXT") != ""
}

func (m *meshContextBuilder) Start(stop <-chan struct{}) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
	}()

	l := log.AddFieldsFromCtx(logger, ctx, context.Background())

	listener := m.eventBus.Subscribe(func(event events.Event) bool {
		resChange, ok := event.(events.ResourceChangedEvent)
		if !ok {
			return false
		}

		// if resChange.TenantID != tenantID {
		// 	return false
		// }

		_, ok = m.typeSet[resChange.Type]
		return ok
	})

	for {
		select {
		case <-stop:
			return nil

		case event := <-listener.Recv():
			if useReactiveBuildBaseMeshContext() {
				resChange := event.(events.ResourceChangedEvent)
				l.Info("Received", "ResourceChangedEvent", resChange)
				mesh := resChange.Key.Mesh
				if mesh != "" {
					l.Info("Type has changed for mesh", "type", resChange.Type, "mesh", mesh)
					m.setTypeChanged(mesh, resChange.Type)
				}
			}
		}
	}
}

func (m *meshContextBuilder) NeedLeaderElection() bool {
	return false
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

	// Check hashCache first for an existing mesh latestContext
	var latestBaseMeshContext *BaseMeshContext
	if m.hashCacheBaseMeshContext != nil {
		if cached, ok := m.hashCacheBaseMeshContext.Get(meshName); ok {
			latestBaseMeshContext = cached.(*BaseMeshContext)
		}
	}

	var baseMeshContext *BaseMeshContext
	if useReactiveBuildBaseMeshContext() {
		baseMeshContext, err = m.BuildBaseMeshContextIfChangedV2(ctx, meshName, latestBaseMeshContext)
		if err != nil {
			return nil, err
		}
	} else {
		baseMeshContext, err = m.BuildBaseMeshContextIfChanged(ctx, meshName, latestBaseMeshContext)
		if err != nil {
			return nil, err
		}
	}

	// By always setting the mesh context, we refresh the TTL
	// with the effect that often used contexts remain in the cache while no
	// longer used contexts are evicted.
	m.hashCacheBaseMeshContext.SetDefault(meshName, baseMeshContext)

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

	virtualOutboundView, err := m.vipsPersistence.GetByMesh(ctx, meshName)
	if err != nil {
		return nil, errors.Wrap(err, "could not fetch vips")
	}
	// resolve all the domains
	domains, outbounds := xds_topology.VIPOutbounds(virtualOutboundView, m.topLevelDomain, m.vipPort)
	loader := datasource.NewStaticLoader(resources.Secrets().Items)

	mesh := baseMeshContext.Mesh
	zoneIngresses := resources.ZoneIngresses().Items
	zoneEgresses := resources.ZoneEgresses().Items
	externalServices := resources.ExternalServices().Items
	endpointMap := xds_topology.BuildEdsEndpointMap(mesh, m.zone, dataplanes, zoneIngresses, zoneEgresses, externalServices)
	esEndpointMap := xds_topology.BuildExternalServicesEndpointMap(ctx, mesh, externalServices, loader, m.zone)

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
		Hash:                        newHash,
		Resource:                    mesh,
		Resources:                   resources,
		DataplanesByName:            dataplanesByName,
		EndpointMap:                 endpointMap,
		ExternalServicesEndpointMap: esEndpointMap,
		CrossMeshEndpoints:          crossMeshEndpointMap,
		VIPDomains:                  domains,
		VIPOutbounds:                outbounds,
		ServicesInformation:         m.generateServicesInformation(mesh, resources.ServiceInsights(), endpointMap, esEndpointMap),
		DataSourceLoader:            loader,
		ReachableServicesGraph:      m.rsGraphBuilder(meshName, resources),
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

func (m *meshContextBuilder) BuildBaseMeshContextIfChangedV2(ctx context.Context, meshName string, latest *BaseMeshContext) (*BaseMeshContext, error) {
	l := log.AddFieldsFromCtx(logger, ctx, context.Background())
	if m.changedTypesByMesh == nil || latest == nil {
		l.Info("no latest base mesh context to use or not using reactive method. Fallabck to default BuildBaseMeshContextIfChanged", "mesh", meshName)
		m.clearTypeChanged(meshName)
		return m.BuildBaseMeshContextIfChanged(ctx, meshName, nil)
	}

	changedTypes, ok := m.changedTypesByMesh[meshName]
	if !ok || len(changedTypes) == 0 {
		// No occurence of this mesh in changed types. Let's re-use latest base mesh context
		l.Info("no resource changed, re-using latest base mesh context to build mesh context", "mesh", meshName)
		return latest, nil
	}

	rmap := ResourceMap{}

	// Find mesh, either in last base context if it hasn't changed since or in the store
	mesh := core_mesh.NewMeshResource()
	_, meshChanged := changedTypes[core_mesh.MeshType]
	if !meshChanged && latest != nil {
		meshList := latest.ResourceMap[core_mesh.MeshType].(*core_mesh.MeshResourceList)
		mesh = meshList.Items[0]
	} else {
		if err := m.rm.Get(ctx, mesh, core_store.GetByKey(meshName, core_model.NoMesh)); err != nil {
			return nil, errors.Wrapf(err, "could not fetch mesh %s", meshName)
		}

	}

	// Add mesh to resource map
	rmap[core_mesh.MeshType] = mesh.Descriptor().NewList()
	_ = rmap[core_mesh.MeshType].AddItem(mesh)

	for t := range m.typeSet {
		desc, err := registry.Global().DescriptorFor(t)
		if err != nil {
			return nil, err
		}

		switch {
		case desc.IsPolicy || desc.Name == core_mesh.MeshGatewayType || desc.Name == core_mesh.ExternalServiceType:
			rmap[t], err = m.fetchResourceListIfChanged(ctx, latest, t, mesh, nil)
		case desc.Name == system.ConfigType:
			rmap[t], err = m.fetchResourceListIfChanged(ctx, latest, t, mesh, func(rs core_model.Resource) bool {
				return rs.GetMeta().GetName() == vips.ConfigKey(meshName)
			})
		default:
			// Do nothing we're not interested in this type
		}
		if err != nil {
			return nil, errors.Wrap(err, "failed to build base mesh context")
		}
	}

	// Reset changed types for this mesh
	m.clearTypeChanged(meshName)

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

func (m *meshContextBuilder) setTypeChanged(mesh string, t core_model.ResourceType) {
	_, ok := m.changedTypesByMesh[mesh]
	if !ok {
		m.changedTypesByMesh[mesh] = map[core_model.ResourceType]struct{}{}
	}

	m.changedTypesByMesh[mesh][t] = struct{}{}
}

func (m *meshContextBuilder) clearTypeChanged(mesh string) {
	if m.changedTypesByMesh != nil {
		m.changedTypesByMesh[mesh] = map[core_model.ResourceType]struct{}{}
	}
}

func (m *meshContextBuilder) BuildBaseMeshContextIfChanged(ctx context.Context, meshName string, latest *BaseMeshContext) (*BaseMeshContext, error) {
	mesh := core_mesh.NewMeshResource()
	if err := m.rm.Get(ctx, mesh, core_store.GetByKey(meshName, core_model.NoMesh)); err != nil {
		return nil, errors.Wrapf(err, "could not fetch mesh %s", meshName)
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

		switch {
		case desc.IsPolicy || desc.Name == core_mesh.MeshGatewayType || desc.Name == core_mesh.ExternalServiceType:
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

// fetch resource from latest base mesh context if it hasn't changed. Else, pull it from the store
func (m *meshContextBuilder) fetchResourceListIfChanged(ctx context.Context, latest *BaseMeshContext, resType core_model.ResourceType, mesh *core_mesh.MeshResource, filterFn filterFn) (core_model.ResourceList, error) {
	if latest == nil {
		return m.fetchResourceList(ctx, resType, mesh, filterFn)
	}

	meshName := mesh.GetMeta().GetName()
	changedTypes, ok := m.changedTypesByMesh[meshName]
	if !ok {
		return latest.ResourceMap[core_mesh.MeshType], nil
	}

	_, hasChanged := changedTypes[resType]
	if !hasChanged {
		return latest.ResourceMap[resType], nil
	}

	return m.fetchResourceList(ctx, resType, mesh, nil)
}

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
			zi, err := xds_topology.ResolveZoneIngressPublicAddress(m.ipFunc, zi)
			if err != nil {
				l.Error(err, "failed to resolve zoneIngress's domain name, ignoring zoneIngress", "name", zi.GetMeta().GetName())
				return nil, nil
			}
			return zi, nil
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
	for _, endpoint := range endpoints[1:] {
		endpointProtocol := core_mesh.ParseProtocol(endpoint.Tags[mesh_proto.ProtocolTag])
		serviceProtocol = util_protocol.GetCommonProtocol(serviceProtocol, endpointProtocol)
	}
	return serviceProtocol
}
