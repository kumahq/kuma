package context

import (
	"fmt"
	"strconv"
	"strings"

	common_api "github.com/kumahq/kuma/v3/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
)

// NormalizeBackendRefTarget converts legacy backend-ref fields into label and
// section-name selectors shared by dataplane outbounds and reachable-backend resolution.
func NormalizeBackendRefTarget(kind, name, namespace string, port *uint32, labels map[string]string, defaultNamespace string) (map[string]string, string) {
	sectionName := ""
	if port != nil && *port > 0 {
		sectionName = fmt.Sprintf("%d", *port)
	}
	if len(labels) > 0 {
		return labels, sectionName
	}
	if name == "" {
		return nil, sectionName
	}

	if common_api.TargetRefKind(kind) == common_api.MeshService && namespace == "" {
		if service, parsedNamespace, parsedPort, ok := parseLegacyMeshServiceTag(name); ok {
			name = service
			namespace = parsedNamespace
			if sectionName == "" && parsedPort > 0 {
				sectionName = fmt.Sprintf("%d", parsedPort)
			}
		}
	}

	normalized := map[string]string{
		mesh_proto.DisplayName: name,
	}
	switch common_api.TargetRefKind(kind) {
	case common_api.MeshService, common_api.MeshExternalService, common_api.MeshMultiZoneService:
		if namespace == "" {
			namespace = defaultNamespace
		}
		if namespace != "" {
			normalized[mesh_proto.KubeNamespaceTag] = namespace
		}
	}
	return normalized, sectionName
}

func parseLegacyMeshServiceTag(name string) (string, string, int32, bool) {
	segments := strings.Split(name, "_")
	if len(segments) != 4 || segments[2] != "svc" {
		return "", "", 0, false
	}

	port, err := strconv.ParseInt(segments[3], 10, 32)
	if err != nil {
		return "", "", 0, false
	}
	return segments[0], segments[1], int32(port), true
}
