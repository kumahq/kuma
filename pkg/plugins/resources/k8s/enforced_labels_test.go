package k8s

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	common_api "github.com/kumahq/kuma/v3/api/common/v1alpha1"
	"github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	config_core "github.com/kumahq/kuma/v3/pkg/config/core"
	workload_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/workload/api/v1alpha1"
	workload_k8s "github.com/kumahq/kuma/v3/pkg/core/resources/apis/workload/k8s/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/resources/labels"
	k8s_common "github.com/kumahq/kuma/v3/pkg/plugins/common/k8s"
	meshtimeout_api "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshtimeout/api/v1alpha1"
	meshtimeout_k8s "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshtimeout/k8s/v1alpha1"
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
			Expect(out.SetSpec(obj.Spec)).To(Succeed())

			adapter := newMetaAdapter(obj, out, systemNamespaceForTest, labels.ControlPlane{})

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
		Expect(out.SetSpec(obj.Spec)).To(Succeed())

		adapter := newMetaAdapter(obj, out, systemNamespaceForTest, labels.ControlPlane{})

		Expect(adapter.GetLabels()).NotTo(HaveKey(v1alpha1.KubeNamespaceTag))
	})
})

var _ = Describe("enforced label derivation through the converters", func() {
	// A policy stored in a namespace the admission webhooks never covered, so
	// labels.Compute never ran on it: no namespace label and no role label.
	policyIn := func(namespace string, stored map[string]string) *meshtimeout_k8s.MeshTimeout {
		return &meshtimeout_k8s.MeshTimeout{
			APIVersion:      meshtimeout_k8s.GroupVersion.String(),
			Kind:            "MeshTimeout",
			Namespace:       namespace,
			Name:            "idle-timeout",
			ResourceVersion: "1",
			Labels:          stored,
			Spec: &meshtimeout_api.MeshTimeout{
				TargetRef: &common_api.TopLevelTargetRef{Kind: common_api.TopLevelTargetRefKindMesh},
			},
		}
	}

	labelsOf := func(converter k8s_common.Converter, obj *meshtimeout_k8s.MeshTimeout) map[string]string {
		out := meshtimeout_api.NewMeshTimeoutResource()
		Expect(converter.ToCoreResource(obj, out)).To(Succeed())
		return out.GetMeta().GetLabels()
	}

	simple := func() k8s_common.Converter { return NewSimpleConverter(systemNamespaceForTest, labels.ControlPlane{}) }
	caching := func() k8s_common.Converter {
		return NewCachingConverter(5*time.Minute, systemNamespaceForTest, labels.ControlPlane{})
	}

	stale := map[string]string{
		v1alpha1.KubeNamespaceTag:    "other-ns",
		v1alpha1.ResourceOriginLabel: string(v1alpha1.GlobalResourceOrigin),
	}

	DescribeTable("should derive the namespace label from the object namespace",
		func(newConverter func() k8s_common.Converter, namespace string, stored map[string]string, expected string) {
			Expect(labelsOf(newConverter(), policyIn(namespace, stored))).To(HaveKeyWithValue(v1alpha1.KubeNamespaceTag, expected))
		},
		Entry("SimpleConverter overwrites a stale label", simple, "app-ns", stale, "app-ns"),
		Entry("CachingConverter overwrites a stale label", caching, "app-ns", stale, "app-ns"),
		Entry("SimpleConverter derives the label a missing webhook never wrote", simple, "app-ns", nil, "app-ns"),
		Entry("CachingConverter derives the label a missing webhook never wrote", caching, "app-ns", nil, "app-ns"),
		Entry("SimpleConverter keeps the stored label in the system namespace", simple, systemNamespaceForTest, stale, "other-ns"),
		Entry("CachingConverter keeps the stored label in the system namespace", caching, systemNamespaceForTest, stale, "other-ns"),
	)

	// On a cache hit the adapter is handed the labels stored on the miss, so the
	// derivation has to already be baked into the cached entry.
	It("should return the derived labels on a CachingConverter cache hit", func() {
		converter := NewCachingConverter(5*time.Minute, systemNamespaceForTest, labels.ControlPlane{})
		obj := policyIn("app-ns", stale)

		miss := labelsOf(converter, obj)
		hit := labelsOf(converter, obj)

		Expect(miss).To(HaveKeyWithValue(v1alpha1.KubeNamespaceTag, "app-ns"))
		Expect(hit).To(Equal(miss))
	})

	zoneCP := labels.ControlPlane{Mode: config_core.Zone, Zone: "zone-1"}
	simpleOnZone := func() k8s_common.Converter { return NewSimpleConverter(systemNamespaceForTest, zoneCP) }
	cachingOnZone := func() k8s_common.Converter {
		return NewCachingConverter(5*time.Minute, systemNamespaceForTest, zoneCP)
	}
	importedFromGlobal := map[string]string{
		v1alpha1.ResourceOriginLabel: string(v1alpha1.GlobalResourceOrigin),
		v1alpha1.ZoneTag:             "other-zone",
	}
	local := map[string]string{
		v1alpha1.ResourceOriginLabel: string(v1alpha1.ZoneResourceOrigin),
		v1alpha1.ZoneTag:             "zone-1",
	}

	DescribeTable("should enforce origin and zone from the CP mode",
		func(newConverter func() k8s_common.Converter, namespace string, stored map[string]string, expected map[string]string) {
			got := labelsOf(newConverter(), policyIn(namespace, stored))
			for _, key := range []string{v1alpha1.ResourceOriginLabel, v1alpha1.ZoneTag} {
				if value, ok := expected[key]; ok {
					Expect(got).To(HaveKeyWithValue(key, value))
				} else {
					Expect(got).NotTo(HaveKey(key))
				}
			}
		},
		Entry("SimpleConverter overrides a claimed import in an app namespace", simpleOnZone, "app-ns", importedFromGlobal, local),
		Entry("CachingConverter overrides a claimed import in an app namespace", cachingOnZone, "app-ns", importedFromGlobal, local),
		Entry("SimpleConverter keeps an import in the system namespace", simpleOnZone, systemNamespaceForTest, importedFromGlobal, importedFromGlobal),
		Entry("CachingConverter keeps an import in the system namespace", cachingOnZone, systemNamespaceForTest, importedFromGlobal, importedFromGlobal),
		Entry("SimpleConverter fills in absent labels in the system namespace", simpleOnZone, systemNamespaceForTest, nil, local),
		Entry("CachingConverter fills in absent labels in the system namespace", cachingOnZone, systemNamespaceForTest, nil, local),
		// The admission webhooks' converter has no mode, so the labels it hands to
		// validation are the ones the user supplied.
		Entry("SimpleConverter without a mode leaves both labels alone", simple, "app-ns", nil, nil),
		Entry("CachingConverter without a mode leaves both labels alone", caching, "app-ns", nil, nil),
	)

	// A policy in the system namespace applies mesh-wide and carries no namespace
	// label. One an older control plane stored would scope it to the system namespace,
	// so the read drops it; an import keeps the namespace it came with.
	DescribeTable("should drop a stale namespace label from a local policy in the system namespace",
		func(newConverter func() k8s_common.Converter, stored map[string]string, expected string) {
			got := labelsOf(newConverter(), policyIn(systemNamespaceForTest, stored))
			if expected == "" {
				Expect(got).NotTo(HaveKey(v1alpha1.KubeNamespaceTag))
			} else {
				Expect(got).To(HaveKeyWithValue(v1alpha1.KubeNamespaceTag, expected))
			}
		},
		Entry("SimpleConverter drops it from a local policy", simpleOnZone, map[string]string{
			v1alpha1.KubeNamespaceTag:    systemNamespaceForTest,
			v1alpha1.ResourceOriginLabel: string(v1alpha1.ZoneResourceOrigin),
		}, ""),
		Entry("CachingConverter drops it from a local policy", cachingOnZone, map[string]string{
			v1alpha1.KubeNamespaceTag:    systemNamespaceForTest,
			v1alpha1.ResourceOriginLabel: string(v1alpha1.ZoneResourceOrigin),
		}, ""),
		Entry("SimpleConverter drops it from a local policy with no stored origin", simpleOnZone, map[string]string{
			v1alpha1.KubeNamespaceTag: systemNamespaceForTest,
		}, ""),
		Entry("SimpleConverter keeps it on an import", simpleOnZone, map[string]string{
			v1alpha1.KubeNamespaceTag:    "app-ns",
			v1alpha1.ResourceOriginLabel: string(v1alpha1.GlobalResourceOrigin),
		}, "app-ns"),
		Entry("CachingConverter keeps it on an import", cachingOnZone, map[string]string{
			v1alpha1.KubeNamespaceTag:    "app-ns",
			v1alpha1.ResourceOriginLabel: string(v1alpha1.GlobalResourceOrigin),
		}, "app-ns"),
		Entry("SimpleConverter without a mode leaves it alone", simple, map[string]string{
			v1alpha1.KubeNamespaceTag: systemNamespaceForTest,
		}, systemNamespaceForTest),
	)

	It("should return the enforced origin and zone on a CachingConverter cache hit", func() {
		converter := cachingOnZone()
		obj := policyIn("app-ns", importedFromGlobal)

		miss := labelsOf(converter, obj)
		hit := labelsOf(converter, obj)

		Expect(miss).To(HaveKeyWithValue(v1alpha1.ResourceOriginLabel, string(v1alpha1.ZoneResourceOrigin)))
		Expect(miss).To(HaveKeyWithValue(v1alpha1.ZoneTag, "zone-1"))
		Expect(hit).To(Equal(miss))
	})

	// A policy the webhook never validated can have no spec at all; GetSpec then hands
	// back a typed nil, and a read must not dereference it.
	DescribeTable("should not panic on a stored policy with no spec",
		func(newConverter func() k8s_common.Converter) {
			obj := policyIn("app-ns", nil)
			obj.Spec = nil
			out := meshtimeout_api.NewMeshTimeoutResource()

			Expect(newConverter().ToCoreResource(obj, out)).To(Succeed())
			Expect(out.GetMeta().GetLabels()).To(HaveKeyWithValue(v1alpha1.KubeNamespaceTag, "app-ns"))
		},
		Entry("SimpleConverter", simple),
		Entry("CachingConverter", caching),
	)
})
