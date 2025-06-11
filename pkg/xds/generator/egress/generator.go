package egress

import (
	"context"

	envoy_listener_v3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/kri"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/generator"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	envoy_listeners "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
	generator_secrets "github.com/kumahq/kuma/pkg/xds/generator/secrets"
)

const (
	EgressProxy = "egress-proxy"

	// OriginEgress is a marker to indicate by which ProxyGenerator resources
	// were generated.
	OriginEgress = "egress"
)

var log = core.Log.WithName("xds").WithName("egress-proxy-generator")

// ZoneEgressGenerator is responsible for generating xDS resources for
// a single ZoneEgress.
type ZoneEgressGenerator interface {
	Generate(context.Context, xds_context.Context, *core_xds.Proxy, *envoy_listeners.ListenerBuilder, *core_xds.MeshResources) (*core_xds.ResourceSet, error)
}

// Generator generates xDS resources for an entire ZoneEgress.
type Generator struct {
	// These generators add to the listener builder
	ZoneEgressGenerators []ZoneEgressGenerator
	// These generators depend on the config being built
	SecretGenerator *generator_secrets.Generator
}

func makeListenerBuilder(
	apiVersion core_xds.APIVersion,
	zoneEgress *core_mesh.ZoneEgressResource,
) *envoy_listeners.ListenerBuilder {
	networking := zoneEgress.Spec.GetNetworking()

	address := networking.GetAddress()
	port := networking.GetPort()

	name := kri.From(zoneEgress, "").String()

	return envoy_listeners.
		NewInboundListenerBuilder(apiVersion, address, port, core_xds.SocketAddressProtocolTCP).
		Configure(envoy_listeners.TLSInspector()).
		WithOverwriteName(name)
}

func (g Generator) Generate(
	ctx context.Context,
	_ *core_xds.ResourceSet,
	xdsCtx xds_context.Context,
	proxy *core_xds.Proxy,
) (*core_xds.ResourceSet, error) {
	resources := core_xds.NewResourceSet()

	listenerBuilder := makeListenerBuilder(
		proxy.APIVersion,
		proxy.ZoneEgressProxy.ZoneEgressResource,
	)

	for _, meshResources := range proxy.ZoneEgressProxy.MeshResourcesList {
		meshName := meshResources.Mesh.GetMeta().GetName()

		// Secrets are generated in relation to a mesh so we need to create a new tracker
		secretsTracker := envoy_common.NewSecretsTracker(meshName, []string{meshName})
		proxy.SecretsTracker = secretsTracker

		for _, zoneEgressGenerator := range g.ZoneEgressGenerators {
			rs, err := zoneEgressGenerator.Generate(ctx, xdsCtx, proxy, listenerBuilder, meshResources)
			if err != nil {
				err := errors.Wrapf(
					err,
					"%T failed to generate resources for zone egress %q",
					zoneEgressGenerator,
					proxy.Id,
				)
				return nil, err
			}

			resources.AddSet(rs)
		}

		listener, err := listenerBuilder.Build()
		if err != nil {
			return nil, err
		}
		core.Log.Info("check listener", "len(listener.(*envoy_listener_v3.Listener).FilterChains)", len(listener.(*envoy_listener_v3.Listener).FilterChains))
		if len(listener.(*envoy_listener_v3.Listener).FilterChains) > 0 {
			// Envoy rejects listener with no filter chains, so there is no point in sending it.
			resources.Add(&core_xds.Resource{
				Name:     listener.GetName(),
				Origin:   OriginEgress,
				Resource: listener,
			})
		}

		rs, err := generator.NewGenerator().Generate(ctx, resources, xdsCtx, proxy)
		if err != nil {
			return nil, err
		}
		resources.AddSet(rs)

		rs, err = g.SecretGenerator.GenerateForZoneEgress(
			ctx, xdsCtx, proxy.Id, proxy.ZoneEgressProxy.ZoneEgressResource, secretsTracker, meshResources.Mesh,
		)
		if err != nil {
			err := errors.Wrapf(
				err,
				"%T failed to generate resources for zone egress %q",
				g.SecretGenerator,
				proxy.Id,
			)
			return nil, err
		}

		resources.AddSet(rs)
	}
	return resources, nil
}
