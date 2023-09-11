package sync

import (
	"context"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	"github.com/kumahq/kuma/pkg/xds/template"
)

type IngressGatewayProxyBuilder struct {
	*IngressProxyBuilder
}

func (p *IngressGatewayProxyBuilder) Build(
	ctx context.Context,
	key core_model.ResourceKey,
	aggregatedMeshCtxs xds_context.AggregatedMeshContexts,
) (*core_xds.Proxy, error) {
	meshContext := aggregatedMeshCtxs.MustGetMeshContext(key.Mesh)

	dp, found := meshContext.DataplanesByName[key.Name]
	if !found {
		return nil, core_store.ErrorResourceNotFound(core_mesh.DataplaneType, key.Name, key.Mesh)
	}

	// NOTE(nicoche)
	// We want to fetch a lot of context related to available endpoints
	// in the current zone, for all meshes.
	// However, Dataplanes are scoped to a single mesh. We hack around
	// by constructing a ZoneIngressProxy, which turns out to contain
	// pretty much all we need. This allows us to grab everything we
	// need without having to write a whole builder from scratch.

	// Here, reference an existing ZoneIngress to take it from DB and
	// update its AvailableServices field that is full of the info we
	// need to build the ZoneIngressProxy
	zoneIngressKey := core_model.ResourceKey{
		// NOTE(nicoche): change that
		Name: "zoneingress-par1",
		// A ZoneIngress is not part of any mesh
		Mesh: "",
	}
	zoneIngress, err := p.getZoneIngress(ctx, zoneIngressKey, aggregatedMeshCtxs)
	if err != nil {
		return nil, err
	}

	meshName := meshContext.Resource.GetMeta().GetName()

	allMeshNames := []string{meshName}
	for _, mesh := range meshContext.Resources.OtherMeshes().Items {
		allMeshNames = append(allMeshNames, mesh.GetMeta().GetName())
	}

	secretsTracker := envoy.NewSecretsTracker(meshName, allMeshNames)

	matchedPolicies, err := p.matchPolicies(meshContext, dp)
	if err != nil {
		return nil, err
	}

	proxy := &core_xds.Proxy{
		Id:               core_xds.FromResourceKey(key),
		APIVersion:       p.apiVersion,
		Dataplane:        dp,
		ZoneIngressProxy: p.buildZoneIngressProxy(zoneIngress, aggregatedMeshCtxs),
		SecretsTracker:   secretsTracker,
		Metadata:         &core_xds.DataplaneMetadata{},
		Policies:         *matchedPolicies,
	}
	return proxy, nil
}

// NOTE(nicoche) This is copy/pasted from DataplaneProxyBuilder.matchPolicies
// We need this for e.g the IngressGateway because we want the Proxy object to contain
// the "plugin" policies. "plugin" policies are a newer style of policies like `MeshTrace`
// which are dynamic
func (p *IngressGatewayProxyBuilder) matchPolicies(meshContext xds_context.MeshContext, dataplane *core_mesh.DataplaneResource) (*core_xds.MatchedPolicies, error) {
	// additionalInbounds, err := manager_dataplane.AdditionalInbounds(dataplane, meshContext.Resource)
	// if err != nil {
	// 	return nil, errors.Wrap(err, "could not fetch additional inbounds")
	// }
	// inbounds := append(dataplane.Spec.GetNetworking().GetInbound(), additionalInbounds...)

	resources := meshContext.Resources
	// ratelimits := ratelimits.BuildRateLimitMap(dataplane, inbounds, resources.RateLimits().Items)
	matchedPolicies := &core_xds.MatchedPolicies{
		// TrafficPermissions: permissions.BuildTrafficPermissionMap(dataplane, inbounds, resources.TrafficPermissions().Items),
		// TrafficLogs:        logs.BuildTrafficLogMap(dataplane, resources.TrafficLogs().Items),
		// HealthChecks:       xds_topology.BuildHealthCheckMap(dataplane, outboundSelectors, resources.HealthChecks().Items),
		// CircuitBreakers:    xds_topology.BuildCircuitBreakerMap(dataplane, outboundSelectors, resources.CircuitBreakers().Items),
		// TrafficTrace:       xds_topology.SelectTrafficTrace(dataplane, resources.TrafficTraces().Items),
		// FaultInjections:    faultinjections.BuildFaultInjectionMap(dataplane, inbounds, resources.FaultInjections().Items),
		// Retries:            xds_topology.BuildRetryMap(dataplane, resources.Retries().Items, outboundSelectors),
		// Timeouts:           xds_topology.BuildTimeoutMap(dataplane, resources.Timeouts().Items),
		// RateLimitsInbound:  ratelimits.Inbound,
		// RateLimitsOutbound: ratelimits.Outbound,
		ProxyTemplate: template.SelectProxyTemplate(dataplane, resources.ProxyTemplates().Items),
		Dynamic:       core_xds.PluginOriginatedPolicies{},
	}
	for name, p := range plugins.Plugins().PolicyPlugins() {
		res, err := p.MatchedPolicies(dataplane, resources)
		if err != nil {
			return nil, errors.Wrapf(err, "could not apply policy plugin %s", name)
		}
		if res.Type == "" {
			return nil, errors.Wrapf(err, "matched policy didn't set type for policy plugin %s", name)
		}
		matchedPolicies.Dynamic[res.Type] = res
	}
	return matchedPolicies, nil
}
