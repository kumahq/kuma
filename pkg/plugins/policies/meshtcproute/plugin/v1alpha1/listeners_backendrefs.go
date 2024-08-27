package v1alpha1

import (
	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	meshroute_xds "github.com/kumahq/kuma/pkg/plugins/policies/core/xds/meshroute"
	meshhttproute "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/plugin/v1alpha1"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshtcproute/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/util/pointer"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
)

func computeConf(toRules core_xds.ToRules, svc meshroute_xds.DestinationService, meshCtx xds_context.MeshContext) *api.Rule {
	// compute for old MeshService
	conf := core_xds.ComputeConf[api.Rule](toRules.Rules, core_xds.MeshService(svc.ServiceName))
	// check if there is configuration for real MeshService and prioritize it
	if svc.OwnerResource != nil {
		resourceConf := toRules.ResourceRules.Compute(*svc.OwnerResource, meshCtx.Resources)
		if resourceConf != nil && len(resourceConf.Conf) != 0 {
			conf = pointer.To(resourceConf.Conf[0].(api.Rule))
		}
	}
	return conf
}

func getBackendRefs(toRulesTCP core_xds.ToRules, toRulesHTTP core_xds.ToRules, svc meshroute_xds.DestinationService, protocol core_mesh.Protocol, backendRef common_api.BackendRef, meshCtx xds_context.MeshContext) []common_api.BackendRef {
	tcpConf := computeConf(toRulesTCP, svc, meshCtx)

	// If the outbounds protocol is http-like and there exists MeshHTTPRoute
	// with rule targeting the same MeshService as MeshTCPRoute, it should take
	// precedence over the latter
	switch protocol {
	case core_mesh.ProtocolHTTP, core_mesh.ProtocolHTTP2, core_mesh.ProtocolGRPC:
		// If we have an >= HTTP service, don't manage routing with
		// MeshTCPRoutes if we either don't have any MeshTCPRoutes or we have
		// MeshHTTPRoutes
		httpConf := meshhttproute.ComputeHTTPRouteConf(toRulesHTTP, svc, meshCtx)
		if tcpConf == nil || httpConf != nil {
			return nil
		}
	default:
	}

	if tcpConf != nil {
		return tcpConf.Default.BackendRefs
	}
	return []common_api.BackendRef{
		backendRef,
	}
}
