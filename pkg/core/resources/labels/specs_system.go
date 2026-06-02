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
		Key:         metadata.KumaServiceAccount,
		Description: "Kubernetes ServiceAccount of the workload behind a proxy. CP-set from the Pod.",
		Owner:       OwnerSystem,
	})

	register(LabelSpec{
		Key:           metadata.KumaWorkload,
		Description:   "Workload identifier of a proxy. CP-set from the Pod owner on K8s; user-supplied on Universal Dataplanes (privileged paths bypass this validator).",
		Owner:         OwnerControlPlane,
		OpenValue:     true,
		AppliesToExpr: "proxy resources on a zone CP (Universal-set or K8s-controller-set)",
		Expected: func(ctx ValidationContext) (string, bool) {
			if ctx.Mode != config_core.Zone || !ctx.Descriptor.IsProxy {
				return "", false
			}
			return "", true
		},
	})

	register(LabelSpec{
		Key:         mesh_proto.ListenerZoneIngressLabel,
		Description: "Marker set by the CP when a Dataplane has a ZoneIngress listener.",
		Owner:       OwnerSystem,
	})

	register(LabelSpec{
		Key:         mesh_proto.ListenerZoneEgressLabel,
		Description: "Marker set by the CP when a Dataplane has a ZoneEgress listener.",
		Owner:       OwnerSystem,
	})

	register(LabelSpec{
		Key:         mesh_proto.ManagedByLabel,
		Description: "Set by the CP on auto-generated resources (e.g. MeshServices).",
		Owner:       OwnerSystem,
	})

	register(LabelSpec{
		Key:         mesh_proto.DeletionGracePeriodStartedLabel,
		Description: "CP-managed lifecycle marker for graceful deletion.",
		Owner:       OwnerSystem,
	})
}
