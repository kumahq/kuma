package labels_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/config/core"
	"github.com/kumahq/kuma/v2/pkg/core/resources/apis/mesh"
	resource_labels "github.com/kumahq/kuma/v2/pkg/core/resources/labels"
	core_model "github.com/kumahq/kuma/v2/pkg/core/resources/model"
	meshtimeout_api "github.com/kumahq/kuma/v2/pkg/plugins/policies/meshtimeout/api/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/test/kds/samples"
	"github.com/kumahq/kuma/v2/pkg/test/resources/builders"
	test_model "github.com/kumahq/kuma/v2/pkg/test/resources/model"
)

var _ = Describe("ComputePolicyRole", func() {
	type testCase struct {
		policy       core_model.Policy
		namespace    resource_labels.Namespace
		expectedRole mesh_proto.PolicyRole
		expectedErr  string
	}

	DescribeTable("should compute the correct policy role",
		func(given testCase) {
			role, err := resource_labels.ComputePolicyRole(given.policy, given.namespace)
			if given.expectedErr != "" {
				Expect(err.Error()).To(Equal(given.expectedErr))
			} else {
				Expect(err).ToNot(HaveOccurred())
			}
			Expect(role).To(Equal(given.expectedRole))
		},
		Entry("consumer policy", testCase{
			policy: builders.MeshTimeout().
				WithMesh("mesh-1").WithName("name-1").
				WithTargetRef(builders.TargetRefMesh()).
				AddTo(builders.TargetRefMesh(), meshtimeout_api.Conf{
					IdleTimeout: &kube_meta.Duration{Duration: 123 * time.Second},
				}).
				Build().Spec,
			namespace:    resource_labels.NewNamespace("kuma-demo", false),
			expectedRole: mesh_proto.ConsumerPolicyRole,
		}),
		Entry("consumer policy with labels", testCase{
			policy: builders.MeshTimeout().
				WithMesh("mesh-1").WithName("name-1").
				WithTargetRef(builders.TargetRefMesh()).
				AddTo(builders.TargetRefMeshServiceLabels(map[string]string{
					"kuma.io/display-name":  "test",
					"kuma.io/zone":          "zone-1",
					"k8s.kuma.io/namespace": "kuma-demo",
				}, ""), meshtimeout_api.Conf{
					IdleTimeout: &kube_meta.Duration{Duration: 123 * time.Second},
				}).
				Build().Spec,
			namespace:    resource_labels.NewNamespace("kuma-demo", false),
			expectedRole: mesh_proto.ConsumerPolicyRole,
		}),
		Entry("producer policy", testCase{
			policy: builders.MeshTimeout().
				WithMesh("mesh-1").WithName("name-1").
				WithTargetRef(builders.TargetRefMesh()).
				AddTo(builders.TargetRefMeshService("backend", "kuma-demo", ""), meshtimeout_api.Conf{
					IdleTimeout: &kube_meta.Duration{Duration: 123 * time.Second},
				}).
				Build().Spec,
			namespace:    resource_labels.NewNamespace("kuma-demo", false),
			expectedRole: mesh_proto.ProducerPolicyRole,
		}),
		Entry("producer policy with no namespace in to[]", testCase{
			policy: builders.MeshTimeout().
				WithMesh("mesh-1").WithName("name-1").
				WithTargetRef(builders.TargetRefMesh()).
				AddTo(builders.TargetRefMeshService("backend", "", ""), meshtimeout_api.Conf{
					IdleTimeout: &kube_meta.Duration{Duration: 123 * time.Second},
				}).
				Build().Spec,
			namespace:    resource_labels.NewNamespace("kuma-demo", false),
			expectedRole: mesh_proto.ProducerPolicyRole,
		}),
		Entry("producer policy for MeshHTTPRoute", testCase{
			policy: builders.MeshTimeout().
				WithMesh("mesh-1").WithName("name-1").
				WithTargetRef(builders.TargetRefMesh()).
				AddTo(builders.TargetRefMeshHTTPRoute("route-1", "kuma-demo"), meshtimeout_api.Conf{
					IdleTimeout: &kube_meta.Duration{Duration: 123 * time.Second},
				}).
				Build().Spec,
			namespace:    resource_labels.NewNamespace("kuma-demo", false),
			expectedRole: mesh_proto.ProducerPolicyRole,
		}),
		Entry("workload-owner policy with from", testCase{
			policy: builders.MeshTimeout().
				WithMesh("mesh-1").WithName("name-1").
				WithTargetRef(builders.TargetRefMesh()).
				AddFrom(builders.TargetRefMesh(), meshtimeout_api.Conf{
					IdleTimeout: &kube_meta.Duration{Duration: 123 * time.Second},
				}).
				Build().Spec,
			namespace:    resource_labels.NewNamespace("kuma-demo", false),
			expectedRole: mesh_proto.WorkloadOwnerPolicyRole,
		}),
		Entry("workload-owner policy with both from and to", testCase{
			policy: builders.MeshTimeout().
				WithMesh("mesh-1").WithName("name-1").
				WithTargetRef(builders.TargetRefMesh()).
				AddTo(builders.TargetRefMesh(), meshtimeout_api.Conf{
					IdleTimeout: &kube_meta.Duration{Duration: 123 * time.Second},
				}).
				AddFrom(builders.TargetRefMesh(), meshtimeout_api.Conf{
					IdleTimeout: &kube_meta.Duration{Duration: 123 * time.Second},
				}).
				Build().Spec,
			namespace:   resource_labels.NewNamespace("kuma-demo", false),
			expectedErr: "it's not allowed to mix 'to' and 'from' arrays in the same policy",
		}),
		Entry("consumer policy with from", testCase{
			policy: builders.MeshTimeout().
				WithMesh("mesh-1").WithName("name-1").
				WithTargetRef(builders.TargetRefMesh()).
				AddTo(builders.TargetRefMeshService("backend", "backend-ns", ""), meshtimeout_api.Conf{
					IdleTimeout: &kube_meta.Duration{Duration: 123 * time.Second},
				}).
				AddFrom(builders.TargetRefMesh(), meshtimeout_api.Conf{
					IdleTimeout: &kube_meta.Duration{Duration: 123 * time.Second},
				}).
				Build().Spec,
			namespace:   resource_labels.NewNamespace("kuma-demo", false),
			expectedErr: "it's not allowed to mix 'to' and 'from' arrays in the same policy",
		}),
		Entry("system policy with both from and to", testCase{
			policy: builders.MeshTimeout().
				WithMesh("mesh-1").WithName("name-1").
				WithTargetRef(builders.TargetRefMesh()).
				AddTo(builders.TargetRefMesh(), meshtimeout_api.Conf{
					IdleTimeout: &kube_meta.Duration{Duration: 123 * time.Second},
				}).
				AddFrom(builders.TargetRefMesh(), meshtimeout_api.Conf{
					IdleTimeout: &kube_meta.Duration{Duration: 123 * time.Second},
				}).
				Build().Spec,
			namespace:    resource_labels.NewNamespace("kuma-system", true),
			expectedRole: mesh_proto.SystemPolicyRole,
		}),
		Entry("policy with consumer and producer to-items", testCase{
			policy: builders.MeshTimeout().
				WithMesh("mesh-1").WithName("name-1").
				WithTargetRef(builders.TargetRefMesh()).
				AddTo(builders.TargetRefMeshService("backend-1", "backend-1-ns", ""), meshtimeout_api.Conf{
					IdleTimeout: &kube_meta.Duration{Duration: 123 * time.Second},
				}).
				AddTo(builders.TargetRefMeshService("backend-2", "backend-2-ns", ""), meshtimeout_api.Conf{
					IdleTimeout: &kube_meta.Duration{Duration: 123 * time.Second},
				}).
				Build().Spec,
			namespace:   resource_labels.NewNamespace("backend-1-ns", false),
			expectedErr: "it's not allowed to mix producer and consumer items in the same policy",
		}),
	)
})

var _ = Describe("Compute", func() {
	type testCase struct {
		r              core_model.Resource
		mode           core.CpMode
		isK8s          bool
		localZone      string
		expectedLabels map[string]string
	}

	DescribeTable("should return correct label map",
		func(given testCase) {
			labels, err := resource_labels.Compute(
				given.r.Descriptor(),
				given.r.GetSpec(),
				given.r.GetMeta().GetLabels(),
				given.r.GetMeta().GetMesh(),
				resource_labels.WithNamespace(resource_labels.GetNamespace(given.r.GetMeta(), "kuma-system")),
				resource_labels.WithMode(given.mode),
				resource_labels.WithK8s(given.isK8s),
				resource_labels.WithZone(given.localZone),
			)
			Expect(err).ToNot(HaveOccurred())
			Expect(labels).To(Equal(given.expectedLabels))
		},
		Entry("plugin originated policy on zone-k8s", testCase{
			mode:      core.Zone,
			isK8s:     true,
			localZone: "zone-1",
			r: builders.MeshTimeout().
				WithMesh("mesh-1").WithName("idle-timeout").
				WithTargetRef(builders.TargetRefMesh()).
				AddTo(builders.TargetRefMesh(), meshtimeout_api.Conf{
					IdleTimeout: &kube_meta.Duration{Duration: 123 * time.Second},
				}).
				Build(),
			expectedLabels: map[string]string{
				"kuma.io/env":    "kubernetes",
				"kuma.io/mesh":   "mesh-1",
				"kuma.io/origin": "zone",
				"kuma.io/zone":   "zone-1",
			},
		}),
		Entry("source/destination policy on zone-k8s", testCase{
			mode:      core.Zone,
			isK8s:     true,
			localZone: "zone-1",
			r: &mesh.TimeoutResource{
				Spec: samples.Timeout,
				Meta: &test_model.ResourceMeta{Mesh: "mesh-1", Name: "sample-timeout"},
			},
			expectedLabels: map[string]string{
				"kuma.io/mesh": "mesh-1",
			},
		}),
		Entry("mesh resource on non-federated zone", testCase{
			mode:      core.Zone,
			isK8s:     true,
			localZone: "zone-1",
			r: &mesh.MeshResource{
				Spec: samples.Mesh1,
				Meta: &test_model.ResourceMeta{Mesh: core_model.NoMesh, Name: "mesh-1"},
			},
			expectedLabels: map[string]string{},
		}),
		Entry("plugin originated policy on zone-k8s on custom namespace", testCase{
			mode:      core.Zone,
			isK8s:     true,
			localZone: "zone-1",
			r: builders.MeshTimeout().
				WithMesh("mesh-1").
				WithName("idle-timeout").
				WithNamespace("custom-ns").
				WithTargetRef(builders.TargetRefMesh()).
				AddTo(builders.TargetRefMesh(), meshtimeout_api.Conf{
					IdleTimeout: &kube_meta.Duration{Duration: 123 * time.Second},
				}).
				Build(),
			expectedLabels: map[string]string{
				"k8s.kuma.io/namespace": "custom-ns",
				"kuma.io/policy-role":   "consumer",
				"kuma.io/mesh":          "mesh-1",
				"kuma.io/origin":        "zone",
				"kuma.io/zone":          "zone-1",
				"kuma.io/env":           "kubernetes",
			},
		}),
		Entry("gateway dataplane proxy", testCase{
			mode:      core.Zone,
			isK8s:     true,
			localZone: "zone-1",
			r: builders.Dataplane().
				WithMesh("mesh-1").
				WithBuiltInGateway("test-gateway").
				Build(),
			expectedLabels: map[string]string{
				"kuma.io/mesh":       "mesh-1",
				"kuma.io/origin":     "zone",
				"kuma.io/zone":       "zone-1",
				"kuma.io/env":        "kubernetes",
				"kuma.io/proxy-type": "gateway",
			},
		}),
		Entry("dataplane proxy", testCase{
			mode:      core.Zone,
			isK8s:     true,
			localZone: "zone-1",
			r: builders.Dataplane().
				WithName("backend-1").
				WithServices("backend").
				WithMesh("mesh-1").
				Build(),
			expectedLabels: map[string]string{
				"kuma.io/mesh":       "mesh-1",
				"kuma.io/origin":     "zone",
				"kuma.io/zone":       "zone-1",
				"kuma.io/env":        "kubernetes",
				"kuma.io/proxy-type": "sidecar",
			},
		}),
		Entry("zone egress proxy", testCase{
			mode:      core.Zone,
			isK8s:     true,
			localZone: "zone-1",
			r: builders.ZoneEgress().
				WithPort(1001).
				Build(),
			expectedLabels: map[string]string{
				"kuma.io/origin":     "zone",
				"kuma.io/zone":       "zone-1",
				"kuma.io/env":        "kubernetes",
				"kuma.io/proxy-type": "zoneegress",
			},
		}),
	)
})
