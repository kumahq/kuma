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

type MeshCa struct {
	Mesh     string
	CaSecret *core_xds.CaSecret
}

func CreateSecretResources(
	mesh *core_mesh.MeshResource,
	identity *core_xds.IdentitySecret,
	cas []xds_secrets.MeshCa,
) *core_xds.ResourceSet {
	resources := core_xds.NewResourceSet()

	meshName := mesh.GetMeta().GetName()
	identitySecret := envoy_secrets.CreateIdentitySecret(identity, meshName)

	var caResources []*core_xds.Resource
	for _, ca := range cas {
		caSecret := envoy_secrets.CreateCaSecret(ca.CaSecret, ca.Mesh)
		caResources = append(
			caResources,
			&core_xds.Resource{
				Name:     caSecret.Name,
				Origin:   OriginSecrets,
				Resource: caSecret,
			},
		)
	}

	resources.Add(
		&core_xds.Resource{
			Name:     identitySecret.Name,
			Origin:   OriginSecrets,
			Resource: identitySecret,
		},
	)
	resources.Add(
		caResources...,
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

	var otherMeshes []*core_mesh.MeshResource
	for _, otherMesh := range ctx.Mesh.Resources.CrossMeshGateways(mesh) {
		if otherMesh.Mesh.GetMeta().GetName() != mesh.GetMeta().GetName() {
			otherMeshes = append(otherMeshes, otherMesh.Mesh)
		}
	}

	identity, cas, err := ctx.ControlPlane.Secrets.GetForDataPlane(
		proxy.Dataplane,
		mesh,
		otherMeshes,
	)
	if err != nil {
		return nil, err
	}

	return CreateSecretResources(mesh, identity, cas), nil
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

		identity, cas, err := ctx.ControlPlane.Secrets.GetForZoneEgress(
			proxy.ZoneEgressProxy.ZoneEgressResource,
			meshResources.Mesh,
		)
		if err != nil {
			return nil, err
		}

		resources.AddSet(CreateSecretResources(meshResources.Mesh, identity, cas))
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
