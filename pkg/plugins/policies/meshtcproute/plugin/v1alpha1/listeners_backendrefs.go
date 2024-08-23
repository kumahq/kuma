package v1alpha1

import (
	"fmt"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	meshhttproute_api "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshtcproute/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/util/pointer"
)

func getBackendRefs(toRulesTCP core_xds.Rules, toRulesHTTP core_xds.Rules, serviceName string, protocol core_mesh.Protocol, fallbackBackendRef core_model.ResolvedBackendRef) []core_model.ResolvedBackendRef {
	service := core_xds.MeshService(serviceName)

	ruleTCP := toRulesTCP.Compute(service)
	var tcpConf *api.Rule
	if ruleTCP != nil {
		tcpConf = pointer.To(ruleTCP.Conf.(api.Rule))
	}

	// If the outbounds protocol is http-like and there exists MeshHTTPRoute
	// with rule targeting the same MeshService as MeshTCPRoute, it should take
	// precedence over the latter
	switch protocol {
	case core_mesh.ProtocolHTTP, core_mesh.ProtocolHTTP2, core_mesh.ProtocolGRPC:
		// If we have an >= HTTP service, don't manage routing with
		// MeshTCPRoutes if we either don't have any MeshTCPRoutes or we have
		// MeshHTTPRoutes
		ruleHTTP := toRulesHTTP.Compute(service)
		var httpConf *meshhttproute_api.PolicyDefault
		if ruleHTTP != nil {
			httpConf = pointer.To(ruleHTTP.Conf.(meshhttproute_api.PolicyDefault))
		}

		if tcpConf == nil || httpConf != nil {
			return nil
		}
	default:
	}

	var backendRefs []core_model.ResolvedBackendRef
	if tcpConf != nil {
		for _, br := range tcpConf.Default.BackendRefs {
			if origin, ok := ruleTCP.GetBackendRefOrigin(core_xds.EmptyMatches); ok {
				backendRefs = append(backendRefs, resolveBackendRef(origin, br))
			} else {
				backendRefs = append(backendRefs, core_model.ResolvedBackendRef{LegacyBackendRef: &br})
			}
		}
	} else {
		return []core_model.ResolvedBackendRef{fallbackBackendRef}
	}

	return backendRefs
}

func resolveBackendRef(meta core_model.ResourceMeta, br common_api.BackendRef) core_model.ResolvedBackendRef {
	resolved := core_model.ResolvedBackendRef{LegacyBackendRef: &br}

	switch {
	case br.Kind == common_api.MeshService && br.ReferencesRealObject():
	case br.Kind == common_api.MeshExternalService:
	case br.Kind == common_api.MeshMultiZoneService:
	default:
		return resolved
	}

	resolved.Resource = &core_model.TypedResourceIdentifier{
		ResourceIdentifier: core_model.TargetRefToResourceIdentifier(meta, br.TargetRef),
		ResourceType:       core_model.ResourceType(br.Kind),
		SectionName:        fmt.Sprintf("%d", br.Port),
	}
	return resolved
}
