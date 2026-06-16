package labels

import (
	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	config_core "github.com/kumahq/kuma/v2/pkg/config/core"
	core_mesh "github.com/kumahq/kuma/v2/pkg/core/resources/apis/mesh"
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
		Expected: func(ctx ValidationContext) (string, error) {
			return ctx.ResourceMesh, nil
		},
	})

	register(LabelSpec{
		Key:   mesh_proto.ZoneTag,
		Owner: OwnerControlPlane,
		RequiredOn: RequiredOn{
			Modes:    []config_core.CpMode{config_core.Zone},
			KDSFlags: []core_model.KDSFlagType{core_model.ProvidedByZoneFlag},
		},
		Expected: func(ctx ValidationContext) (string, error) {
			return ctx.ZoneName, nil
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
		Expected: func(ctx ValidationContext) (string, error) {
			if ctx.Env == config_core.KubernetesEnvironment {
				return mesh_proto.KubernetesEnvironment, nil
			}
			return mesh_proto.UniversalEnvironment, nil
		},
	})

	register(LabelSpec{
		Key:   mesh_proto.DisplayName,
		Owner: OwnerControlPlane,
		Expected: func(ctx ValidationContext) (string, error) {
			return ctx.ResourceName, nil
		},
	})

	register(LabelSpec{
		Key:   mesh_proto.PolicyRoleLabel,
		Owner: OwnerControlPlane,
		AllowedValues: []string{
			string(mesh_proto.SystemPolicyRole),
			string(mesh_proto.ProducerPolicyRole),
			string(mesh_proto.ConsumerPolicyRole),
			string(mesh_proto.WorkloadOwnerPolicyRole),
		},
		RequiredOn: RequiredOn{
			Policy: true,
		},
		Expected: func(ctx ValidationContext) (string, error) {
			pol, ok := ctx.Spec.(core_model.Policy)
			if !ok {
				return "", nil
			}
			role, err := ComputePolicyRole(pol, ctx.Namespace)
			if err != nil {
				return "", err
			}
			return string(role), nil
		},
	})

	register(LabelSpec{
		Key:   mesh_proto.KubeNamespaceTag,
		Owner: OwnerControlPlane,
		RequiredOn: RequiredOn{
			Environments:      []config_core.EnvironmentType{config_core.KubernetesEnvironment},
			RequiresNamespace: true,
		},
		Expected: func(ctx ValidationContext) (string, error) {
			return ctx.Namespace.value, nil
		},
	})

	register(LabelSpec{
		Key:   mesh_proto.ProxyTypeLabel,
		Owner: OwnerControlPlane,
		RequiredOn: RequiredOn{
			ResourceTypes: []core_model.ResourceType{
				core_mesh.DataplaneType,
				core_mesh.ZoneIngressType,
				core_mesh.ZoneEgressType,
			},
		},
		Expected: func(ctx ValidationContext) (string, error) {
			proxy, ok := ctx.Spec.(core_model.ProxyResource)
			if !ok {
				return "", nil
			}
			return string(proxy.GetProxyType()), nil
		},
	})

	// System-owned labels.
	register(LabelSpec{
		Key:   metadata.KumaServiceAccount,
		Owner: OwnerSystem,
		RequiredOn: RequiredOn{
			Environments:  []config_core.EnvironmentType{config_core.KubernetesEnvironment},
			ResourceTypes: []core_model.ResourceType{core_mesh.DataplaneType},
		},
	})

	register(LabelSpec{
		Key:   metadata.KumaWorkload,
		Owner: OwnerSystem,
		RequiredOn: RequiredOn{
			Modes:         []config_core.CpMode{config_core.Zone},
			Environments:  []config_core.EnvironmentType{config_core.KubernetesEnvironment},
			ResourceTypes: []core_model.ResourceType{core_mesh.DataplaneType},
		},
	})

	register(LabelSpec{
		Key:   mesh_proto.ListenerZoneIngressLabel,
		Owner: OwnerControlPlane,
		RequiredOn: RequiredOn{
			ResourceTypes: []core_model.ResourceType{core_mesh.DataplaneType},
			SpecTraits:    []SpecTrait{HasZoneIngressListener},
		},
		Expected: func(ctx ValidationContext) (string, error) {
			return "enabled", nil
		},
	})

	register(LabelSpec{
		Key:   mesh_proto.ListenerZoneEgressLabel,
		Owner: OwnerControlPlane,
		RequiredOn: RequiredOn{
			ResourceTypes: []core_model.ResourceType{core_mesh.DataplaneType},
			SpecTraits:    []SpecTrait{HasZoneEgressListener},
		},
		Expected: func(ctx ValidationContext) (string, error) {
			return "enabled", nil
		},
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
			Modes:         []config_core.CpMode{config_core.Zone},
			Environments:  []config_core.EnvironmentType{config_core.UniversalEnvironment},
			ResourceTypes: []core_model.ResourceType{core_mesh.DataplaneType},
		},
	})
}
