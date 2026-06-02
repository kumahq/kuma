package labels

import (
	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	config_core "github.com/kumahq/kuma/v2/pkg/config/core"
	"github.com/kumahq/kuma/v2/pkg/plugins/runtime/k8s/metadata"
)

// System-managed labels. The CP sets these from external sources (Pod metadata,
// lifecycle events, dataplane spec) — non-privileged callers should never
// supply them on apply. KDS sync and the K8s controller go through the
// Privileged bypass and skip validation entirely.

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
