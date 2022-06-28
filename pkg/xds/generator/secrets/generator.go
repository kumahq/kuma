package generator

import (
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/core"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_secrets "github.com/kumahq/kuma/pkg/xds/envoy/secrets/v3"
	generator_core "github.com/kumahq/kuma/pkg/xds/generator/core"
)

// OriginSecrets is a marker to indicate by which ProxyGenerator resources were generated.
const OriginSecrets = "secrets"

type Generator struct {
}

var _ generator_core.ResourceGenerator = Generator{}

func CreateCaSecretResource(name string, ca *core_xds.CaSecret) *core_xds.Resource {
	caSecret := envoy_secrets.CreateCaSecret(ca, name)
	return &core_xds.Resource{
		Name:     caSecret.Name,
		Origin:   OriginSecrets,
		Resource: caSecret,
	}
}

func CreateIdentitySecretResource(name string, identity *core_xds.IdentitySecret) *core_xds.Resource {
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
	ctx xds_context.Context,
	proxyId core_xds.ProxyId,
	zoneEgressResource *core_mesh.ZoneEgressResource,
	secretsTracker core_xds.SecretsTracker,
	mesh *core_mesh.MeshResource,
) (*core_xds.ResourceSet, error) {
	resources := core_xds.NewResourceSet()

	log := core.Log.WithName("secrets-generator").WithValues("proxyID", proxyId.String())

	if !mesh.MTLSEnabled() {
		return nil, nil
	}

	meshName := mesh.GetMeta().GetName()

	usedIdentity := secretsTracker.UsedIdentity()

	identity, ca, err := ctx.ControlPlane.Secrets.GetForZoneEgress(zoneEgressResource, mesh)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate ZoneEgress secrets")
	}

	if usedIdentity {
		log.V(1).Info("added identity", "mesh", meshName)
		resources.Add(CreateIdentitySecretResource(secretsTracker.RequestIdentityCert().Name(), identity))
	}

	if _, ok := secretsTracker.UsedCas()[meshName]; ok {
		log.V(1).Info("added mesh CA resources", "mesh", meshName)
		resources.Add(CreateCaSecretResource(secretsTracker.RequestCa(meshName).Name(), ca))
	}

	return resources, nil
}

// Generate uses the SecretsTracker on Proxy and
// generates whatever secrets were used in the config generation.
func (g Generator) Generate(
	ctx xds_context.Context,
	proxy *core_xds.Proxy,
) (*core_xds.ResourceSet, error) {
	resources := core_xds.NewResourceSet()

	log := core.Log.WithName("secrets-generator").WithValues("proxyID", proxy.Id.String())

	if proxy.Dataplane != nil {
		log = log.WithValues("mesh", ctx.Mesh.Resource.GetMeta().GetName())
	}

	usedIdentity := proxy.SecretsTracker.UsedIdentity()
	usedCas := proxy.SecretsTracker.UsedCas()
	usedAllInOne := proxy.SecretsTracker.UsedAllInOne()

	if usedAllInOne {
		otherMeshes := ctx.Mesh.Resources.OtherMeshes().Items
		identity, allInOneCa, err := ctx.ControlPlane.Secrets.GetAllInOne(ctx.Mesh.Resource, proxy.Dataplane, otherMeshes)
		if err != nil {
			return nil, errors.Wrap(err, "failed to generate all in one CA")
		}

		resources.Add(CreateCaSecretResource(proxy.SecretsTracker.RequestAllInOneCa().Name(), allInOneCa))
		resources.Add(CreateIdentitySecretResource(proxy.SecretsTracker.RequestIdentityCert().Name(), identity))
		log.V(1).Info("added all in one CA resources")
	}

	if usedIdentity || len(usedCas) > 0 {
		otherMeshes := ctx.Mesh.Resources.OtherMeshes().Items
		identity, meshCas, err := ctx.ControlPlane.Secrets.GetForDataPlane(proxy.Dataplane, ctx.Mesh.Resource, otherMeshes)

		if err != nil {
			return nil, errors.Wrap(err, "failed to generate dataplane identity cert and CAs")
		}

		resources.Add(CreateIdentitySecretResource(proxy.SecretsTracker.RequestIdentityCert().Name(), identity))

		var addedCas []string
		for mesh := range usedCas {
			if ca, ok := meshCas[mesh]; ok {
				resources.Add(CreateCaSecretResource(proxy.SecretsTracker.RequestCa(mesh).Name(), ca))
			} else {
				// We need to add _something_ here so that Envoy syncs the
				// config
				emptyCa := &core_xds.CaSecret{}
				resources.Add(CreateCaSecretResource(proxy.SecretsTracker.RequestCa(mesh).Name(), emptyCa))
			}
			addedCas = append(addedCas, mesh)
		}
		log.V(1).Info("added identity and mesh CAs resources", "cas", addedCas)
	}

	return resources, nil
}
