package generator

import (
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_secrets "github.com/kumahq/kuma/pkg/xds/envoy/secrets/v3"
	xds_secrets "github.com/kumahq/kuma/pkg/xds/secrets"
)

// OriginSecrets is a marker to indicate by which ProxyGenerator resources were generated.
const OriginSecrets = "secrets"

type SecretsProxyGenerator struct {
}

var _ ResourceGenerator = SecretsProxyGenerator{}

func createSecrets(
	resources *core_xds.ResourceSet,
	secrets xds_secrets.Secrets,
	proxy *core_xds.Proxy,
	mesh *core_mesh.MeshResource,
) error {
	identity, ca, err := secrets.Get(proxy.Dataplane, mesh)
	if err != nil {
		return err
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

	return nil
}

func (s SecretsProxyGenerator) Generate(
	ctx xds_context.Context,
	proxy *core_xds.Proxy,
) (*core_xds.ResourceSet, error) {
	resources := core_xds.NewResourceSet()

	if proxy.GetMeshes != nil {
		for _, mesh := range proxy.GetMeshes().Items {
			if !mesh.MTLSEnabled() {
				continue
			}

			if err := createSecrets(resources, ctx.ControlPlane.Secrets, proxy, mesh); err != nil {
				return nil, err
			}
		}

		return resources, nil
	}

	mesh := ctx.Mesh.Resource

	if !mesh.MTLSEnabled() {
		return nil, nil
	}

	if err := createSecrets(resources, ctx.ControlPlane.Secrets, proxy, mesh); err != nil {
		return nil, err
	}

	return resources, nil
}
