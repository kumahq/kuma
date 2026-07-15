package topology

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	meshexternalservice_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshexternalservice/api/v1alpha1"
	meshmzservice_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshmultizoneservice/api/v1alpha1"
	meshservice_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshservice/api/v1alpha1"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	xds_types "github.com/kumahq/kuma/v3/pkg/core/xds/types"
)

const legacyMeshDomain = "mesh"

// LegacyDomains restores legacy service-tag DNS aliases from the new
// MeshService-family resources so dataplanes and legacy zone proxies keep
// resolving the old hostnames while xDS no longer depends on VirtualOutbound.
func LegacyDomains(
	meshServices []*meshservice_api.MeshServiceResource,
	meshExternalServices []*meshexternalservice_api.MeshExternalServiceResource,
	meshMultiZoneServices []*meshmzservice_api.MeshMultiZoneServiceResource,
) []xds_types.VIPDomains {
	domainsByAddress := map[string]map[string]struct{}{}

	add := func(address string, domains ...string) {
		if address == "" {
			return
		}
		entries, ok := domainsByAddress[address]
		if !ok {
			entries = map[string]struct{}{}
			domainsByAddress[address] = entries
		}
		for _, domain := range domains {
			if domain == "" {
				continue
			}
			entries[domain] = struct{}{}
		}
	}

	for _, ms := range meshServices {
		add(firstMeshServiceVIP(ms), legacyMeshServiceDomains(ms)...)
	}
	for _, mes := range meshExternalServices {
		add(firstMeshExternalServiceVIP(mes), legacyMeshExternalServiceDomains(mes)...)
	}
	for _, mzs := range meshMultiZoneServices {
		add(firstMeshMultiZoneServiceVIP(mzs), legacyMeshMultiZoneServiceDomains(mzs)...)
	}

	addresses := slices.Sorted(maps.Keys(domainsByAddress))

	var result []xds_types.VIPDomains
	for _, address := range addresses {
		domains := slices.Sorted(maps.Keys(domainsByAddress[address]))
		result = append(result, xds_types.VIPDomains{
			Address: address,
			Domains: domains,
		})
	}

	return result
}

func legacyMeshServiceDomains(ms *meshservice_api.MeshServiceResource) []string {
	displayName := core_model.GetDisplayName(ms.GetMeta())
	if displayName == "" {
		return nil
	}

	namespace := meshServiceNamespace(ms)
	if ms.GetMeta().GetLabels()[mesh_proto.EnvTag] != mesh_proto.KubernetesEnvironment || namespace == "" {
		return []string{legacyMeshServiceName(displayName, namespace, 0) + "." + legacyMeshDomain}
	}

	unique := map[string]struct{}{}
	for _, port := range ms.Spec.Ports {
		serviceTag := legacyMeshServiceName(displayName, namespace, port.Port)
		unique[serviceTag+"."+legacyMeshDomain] = struct{}{}
		unique[strings.ReplaceAll(serviceTag, "_", ".")+"."+legacyMeshDomain] = struct{}{}
	}

	return slices.Sorted(maps.Keys(unique))
}

func legacyMeshServiceEntryName(ms *meshservice_api.MeshServiceResource, port int32) string {
	return legacyMeshServiceName(core_model.GetDisplayName(ms.GetMeta()), meshServiceNamespace(ms), port)
}

func legacyMeshServiceName(displayName, namespace string, port int32) string {
	if displayName == "" {
		return ""
	}
	if namespace == "" || port == 0 {
		return displayName
	}
	return fmt.Sprintf("%s_%s_svc_%d", displayName, namespace, port)
}

func meshServiceNamespace(ms *meshservice_api.MeshServiceResource) string {
	if labels := ms.GetMeta().GetLabels(); labels != nil {
		if namespace := labels[mesh_proto.KubeNamespaceTag]; namespace != "" {
			return namespace
		}
	}
	if ms.Spec.Selector.DataplaneTags != nil {
		return (*ms.Spec.Selector.DataplaneTags)[mesh_proto.KubeNamespaceTag]
	}
	return ""
}

func legacyMeshExternalServiceDomains(mes *meshexternalservice_api.MeshExternalServiceResource) []string {
	displayName := core_model.GetDisplayName(mes.GetMeta())
	if displayName == "" {
		return nil
	}
	return []string{displayName + "." + legacyMeshDomain}
}

func legacyMeshMultiZoneServiceDomains(mzs *meshmzservice_api.MeshMultiZoneServiceResource) []string {
	displayName := core_model.GetDisplayName(mzs.GetMeta())
	if displayName == "" {
		return nil
	}
	return []string{displayName + "." + legacyMeshDomain}
}

func firstMeshServiceVIP(ms *meshservice_api.MeshServiceResource) string {
	if len(ms.Status.VIPs) == 0 {
		return ""
	}
	return ms.Status.VIPs[0].IP
}

func firstMeshExternalServiceVIP(mes *meshexternalservice_api.MeshExternalServiceResource) string {
	return mes.Status.VIP.IP
}

func firstMeshMultiZoneServiceVIP(mzs *meshmzservice_api.MeshMultiZoneServiceResource) string {
	if len(mzs.Status.VIPs) == 0 {
		return ""
	}
	return mzs.Status.VIPs[0].IP
}
