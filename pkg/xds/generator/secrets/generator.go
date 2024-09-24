package secrets

import (
	"context"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/core"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_secrets "github.com/kumahq/kuma/pkg/xds/envoy/secrets/v3"
	generator_core "github.com/kumahq/kuma/pkg/xds/generator/core"
)

var generatorLogger = core.Log.WithName("secrets-generator")

// OriginSecrets is a marker to indicate by which ProxyGenerator resources were generated.
const OriginSecrets = "secrets"

type Generator struct{}

var _ generator_core.ResourceGenerator = Generator{}

func createCaSecretResource(name string, ca *core_xds.CaSecret) *core_xds.Resource {
	caSecret := envoy_secrets.CreateCaSecret(ca, name)
	return &core_xds.Resource{
		Name:     caSecret.Name,
		Origin:   OriginSecrets,
		Resource: caSecret,
	}
}

func createIdentitySecretResource(name string, identity *core_xds.IdentitySecret) *core_xds.Resource {
	identitySecret := envoy_secrets.CreateIdentitySecret(identity, name)
	return &core_xds.Resource{
		Name:     identitySecret.Name,
		Origin:   OriginSecrets,
		Resource: identitySecret,
	}
}

// GenerateForZoneEgress generates whatever secrets were referenced in the
// zone egress config generation.
func (g Generator) GenerateForZoneEgress(
	ctx context.Context,
	xdsCtx xds_context.Context,
	proxyId core_xds.ProxyId,
	zoneEgressResource *core_mesh.ZoneEgressResource,
	secretsTracker core_xds.SecretsTracker,
	mesh *core_mesh.MeshResource,
) (*core_xds.ResourceSet, error) {
	resources := core_xds.NewResourceSet()

	log := generatorLogger.WithValues("proxyID", proxyId.String())

	if !mesh.MTLSEnabled() {
		return nil, nil
	}

	meshName := mesh.GetMeta().GetName()

	usedIdentity := secretsTracker.UsedIdentity()

	identity, ca, err := xdsCtx.ControlPlane.Secrets.GetForZoneEgress(ctx, zoneEgressResource, mesh)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate ZoneEgress secrets")
	}

	if usedIdentity {
		log.V(1).Info("added identity", "mesh", meshName)
		resources.Add(createIdentitySecretResource(secretsTracker.RequestIdentityCert().Name(), identity))
	}

	if _, ok := secretsTracker.UsedCas()[meshName]; ok {
		log.V(1).Info("added mesh CA resources", "mesh", meshName)
		resources.Add(createCaSecretResource(secretsTracker.RequestCa(meshName).Name(), ca))
	}

	return resources, nil
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

	usedIdentity := proxy.SecretsTracker.UsedIdentity()
	usedCAs := proxy.SecretsTracker.UsedCas()
	usedAllInOne := proxy.SecretsTracker.UsedAllInOne()

	otherMeshes := xdsCtx.Mesh.Resources.OtherMeshes(xdsCtx.Mesh.Resource.GetMeta().GetName()).Items

	if usedAllInOne {
		identity, allInOneCa, err := xdsCtx.ControlPlane.Secrets.GetAllInOne(ctx, xdsCtx.Mesh.Resource, proxy.Dataplane, otherMeshes)
		if err != nil {
			return nil, errors.Wrap(err, "failed to generate all in one CA")
		}

		resources.Add(createCaSecretResource(proxy.SecretsTracker.RequestAllInOneCa().Name(), allInOneCa))
		resources.Add(createIdentitySecretResource(proxy.SecretsTracker.RequestIdentityCert().Name(), identity))
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

		resources.Add(createIdentitySecretResource(proxy.SecretsTracker.RequestIdentityCert().Name(), identity))

		var addedCas []string
		for mesh := range usedCAs {
			if ca, ok := generatedMeshCAs[mesh]; ok {
				resources.Add(createCaSecretResource(proxy.SecretsTracker.RequestCa(mesh).Name(), ca))
			} else {
				// We need to add _something_ here so that Envoy syncs the
				// config
				emptyCa := &core_xds.CaSecret{}
				resources.Add(createCaSecretResource(proxy.SecretsTracker.RequestCa(mesh).Name(), emptyCa))
			}
			addedCas = append(addedCas, mesh)
		}
		log.V(1).Info("added identity and mesh CAs resources", "cas", addedCas)
	}

	return resources, nil
}
