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

	It("returns no violations for a clean apply on a zone CP", func() {
		ctx := mtDesc()
		ctx.FederatedZone = true
		violations := labels.Validate(map[string]string{
			mesh_proto.ResourceOriginLabel: "zone",
			mesh_proto.MeshTag:             "mesh-1",
			mesh_proto.ZoneTag:             "kuma-zone",
		}, ctx)
		Expect(violations).To(BeEmpty())
	})

	It("rejects mismatched mesh label", func() {
		ctx := mtDesc()
		ctx.FederatedZone = true
		violations := labels.Validate(map[string]string{
			mesh_proto.ResourceOriginLabel: "zone",
			mesh_proto.MeshTag:             "other-mesh",
		}, ctx)
		Expect(violations).To(ContainElement(labels.Violation{
			Key:    mesh_proto.MeshTag,
			Reason: "kuma.io/mesh should be 'mesh-1', got 'other-mesh'",
		}))
	})

	It("rejects mismatched zone label", func() {
		ctx := mtDesc()
		ctx.FederatedZone = true
		violations := labels.Validate(map[string]string{
			mesh_proto.ResourceOriginLabel: "zone",
			mesh_proto.ZoneTag:             "other-zone",
		}, ctx)
		Expect(violations).To(ContainElement(labels.Violation{
			Key:    mesh_proto.ZoneTag,
			Reason: "kuma.io/zone should be 'kuma-zone', got 'other-zone'",
		}))
	})

	It("requires origin on a federated zone CP (REST)", func() {
		ctx := mtDesc()
		ctx.FederatedZone = true
		violations := labels.Validate(map[string]string{}, ctx)
		Expect(violations).To(ContainElement(labels.Violation{
			Key:    mesh_proto.ResourceOriginLabel,
			Reason: "the kuma.io/origin label must be set to 'zone'",
		}))
	})

	It("rejects origin=global on a zone CP", func() {
		ctx := mtDesc()
		ctx.FederatedZone = true
		violations := labels.Validate(map[string]string{
			mesh_proto.ResourceOriginLabel: "global",
		}, ctx)
		Expect(violations).To(ContainElement(labels.Violation{
			Key:    mesh_proto.ResourceOriginLabel,
			Reason: "kuma.io/origin should be 'zone', got 'global'",
		}))
	})

	It("rejects user-set kuma.io/env on global CP (not applicable)", func() {
		ctx := mtDesc()
		ctx.Mode = config_core.Global
		violations := labels.Validate(map[string]string{
			mesh_proto.EnvTag: "universal",
		}, ctx)
		Expect(violations).To(ContainElement(labels.Violation{
			Key:    mesh_proto.EnvTag,
			Reason: "is a reserved label managed by the control plane and cannot be set on apply",
		}))
	})

	It("accepts kuma.io/env=universal on Universal zone CP", func() {
		ctx := mtDesc()
		ctx.FederatedZone = true
		violations := labels.Validate(map[string]string{
			mesh_proto.ResourceOriginLabel: "zone",
			mesh_proto.EnvTag:              "universal",
		}, ctx)
		Expect(violations).To(BeEmpty())
	})

	It("rejects kuma.io/env=universal on K8s zone CP (mismatch — expects 'kubernetes')", func() {
		ctx := mtDesc()
		ctx.FederatedZone = true
		ctx.IsK8s = true
		ctx.Namespace = labels.NewNamespace("kuma-system", true)
		violations := labels.Validate(map[string]string{
			mesh_proto.ResourceOriginLabel: "zone",
			mesh_proto.EnvTag:              "universal",
		}, ctx)
		Expect(violations).To(ContainElement(labels.Violation{
			Key:    mesh_proto.EnvTag,
			Reason: "kuma.io/env should be 'kubernetes', got 'universal'",
		}))
	})

	It("rejects mismatched display-name", func() {
		ctx := mtDesc()
		ctx.FederatedZone = true
		violations := labels.Validate(map[string]string{
			mesh_proto.ResourceOriginLabel: "zone",
			mesh_proto.DisplayName:         "wrong-name",
		}, ctx)
		Expect(violations).To(ContainElement(labels.Violation{
			Key:    mesh_proto.DisplayName,
			Reason: "kuma.io/display-name should be 'mtp-1', got 'wrong-name'",
		}))
	})

	It("rejects arbitrary reserved keys not in the registry", func() {
		ctx := mtDesc()
		ctx.FederatedZone = true
		violations := labels.Validate(map[string]string{
			mesh_proto.ResourceOriginLabel: "zone",
			"kuma.io/foo":                  "bar",
			"k8s.kuma.io/baz":              "qux",
		}, ctx)
		var keys []string
		for _, v := range violations {
			keys = append(keys, v.Key)
		}
		Expect(keys).To(ConsistOf("k8s.kuma.io/baz", "kuma.io/foo"))
	})

	It("allows non-reserved user labels", func() {
		ctx := mtDesc()
		ctx.FederatedZone = true
		violations := labels.Validate(map[string]string{
			mesh_proto.ResourceOriginLabel: "zone",
			"example.com/team":             "platform",
			"app":                          "billing",
		}, ctx)
		Expect(violations).To(BeEmpty())
	})

	It("allows OwnerUser flags with allowed values", func() {
		ctx := mtDesc()
		ctx.FederatedZone = true
		violations := labels.Validate(map[string]string{
			mesh_proto.ResourceOriginLabel: "zone",
			mesh_proto.EffectLabel:         "shadow",
			mesh_proto.KDSSyncLabel:        "disabled",
		}, ctx)
		Expect(violations).To(BeEmpty())
	})

	It("rejects OwnerUser flag values outside the allowed set", func() {
		ctx := mtDesc()
		ctx.FederatedZone = true
		violations := labels.Validate(map[string]string{
			mesh_proto.ResourceOriginLabel: "zone",
			mesh_proto.EffectLabel:         "loud",
		}, ctx)
		Expect(violations).To(ContainElement(labels.Violation{
			Key:    mesh_proto.EffectLabel,
			Reason: "must be one of ['', 'shadow']",
		}))
	})

	It("allows user-set kuma.io/workload on Universal Dataplane", func() {
		ctx := dpDesc()
		ctx.FederatedZone = true
		violations := labels.Validate(map[string]string{
			mesh_proto.ResourceOriginLabel: "zone",
			metadata.KumaWorkload:          "my-workload",
		}, ctx)
		Expect(violations).To(BeEmpty())
	})

	It("rejects kuma.io/workload on a policy (not a proxy)", func() {
		ctx := mtDesc()
		ctx.FederatedZone = true
		violations := labels.Validate(map[string]string{
			mesh_proto.ResourceOriginLabel: "zone",
			metadata.KumaWorkload:          "my-workload",
		}, ctx)
		Expect(violations).To(ContainElement(labels.Violation{
			Key:    metadata.KumaWorkload,
			Reason: "is a reserved label managed by the control plane and cannot be set on apply",
		}))
	})

	It("rejects system-only labels (managed-by) from user input", func() {
		ctx := mtDesc()
		ctx.FederatedZone = true
		violations := labels.Validate(map[string]string{
			mesh_proto.ResourceOriginLabel: "zone",
			mesh_proto.ManagedByLabel:      "anything",
		}, ctx)
		Expect(violations).To(ContainElement(labels.Violation{
			Key:    mesh_proto.ManagedByLabel,
			Reason: "is a reserved label managed by the control plane and cannot be set on apply",
		}))
	})

	It("bypasses validation entirely when Privileged is true", func() {
		ctx := mtDesc()
		ctx.FederatedZone = true
		ctx.Privileged = true
		// This input would otherwise produce several violations.
		violations := labels.Validate(map[string]string{
			mesh_proto.ResourceOriginLabel: "global",
			mesh_proto.MeshTag:             "other-mesh",
			mesh_proto.DisplayName:         "wrong-name",
			mesh_proto.ManagedByLabel:      "kds",
		}, ctx)
		Expect(violations).To(BeEmpty())
	})

	It("honors DisableOriginLabelValidation for the origin spec", func() {
		ctx := mtDesc()
		ctx.FederatedZone = true
		ctx.DisableOriginLabelValidation = true
		violations := labels.Validate(map[string]string{}, ctx)
		Expect(violations).To(BeEmpty())
	})
})
