package k8s

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	common_api "github.com/kumahq/kuma/v2/api/common/v1alpha1"
	"github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	workload_api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/workload/api/v1alpha1"
	workload_k8s "github.com/kumahq/kuma/v2/pkg/core/resources/apis/workload/k8s/v1alpha1"
	k8s_common "github.com/kumahq/kuma/v2/pkg/plugins/common/k8s"
	meshtimeout_api "github.com/kumahq/kuma/v2/pkg/plugins/policies/meshtimeout/api/v1alpha1"
	meshtimeout_k8s "github.com/kumahq/kuma/v2/pkg/plugins/policies/meshtimeout/k8s/v1alpha1"
)

const systemNamespaceForTest = "kuma-system"

var _ = Describe("newMetaAdapter", func() {
	DescribeTable("namespace label",
		func(namespace string, stored map[string]string, expected string) {
			obj := &workload_k8s.Workload{
				Name: "res-1", Namespace: namespace, Labels: stored,
				Spec: &workload_api.Workload{},
			}
			out := workload_api.NewWorkloadResource()
			adapter := newMetaAdapter(obj, systemNamespaceForTest, out.Descriptor(), obj.Spec)

			Expect(adapter.GetLabels()).To(HaveKeyWithValue(v1alpha1.KubeNamespaceTag, expected))
		},
		Entry("overwrites a stored label that disagrees with the namespace",
			"app-ns",
			map[string]string{v1alpha1.KubeNamespaceTag: "other-ns"},
			"app-ns"),
		Entry("sets the label when it is absent (pre-2.9 resource)",
			"app-ns", nil, "app-ns"),
		Entry("overwrites a stored label even when the origin label says global",
			"app-ns",
			map[string]string{
				v1alpha1.KubeNamespaceTag:    "other-ns",
				v1alpha1.ResourceOriginLabel: string(v1alpha1.GlobalResourceOrigin),
			},
			"app-ns"),
		Entry("keeps the label of a resource imported over KDS",
			systemNamespaceForTest,
			map[string]string{
				v1alpha1.KubeNamespaceTag:    "app-ns-on-the-other-cp",
				v1alpha1.ResourceOriginLabel: string(v1alpha1.GlobalResourceOrigin),
			},
			"app-ns-on-the-other-cp"),
	)

	It("should not add the namespace label to a cluster-scoped object", func() {
		obj := &workload_k8s.Workload{
			Name: "default",
			Spec: &workload_api.Workload{},
		}
		out := workload_api.NewWorkloadResource()

		Expect(newMetaAdapter(obj, systemNamespaceForTest, out.Descriptor(), obj.Spec).GetLabels()).
			NotTo(HaveKey(v1alpha1.KubeNamespaceTag))
	})
})

var _ = Describe("enforced label derivation through the converters", func() {
	// A policy stored in a namespace the admission webhooks never covered, so
	// ComputeLabels never ran on it: no namespace label and no role label.
	policyIn := func(namespace string, stored map[string]string) *meshtimeout_k8s.MeshTimeout {
		return &meshtimeout_k8s.MeshTimeout{
			APIVersion:      meshtimeout_k8s.GroupVersion.String(),
			Kind:            "MeshTimeout",
			Namespace:       namespace,
			Name:            "idle-timeout",
			ResourceVersion: "1",
			Labels:          stored,
			Spec: &meshtimeout_api.MeshTimeout{
				TargetRef: &common_api.TargetRef{Kind: common_api.Mesh},
			},
		}
	}

	labelsOf := func(converter k8s_common.Converter, obj *meshtimeout_k8s.MeshTimeout) map[string]string {
		out := meshtimeout_api.NewMeshTimeoutResource()
		Expect(converter.ToCoreResource(obj, out)).To(Succeed())
		return out.GetMeta().GetLabels()
	}

	simple := func() k8s_common.Converter { return NewSimpleConverter(systemNamespaceForTest) }
	caching := func() k8s_common.Converter { return NewCachingConverter(5*time.Minute, systemNamespaceForTest) }

	stale := map[string]string{
		v1alpha1.KubeNamespaceTag:    "other-ns",
		v1alpha1.PolicyRoleLabel:     string(v1alpha1.SystemPolicyRole),
		v1alpha1.ResourceOriginLabel: string(v1alpha1.GlobalResourceOrigin),
	}

	DescribeTable("should derive both labels from the object namespace",
		func(newConverter func() k8s_common.Converter, namespace string, stored map[string]string, expected map[string]string) {
			Expect(labelsOf(newConverter(), policyIn(namespace, stored))).To(SatisfyAll(
				HaveKeyWithValue(v1alpha1.KubeNamespaceTag, expected[v1alpha1.KubeNamespaceTag]),
				HaveKeyWithValue(v1alpha1.PolicyRoleLabel, expected[v1alpha1.PolicyRoleLabel]),
			))
		},
		Entry("SimpleConverter overwrites stale labels", simple, "app-ns", stale, map[string]string{
			v1alpha1.KubeNamespaceTag: "app-ns",
			v1alpha1.PolicyRoleLabel:  string(v1alpha1.WorkloadOwnerPolicyRole),
		}),
		Entry("CachingConverter overwrites stale labels", caching, "app-ns", stale, map[string]string{
			v1alpha1.KubeNamespaceTag: "app-ns",
			v1alpha1.PolicyRoleLabel:  string(v1alpha1.WorkloadOwnerPolicyRole),
		}),
		Entry("SimpleConverter derives labels a missing webhook never wrote", simple, "app-ns", nil, map[string]string{
			v1alpha1.KubeNamespaceTag: "app-ns",
			v1alpha1.PolicyRoleLabel:  string(v1alpha1.WorkloadOwnerPolicyRole),
		}),
		Entry("CachingConverter derives labels a missing webhook never wrote", caching, "app-ns", nil, map[string]string{
			v1alpha1.KubeNamespaceTag: "app-ns",
			v1alpha1.PolicyRoleLabel:  string(v1alpha1.WorkloadOwnerPolicyRole),
		}),
		Entry("SimpleConverter keeps the stored labels in the system namespace", simple, systemNamespaceForTest, stale, map[string]string{
			v1alpha1.KubeNamespaceTag: "other-ns",
			v1alpha1.PolicyRoleLabel:  string(v1alpha1.SystemPolicyRole),
		}),
		Entry("CachingConverter keeps the stored labels in the system namespace", caching, systemNamespaceForTest, stale, map[string]string{
			v1alpha1.KubeNamespaceTag: "other-ns",
			v1alpha1.PolicyRoleLabel:  string(v1alpha1.SystemPolicyRole),
		}),
	)

	// On a cache hit only the spec comes from the cache; the labels still have to be
	// derived from the object, enforcement included.
	It("should return the derived labels on a CachingConverter cache hit", func() {
		converter := NewCachingConverter(5*time.Minute, systemNamespaceForTest)
		obj := policyIn("app-ns", stale)

		miss := labelsOf(converter, obj)
		hit := labelsOf(converter, obj)

		Expect(miss).To(HaveKeyWithValue(v1alpha1.KubeNamespaceTag, "app-ns"))
		Expect(miss).To(HaveKeyWithValue(v1alpha1.PolicyRoleLabel, string(v1alpha1.WorkloadOwnerPolicyRole)))
		Expect(hit).To(Equal(miss))
	})

	// The same missing webhook that leaves the role label off also leaves the spec
	// unvalidated, so a stored policy can have no spec at all. GetSpec hands back a
	// typed nil for it, and deriving a role from that would dereference a nil policy
	// on every read, panicking every conversion of that type.
	DescribeTable("should not panic on a stored policy with no spec",
		func(newConverter func() k8s_common.Converter) {
			obj := policyIn("app-ns", nil)
			obj.Spec = nil
			out := meshtimeout_api.NewMeshTimeoutResource()

			Expect(newConverter().ToCoreResource(obj, out)).To(Succeed())
			Expect(out.GetMeta().GetLabels()).To(SatisfyAll(
				HaveKeyWithValue(v1alpha1.KubeNamespaceTag, "app-ns"),
				HaveKeyWithValue(v1alpha1.PolicyRoleLabel, string(v1alpha1.WorkloadOwnerPolicyRole)),
			))
		},
		Entry("SimpleConverter", simple),
		Entry("CachingConverter", caching),
	)
})
