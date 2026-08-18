package labels

import (
	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	config_core "github.com/kumahq/kuma/v3/pkg/config/core"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
)

// registry declares every label the control plane computes. Duplicate keys are
// a compile error, so exactly one spec owns a key.
var registry = map[string]LabelSpec{
	mesh_proto.ResourceOriginLabel: {
		// Standalone is normalized to zone at CP startup, so Global and Zone
		// cover every real CP mode.
		RequiredOn: RequiredOn{
			Modes: []config_core.CpMode{config_core.Global, config_core.Zone},
		},
		Expected: func(ctx ValidationContext) (string, error) {
			if ctx.Mode == config_core.Global {
				return string(mesh_proto.GlobalResourceOrigin), nil
			}
			return string(mesh_proto.ZoneResourceOrigin), nil
		},
	},

	mesh_proto.MeshTag: {
		RequiredOn: RequiredOn{
			ResourceScopes: []core_model.ResourceScope{core_model.ScopeMesh},
		},
		Expected: func(ctx ValidationContext) (string, error) {
			return ctx.ResourceMesh, nil
		},
	},

	mesh_proto.ZoneTag: {
		// If a resource can't be created on a zone (like Mesh), there is no
		// point in adding 'kuma.io/zone' and 'kuma.io/env'.
		RequiredOn: RequiredOn{
			Modes:    []config_core.CpMode{config_core.Zone},
			KDSFlags: []core_model.KDSFlagType{core_model.ProvidedByZoneFlag},
		},
		Expected: func(ctx ValidationContext) (string, error) {
			return ctx.ZoneName, nil
		},
	},

	mesh_proto.EnvTag: {
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
	},

	mesh_proto.DisplayName: {
		Expected: func(ctx ValidationContext) (string, error) {
			return ctx.ResourceName, nil
		},
	},

	mesh_proto.PolicyRoleLabel: {
		RequiredOn: RequiredOn{
			Policy:            true,
			RequiresNamespace: true,
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
	},

	mesh_proto.KubeNamespaceTag: {
		RequiredOn: RequiredOn{
			Environments:      []config_core.EnvironmentType{config_core.KubernetesEnvironment},
			RequiresNamespace: true,
		},
		Expected: func(ctx ValidationContext) (string, error) {
			return ctx.Namespace.value, nil
		},
	},

	mesh_proto.ProxyTypeLabel: {
		RequiredOn: RequiredOn{Proxy: true},
		Expected: func(ctx ValidationContext) (string, error) {
			proxy, ok := ctx.Spec.(core_model.ProxyResource)
			if !ok {
				return "", nil
			}
			return string(proxy.GetProxyType()), nil
		},
	},

	mesh_proto.ListenerZoneIngressLabel: {
		RequiredOn: RequiredOn{
			ResourceTypes: []core_model.ResourceType{core_mesh.DataplaneType},
			SpecTraits:    []SpecTrait{HasZoneIngressListener},
		},
		Expected: func(ctx ValidationContext) (string, error) {
			return "enabled", nil
		},
	},

	mesh_proto.ListenerZoneEgressLabel: {
		RequiredOn: RequiredOn{
			ResourceTypes: []core_model.ResourceType{core_mesh.DataplaneType},
			SpecTraits:    []SpecTrait{HasZoneEgressListener},
		},
		Expected: func(ctx ValidationContext) (string, error) {
			return "enabled", nil
		},
	},
}
