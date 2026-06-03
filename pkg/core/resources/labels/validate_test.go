package labels_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	config_core "github.com/kumahq/kuma/v2/pkg/config/core"
	core_mesh "github.com/kumahq/kuma/v2/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/v2/pkg/core/resources/labels"
	mtp_api "github.com/kumahq/kuma/v2/pkg/plugins/policies/meshtrafficpermission/api/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/plugins/runtime/k8s/metadata"
	"github.com/kumahq/kuma/v2/pkg/test/resources/builders"
)

var _ = Describe("Validate", func() {
	mtDesc := func() labels.ValidationContext {
		// MeshTrafficPermission descriptor: IsPolicy=true, IsPluginOriginated=true.
		mtp := builders.MeshTrafficPermission().
			WithTargetRef(builders.TargetRefMesh()).
			AddFrom(builders.TargetRefMesh(), mtp_api.Allow).
			Build()
		return labels.ValidationContext{
			Mode:         config_core.Zone,
			ZoneName:     "kuma-zone",
			Descriptor:   mtp.Descriptor(),
			Spec:         mtp.Spec,
			ResourceName: "mtp-1",
			ResourceMesh: "mesh-1",
		}
	}

	dpDesc := func() labels.ValidationContext {
		dp := core_mesh.NewDataplaneResource()
		return labels.ValidationContext{
			Mode:         config_core.Zone,
			ZoneName:     "kuma-zone",
			Descriptor:   dp.Descriptor(),
			Spec:         dp.Spec,
			ResourceName: "dp-1",
			ResourceMesh: "mesh-1",
		}
	}

	It("returns nothing for a clean apply on a zone CP", func() {
		ctx := mtDesc()
		ctx.FederatedZone = true
		r := labels.Validate(map[string]string{
			mesh_proto.ResourceOriginLabel: "zone",
			mesh_proto.MeshTag:             "mesh-1",
			mesh_proto.ZoneTag:             "kuma-zone",
		}, ctx)
		Expect(r.Errors).To(BeEmpty())
		Expect(r.Warnings).To(BeEmpty())
	})

	It("warns on mismatched mesh label and replaces with the expected value", func() {
		ctx := mtDesc()
		ctx.FederatedZone = true
		r := labels.Validate(map[string]string{
			mesh_proto.ResourceOriginLabel: "zone",
			mesh_proto.MeshTag:             "other-mesh",
		}, ctx)
		Expect(r.Errors).To(BeEmpty())
		Expect(r.Warnings).To(ContainElement(labels.Violation{
			Key:    mesh_proto.MeshTag,
			Reason: "kuma.io/mesh is computed by the control plane (expected 'mesh-1'); the supplied value 'other-mesh' was overridden",
		}))
		Expect(r.Sanitized).To(HaveKeyWithValue(mesh_proto.MeshTag, "mesh-1"))
	})

	It("warns on mismatched zone label and replaces with the expected value", func() {
		ctx := mtDesc()
		ctx.FederatedZone = true
		r := labels.Validate(map[string]string{
			mesh_proto.ResourceOriginLabel: "zone",
			mesh_proto.ZoneTag:             "other-zone",
		}, ctx)
		Expect(r.Errors).To(BeEmpty())
		Expect(r.Warnings).To(ContainElement(labels.Violation{
			Key:    mesh_proto.ZoneTag,
			Reason: "kuma.io/zone is computed by the control plane (expected 'kuma-zone'); the supplied value 'other-zone' was overridden",
		}))
		Expect(r.Sanitized).To(HaveKeyWithValue(mesh_proto.ZoneTag, "kuma-zone"))
	})

	It("does not require origin to be present (Compute will populate it)", func() {
		ctx := mtDesc()
		ctx.FederatedZone = true
		r := labels.Validate(map[string]string{}, ctx)
		Expect(r.Errors).To(BeEmpty())
		Expect(r.Warnings).To(BeEmpty())
	})

	It("warns on origin=global on a zone CP and replaces with the expected value", func() {
		ctx := mtDesc()
		ctx.FederatedZone = true
		r := labels.Validate(map[string]string{
			mesh_proto.ResourceOriginLabel: "global",
		}, ctx)
		Expect(r.Errors).To(BeEmpty())
		Expect(r.Warnings).To(ContainElement(labels.Violation{
			Key:    mesh_proto.ResourceOriginLabel,
			Reason: "kuma.io/origin is computed by the control plane (expected 'zone'); the supplied value 'global' was overridden",
		}))
		Expect(r.Sanitized).To(HaveKeyWithValue(mesh_proto.ResourceOriginLabel, "zone"))
	})

	It("warns on user-set kuma.io/env on global CP (not applicable)", func() {
		ctx := mtDesc()
		ctx.Mode = config_core.Global
		r := labels.Validate(map[string]string{
			mesh_proto.EnvTag: "universal",
		}, ctx)
		Expect(r.Errors).To(BeEmpty())
		Expect(r.Warnings).To(ContainElement(labels.Violation{
			Key:    mesh_proto.EnvTag,
			Reason: "kuma.io/env is managed by the control plane and is not applicable in this context; the supplied value 'universal' was removed",
		}))
		Expect(r.Sanitized).NotTo(HaveKey(mesh_proto.EnvTag))
	})

	It("accepts kuma.io/env=universal on Universal zone CP", func() {
		ctx := mtDesc()
		ctx.FederatedZone = true
		r := labels.Validate(map[string]string{
			mesh_proto.ResourceOriginLabel: "zone",
			mesh_proto.EnvTag:              "universal",
		}, ctx)
		Expect(r.Errors).To(BeEmpty())
		Expect(r.Warnings).To(BeEmpty())
	})

	It("warns on kuma.io/env=universal on K8s zone CP (mismatch — expects 'kubernetes')", func() {
		ctx := mtDesc()
		ctx.FederatedZone = true
		ctx.IsK8s = true
		ctx.Namespace = labels.NewNamespace("kuma-system", true)
		r := labels.Validate(map[string]string{
			mesh_proto.ResourceOriginLabel: "zone",
			mesh_proto.EnvTag:              "universal",
		}, ctx)
		Expect(r.Errors).To(BeEmpty())
		Expect(r.Warnings).To(ContainElement(labels.Violation{
			Key:    mesh_proto.EnvTag,
			Reason: "kuma.io/env is computed by the control plane (expected 'kubernetes'); the supplied value 'universal' was overridden",
		}))
	})

	It("warns on mismatched display-name and replaces with the expected value", func() {
		ctx := mtDesc()
		ctx.FederatedZone = true
		r := labels.Validate(map[string]string{
			mesh_proto.ResourceOriginLabel: "zone",
			mesh_proto.DisplayName:         "wrong-name",
		}, ctx)
		Expect(r.Errors).To(BeEmpty())
		Expect(r.Warnings).To(ContainElement(labels.Violation{
			Key:    mesh_proto.DisplayName,
			Reason: "kuma.io/display-name is computed by the control plane (expected 'mtp-1'); the supplied value 'wrong-name' was overridden",
		}))
		Expect(r.Sanitized).To(HaveKeyWithValue(mesh_proto.DisplayName, "mtp-1"))
	})

	It("rejects arbitrary reserved keys not in the registry", func() {
		ctx := mtDesc()
		ctx.FederatedZone = true
		r := labels.Validate(map[string]string{
			mesh_proto.ResourceOriginLabel: "zone",
			"kuma.io/foo":                  "bar",
			"k8s.kuma.io/baz":              "qux",
		}, ctx)
		var keys []string
		for _, v := range r.Errors {
			keys = append(keys, v.Key)
		}
		Expect(keys).To(ConsistOf("k8s.kuma.io/baz", "kuma.io/foo"))
		Expect(r.Warnings).To(BeEmpty())
	})

	It("allows non-reserved user labels", func() {
		ctx := mtDesc()
		ctx.FederatedZone = true
		r := labels.Validate(map[string]string{
			mesh_proto.ResourceOriginLabel: "zone",
			"example.com/team":             "platform",
			"app":                          "billing",
		}, ctx)
		Expect(r.Errors).To(BeEmpty())
		Expect(r.Warnings).To(BeEmpty())
	})

	It("allows OwnerUser flags with allowed values", func() {
		ctx := mtDesc()
		ctx.FederatedZone = true
		r := labels.Validate(map[string]string{
			mesh_proto.ResourceOriginLabel: "zone",
			mesh_proto.EffectLabel:         "shadow",
			mesh_proto.KDSSyncLabel:        "disabled",
		}, ctx)
		Expect(r.Errors).To(BeEmpty())
		Expect(r.Warnings).To(BeEmpty())
	})

	It("rejects OwnerUser flag values outside the allowed set", func() {
		ctx := mtDesc()
		ctx.FederatedZone = true
		r := labels.Validate(map[string]string{
			mesh_proto.ResourceOriginLabel: "zone",
			mesh_proto.EffectLabel:         "loud",
		}, ctx)
		Expect(r.Errors).To(ContainElement(labels.Violation{
			Key:    mesh_proto.EffectLabel,
			Reason: "must be one of ['', 'shadow']",
		}))
		Expect(r.Warnings).To(BeEmpty())
	})

	It("allows user-set kuma.io/workload on Universal Dataplane", func() {
		ctx := dpDesc()
		ctx.FederatedZone = true
		r := labels.Validate(map[string]string{
			mesh_proto.ResourceOriginLabel: "zone",
			metadata.KumaWorkload:          "my-workload",
		}, ctx)
		Expect(r.Errors).To(BeEmpty())
		Expect(r.Warnings).To(BeEmpty())
	})

	It("warns on kuma.io/workload on a policy (not a proxy) and strips it", func() {
		ctx := mtDesc()
		ctx.FederatedZone = true
		r := labels.Validate(map[string]string{
			mesh_proto.ResourceOriginLabel: "zone",
			metadata.KumaWorkload:          "my-workload",
		}, ctx)
		Expect(r.Errors).To(BeEmpty())
		Expect(r.Warnings).To(ContainElement(labels.Violation{
			Key:    metadata.KumaWorkload,
			Reason: "kuma.io/workload is managed by the control plane and is not applicable in this context; the supplied value 'my-workload' was removed",
		}))
		Expect(r.Sanitized).NotTo(HaveKey(metadata.KumaWorkload))
	})

	It("warns on system-only labels (managed-by) from user input and strips them", func() {
		ctx := mtDesc()
		ctx.FederatedZone = true
		r := labels.Validate(map[string]string{
			mesh_proto.ResourceOriginLabel: "zone",
			mesh_proto.ManagedByLabel:      "anything",
		}, ctx)
		Expect(r.Errors).To(BeEmpty())
		Expect(r.Warnings).To(ContainElement(labels.Violation{
			Key:    mesh_proto.ManagedByLabel,
			Reason: "kuma.io/managed-by is set by the control plane; the supplied value 'anything' was overridden",
		}))
		Expect(r.Sanitized).NotTo(HaveKey(mesh_proto.ManagedByLabel))
	})

	It("bypasses validation entirely when Privileged is true", func() {
		ctx := mtDesc()
		ctx.FederatedZone = true
		ctx.Privileged = true
		// This input would otherwise produce several findings.
		input := map[string]string{
			mesh_proto.ResourceOriginLabel: "global",
			mesh_proto.MeshTag:             "other-mesh",
			mesh_proto.DisplayName:         "wrong-name",
			mesh_proto.ManagedByLabel:      "kds",
		}
		r := labels.Validate(input, ctx)
		Expect(r.Errors).To(BeEmpty())
		Expect(r.Warnings).To(BeEmpty())
		Expect(r.Sanitized).To(Equal(input))
	})

	It("honors DisableOriginLabelValidation for the origin spec", func() {
		ctx := mtDesc()
		ctx.FederatedZone = true
		ctx.DisableOriginLabelValidation = true
		r := labels.Validate(map[string]string{}, ctx)
		Expect(r.Errors).To(BeEmpty())
		Expect(r.Warnings).To(BeEmpty())
	})
})
