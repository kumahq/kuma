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

func addSecretsToResources(
	mesh *core_mesh.MeshResource,
	identity *core_xds.IdentitySecret,
	ca *core_xds.CaSecret,
	resources *core_xds.ResourceSet,
) *core_xds.ResourceSet {
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

	return addSecretsToResources(
		mesh,
		identity,
		ca,
		core_xds.NewResourceSet(),
	), nil
}

func createZoneEgressSecrets(
	ctx xds_context.Context,
	proxy *core_xds.Proxy,
) (*core_xds.ResourceSet, error) {
	var resources *core_xds.ResourceSet

	zoneEgressProxy := proxy.ZoneEgressProxy

	for _, mesh := range zoneEgressProxy.Meshes {
		if !mesh.MTLSEnabled() {
			continue
		}

		resources = core_xds.NewResourceSet()

		identity, ca, err := ctx.ControlPlane.Secrets.GetForZoneEgress(
			zoneEgressProxy.ZoneEgressResource,
			mesh,
		)
		if err != nil {
			return nil, err
		}

		addSecretsToResources(mesh, identity, ca, resources)
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
