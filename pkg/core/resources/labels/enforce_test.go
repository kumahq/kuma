package labels_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	config_core "github.com/kumahq/kuma/v3/pkg/config/core"
	meshservice_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshservice/api/v1alpha1"
	resource_labels "github.com/kumahq/kuma/v3/pkg/core/resources/labels"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	meshaccesslog_api "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshaccesslog/api/v1alpha1"
	meshtimeout_api "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshtimeout/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/test/resources/builders"
)

var _ = Describe("EnforcedReadLabels", func() {
	idleTimeout := meshtimeout_api.Conf{
		IdleTimeout: &kube_meta.Duration{Duration: 123 * time.Second},
	}
	meshWideTimeout := func() core_model.Resource {
		return builders.MeshTimeout().
			WithTargetRef(builders.TargetRefMesh()).
			AddTo(builders.TargetRefMesh(), idleTimeout).
			Build()
	}
	appNamespace := resource_labels.WithNamespace(resource_labels.NewNamespace("kuma-demo", false))
	systemNamespace := resource_labels.WithNamespace(resource_labels.NewNamespace("kuma-system", true))
	universal := resource_labels.WithNamespace(resource_labels.UnsetNamespace)
	zoneCP := []resource_labels.Option{resource_labels.WithMode(config_core.Zone), resource_labels.WithZone("zone-1")}
	globalCP := []resource_labels.Option{resource_labels.WithMode(config_core.Global)}

	type testCase struct {
		r        core_model.Resource
		isLocal  bool
		opts     []resource_labels.Option
		expected map[string]string
	}

	DescribeTable("should recompute the control-plane-owned labels",
		func(given testCase) {
			Expect(resource_labels.EnforcedReadLabels(
				given.r.Descriptor(),
				given.r.GetSpec(),
				given.isLocal,
				given.opts...,
			)).To(Equal(given.expected))
		},
		Entry("workload-owner policy in an app namespace", testCase{
			r:       meshWideTimeout(),
			isLocal: true,
			opts:    []resource_labels.Option{appNamespace},
			expected: map[string]string{
				mesh_proto.KubeNamespaceTag: "kuma-demo",
				mesh_proto.PolicyRoleLabel:  string(mesh_proto.ConsumerPolicyRole),
			},
		}),
		Entry("rules-based policy with no to[] is workload-owner", testCase{
			r: builders.MeshAccessLog().
				WithTargetRef(builders.TargetRefMesh()).
				AddRule(builders.MeshAccessLogConf().
					AddBackends(make([]meshaccesslog_api.Backend, 0))).
				Build(),
			isLocal: true,
			opts:    []resource_labels.Option{appNamespace},
			expected: map[string]string{
				mesh_proto.KubeNamespaceTag: "kuma-demo",
				mesh_proto.PolicyRoleLabel:  string(mesh_proto.WorkloadOwnerPolicyRole),
			},
		}),
		Entry("a producer policy stays producer", testCase{
			r: builders.MeshTimeout().
				WithTargetRef(builders.TargetRefMesh()).
				AddTo(builders.TargetRefMeshServiceLabels(map[string]string{
					mesh_proto.DisplayName:      "backend",
					mesh_proto.KubeNamespaceTag: "kuma-demo",
				}, ""), idleTimeout).
				Build(),
			isLocal: true,
			opts:    []resource_labels.Option{appNamespace},
			expected: map[string]string{
				mesh_proto.KubeNamespaceTag: "kuma-demo",
				mesh_proto.PolicyRoleLabel:  string(mesh_proto.ProducerPolicyRole),
			},
		}),
		// A policy that mixes producer and consumer items is rejected at admission, so
		// it can only be stored in a namespace the webhooks never covered. Fall back to
		// the narrowest role rather than propagating an error out of every read.
		Entry("mixed producer and consumer to[] falls back to workload-owner", testCase{
			r: builders.MeshTimeout().
				WithTargetRef(builders.TargetRefMesh()).
				AddTo(builders.TargetRefMeshServiceLabels(map[string]string{
					mesh_proto.DisplayName:      "backend-1",
					mesh_proto.KubeNamespaceTag: "kuma-demo",
				}, ""), idleTimeout).
				AddTo(builders.TargetRefMeshServiceLabels(map[string]string{
					mesh_proto.DisplayName:      "backend-2",
					mesh_proto.KubeNamespaceTag: "other-ns",
				}, ""), idleTimeout).
				Build(),
			isLocal: true,
			opts:    []resource_labels.Option{appNamespace},
			expected: map[string]string{
				mesh_proto.KubeNamespaceTag: "kuma-demo",
				mesh_proto.PolicyRoleLabel:  string(mesh_proto.WorkloadOwnerPolicyRole),
			},
		}),
		Entry("nothing is enforced in the system namespace without a mode", testCase{
			r:        meshWideTimeout(),
			isLocal:  true,
			opts:     []resource_labels.Option{systemNamespace},
			expected: nil,
		}),
		Entry("nothing is enforced on Universal without a mode", testCase{
			r:        meshWideTimeout(),
			isLocal:  false,
			opts:     []resource_labels.Option{universal},
			expected: nil,
		}),
		Entry("a non-policy resource gets no role", testCase{
			r:       meshservice_api.NewMeshServiceResource(),
			isLocal: true,
			opts:    []resource_labels.Option{appNamespace},
			expected: map[string]string{
				mesh_proto.KubeNamespaceTag: "kuma-demo",
			},
		}),
		Entry("local resource on a zone gets origin and zone", testCase{
			r:       meshWideTimeout(),
			isLocal: true,
			opts:    append([]resource_labels.Option{universal}, zoneCP...),
			expected: map[string]string{
				mesh_proto.ResourceOriginLabel: string(mesh_proto.ZoneResourceOrigin),
				mesh_proto.ZoneTag:             "zone-1",
			},
		}),
		Entry("local resource in a k8s app namespace on a zone gets everything", testCase{
			r:       meshWideTimeout(),
			isLocal: true,
			opts:    append([]resource_labels.Option{resource_labels.WithK8s(true), appNamespace}, zoneCP...),
			expected: map[string]string{
				mesh_proto.KubeNamespaceTag:    "kuma-demo",
				mesh_proto.PolicyRoleLabel:     string(mesh_proto.ConsumerPolicyRole),
				mesh_proto.ResourceOriginLabel: string(mesh_proto.ZoneResourceOrigin),
				mesh_proto.ZoneTag:             "zone-1",
			},
		}),
		Entry("import on a zone gets the global origin and no zone", testCase{
			r:       meshWideTimeout(),
			isLocal: false,
			opts:    append([]resource_labels.Option{universal}, zoneCP...),
			expected: map[string]string{
				mesh_proto.ResourceOriginLabel: string(mesh_proto.GlobalResourceOrigin),
			},
		}),
		Entry("local resource on global gets the global origin and no zone", testCase{
			r:       meshWideTimeout(),
			isLocal: true,
			opts:    append([]resource_labels.Option{universal}, globalCP...),
			expected: map[string]string{
				mesh_proto.ResourceOriginLabel: string(mesh_proto.GlobalResourceOrigin),
			},
		}),
		Entry("import on global gets the zone origin and no zone", testCase{
			r:       meshWideTimeout(),
			isLocal: false,
			opts:    append([]resource_labels.Option{universal}, globalCP...),
			expected: map[string]string{
				mesh_proto.ResourceOriginLabel: string(mesh_proto.ZoneResourceOrigin),
			},
		}),
		Entry("zone is not enforced on a type the zone does not provide", testCase{
			r:       builders.Mesh().Build(),
			isLocal: true,
			opts:    append([]resource_labels.Option{universal}, zoneCP...),
			expected: map[string]string{
				mesh_proto.ResourceOriginLabel: string(mesh_proto.ZoneResourceOrigin),
			},
		}),
		Entry("zone is not enforced when the zone has no name", testCase{
			r:       meshWideTimeout(),
			isLocal: true,
			opts:    []resource_labels.Option{universal, resource_labels.WithMode(config_core.Zone)},
			expected: map[string]string{
				mesh_proto.ResourceOriginLabel: string(mesh_proto.ZoneResourceOrigin),
			},
		}),
	)
})
