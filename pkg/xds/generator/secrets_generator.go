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

func createSecrets(
	ctx xds_context.Context,
	proxy *core_xds.Proxy,
	meshes []*core_mesh.MeshResource,
) (*core_xds.ResourceSet, error) {
	var resources *core_xds.ResourceSet

	for _, mesh := range meshes {
		if !mesh.MTLSEnabled() {
			continue
		}

		if resources == nil {
			resources = core_xds.NewResourceSet()
		}

		identity, ca, err := ctx.ControlPlane.Secrets.Get(proxy.Dataplane, mesh)
		if err != nil {
			return nil, err
		}

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
	}

	return resources, nil
}

func (s SecretsProxyGenerator) Generate(
	ctx xds_context.Context,
	proxy *core_xds.Proxy,
) (*core_xds.ResourceSet, error) {
	if proxy.ZoneEgressProxy != nil {
		return createSecrets(ctx, proxy, proxy.ZoneEgressProxy.Meshes)
	}

	return createSecrets(ctx, proxy, []*core_mesh.MeshResource{ctx.Mesh.Resource})
}
