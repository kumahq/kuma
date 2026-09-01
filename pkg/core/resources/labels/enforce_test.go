package labels_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	meshservice_api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshservice/api/v1alpha1"
	resource_labels "github.com/kumahq/kuma/v2/pkg/core/resources/labels"
	core_model "github.com/kumahq/kuma/v2/pkg/core/resources/model"
	meshaccesslog_api "github.com/kumahq/kuma/v2/pkg/plugins/policies/meshaccesslog/api/v1alpha1"
	meshtimeout_api "github.com/kumahq/kuma/v2/pkg/plugins/policies/meshtimeout/api/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/test/resources/builders"
)

var _ = Describe("EnforcedReadLabels", func() {
	idleTimeout := meshtimeout_api.Conf{
		IdleTimeout: &kube_meta.Duration{Duration: 123 * time.Second},
	}

	type testCase struct {
		r         core_model.Resource
		namespace resource_labels.Namespace
		expected  map[string]string
	}

	DescribeTable("should recompute the control-plane-owned labels",
		func(given testCase) {
			Expect(resource_labels.EnforcedReadLabels(
				given.r.Descriptor(),
				given.r.GetSpec(),
				given.namespace,
			)).To(Equal(given.expected))
		},
		Entry("workload-owner policy in an app namespace", testCase{
			r: builders.MeshTimeout().
				WithTargetRef(builders.TargetRefMesh()).
				AddTo(builders.TargetRefMesh(), idleTimeout).
				Build(),
			namespace: resource_labels.NewNamespace("kuma-demo", false),
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
			namespace: resource_labels.NewNamespace("kuma-demo", false),
			expected: map[string]string{
				mesh_proto.KubeNamespaceTag: "kuma-demo",
				mesh_proto.PolicyRoleLabel:  string(mesh_proto.WorkloadOwnerPolicyRole),
			},
		}),
		Entry("a producer policy stays producer", testCase{
			r: builders.MeshTimeout().
				WithTargetRef(builders.TargetRefMesh()).
				AddTo(builders.TargetRefMeshService("backend", "kuma-demo", ""), idleTimeout).
				Build(),
			namespace: resource_labels.NewNamespace("kuma-demo", false),
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
				AddTo(builders.TargetRefMeshService("backend-1", "kuma-demo", ""), idleTimeout).
				AddTo(builders.TargetRefMeshService("backend-2", "other-ns", ""), idleTimeout).
				Build(),
			namespace: resource_labels.NewNamespace("kuma-demo", false),
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
			namespace: resource_labels.NewNamespace("kuma-system", true),
			expected:  nil,
		}),
		Entry("nothing is enforced on Universal", testCase{
			r: builders.MeshTimeout().
				WithTargetRef(builders.TargetRefMesh()).
				AddTo(builders.TargetRefMesh(), idleTimeout).
				Build(),
			namespace: resource_labels.UnsetNamespace,
			expected:  nil,
		}),
		Entry("a non-policy resource gets no role", testCase{
			r:         meshservice_api.NewMeshServiceResource(),
			namespace: resource_labels.NewNamespace("kuma-demo", false),
			expected: map[string]string{
				mesh_proto.KubeNamespaceTag: "kuma-demo",
			},
		}),
	)
})
