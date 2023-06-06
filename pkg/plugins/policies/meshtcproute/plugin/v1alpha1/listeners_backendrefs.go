package v1alpha1

import (
	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	meshhttproute_api "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshtcproute/api/v1alpha1"
)

func matchingHTTPRuleExist(
	toRulesHTTP core_xds.Rules,
	service core_xds.Subset,
	protocol core_mesh.Protocol,
) bool {
	switch protocol {
	case core_mesh.ProtocolHTTP, core_mesh.ProtocolHTTP2, core_mesh.ProtocolGRPC:
	default:
		return false
	}

	return core_xds.ComputeConf[meshhttproute_api.PolicyDefault](
		toRulesHTTP,
		service,
	) != nil
}

func getTCPBackendRefs(
	toRulesTCP core_xds.Rules,
	service core_xds.Subset,
) []common_api.BackendRef {
	conf := core_xds.ComputeConf[api.Rule](toRulesTCP, service)
	if conf != nil {
		return conf.Default.BackendRefs
	}

	return nil
}

func getBackendRefs(
	toRulesTCP core_xds.Rules,
	toRulesHTTP core_xds.Rules,
	serviceName string,
	protocol core_mesh.Protocol,
) []common_api.BackendRef {
	service := core_xds.MeshService(serviceName)

	// If the outbounds protocol is http-like and there exists MeshHTTPRoute
	// with rule targeting the same MeshService as MeshTCPRoute, it should take
	// precedence over the latter
	if matchingHTTPRuleExist(toRulesHTTP, service, protocol) {
		return nil
	}

	return getTCPBackendRefs(toRulesTCP, service)
}
