package labels

import (
	"strings"

	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	config_core "github.com/kumahq/kuma/v2/pkg/config/core"
	core_model "github.com/kumahq/kuma/v2/pkg/core/resources/model"
)

// Control-plane-managed labels. Each one is computed by the CP for the given
// resource and context; users may set them only if their value matches.

func init() {
	register(LabelSpec{
		Key:                       mesh_proto.ResourceOriginLabel,
		Owner:                     OwnerControlPlane,
		AllowedValues:             []string{string(mesh_proto.GlobalResourceOrigin), string(mesh_proto.ZoneResourceOrigin)},
		AllowAnyWhenNotApplicable: true,
		// The CP claims authority over kuma.io/origin wherever the resource
		// can be locally originated: Global (always 'global'), or Zone for
		// any ProvidedByZone-flagged type ('zone'). The narrower "user must
		// consciously set it" gate is in RequirePresence below.
		Expected: func(ctx ValidationContext) (string, bool) {
			if ctx.DisableOriginLabelValidation {
				return "", false
			}
			if ctx.Mode == config_core.Global {
				return string(mesh_proto.GlobalResourceOrigin), true
			}
			if ctx.Descriptor.KDSFlags.Has(core_model.ProvidedByZoneFlag) {
				return string(mesh_proto.ZoneResourceOrigin), true
			}
			return "", false
		},
		RequirePresence: func(ctx ValidationContext) bool {
			if ctx.DisableOriginLabelValidation {
				return false
			}
			if ctx.Mode != config_core.Zone {
				return false
			}
			if ctx.IsK8s {
				return ctx.Namespace.system
			}
			return ctx.FederatedZone && ctx.Descriptor.IsPluginOriginated
		},
	})

	register(LabelSpec{
		Key:   mesh_proto.MeshTag,
		Owner: OwnerControlPlane,
		Expected: func(ctx ValidationContext) (string, bool) {
			// Only Mesh-scoped resources have a meaningful kuma.io/mesh; for
			// Global-scoped resources (Mesh, Zone, ...) the label is not applicable.
			if ctx.Descriptor.Scope != core_model.ScopeMesh {
				return "", false
			}
			return ctx.ResourceMesh, true
		},
	})

	register(LabelSpec{
		Key:   mesh_proto.ZoneTag,
		Owner: OwnerControlPlane,
		Expected: func(ctx ValidationContext) (string, bool) {
			if ctx.Mode != config_core.Zone {
				return "", false
			}
			if !ctx.Descriptor.KDSFlags.Has(core_model.ProvidedByZoneFlag) {
				return "", false
			}
			return ctx.ZoneName, true
		},
	})

	register(LabelSpec{
		Key:           mesh_proto.EnvTag,
		Owner:         OwnerControlPlane,
		AllowedValues: []string{mesh_proto.KubernetesEnvironment, mesh_proto.UniversalEnvironment},
		Expected: func(ctx ValidationContext) (string, bool) {
			if ctx.Mode != config_core.Zone {
				return "", false
			}
			if !ctx.Descriptor.KDSFlags.Has(core_model.ProvidedByZoneFlag) {
				return "", false
			}
			if ctx.IsK8s {
				return mesh_proto.KubernetesEnvironment, true
			}
			return mesh_proto.UniversalEnvironment, true
		},
	})

	register(LabelSpec{
		Key:   mesh_proto.DisplayName,
		Owner: OwnerControlPlane,
		Expected: func(ctx ValidationContext) (string, bool) {
			if ctx.ResourceName == "" {
				return "", false
			}
			return ctx.ResourceName, true
		},
	})

	register(LabelSpec{
		Key:           mesh_proto.PolicyRoleLabel,
		Owner:         OwnerControlPlane,
		AllowedValues: []string{string(mesh_proto.SystemPolicyRole), string(mesh_proto.ProducerPolicyRole), string(mesh_proto.ConsumerPolicyRole), string(mesh_proto.WorkloadOwnerPolicyRole)},
		Expected: func(ctx ValidationContext) (string, bool) {
			if !ctx.Descriptor.IsPolicy || !ctx.Descriptor.IsPluginOriginated {
				return "", false
			}
			pol, ok := ctx.Spec.(core_model.Policy)
			if !ok {
				return "", false
			}
			role, err := ComputePolicyRole(pol, ctx.Namespace)
			if err != nil {
				// Bad policy spec; leave it to spec validators. Treat as
				// "not applicable" so we do not double-report here.
				return "", false
			}
			return string(role), true
		},
	})

	register(LabelSpec{
		Key:   mesh_proto.ProxyTypeLabel,
		Owner: OwnerControlPlane,
		Expected: func(ctx ValidationContext) (string, bool) {
			if !ctx.Descriptor.IsProxy {
				return "", false
			}
			proxy, ok := ctx.Spec.(core_model.ProxyResource)
			if !ok {
				return "", false
			}
			return strings.ToLower(string(proxy.GetProxyType())), true
		},
	})

	register(LabelSpec{
		Key:   mesh_proto.KubeNamespaceTag,
		Owner: OwnerControlPlane,
		Expected: func(ctx ValidationContext) (string, bool) {
			if !ctx.IsK8s || ctx.Namespace.value == "" {
				return "", false
			}
			return ctx.Namespace.value, true
		},
	})
}
