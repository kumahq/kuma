package labels

import (
	"strings"

	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	config_core "github.com/kumahq/kuma/v2/pkg/config/core"
	core_model "github.com/kumahq/kuma/v2/pkg/core/resources/model"
	"github.com/kumahq/kuma/v2/pkg/plugins/runtime/k8s/metadata"
)

func init() {
	// Control-plane-owned labels.
	register(LabelSpec{
		Key:   mesh_proto.MeshTag,
		Owner: OwnerControlPlane,
		RequiredOn: RequiredOn{
			ResourceScopes: []core_model.ResourceScope{core_model.ScopeMesh},
		},
		Expected: func(ctx ValidationContext) string {
			return ctx.ResourceMesh
		},
	})

	register(LabelSpec{
		Key:   mesh_proto.ZoneTag,
		Owner: OwnerControlPlane,
		RequiredOn: RequiredOn{
			Modes:    []config_core.CpMode{config_core.Zone},
			KDSFlags: []core_model.KDSFlagType{core_model.ProvidedByZoneFlag},
		},
		Expected: func(ctx ValidationContext) string {
			return ctx.ZoneName
		},
	})

	register(LabelSpec{
		Key:           mesh_proto.EnvTag,
		Owner:         OwnerControlPlane,
		AllowedValues: []string{mesh_proto.KubernetesEnvironment, mesh_proto.UniversalEnvironment},
		RequiredOn: RequiredOn{
			Modes:    []config_core.CpMode{config_core.Zone},
			KDSFlags: []core_model.KDSFlagType{core_model.ProvidedByZoneFlag},
		},
		Expected: func(ctx ValidationContext) string {
			if ctx.Env == config_core.KubernetesEnvironment {
				return mesh_proto.KubernetesEnvironment
			}
			return mesh_proto.UniversalEnvironment
		},
	})

	register(LabelSpec{
		Key:   mesh_proto.DisplayName,
		Owner: OwnerControlPlane,
		Expected: func(ctx ValidationContext) string {
			return ctx.ResourceName
		},
	})

	register(LabelSpec{
		Key:           mesh_proto.PolicyRoleLabel,
		Owner:         OwnerControlPlane,
		AllowedValues: []string{string(mesh_proto.SystemPolicyRole), string(mesh_proto.ProducerPolicyRole), string(mesh_proto.ConsumerPolicyRole), string(mesh_proto.WorkloadOwnerPolicyRole)},
		RequiredOn: RequiredOn{
			ResourceTraits: []ResourceTrait{TraitPolicy, TraitPluginOriginated},
		},
		Expected: func(ctx ValidationContext) string {
			pol, ok := ctx.Spec.(core_model.Policy)
			if !ok {
				return ""
			}
			role, err := ComputePolicyRole(pol, ctx.Namespace)
			if err != nil {
				return ""
			}
			return string(role)
		},
	})

	register(LabelSpec{
		Key:   mesh_proto.ProxyTypeLabel,
		Owner: OwnerControlPlane,
		RequiredOn: RequiredOn{
			ResourceTraits: []ResourceTrait{TraitProxy},
		},
		Expected: func(ctx ValidationContext) string {
			proxy, ok := ctx.Spec.(core_model.ProxyResource)
			if !ok {
				return ""
			}
			return strings.ToLower(string(proxy.GetProxyType()))
		},
	})

	register(LabelSpec{
		Key:   mesh_proto.KubeNamespaceTag,
		Owner: OwnerControlPlane,
		RequiredOn: RequiredOn{
			Environments:      []config_core.EnvironmentType{config_core.KubernetesEnvironment},
			RequiresNamespace: true,
		},
		Expected: func(ctx ValidationContext) string {
			return ctx.Namespace.value
		},
	})

	// System-owned labels.
	register(LabelSpec{
		Key:   metadata.KumaServiceAccount,
		Owner: OwnerSystem,
		RequiredOn: RequiredOn{
			Environments:   []config_core.EnvironmentType{config_core.KubernetesEnvironment},
			ResourceTraits: []ResourceTrait{TraitProxy},
		},
	})

	register(LabelSpec{
		Key:   metadata.KumaWorkload,
		Owner: OwnerSystem,
		RequiredOn: RequiredOn{
			Modes:          []config_core.CpMode{config_core.Zone},
			Environments:   []config_core.EnvironmentType{config_core.KubernetesEnvironment},
			ResourceTraits: []ResourceTrait{TraitProxy},
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

	// User-owned labels.
	register(LabelSpec{
		Key:           mesh_proto.EffectLabel,
		Owner:         OwnerUser,
		AllowedValues: []string{"", "shadow"},
	})

	register(LabelSpec{
		Key:           mesh_proto.KDSSyncLabel,
		Owner:         OwnerUser,
		AllowedValues: []string{"", "enabled", "disabled"},
	})

	register(LabelSpec{
		Key:   metadata.KumaWorkload,
		Owner: OwnerUser,
		RequiredOn: RequiredOn{
			Modes:          []config_core.CpMode{config_core.Zone},
			Environments:   []config_core.EnvironmentType{config_core.UniversalEnvironment},
			ResourceTraits: []ResourceTrait{TraitProxy},
		},
	})
}
