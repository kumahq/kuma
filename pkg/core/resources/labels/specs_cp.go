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
		Key:               mesh_proto.ResourceOriginLabel,
		Description:       "Origin of the resource. 'global' on the global CP, 'zone' on a zone CP.",
		Owner:             OwnerControlPlane,
		AllowedValues:     []string{string(mesh_proto.GlobalResourceOrigin), string(mesh_proto.ZoneResourceOrigin)},
		ExpectedValueExpr: "'global' on global CP; 'zone' on zone CP",
		RequiredWhenExpr:  "zone CP, plugin-originated resources (REST) or system namespace (K8s)",
		Expected: func(ctx ValidationContext) (string, bool) {
			if ctx.Mode == config_core.Global {
				return string(mesh_proto.GlobalResourceOrigin), true
			}
			return string(mesh_proto.ZoneResourceOrigin), true
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
		Key:               mesh_proto.MeshTag,
		Description:       "Mesh the resource belongs to.",
		Owner:             OwnerControlPlane,
		ExpectedValueExpr: "= resource's mesh field",
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
		Key:               mesh_proto.ZoneTag,
		Description:       "Zone the resource belongs to.",
		Owner:             OwnerControlPlane,
		ExpectedValueExpr: "current zone name (zone CP only)",
		AppliesToExpr:     "zone CP only",
		Expected: func(ctx ValidationContext) (string, bool) {
			if ctx.Mode != config_core.Zone {
				return "", false
			}
			return ctx.ZoneName, true
		},
	})

	register(LabelSpec{
		Key:               mesh_proto.EnvTag,
		Description:       "Environment (kubernetes or universal) the resource was created on.",
		Owner:             OwnerControlPlane,
		AllowedValues:     []string{mesh_proto.KubernetesEnvironment, mesh_proto.UniversalEnvironment},
		ExpectedValueExpr: "'kubernetes' on K8s zone, 'universal' on universal zone",
		AppliesToExpr:     "zone CP only",
		Expected: func(ctx ValidationContext) (string, bool) {
			if ctx.Mode != config_core.Zone {
				return "", false
			}
			if ctx.IsK8s {
				return mesh_proto.KubernetesEnvironment, true
			}
			return mesh_proto.UniversalEnvironment, true
		},
	})

	register(LabelSpec{
		Key:               mesh_proto.DisplayName,
		Description:       "Human-readable name of the resource (used by UIs and KDS sync).",
		Owner:             OwnerControlPlane,
		ExpectedValueExpr: "= resource name (K8s metadata.name without .namespace suffix)",
		Expected: func(ctx ValidationContext) (string, bool) {
			if ctx.ResourceName == "" {
				return "", false
			}
			return ctx.ResourceName, true
		},
	})

	register(LabelSpec{
		Key:               mesh_proto.PolicyRoleLabel,
		Description:       "Role of the policy: system, producer, consumer or workload-owner.",
		Owner:             OwnerControlPlane,
		AllowedValues:     []string{string(mesh_proto.SystemPolicyRole), string(mesh_proto.ProducerPolicyRole), string(mesh_proto.ConsumerPolicyRole), string(mesh_proto.WorkloadOwnerPolicyRole)},
		ExpectedValueExpr: "ComputePolicyRole(spec, namespace)",
		AppliesToExpr:     "plugin-originated policies",
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
		Key:               mesh_proto.ProxyTypeLabel,
		Description:       "Type of the proxy (sidecar, gateway, zoneingress, zoneegress).",
		Owner:             OwnerControlPlane,
		ExpectedValueExpr: "spec.GetProxyType()",
		AppliesToExpr:     "proxy resources (IsProxy=true)",
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
		Key:               mesh_proto.KubeNamespaceTag,
		Description:       "Kubernetes namespace the resource was applied in.",
		Owner:             OwnerControlPlane,
		ExpectedValueExpr: "current namespace (K8s only)",
		AppliesToExpr:     "K8s resources with a namespace",
		Expected: func(ctx ValidationContext) (string, bool) {
			if !ctx.IsK8s || ctx.Namespace.value == "" {
				return "", false
			}
			return ctx.Namespace.value, true
		},
	})
}
