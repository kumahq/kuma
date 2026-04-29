package labels

import (
	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/plugins/runtime/k8s/metadata"
)

// if changed sync with:
// https://github.com/Kong/shared-speakeasy/blob/b3ddd3ef1f31e42bfe71b96ea473493072f9742c/customtypes/kumalabels/kumalabels.go#L15
var AllComputedLabels = map[string]struct{}{
	metadata.KumaMeshLabel:         {}, // can be set by user, on Universal must be equal to 'meta.mesh'
	mesh_proto.ResourceOriginLabel: {}, // must be set on zone if special flag is on, should be 'global' on 'global' and 'zone' on 'zone'
	mesh_proto.ZoneTag:             {}, // only on zone, should be equal to zone
	mesh_proto.EnvTag:              {}, // uni on uni and k8s on k8s
	mesh_proto.KubeNamespaceTag:    {}, // only on k8s, equal to the namespace
	mesh_proto.PolicyRoleLabel:     {}, // only on policies, defined by the 'spec'
	mesh_proto.ProxyTypeLabel:      {}, // only on Dataplane
	metadata.KumaServiceAccount:    {}, // only on Dataplane
	metadata.KumaWorkload:          {}, // only on Dataplane
	mesh_proto.DisplayName:         {}, // can be set but must be equal to 'meta.name', unless the user is priviledged
}
