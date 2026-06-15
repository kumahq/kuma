package labels

import (
	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	config_core "github.com/kumahq/kuma/v2/pkg/config/core"
	"github.com/kumahq/kuma/v2/pkg/plugins/runtime/k8s/metadata"
)

func init() {
	register(LabelSpec{
		Key:   metadata.KumaServiceAccount,
		Owner: OwnerSystem,
	})

	register(LabelSpec{
		Key:       metadata.KumaWorkload,
		Owner:     OwnerControlPlane,
		OpenValue: true,
		Expected: func(ctx ValidationContext) (string, bool) {
			if ctx.Mode != config_core.Zone || !ctx.Descriptor.IsProxy {
				return "", false
			}
			return "", true
		},
	})

	register(LabelSpec{
		Key:   mesh_proto.ListenerZoneIngressLabel,
		Owner: OwnerSystem,
	})

	register(LabelSpec{
		Key:   mesh_proto.ListenerZoneEgressLabel,
		Owner: OwnerSystem,
	})

	register(LabelSpec{
		Key:   mesh_proto.ManagedByLabel,
		Owner: OwnerSystem,
	})

	register(LabelSpec{
		Key:   mesh_proto.DeletionGracePeriodStartedLabel,
		Owner: OwnerSystem,
	})
}
