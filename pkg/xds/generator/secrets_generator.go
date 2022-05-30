package generator

import (
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/core"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_secrets "github.com/kumahq/kuma/pkg/xds/envoy/secrets/v3"
	generator_core "github.com/kumahq/kuma/pkg/xds/generator/core"
)

// OriginSecrets is a marker to indicate by which ProxyGenerator resources were generated.
const OriginSecrets = "secrets"

type SecretsProxyGenerator struct {
}

var _ generator_core.ResourceGenerator = SecretsProxyGenerator{}

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

// GenerateSecretsFromTracker takes uses the SecretsTracker on Proxy and
// generates whatever secrets were used in the config generation.
func (g SecretsProxyGenerator) Generate(
	ctx xds_context.Context,
	proxy *core_xds.Proxy,
) (*core_xds.ResourceSet, error) {
	resources := core_xds.NewResourceSet()

	log := core.Log.WithName("secrets-generator")
	switch {
	case proxy.Dataplane != nil:
		log = log.WithValues("proxy", proxy.Dataplane.GetMeta().GetName())
	case proxy.ZoneIngressProxy != nil:
		log = log.WithValues("zoneingress", proxy.ZoneIngress.GetMeta().GetName())
	case proxy.ZoneEgressProxy != nil:
		log = log.WithValues("zoneingress", proxy.ZoneEgressProxy.ZoneEgressResource.GetMeta().GetName())
	}

	// We don't have a secrets tracker if we don't have a mesh (zone ingress/egress)
	if proxy.SecretsTracker == nil {
		return nil, nil
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
		var identity *core_xds.IdentitySecret
		meshCas := map[string]*core_xds.CaSecret{}
		var err error

		if proxy.ZoneEgressProxy != nil {
			for _, meshResources := range proxy.ZoneEgressProxy.MeshResourcesList {
				if !meshResources.Mesh.MTLSEnabled() {
					continue
				}

				// TODO this method could use a refactor, the identity secret remains the same
				zoneEgressIdentity, ca, err := ctx.ControlPlane.Secrets.GetForZoneEgress(proxy.ZoneEgressProxy.ZoneEgressResource, meshResources.Mesh)
				if err != nil {
					return nil, errors.Wrap(err, "failed to generate ZoneEgress secrets")
				}

				identity = zoneEgressIdentity

				meshCas[meshResources.Mesh.GetMeta().GetName()] = ca
			}
		} else {
			otherMeshes := ctx.Mesh.Resources.OtherMeshes().Items
			identity, meshCas, err = ctx.ControlPlane.Secrets.GetForDataPlane(proxy.Dataplane, ctx.Mesh.Resource, otherMeshes)
		}

		if err != nil {
			return nil, errors.Wrap(err, "failed to generate dataplane identity cert and CAs")
		}

		resources.Add(CreateIdentitySecretResource(proxy.SecretsTracker.RequestIdentityCert().Name(), identity))

		var addedCas []string
		for mesh, ca := range meshCas {
			if _, ok := usedCas[mesh]; ok {
				addedCas = append(addedCas, mesh)
				resources.Add(CreateCaSecretResource(proxy.SecretsTracker.RequestCa(mesh).Name(), ca))
			}
		}
		log.V(1).Info("added identity and mesh CAs resources", "cas", addedCas)
	}

	return resources, nil
}
