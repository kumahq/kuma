package v1alpha1

import (
	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	meshhttproute_api "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshtcproute/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/util/pointer"
)

func getBackendRefs(
	toRulesTCP core_xds.Rules,
	toRulesHTTP core_xds.Rules,
	serviceName string,
	protocol core_mesh.Protocol,
	tags map[string]string,
) []common_api.BackendRef {
	serviceElement := core_xds.MeshServiceElement(serviceName)

	tcpConf := core_xds.ComputeConf[api.Rule](toRulesTCP, serviceElement)

	// If the outbounds protocol is http-like and there exists MeshHTTPRoute
	// with rule targeting the same MeshService as MeshTCPRoute, it should take
	// precedence over the latter
	switch protocol {
	case core_mesh.ProtocolHTTP, core_mesh.ProtocolHTTP2, core_mesh.ProtocolGRPC:
		// If we have an >= HTTP service, don't manage routing with
		// MeshTCPRoutes if we either don't have any MeshTCPRoutes or we have
		// MeshHTTPRoutes
		httpConf := core_xds.ComputeConf[meshhttproute_api.PolicyDefault](
			toRulesHTTP,
			serviceElement,
		)
		if tcpConf == nil || httpConf != nil {
			return nil
		}
	default:
	}

	if tcpConf != nil {
		return tcpConf.Default.BackendRefs
	}
	return []common_api.BackendRef{
		{
			TargetRef: common_api.TargetRef{
				Kind: common_api.MeshService,
				Name: serviceName,
				Tags: tags,
			},
			Weight: pointer.To(uint(100)),
		},
	}
}
