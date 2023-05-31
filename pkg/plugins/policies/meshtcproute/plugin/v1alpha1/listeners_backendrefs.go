package v1alpha1

import (
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
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

	for _, httpRule := range toRulesHTTP {
		if httpRule.Subset.IsSubset(service) {
			return true
		}
	}

	return false
}

func getTCPBackendRefs(
	toRulesTCP core_xds.Rules,
	service core_xds.Subset,
) []api.BackendRef {
	for _, tcpRule := range toRulesTCP {
		if tcpRule.Subset.IsSubset(service) {
			return tcpRule.Conf.(api.Rule).Default.BackendRefs
		}
	}

	return nil
}

func getBackendRefs(
	toRulesTCP core_xds.Rules,
	toRulesHTTP core_xds.Rules,
	serviceName string,
	protocol core_mesh.Protocol,
) []api.BackendRef {
	service := core_xds.MeshService(serviceName)

	// If the outbounds protocol is http-like and there exists MeshHTTPRoute
	// with rule targeting the same MeshService as MeshTCPRoute, it should take
	// precedence over the latter
	if matchingHTTPRuleExist(toRulesHTTP, service, protocol) {
		return nil
	}

	return getTCPBackendRefs(toRulesTCP, service)
}
