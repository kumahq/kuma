package generator

import (
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_secrets "github.com/kumahq/kuma/pkg/xds/envoy/secrets/v3"
)

// OriginSecrets is a marker to indicate by which ProxyGenerator resources were generated.
const OriginSecrets = "secrets"

type SecretsProxyGenerator struct {
}

var _ ResourceGenerator = SecretsProxyGenerator{}

func (s SecretsProxyGenerator) Generate(ctx xds_context.Context, proxy *core_xds.Proxy) (*core_xds.ResourceSet, error) {
	if !ctx.Mesh.Resource.MTLSEnabled() {
		return nil, nil
	}
	identity, ca, err := ctx.ControlPlane.Secrets.Get(proxy.Dataplane, ctx.Mesh.Resource)
	if err != nil {
		return nil, err
	}

	resources := core_xds.NewResourceSet()
	identitySecret := envoy_secrets.CreateIdentitySecret(identity)
	caSecret := envoy_secrets.CreateCaSecret(ca)
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
	return resources, nil
}
