package labels

import (
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/metadata"
)

var AllComputedLabels = map[string]struct{}{
	metadata.KumaMeshLabel:         {},
	mesh_proto.ResourceOriginLabel: {},
	mesh_proto.ZoneTag:             {},
	mesh_proto.EnvTag:              {},
	mesh_proto.KubeNamespaceTag:    {},
	mesh_proto.PolicyRoleLabel:     {},
	mesh_proto.ProxyTypeLabel:      {},
}
