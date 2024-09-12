package v1alpha1

import (
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
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

	ruleTCP := toRules.Rules.Compute(core_xds.MeshService(svc.ServiceName))
	if ruleTCP != nil {
		tcpConf = pointer.To(ruleTCP.Conf.(api.Rule))
		if o, ok := ruleTCP.GetBackendRefOrigin(core_xds.EmptyMatches); ok {
			origin = o
		}
	}
	// check if there is configuration for real MeshService and prioritize it
	if svc.Resource != nil {
		resourceConf := toRules.ResourceRules.Compute(*svc.Resource, meshCtx.Resources)
		if resourceConf != nil && len(resourceConf.Conf) != 0 {
			tcpConf = pointer.To(resourceConf.Conf[0].(api.Rule))
			if o, ok := resourceConf.GetBackendRefOrigin(core_xds.EmptyMatches); ok {
				origin = o
			}
		}
	}
	return tcpConf, origin
}

func getBackendRefs(
	toRulesTCP core_xds.ToRules,
	toRulesHTTP core_xds.ToRules,
	svc meshroute_xds.DestinationService,
	protocol core_mesh.Protocol,
	fallbackBackendRef core_model.ResolvedBackendRef,
	meshCtx xds_context.MeshContext,
) []core_model.ResolvedBackendRef {
	tcpConf, backendRefOrigin := computeConf(toRulesTCP, svc, meshCtx)

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

	var backendRefs []core_model.ResolvedBackendRef
	if tcpConf != nil {
		for _, br := range tcpConf.Default.BackendRefs {
			if backendRefOrigin != nil {
				if resolved := core_model.ResolveBackendRef(backendRefOrigin, br, meshCtx.ResolveResourceIdentifier); resolved != nil {
					backendRefs = append(backendRefs, *resolved)
				}
			} else {
				backendRefs = append(backendRefs, core_model.ResolvedBackendRef{LegacyBackendRef: &br})
			}
		}
	} else {
		return []core_model.ResolvedBackendRef{fallbackBackendRef}
	}

	return backendRefs
}
