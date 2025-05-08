package v1alpha1

import (
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
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

func computeConf(toRules core_xds.ToRules, svc meshroute_xds.DestinationService, meshCtx xds_context.MeshContext) (*api.Rule, core_model.ResourceMeta) {
	// compute for old MeshService
	var tcpConf *api.Rule
	var origin core_model.ResourceMeta

	ruleTCP := toRules.Rules.Compute(subsetutils.MeshServiceElement(svc.ServiceName))
	if ruleTCP != nil {
		tcpConf = pointer.To(ruleTCP.Conf.(api.Rule))
		origin = impactfulMeta(ruleTCP.Origin)
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
func impactfulOrigin(os []common.Origin) core_model.ResourceMeta {
	if len(os) == 0 {
		return nil
	}
	return os[len(os)-1].Resource
}

// impactfulMeta returns the origin policy that contributed the backendRefs
func impactfulMeta(ms []core_model.ResourceMeta) core_model.ResourceMeta {
	if len(ms) == 0 {
		return nil
	}
	return ms[len(ms)-1]
}

func getBackendRefs(
	toRulesTCP core_xds.ToRules,
	toRulesHTTP core_xds.ToRules,
	svc meshroute_xds.DestinationService,
	protocol core_mesh.Protocol,
	meshCtx xds_context.MeshContext,
) []resolve.ResolvedBackendRef {
	tcpConf, origin := computeConf(toRulesTCP, svc, meshCtx)

	// If the outbounds protocol is http-like and there exists MeshHTTPRoute
	// with rule targeting the same MeshService as MeshTCPRoute, it should take
	// precedence over the latter
	switch protocol {
	case core_mesh.ProtocolHTTP, core_mesh.ProtocolHTTP2, core_mesh.ProtocolGRPC:
		// If we have an >= HTTP service, don't manage routing with
		// MeshTCPRoutes if we either don't have any MeshTCPRoutes or we have
		// MeshHTTPRoutes
		httpConf, _ := meshhttproute.ComputeHTTPRouteConf(toRulesHTTP, svc, meshCtx)
		if tcpConf == nil || httpConf != nil {
			return nil
		}
	default:
	}

	var backendRefs []resolve.ResolvedBackendRef
	if tcpConf != nil {
		for _, br := range pointer.Deref(tcpConf.Default.BackendRefs) {
			if resolved := resolve.BackendRefOrNil(origin, br, meshCtx.ResolveResourceIdentifier); resolved != nil {
				backendRefs = append(backendRefs, *resolved)
			}
		}
	} else {
		return []resolve.ResolvedBackendRef{*svc.DefaultBackendRef()}
	}

	return backendRefs
}
