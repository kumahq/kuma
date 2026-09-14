package context

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"hash/fnv"
	"slices"

	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	config_core "github.com/kumahq/kuma/v3/pkg/config/core"
	"github.com/kumahq/kuma/v3/pkg/core/datasource"
	"github.com/kumahq/kuma/v3/pkg/core/dns/lookup"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	meshexternalservice_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshexternalservice/api/v1alpha1"
	meshidentity_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshidentity/api/v1alpha1"
	meshmzservice_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshmultizoneservice/api/v1alpha1"
	meshservice_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshservice/api/v1alpha1"
	meshtrust_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshtrust/api/v1alpha1"
	meshzoneaddress_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshzoneaddress/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/v3/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/core/resources/registry"
	core_store "github.com/kumahq/kuma/v3/pkg/core/resources/store"
	"github.com/kumahq/kuma/v3/pkg/core/xds"
	xds_types "github.com/kumahq/kuma/v3/pkg/core/xds/types"
	"github.com/kumahq/kuma/v3/pkg/log"
	xds_topology "github.com/kumahq/kuma/v3/pkg/xds/topology"
)

type meshContextBuilder struct {
	rm                     manager.ReadOnlyResourceManager
	typeSet                map[core_model.ResourceType]struct{}
	ipFunc                 lookup.LookupIPFunc
	zone                   string
	withPolicyMatchingHash bool
}

// MeshContextBuilderOption configures optional behavior of the MeshContextBuilder.
type MeshContextBuilderOption func(*meshContextBuilder)

// WithPolicyMatchingHash enables computing MeshContext.PolicyMatchingHash (the MatchedPolicies
// cache key). It is skipped by default to avoid hashing work when the cache is disabled.
func WithPolicyMatchingHash() MeshContextBuilderOption {
	return func(m *meshContextBuilder) {
		m.withPolicyMatchingHash = true
	}
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
	opts ...MeshContextBuilderOption,
) MeshContextBuilder {
	typeSet := map[core_model.ResourceType]struct{}{}
	for _, typ := range types {
		typeSet[typ] = struct{}{}
	}

	builder := &meshContextBuilder{
		rm:      rm,
		typeSet: typeSet,
		ipFunc:  ipFunc,
		zone:    zone,
	}
	for _, opt := range opts {
		opt(builder)
	}
	return builder
}

func (m *meshContextBuilder) Build(ctx context.Context, meshName string) (MeshContext, error) {
	meshCtx, err := m.BuildIfChanged(ctx, meshName, nil)
	if err != nil {
		return MeshContext{}, err
	}
	return *meshCtx, nil
}

func (m *meshContextBuilder) BuildIfChanged(ctx context.Context, meshName string, latestMeshCtx *MeshContext) (*MeshContext, error) {
	var latestGlobal *GlobalContext
	var latestBase *BaseMeshContext
	if latestMeshCtx != nil {
		latestGlobal = latestMeshCtx.globalContext
		latestBase = latestMeshCtx.BaseMeshContext
	}
	globalContext, err := m.BuildGlobalContextIfChanged(ctx, latestGlobal)
	if err != nil {
		return nil, err
	}
	baseMeshContext, err := m.BuildBaseMeshContextIfChanged(ctx, meshName, latestBase)
	if err != nil {
		return nil, err
	}

	var managedTypes []core_model.ResourceType // The types not managed by global nor baseMeshContext
	resources := NewResources()
	// Build all the local entities from the parent contexts
	for resType := range m.typeSet {
		rl, ok := globalContext.ResourceMap[resType]
		if ok { // Exists in global context take it from there
			resources.MeshLocalResources[resType] = rl
		} else {
			rl, ok = baseMeshContext.ResourceMap[resType]
			if ok { // Exist in the baseMeshContext take it from there
				resources.MeshLocalResources[resType] = rl
			} else { // absent from all parent contexts get it now
				managedTypes = append(managedTypes, resType)
				rl, err = m.fetchResourceList(ctx, resType, baseMeshContext.Mesh)
				if err != nil {
					return nil, errors.Wrap(err, fmt.Sprintf("could not fetch resources of type:%s", resType))
				}
				resources.MeshLocalResources[resType] = rl
			}
		}
	}

	slices.Sort(managedTypes)
	managedTypeHashes := resources.MeshLocalResources.typeHashes(managedTypes)
	newHash := base64.StdEncoding.EncodeToString(m.hash(globalContext, baseMeshContext, managedTypeHashes))
	if latestMeshCtx != nil && newHash == latestMeshCtx.Hash {
		return latestMeshCtx, nil
	}

	var policyMatchingHash string
	if m.withPolicyMatchingHash {
		policyMatchingHash = base64.StdEncoding.EncodeToString(m.computePolicyMatchingHash(globalContext, baseMeshContext, managedTypeHashes))
	}

	loader := datasource.NewStaticLoader(resources.Secrets().Items)
	topology := m.topologyContextIfChanged(ctx, latestMeshCtx, baseMeshContext, managedTypeHashes, resources, loader)

	return &MeshContext{
		globalContext:                   globalContext,
		topology:                        topology,
		Hash:                            newHash,
		PolicyMatchingHash:              policyMatchingHash,
		Resource:                        baseMeshContext.Mesh,
		Resources:                       resources,
		BaseMeshContext:                 baseMeshContext,
		DataplanesByName:                topology.DataplanesByName,
		EndpointMap:                     topology.EndpointMap,
		VIPDomains:                      baseMeshContext.VIPDomains,
		VIPOutbounds:                    baseMeshContext.VIPOutbounds,
		DataSourceLoader:                loader,
		CAsByTrustDomain:                getCAsByTrustDomain(resources.MeshTrusts().Items),
		ZoneEgresses:                    topology.ZoneEgresses,
		DataplaneZoneIngressEndpointMap: topology.DataplaneZoneIngressEndpointMap,
		DataplaneZoneEgressEndpointMap:  topology.DataplaneZoneEgressEndpointMap,
	}, nil
}

// topologyInputTypes are the types a TopologyContext is built from. Its hash covers only these,
// so a change to anything else, like a policy, keeps the endpoint maps of the previous context.
var topologyInputTypes = map[core_model.ResourceType]bool{
	core_mesh.DataplaneType:                         true,
	meshzoneaddress_api.MeshZoneAddressType:         true,
	system.SecretType:                               true,
	meshidentity_api.MeshIdentityType:               true,
	meshservice_api.MeshServiceType:                 true,
	meshexternalservice_api.MeshExternalServiceType: true,
	meshmzservice_api.MeshMultiZoneServiceType:      true,
}

func (m *meshContextBuilder) topologyContextIfChanged(
	ctx context.Context,
	latestMeshCtx *MeshContext,
	baseMeshContext *BaseMeshContext,
	managedTypeHashes []typeHash,
	resources Resources,
	loader datasource.Loader,
) *TopologyContext {
	hasher := fnv.New128a()
	for _, hashes := range [][]typeHash{baseMeshContext.typeHashes, managedTypeHashes} {
		for _, th := range hashes {
			if topologyInputTypes[th.resourceType] {
				_, _ = hasher.Write(th.hash)
			}
		}
	}
	hash := hasher.Sum(nil)
	if latestMeshCtx != nil && latestMeshCtx.topology != nil && bytes.Equal(latestMeshCtx.topology.hash, hash) {
		return latestMeshCtx.topology
	}

	dataplanes := resources.Dataplanes().Items
	dataplanesByName := make(map[string]*core_mesh.DataplaneResource, len(dataplanes))
	for _, dp := range dataplanes {
		dataplanesByName[dp.Meta.GetName()] = dp
	}
	meshServices := resources.MeshServices().Items
	meshExternalServices := resources.MeshExternalServices().Items
	identities := resources.MeshIdentities().Items
	zoneEgresses := resolveZoneEgresses(dataplanes, identities, m.zone)
	endpointMap, zoneIngressEndpointMap := xds_topology.BuildDataplaneEndpointMaps(
		ctx,
		m.zone,
		meshServices,
		resources.MeshMultiZoneServices().Items,
		meshExternalServices,
		dataplanes,
		resources.MeshZoneAddresses().Items,
		loader,
		len(identities) > 0,
		zoneEgresses,
	)
	return &TopologyContext{
		DataplanesByName:                dataplanesByName,
		EndpointMap:                     endpointMap,
		ZoneEgresses:                    zoneEgresses,
		DataplaneZoneIngressEndpointMap: zoneIngressEndpointMap,
		DataplaneZoneEgressEndpointMap:  xds_topology.BuildDataplaneZoneEgressEndpointMap(ctx, baseMeshContext.Mesh, meshExternalServices, loader),
		hash:                            hash,
	}
}

func (m *meshContextBuilder) BuildGlobalContextIfChanged(ctx context.Context, latest *GlobalContext) (*GlobalContext, error) {
	rmap := ResourceMap{}
	for t := range m.typeSet {
		desc, err := registry.Global().DescriptorFor(t)
		if err != nil {
			return nil, err
		}
		if desc.Scope == core_model.ScopeGlobal && !meshScopedGlobalTypes[desc.Name] {
			rmap[t], err = m.fetchResourceList(ctx, t, nil)
			if err != nil {
				return nil, errors.Wrap(err, "failed to build global context")
			}
		}
	}

	typeHashes := rmap.hashByType()
	newHash := combineTypeHashes(typeHashes)
	if latest != nil && bytes.Equal(newHash, latest.hash) {
		return latest, nil
	}
	return &GlobalContext{
		hash:        newHash,
		typeHashes:  typeHashes,
		ResourceMap: rmap,
	}, nil
}

// meshScopedGlobalTypes are global types that never go to the global context. Config is left to
// more specific filters, and each mesh only needs its own Mesh, which the base mesh context holds,
// so a change to one Mesh does not invalidate every other mesh.
var meshScopedGlobalTypes = map[core_model.ResourceType]bool{
	system.ConfigType:  true,
	core_mesh.MeshType: true,
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
	var destinations [][]core_model.Resource
	for t := range m.typeSet {
		desc, err := registry.Global().DescriptorFor(t)
		if err != nil {
			return nil, err
		}
		// Only pick the policies, gateways, and the vip config map
		switch {
		case desc.IsDestination:
			rmap[t], err = m.fetchResourceList(ctx, t, mesh)
			if err != nil {
				return nil, errors.Wrap(err, "failed to build base mesh context")
			}
			destinations = append(destinations, rmap[t].GetItems())
		case desc.IsPolicy:
			rmap[t], err = m.fetchResourceList(ctx, t, mesh)
			if err != nil {
				return nil, errors.Wrap(err, "failed to build base mesh context")
			}
		default:
			// DO nothing we're not interested in this type
		}
	}
	typeHashes := rmap.hashByType()
	newHash := combineTypeHashes(typeHashes)
	if latest != nil && bytes.Equal(newHash, latest.hash) {
		return latest, nil
	}

	destinationResources := Resources{MeshLocalResources: rmap}
	return &BaseMeshContext{
		hash:             newHash,
		typeHashes:       typeHashes,
		Mesh:             mesh,
		ResourceMap:      rmap,
		DestinationIndex: NewDestinationIndex(destinations...),
		VIPDomains:       vipDomains(destinationResources),
		VIPOutbounds:     vipOutbounds(destinationResources),
	}, nil
}

func vipOutbounds(resources Resources) xds_types.Outbounds {
	var outbounds xds_types.Outbounds
	outbounds = append(outbounds, xds_topology.Outbounds(resources.MeshServices().Items)...)
	outbounds = append(outbounds, xds_topology.Outbounds(resources.MeshExternalServices().Items)...)
	return append(outbounds, xds_topology.Outbounds(resources.MeshMultiZoneServices().Items)...)
}

func vipDomains(resources Resources) []xds_types.VIPDomains {
	var domains []xds_types.VIPDomains
	domains = append(domains, xds_topology.Domains(resources.MeshServices().Items)...)
	domains = append(domains, xds_topology.Domains(resources.MeshExternalServices().Items)...)
	return append(domains, xds_topology.Domains(resources.MeshMultiZoneServices().Items)...)
}

// fetch all resources of a type with potential filters etc
func (m *meshContextBuilder) fetchResourceList(ctx context.Context, resType core_model.ResourceType, mesh *core_mesh.MeshResource) (core_model.ResourceList, error) {
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
	listOptsFunc = append(listOptsFunc, core_store.ListOrdered())
	list := desc.NewList()
	if err := m.rm.List(ctx, list, listOptsFunc...); err != nil {
		return nil, err
	}
	if resType != core_mesh.DataplaneType && resType != meshzoneaddress_api.MeshZoneAddressType {
		// No post processing stuff so return the list as is
		return list, nil
	}
	list, err = modifyAllEntries(list, func(resource core_model.Resource) (core_model.Resource, error) {
		switch resType {
		case meshzoneaddress_api.MeshZoneAddressType:
			mza, ok := resource.(*meshzoneaddress_api.MeshZoneAddressResource)
			if !ok {
				return nil, errors.New("entry is not a meshZoneAddress this shouldn't happen")
			}

			resolvedMeshZoneAddress, err := xds_topology.ResolveMeshZoneAddressPublicAddress(m.ipFunc, mza)
			if err != nil {
				l.Error(err, "failed to resolve meshZoneAddress's domain name, ignoring meshZoneAddress", "mesh", mza.GetMeta().GetMesh(), "name", mza.GetMeta().GetName())
				return nil, nil
			}
			return resolvedMeshZoneAddress, nil
		case core_mesh.DataplaneType:
			dp, ok := resource.(*core_mesh.DataplaneResource)
			if !ok {
				return nil, errors.New("entry is not a dataplane this shouldn't happen")
			}

			resolvedDataplane, err := xds_topology.ResolveDataplaneAddress(m.ipFunc, dp)
			if err != nil {
				l.Error(err, "failed to resolve dataplane's domain name, ignoring dataplane", "mesh", dp.GetMeta().GetMesh(), "name", dp.GetMeta().GetName())
				return nil, nil
			}
			return resolvedDataplane, nil
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

func (m *meshContextBuilder) hash(globalContext *GlobalContext, baseMeshContext *BaseMeshContext, managedTypeHashes []typeHash) []byte {
	hasher := fnv.New128a()
	_, _ = hasher.Write(globalContext.hash)
	_, _ = hasher.Write(baseMeshContext.hash)
	for _, th := range managedTypeHashes {
		_, _ = hasher.Write(th.hash)
	}
	return hasher.Sum(nil)
}

// affectsPolicyMatching reports whether resources of resType can change which policies match a
// dataplane, per the type's ResourceTypeDescriptor.AffectsPolicyMatching. Unknown types are
// treated as relevant (safe default: spurious cache misses, never stale xDS).
func affectsPolicyMatching(resType core_model.ResourceType) bool {
	descriptor, err := registry.Global().DescriptorFor(resType)
	if err != nil {
		return true
	}
	return descriptor.AffectsPolicyMatching
}

func (m *meshContextBuilder) computePolicyMatchingHash(globalContext *GlobalContext, baseMeshContext *BaseMeshContext, managedTypeHashes []typeHash) []byte {
	hasher := fnv.New128a()
	for _, hashes := range [][]typeHash{globalContext.typeHashes, baseMeshContext.typeHashes, managedTypeHashes} {
		for _, th := range hashes {
			if affectsPolicyMatching(th.resourceType) {
				_, _ = hasher.Write(policyMatchingListHash(th))
			}
		}
	}
	return hasher.Sum(nil)
}

func resolveZoneEgresses(
	dataplanes []*core_mesh.DataplaneResource,
	identities []*meshidentity_api.MeshIdentityResource,
	zone string,
) []xds.ZoneEgressInstance {
	var dpEgresses []xds.ZoneEgressInstance
	for _, dp := range dataplanes {
		listeners := dp.Spec.GetNetworking().GetReadyZoneEgressListeners()
		if len(listeners) == 0 {
			continue
		}
		var san string
		if identity, ok := meshidentity_api.BestMatched(dp.GetMeta().GetLabels(), identities); ok {
			env := config_core.UniversalEnvironment
			if _, isK8s := dp.GetMeta().GetLabels()[mesh_proto.KubeNamespaceTag]; isK8s {
				env = config_core.KubernetesEnvironment
			}
			if trustDomain, err := identity.GetTrustDomain(zone); err != nil {
				logger.Error(err, "failed to compute trust domain for zone egress", "dataplane", dp.GetMeta().GetName())
			} else if spiffeID, err := identity.Spec.GetSpiffeID(trustDomain, dp.GetMeta(), env); err != nil {
				logger.Error(err, "failed to compute SPIFFE ID for zone egress", "dataplane", dp.GetMeta().GetName())
			} else {
				san = spiffeID
			}
		}
		for _, l := range listeners {
			dpEgresses = append(dpEgresses, xds.ZoneEgressInstance{Address: l.GetAddress(), Port: l.GetPort(), SAN: san})
		}
	}
	return dpEgresses
}

func getCAsByTrustDomain(trusts []*meshtrust_api.MeshTrustResource) map[string][]PEMBytes {
	casByTrustDomain := map[string][]PEMBytes{}
	for _, trust := range trusts {
		for _, ca := range trust.Spec.CABundles {
			casByTrustDomain[trust.Spec.TrustDomain] = append(casByTrustDomain[trust.Spec.TrustDomain], PEMBytes(ca.PEM.Value))
		}
	}
	return casByTrustDomain
}
