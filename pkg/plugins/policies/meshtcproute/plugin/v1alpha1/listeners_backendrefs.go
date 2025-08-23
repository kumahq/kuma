package v1alpha1

import (
	"fmt"

	"github.com/kumahq/kuma/pkg/core/kri"
	core_meta "github.com/kumahq/kuma/pkg/core/metadata"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/rules/common"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/rules/resolve"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/rules/subsetutils"
	meshroute_xds "github.com/kumahq/kuma/pkg/plugins/policies/core/xds/meshroute"
	meshhttproute "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/plugin/v1alpha1"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshtcproute/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/util/pointer"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
)

func computeConf(toRules core_xds.ToRules, svc meshroute_xds.DestinationService, meshCtx xds_context.MeshContext) (*api.Rule, common.Origin) {
	// compute for old MeshService
	var tcpConf *api.Rule
	var origin common.Origin

	ruleTCP := toRules.Rules.Compute(subsetutils.KumaServiceTagElement(svc.KumaServiceTagValue))
	if ruleTCP != nil {
		tcpConf = pointer.To(ruleTCP.Conf.(api.Rule))
		origin = impactfulOriginFromMeta(ruleTCP.Origin)
	}
	// check if there is configuration for real MeshService and prioritize it
	if r, ok := svc.Outbound.AssociatedServiceResource(); ok {
		resourceConf := toRules.ResourceRules.Compute(r, meshCtx.Resources)
		if resourceConf != nil && len(resourceConf.Conf) != 0 {
			tcpConf = pointer.To(resourceConf.Conf[0].(api.Rule))
			origin = impactfulOrigin(resourceConf.Origin)
		}
	}
	return tcpConf, origin
}

// impactfulOrigin returns the origin policy that contributed the backendRefs
func impactfulOrigin(os []common.Origin) common.Origin {
	if len(os) == 0 {
		return common.Origin{}
	}
	return os[len(os)-1]
}

// impactfulOriginFromMeta returns the origin policy that contributed the backendRefs
func impactfulOriginFromMeta(ms []core_model.ResourceMeta) common.Origin {
	if len(ms) == 0 {
		return common.Origin{}
	}
	return common.Origin{Resource: ms[len(ms)-1]}
}

func getBackendRefs(
	toRulesTCP core_xds.ToRules,
	toRulesHTTP core_xds.ToRules,
	svc meshroute_xds.DestinationService,
	meshCtx xds_context.MeshContext,
) []resolve.ResolvedBackendRef {
	tcpConf, origin := computeConf(toRulesTCP, svc, meshCtx)

	// If the outbounds protocol is http-like and there exists MeshHTTPRoute
	// with rule targeting the same MeshService as MeshTCPRoute, it should take
	// precedence over the latter
	if core_meta.IsHTTPBased(svc.Protocol) {
		// If we have an >= HTTP service, don't manage routing with
		// MeshTCPRoutes if we either don't have any MeshTCPRoutes or we have
		// MeshHTTPRoutes
		httpConf, _ := meshhttproute.ComputeHTTPRouteConf(toRulesHTTP, svc, meshCtx)
		if tcpConf == nil || httpConf != nil {
			return nil
		}
	}

	if tcpConf == nil {
		return []resolve.ResolvedBackendRef{*svc.DefaultBackendRef()}
	}

	originID := kri.WithSectionName(
		kri.FromResourceMeta(origin.Resource, api.MeshTCPRouteType),
		fmt.Sprintf("rule_%d", origin.RuleIndex),
	)

	var resolved []resolve.ResolvedBackendRef
	for _, ref := range pointer.Deref(tcpConf.Default.BackendRefs) {
		if r, ok := resolve.BackendRef(originID, ref, meshCtx.ResolveResourceIdentifier); ok {
			resolved = append(resolved, r)
		}
	}

	return resolved
}
