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

	type testCase struct {
		r        core_model.Resource
		stored   map[string]string
		opts     []resource_labels.Option
		expected map[string]string
	}

	DescribeTable("should recompute the control-plane-owned labels",
		func(given testCase) {
			Expect(resource_labels.EnforcedReadLabels(
				given.r.Descriptor(),
				given.r.GetSpec(),
				given.stored,
				given.opts...,
			)).To(Equal(given.expected))
		},
		Entry("workload-owner policy in an app namespace", testCase{
			r: builders.MeshTimeout().
				WithTargetRef(builders.TargetRefMesh()).
				AddTo(builders.TargetRefMesh(), idleTimeout).
				Build(),
			opts: []resource_labels.Option{resource_labels.WithNamespace(resource_labels.NewNamespace("kuma-demo", false))},
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
			opts: []resource_labels.Option{resource_labels.WithNamespace(resource_labels.NewNamespace("kuma-demo", false))},
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
			opts: []resource_labels.Option{resource_labels.WithNamespace(resource_labels.NewNamespace("kuma-demo", false))},
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
			opts: []resource_labels.Option{resource_labels.WithNamespace(resource_labels.NewNamespace("kuma-demo", false))},
			expected: map[string]string{
				mesh_proto.KubeNamespaceTag: "kuma-demo",
				mesh_proto.PolicyRoleLabel:  string(mesh_proto.WorkloadOwnerPolicyRole),
			},
		}),
		Entry("nothing is enforced in the system namespace", testCase{
			r: builders.MeshTimeout().
				WithTargetRef(builders.TargetRefMesh()).
				AddTo(builders.TargetRefMesh(), idleTimeout).
				Build(),
			opts:     []resource_labels.Option{resource_labels.WithNamespace(resource_labels.NewNamespace("kuma-system", true))},
			expected: nil,
		}),
		Entry("a stale zone on a zone-originated policy is replaced", testCase{
			r: builders.MeshTimeout().
				WithTargetRef(builders.TargetRefMesh()).
				AddTo(builders.TargetRefMesh(), idleTimeout).
				Build(),
			stored: map[string]string{
				mesh_proto.ResourceOriginLabel: string(mesh_proto.ZoneResourceOrigin),
				mesh_proto.ZoneTag:             "default",
			},
			opts: []resource_labels.Option{
				resource_labels.WithNamespace(resource_labels.NewNamespace("kuma-system", true)),
				resource_labels.WithMode(config_core.Zone),
				resource_labels.WithZone("kuma-2"),
			},
			expected: map[string]string{
				mesh_proto.ZoneTag: "kuma-2",
			},
		}),
		Entry("the zone is enforced alongside the namespace and role", testCase{
			r: builders.MeshTimeout().
				WithTargetRef(builders.TargetRefMesh()).
				AddTo(builders.TargetRefMesh(), idleTimeout).
				Build(),
			stored: map[string]string{
				mesh_proto.ResourceOriginLabel: string(mesh_proto.ZoneResourceOrigin),
				mesh_proto.ZoneTag:             "default",
			},
			opts: []resource_labels.Option{
				resource_labels.WithNamespace(resource_labels.NewNamespace("kuma-demo", false)),
				resource_labels.WithMode(config_core.Zone),
				resource_labels.WithZone("kuma-2"),
			},
			expected: map[string]string{
				mesh_proto.ZoneTag:          "kuma-2",
				mesh_proto.KubeNamespaceTag: "kuma-demo",
				mesh_proto.PolicyRoleLabel:  string(mesh_proto.ConsumerPolicyRole),
			},
		}),
		// Global relays another zone's producer policy with its kuma.io/zone intact
		// but origin rewritten to global; it must not be claimed by the reader.
		Entry("a resource synced from global keeps its zone", testCase{
			r: builders.MeshTimeout().
				WithTargetRef(builders.TargetRefMesh()).
				AddTo(builders.TargetRefMesh(), idleTimeout).
				Build(),
			stored: map[string]string{
				mesh_proto.ResourceOriginLabel: string(mesh_proto.GlobalResourceOrigin),
				mesh_proto.ZoneTag:             "kuma-3",
			},
			opts: []resource_labels.Option{
				resource_labels.WithNamespace(resource_labels.NewNamespace("kuma-system", true)),
				resource_labels.WithMode(config_core.Zone),
				resource_labels.WithZone("kuma-2"),
			},
			expected: nil,
		}),
		Entry("a global CP does not rewrite the zone of a synced resource", testCase{
			r: builders.MeshTimeout().
				WithTargetRef(builders.TargetRefMesh()).
				AddTo(builders.TargetRefMesh(), idleTimeout).
				Build(),
			stored: map[string]string{
				mesh_proto.ResourceOriginLabel: string(mesh_proto.ZoneResourceOrigin),
				mesh_proto.ZoneTag:             "kuma-2",
			},
			opts: []resource_labels.Option{
				resource_labels.WithNamespace(resource_labels.NewNamespace("kuma-system", true)),
				resource_labels.WithMode(config_core.Global),
				resource_labels.WithZone("default"),
			},
			expected: nil,
		}),
		// Pre-2.6 objects carry no origin; nothing verified them, and matching does
		// not consult the zone label without an origin anyway.
		Entry("an object with no origin is not claimed", testCase{
			r: builders.MeshTimeout().
				WithTargetRef(builders.TargetRefMesh()).
				AddTo(builders.TargetRefMesh(), idleTimeout).
				Build(),
			stored: map[string]string{
				mesh_proto.ZoneTag: "default",
			},
			opts: []resource_labels.Option{
				resource_labels.WithNamespace(resource_labels.NewNamespace("kuma-system", true)),
				resource_labels.WithMode(config_core.Zone),
				resource_labels.WithZone("kuma-2"),
			},
			expected: nil,
		}),
		Entry("nothing is enforced on Universal", testCase{
			r: builders.MeshTimeout().
				WithTargetRef(builders.TargetRefMesh()).
				AddTo(builders.TargetRefMesh(), idleTimeout).
				Build(),
			opts:     []resource_labels.Option{resource_labels.WithNamespace(resource_labels.UnsetNamespace)},
			expected: nil,
		}),
		Entry("a non-policy resource gets no role", testCase{
			r:    meshservice_api.NewMeshServiceResource(),
			opts: []resource_labels.Option{resource_labels.WithNamespace(resource_labels.NewNamespace("kuma-demo", false))},
			expected: map[string]string{
				mesh_proto.KubeNamespaceTag: "kuma-demo",
			},
		}),
	)
})
