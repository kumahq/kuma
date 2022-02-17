package generator

import (
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_secrets "github.com/kumahq/kuma/pkg/xds/envoy/secrets/v3"
)

// OriginSecrets is a marker to indicate by which ProxyGenerator resources were generated.
const OriginSecrets = "secrets"

type SecretsProxyGenerator struct {
}

var _ ResourceGenerator = SecretsProxyGenerator{}

func createSecretResources(
	mesh *core_mesh.MeshResource,
	identity *core_xds.IdentitySecret,
	ca *core_xds.CaSecret,
) *core_xds.ResourceSet {
	resources := core_xds.NewResourceSet()

	meshName := mesh.GetMeta().GetName()
	identitySecret := envoy_secrets.CreateIdentitySecret(identity, meshName)
	caSecret := envoy_secrets.CreateCaSecret(ca, meshName)

	resources.Add(
		&core_xds.Resource{
			Name:     identitySecret.Name,
			Origin:   OriginSecrets,
			Resource: identitySecret,
		},
		&core_xds.Resource{
			Name:     caSecret.Name,
			Origin:   OriginSecrets,
			Resource: caSecret,
		},
	)

	return resources
}

func createDataPlaneProxySecrets(
	ctx xds_context.Context,
	proxy *core_xds.Proxy,
) (*core_xds.ResourceSet, error) {
	mesh := ctx.Mesh.Resource

	if !mesh.MTLSEnabled() {
		return nil, nil
	}

	identity, ca, err := ctx.ControlPlane.Secrets.GetForDataPlane(
		proxy.Dataplane,
		mesh,
	)
	if err != nil {
		return nil, err
	}

	return createSecretResources(mesh, identity, ca), nil
}

func createZoneEgressSecrets(
	ctx xds_context.Context,
	proxy *core_xds.Proxy,
) (*core_xds.ResourceSet, error) {
	resources := core_xds.NewResourceSet()

	for _, meshResources := range proxy.ZoneEgressProxy.MeshResourcesList {
		if !meshResources.Mesh.MTLSEnabled() {
			continue
		}

		if len(meshResources.ExternalServices) == 0 {
			// https://github.com/envoyproxy/envoy/issues/9310
			// Envoy has a behavior that if you include a secret over ADS that is not referenced anywhere
			// all secrets are stuck in warming state.
			// We need to only deliver secrets that are used in other parts of config.
			// ZoneEgress only use identity and CA certs for external services.
			// For internal services it just passes the traffic to ZoneIngress through SNI
			continue
		}

		identity, ca, err := ctx.ControlPlane.Secrets.GetForZoneEgress(
			proxy.ZoneEgressProxy.ZoneEgressResource,
			meshResources.Mesh,
		)
		if err != nil {
			return nil, err
		}

		resources.AddSet(createSecretResources(meshResources.Mesh, identity, ca))
	}

	return resources, nil
}

func (s SecretsProxyGenerator) Generate(
	ctx xds_context.Context,
	proxy *core_xds.Proxy,
) (*core_xds.ResourceSet, error) {
	if proxy.ZoneEgressProxy != nil {
		return createZoneEgressSecrets(ctx, proxy)
	}

	return createDataPlaneProxySecrets(ctx, proxy)
}
