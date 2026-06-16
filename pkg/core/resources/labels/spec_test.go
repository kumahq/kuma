package labels_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	config_core "github.com/kumahq/kuma/v2/pkg/config/core"
	core_mesh "github.com/kumahq/kuma/v2/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/v2/pkg/core/resources/labels"
	core_model "github.com/kumahq/kuma/v2/pkg/core/resources/model"
)

var _ = Describe("RequiredOn.Matches", func() {
	baseCtx := func() labels.ValidationContext {
		return labels.ValidationContext{
			Mode: config_core.Zone,
			Env:  config_core.KubernetesEnvironment,
			Descriptor: core_model.ResourceTypeDescriptor{
				Scope:              core_model.ScopeMesh,
				KDSFlags:           core_model.ProvidedByZoneFlag,
				IsPolicy:           true,
				IsProxy:            false,
				IsPluginOriginated: true,
			},
			Namespace: labels.NewNamespace("kuma-ns", false),
		}
	}

	It("matches the zero value against any context", func() {
		Expect(labels.RequiredOn{}.Matches(baseCtx())).To(BeTrue())
	})

	It("enforces Modes as an allowlist", func() {
		r := labels.RequiredOn{Modes: []config_core.CpMode{config_core.Zone}}
		Expect(r.Matches(baseCtx())).To(BeTrue())

		ctx := baseCtx()
		ctx.Mode = config_core.Global
		Expect(r.Matches(ctx)).To(BeFalse())
	})

	It("enforces ResourceScopes as an allowlist", func() {
		r := labels.RequiredOn{ResourceScopes: []core_model.ResourceScope{core_model.ScopeMesh}}
		Expect(r.Matches(baseCtx())).To(BeTrue())

		ctx := baseCtx()
		ctx.Descriptor.Scope = core_model.ScopeGlobal
		Expect(r.Matches(ctx)).To(BeFalse())
	})

	It("requires every listed KDS flag (AND)", func() {
		r := labels.RequiredOn{KDSFlags: []core_model.KDSFlagType{core_model.ProvidedByZoneFlag}}
		Expect(r.Matches(baseCtx())).To(BeTrue())

		ctx := baseCtx()
		ctx.Descriptor.KDSFlags = 0
		Expect(r.Matches(ctx)).To(BeFalse())
	})

	It("requires a plugin-originated policy when Policy is set", func() {
		r := labels.RequiredOn{Policy: true}
		Expect(r.Matches(baseCtx())).To(BeTrue())

		ctx := baseCtx()
		ctx.Descriptor.IsPluginOriginated = false
		Expect(r.Matches(ctx)).To(BeFalse())
	})

	It("enforces ResourceTypes as an allowlist", func() {
		r := labels.RequiredOn{ResourceTypes: []core_model.ResourceType{core_mesh.DataplaneType}}

		ctx := baseCtx()
		ctx.Descriptor.Name = core_mesh.DataplaneType
		Expect(r.Matches(ctx)).To(BeTrue())

		ctx.Descriptor.Name = core_mesh.MeshGatewayType
		Expect(r.Matches(ctx)).To(BeFalse())
	})

	It("requires every listed spec trait (AND)", func() {
		r := labels.RequiredOn{SpecTraits: []labels.SpecTrait{labels.HasZoneIngressListener}}

		ctx := baseCtx()
		ctx.Spec = &mesh_proto.Dataplane{
			Networking: &mesh_proto.Dataplane_Networking{
				Listeners: []*mesh_proto.Dataplane_Networking_Listener{
					{Type: mesh_proto.Dataplane_Networking_Listener_ZoneIngress},
				},
			},
		}
		Expect(r.Matches(ctx)).To(BeTrue())

		ctx.Spec = &mesh_proto.Dataplane{}
		Expect(r.Matches(ctx)).To(BeFalse())
	})

	It("enforces Environments as an allowlist", func() {
		r := labels.RequiredOn{Environments: []config_core.EnvironmentType{config_core.KubernetesEnvironment}}
		Expect(r.Matches(baseCtx())).To(BeTrue())

		ctx := baseCtx()
		ctx.Env = config_core.UniversalEnvironment
		Expect(r.Matches(ctx)).To(BeFalse())
	})

	It("requires a namespace when RequiresNamespace is set", func() {
		r := labels.RequiredOn{RequiresNamespace: true}
		Expect(r.Matches(baseCtx())).To(BeTrue())

		ctx := baseCtx()
		ctx.Namespace = labels.UnsetNamespace
		Expect(r.Matches(ctx)).To(BeFalse())
	})

	It("AND-combines multiple dimensions", func() {
		r := labels.RequiredOn{
			Modes:    []config_core.CpMode{config_core.Zone},
			KDSFlags: []core_model.KDSFlagType{core_model.ProvidedByZoneFlag},
			Policy:   true,
		}
		Expect(r.Matches(baseCtx())).To(BeTrue())

		ctx := baseCtx()
		ctx.Mode = config_core.Global
		Expect(r.Matches(ctx)).To(BeFalse())
	})
})
