package egress

import (
	"context"
	"slices"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/core/kri"
	"github.com/kumahq/kuma/pkg/core/naming"
	"github.com/kumahq/kuma/pkg/core/naming/unified-naming"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	envoy_listeners "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
	envoy_names "github.com/kumahq/kuma/pkg/xds/envoy/names"
	generator_core "github.com/kumahq/kuma/pkg/xds/generator/core"
	"github.com/kumahq/kuma/pkg/xds/generator/metadata"
	generator_secrets "github.com/kumahq/kuma/pkg/xds/generator/secrets"
)

// Generator generates xDS resources for an entire ZoneEgress.
type Generator struct {
	SecretGenerator *generator_secrets.Generator
	PolicyGenerator generator_core.ResourceGenerator
}

func (g Generator) Generate(
	ctx context.Context,
	_ *core_xds.ResourceSet,
	xdsCtx xds_context.Context,
	proxy *core_xds.Proxy,
) (*core_xds.ResourceSet, error) {
	rs := core_xds.NewResourceSet()

	unifiedNaming := unified_naming.Enabled(proxy.Metadata, xdsCtx.Mesh.Resource)
	getName := naming.GetNameOrFallbackFunc(unifiedNaming)

	zoneEgress := proxy.ZoneEgressProxy.ZoneEgressResource
	address := zoneEgress.Spec.GetNetworking().GetAddress()
	port := zoneEgress.Spec.GetNetworking().GetPort()

	listenerName := getName(kri.From(zoneEgress).String(), envoy_names.GetInboundListenerName(address, port))
	statPrefix := getName(naming.MustContextualInboundName(zoneEgress, port), "")

	listener := envoy_listeners.NewListenerBuilder(proxy.APIVersion, listenerName).
		Configure(envoy_listeners.InboundListener(address, port, core_xds.SocketAddressProtocolTCP)).
		Configure(envoy_listeners.StatPrefix(statPrefix)).
		Configure(envoy_listeners.TLSInspector())

	for _, meshResources := range proxy.ZoneEgressProxy.MeshResourcesList {
		mesh := meshResources.Mesh
		meshName := mesh.GetMeta().GetName()

		// Secrets are generated in relation to a mesh so we need to create a new tracker
		secretsTracker := envoy_common.NewSecretsTracker(meshName, []string{meshName})

		internal, internalFCB, err := genInternalResources(proxy, xdsCtx, meshResources)
		if err != nil {
			return nil, err
		}
		rs.AddSet(internal)

		external, externalFCB, err := genExternalResources(proxy, meshResources, secretsTracker, unifiedNaming)
		if err != nil {
			return nil, err
		}
		rs.AddSet(external)

		for _, filterChain := range slices.Concat(internalFCB, externalFCB) {
			listener.Configure(envoy_listeners.FilterChain(filterChain))
		}

		// Envoy rejects listener with no filter chains, so there is no point in sending it
		if len(externalFCB) > 0 || len(internalFCB) > 0 {
			resource, err := listener.Build()
			if err != nil {
				return nil, err
			}

			rs.Add(&core_xds.Resource{
				Name:     resource.GetName(),
				Origin:   metadata.OriginEgress,
				Resource: resource,
			})
		}

		policyResources, err := g.PolicyGenerator.Generate(ctx, rs, xdsCtx, proxy)
		if err != nil {
			return nil, err
		}
		rs.AddSet(policyResources)

		secretResources, err := g.SecretGenerator.GenerateForZoneEgress(ctx, xdsCtx, proxy, secretsTracker, mesh)
		if err != nil {
			return nil, errors.Wrapf(
				err,
				"failed to generate secret resources for zone egress: %s",
				zoneEgress.GetMeta().GetName(),
			)
		}
		rs.AddSet(secretResources)
	}

	return rs, nil
}
