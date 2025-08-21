package labels

import (
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/metadata"
)

// if changed sync with:
// https://github.com/Kong/shared-speakeasy/blob/b3ddd3ef1f31e42bfe71b96ea473493072f9742c/customtypes/kumalabels/kumalabels.go#L15
var AllComputedLabels = map[string]struct{}{
	metadata.KumaMeshLabel:         {},
	mesh_proto.ResourceOriginLabel: {},
	mesh_proto.ZoneTag:             {},
	mesh_proto.EnvTag:              {},
	mesh_proto.KubeNamespaceTag:    {},
	mesh_proto.PolicyRoleLabel:     {},
	mesh_proto.ProxyTypeLabel:      {},
	metadata.KumaServiceAccount:    {},
}
