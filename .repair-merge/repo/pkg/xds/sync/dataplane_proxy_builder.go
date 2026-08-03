package sync

import (
	"context"
	"net"

	"github.com/pkg/errors"

	common_api "github.com/kumahq/kuma/v3/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/kri"
	core_plugins "github.com/kumahq/kuma/v3/pkg/core/plugins"
	core_resources "github.com/kumahq/kuma/v3/pkg/core/resources/apis/core"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/v3/pkg/core/resources/store"
	core_xds "github.com/kumahq/kuma/v3/pkg/core/xds"
	xds_types "github.com/kumahq/kuma/v3/pkg/core/xds/types"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules/resolve"
	tproxy_dp "github.com/kumahq/kuma/v3/pkg/transparentproxy/config/dataplane"
	"github.com/kumahq/kuma/v3/pkg/util/pointer"
	xds_context "github.com/kumahq/kuma/v3/pkg/xds/context"
	"github.com/kumahq/kuma/v3/pkg/xds/envoy"
)

type DataplaneProxyBuilder struct {
	Zone              string
	APIVersion        core_xds.APIVersion
	InternalAddresses []core_xds.InternalAddress
	IncludeShadow     bool
	// nil disables caching (inspect, test, egress paths)
	policyMatchingCache core_plugins.PolicyMatchingCacheAccessor
}

func (p *DataplaneProxyBuilder) WithPolicyMatchingCache(cache core_plugins.PolicyMatchingCacheAccessor) *DataplaneProxyBuilder {
	p.policyMatchingCache = cache
	return p
}

func (p *DataplaneProxyBuilder) Build(ctx context.Context, key core_model.ResourceKey, meta *core_xds.DataplaneMetadata, meshContext xds_context.MeshContext) (*core_xds.Proxy, error) {
	dp, found := meshContext.DataplanesByName[key.Name]
	if !found {
		return nil, core_store.ErrorResourceNotFound(core_mesh.DataplaneType, key.Name, key.Mesh)
	}

	tpEnabled := tproxy_dp.GetDataplaneConfig(dp, meta).Enabled()
	routing, outbounds := p.resolveRouting(meshContext, dp, tpEnabled, meta.HasFeature(xds_types.FeatureBindOutbounds))

	matchedPolicies, err := p.matchPolicies(meshContext, dp)
	if err != nil {
		return nil, errors.Wrap(err, "could not match policies")
	}

	meshName := meshContext.Resource.GetMeta().GetName()

	allMeshNames := []string{}
	for _, mesh := range meshContext.Resources.Meshes().Items {
		allMeshNames = append(allMeshNames, mesh.GetMeta().GetName())
	}

	secretsTracker := envoy.NewSecretsTracker(meshName, allMeshNames)

	proxy := &core_xds.Proxy{
		Id:                core_xds.FromResourceKey(key),
		APIVersion:        p.APIVersion,
		InternalAddresses: p.InternalAddresses,
		Dataplane:         dp,
		Outbounds:         outbounds,
		Routing:           *routing,
		Policies:          *matchedPolicies,
		SecretsTracker:    secretsTracker,
		Metadata:          meta,
		Zone:              p.Zone,
		RuntimeExtensions: map[string]any{},
	}
	for k, pl := range core_plugins.Plugins().ProxyPlugins() {
		err := pl.Apply(ctx, meshContext, proxy)
		if err != nil {
			return nil, errors.Wrapf(err, "failed applying proxy plugin: %s", k)
		}
	}
	return proxy, nil
}

func (p *DataplaneProxyBuilder) resolveRouting(
	meshContext xds_context.MeshContext,
	dataplane *core_mesh.DataplaneResource,
	tpEnabled bool,
	bindOutbounds bool,
) (*core_xds.Routing, []*xds_types.Outbound) {
	outbounds := p.resolveVIPOutbounds(meshContext, dataplane, tpEnabled, bindOutbounds)
	routing := &core_xds.Routing{OutboundTargets: meshContext.EndpointMap}
	return routing, outbounds
}

func (p *DataplaneProxyBuilder) resolveVIPOutbounds(
	meshContext xds_context.MeshContext,
	dataplane *core_mesh.DataplaneResource,
	tpEnabled bool,
	bindOutbounds bool,
) []*xds_types.Outbound {
	if !tpEnabled && !bindOutbounds {
		return asOutbounds(dataplane, meshContext.ResolveResourceIdentifier)
	}
	var reachableBackends map[kri.Identifier]core_resources.Port
	var onlySelectedBackends bool
	if dataplane.Spec.GetNetworking().GetTransparentProxying() != nil {
		reachableBackends, onlySelectedBackends = meshContext.BaseMeshContext.DestinationIndex.GetReachableBackends(dataplane)
	}

	var newOutbounds []*xds_types.Outbound
	for _, outbound := range meshContext.VIPOutbounds {
		if onlySelectedBackends {
			// check if there is an entry with specific port or without port
			_, selected := reachableBackends[outbound.Resource]
			_, selectedBySectionName := reachableBackends[kri.NoSectionName(outbound.Resource)]
			if !selected && !selectedBySectionName {
				// ignore VIP outbound if reachableBackends is defined and not specified
				continue
			}
		}
		if dataplane.UsesInboundInterface(net.ParseIP(outbound.GetAddress()), outbound.GetPort()) {
			// Skip overlapping outbound interface with inbound.
			// This may happen for example with Headless service on Kubernetes (outbound is a PodIP not ClusterIP, so it's the same as inbound).
			continue
		}
		newOutbounds = append(newOutbounds, outbound)
	}
	// VIP outbounds never carry a legacy outbound, so the dataplane's legacy
	// outbound list is always cleared here.
	dataplane.Spec.Networking.Outbound = nil
	return newOutbounds
}

func (p *DataplaneProxyBuilder) matchPolicies(meshContext xds_context.MeshContext, dataplane *core_mesh.DataplaneResource) (*core_xds.MatchedPolicies, error) {
	resources := meshContext.Resources
	matchedPolicies := &core_xds.MatchedPolicies{
		Dynamic: core_xds.PluginOriginatedPolicies{},
	}
	opts := []core_plugins.MatchedPoliciesOption{}
	if p.IncludeShadow {
		opts = append(opts, core_plugins.IncludeShadow())
	}
	if p.policyMatchingCache != nil {
		opts = append(opts, core_plugins.WithCache(p.policyMatchingCache, meshContext.PolicyMatchingHash))
	}
	for _, p := range core_plugins.Plugins().PolicyPlugins() {
		res, err := p.Plugin.MatchedPolicies(dataplane, resources, opts...)
		if err != nil {
			return nil, errors.Wrapf(err, "could not apply policy plugin %s", p.Name)
		}
		if res.Type == "" {
			return nil, errors.Wrapf(err, "matched policy didn't set type for policy plugin %s", p.Name)
		}
		matchedPolicies.Dynamic[res.Type] = res
	}
	return matchedPolicies, nil
}

func asOutbounds(dataplane *core_mesh.DataplaneResource, resolver resolve.LabelResourceIdentifierResolver) xds_types.Outbounds {
	var outbounds xds_types.Outbounds
	for _, o := range dataplane.Spec.Networking.Outbound {
		if o.BackendRef != nil {
			port := o.BackendRef.Port
			labels, sectionName := xds_context.NormalizeBackendRefTarget(
				o.BackendRef.Kind,
				o.BackendRef.Name,
				"",
				&port,
				o.BackendRef.Labels,
				dataplane.GetMeta().GetLabels()[mesh_proto.KubeNamespaceTag],
			)
			// convert proto BackendRef to common_api.BackendRef
			backendRef := common_api.BackendRef{
				TargetRef: common_api.TargetRef{
					Kind:   common_api.TargetRefKind(o.BackendRef.Kind),
					Labels: pointer.To(labels),
				},
				Port: pointer.To(o.BackendRef.Port),
			}
			if sectionName != "" {
				backendRef.SectionName = pointer.To(sectionName)
			}
			ref, ok := resolve.BackendRef(kri.From(dataplane), backendRef, resolver)
			if !ok {
				continue
			}
			if ref.ReferencesRealResource() {
				outbounds = append(outbounds, &xds_types.Outbound{
					Address:  o.Address,
					Port:     o.Port,
					Resource: ref.Resource(),
				})
			}
		} else {
			outbounds = append(outbounds, &xds_types.Outbound{LegacyOutbound: o})
		}
	}
	return outbounds
}
