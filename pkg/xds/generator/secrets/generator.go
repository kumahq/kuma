package secrets

import (
	"context"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/naming"
	unified_naming "github.com/kumahq/kuma/pkg/core/naming/unified-naming"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/system_names"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_secrets "github.com/kumahq/kuma/pkg/xds/envoy/secrets/v3"
	"github.com/kumahq/kuma/pkg/xds/generator/metadata"
)

var generatorLogger = core.Log.WithName("secrets-generator")

type Generator struct{}

func createCaSecretResource(name string, ca *core_xds.CaSecret) *core_xds.Resource {
	caSecret := envoy_secrets.CreateCaSecret(ca, name)
	return &core_xds.Resource{
		Name:     caSecret.Name,
		Origin:   metadata.OriginSecrets,
		Resource: caSecret,
	}
}

func createIdentitySecretResource(name string, identity *core_xds.IdentitySecret) *core_xds.Resource {
	identitySecret := envoy_secrets.CreateIdentitySecret(identity, name)
	return &core_xds.Resource{
		Name:     identitySecret.Name,
		Origin:   metadata.OriginSecrets,
		Resource: identitySecret,
	}
}

// GenerateForZoneEgress generates whatever secrets were referenced in the
// zone egress config generation.
func (g Generator) GenerateForZoneEgress(
	ctx context.Context,
	xdsCtx xds_context.Context,
	proxy *core_xds.Proxy,
	secretsTracker core_xds.SecretsTracker,
	mesh *core_mesh.MeshResource,
) (*core_xds.ResourceSet, error) {
	if !mesh.MTLSEnabled() {
		return nil, nil
	}

	zoneEgress := proxy.ZoneEgressProxy.ZoneEgressResource
	log := generatorLogger.WithValues("ZoneEgress", zoneEgress.GetMeta().GetName())

	identity, ca, err := xdsCtx.ControlPlane.Secrets.GetForZoneEgress(ctx, zoneEgress, mesh)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate ZoneEgress secrets")
	}

	rs := core_xds.NewResourceSet()
	meshName := mesh.GetMeta().GetName()
	unifiedNaming := unified_naming.Enabled(proxy.Metadata, mesh)
	getName := naming.GetNameOrFallbackFunc(unifiedNaming)

	if secretsTracker.UsedIdentity() {
		log.V(1).Info("added identity", "mesh", meshName)

		identitySecretName := getName(
			system_names.AsSystemName("mtls_identity_"+meshName),
			secretsTracker.RequestIdentityCert().Name(),
		)

		rs.Add(createIdentitySecretResource(identitySecretName, identity))
	}

	if _, ok := secretsTracker.UsedCas()[meshName]; ok {
		log.V(1).Info("added mesh CA resources", "mesh", meshName)

		name := getName(
			system_names.AsSystemName("mtls_ca_"+meshName),
			secretsTracker.RequestCa(meshName).Name(),
		)

		rs.Add(createCaSecretResource(name, ca))
	}

	return rs, nil
}

// Generate uses the SecretsTracker on Proxy and
// generates whatever secrets were used in the config generation.
func (g Generator) Generate(
	ctx context.Context,
	_ *core_xds.ResourceSet,
	xdsCtx xds_context.Context,
	proxy *core_xds.Proxy,
) (*core_xds.ResourceSet, error) {
	resources := core_xds.NewResourceSet()

	log := generatorLogger.WithValues("proxyID", proxy.Id.String())

	if proxy.Dataplane != nil {
		log = log.WithValues("mesh", xdsCtx.Mesh.Resource.GetMeta().GetName())
	}
	getNameOrDefault := system_names.GetNameOrDefault(unified_naming.Enabled(proxy.Metadata, xdsCtx.Mesh.Resource))

	usedIdentity := proxy.SecretsTracker.UsedIdentity()
	usedCAs := proxy.SecretsTracker.UsedCas()
	usedAllInOne := proxy.SecretsTracker.UsedAllInOne()

	otherMeshes := xdsCtx.Mesh.Resources.OtherMeshes(xdsCtx.Mesh.Resource.GetMeta().GetName()).Items

	if usedAllInOne {
		identity, allInOneCa, err := xdsCtx.ControlPlane.Secrets.GetAllInOne(ctx, xdsCtx.Mesh.Resource, proxy.Dataplane, otherMeshes)
		if err != nil {
			return nil, errors.Wrap(err, "failed to generate all in one CA")
		}

		caSecretName := getNameOrDefault(
			system_names.AsSystemName("mtls_ca_all_meshes"),
			proxy.SecretsTracker.RequestAllInOneCa().Name(),
		)
		if len(xdsCtx.Mesh.TrustsByTrustDomain) == 0 {
			resources.Add(createCaSecretResource(caSecretName, allInOneCa))
		}
		identitySecretName := getNameOrDefault(
			system_names.AsSystemName("mtls_identity_"+proxy.SecretsTracker.RequestIdentityCert().MeshName()),
			proxy.SecretsTracker.RequestIdentityCert().Name(),
		)
		resources.Add(createIdentitySecretResource(identitySecretName, identity))
		log.V(1).Info("added all in one CA resources")
	}

	if usedIdentity || len(usedCAs) > 0 {
		var usedCAsMeshes []*core_mesh.MeshResource
		for _, otherMesh := range otherMeshes {
			if _, ok := usedCAs[otherMesh.GetMeta().GetName()]; ok {
				usedCAsMeshes = append(usedCAsMeshes, otherMesh)
			}
		}
		identity, generatedMeshCAs, err := xdsCtx.ControlPlane.Secrets.GetForDataPlane(ctx, proxy.Dataplane, xdsCtx.Mesh.Resource, usedCAsMeshes)
		if err != nil {
			return nil, errors.Wrap(err, "failed to generate dataplane identity cert and CAs")
		}

		identitySecretName := getNameOrDefault(
			system_names.AsSystemName("mtls_identity_"+proxy.SecretsTracker.RequestIdentityCert().MeshName()),
			proxy.SecretsTracker.RequestIdentityCert().Name(),
		)
		resources.Add(createIdentitySecretResource(identitySecretName, identity))

		var addedCas []string
		for mesh := range usedCAs {
			// when there is a MeshTrust we create a different secret which includes the mTLS Mesh CA
			if len(xdsCtx.Mesh.TrustsByTrustDomain) > 0 {
				break
			}
			identityName := getNameOrDefault(
				system_names.AsSystemName("mtls_ca_"+mesh),
				proxy.SecretsTracker.RequestCa(mesh).Name(),
			)
			if ca, ok := generatedMeshCAs[mesh]; ok {
				resources.Add(createCaSecretResource(identityName, ca))
			} else {
				// We need to add _something_ here so that Envoy syncs the
				// config
				emptyCa := &core_xds.CaSecret{}
				resources.Add(createCaSecretResource(identityName, emptyCa))
			}
			addedCas = append(addedCas, mesh)
		}
		log.V(1).Info("added identity and mesh CAs resources", "cas", addedCas)
	}

	return resources, nil
}
