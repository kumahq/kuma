package k8s_test

import (
	"reflect"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	workload_api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/workload/api/v1alpha1"
	workload_k8s "github.com/kumahq/kuma/v2/pkg/core/resources/apis/workload/k8s/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/plugins/resources/k8s"
	"github.com/kumahq/kuma/v2/pkg/plugins/runtime/k8s/metadata"
)

var _ = Describe("CachingConverter", func() {
	It("should preserve status on cache hit", func() {
		// setup
		converter := k8s.NewCachingConverter(5 * time.Minute)

		// given - K8s Workload object with status
		k8sWorkload := &workload_k8s.Workload{
			TypeMeta: metav1.TypeMeta{
				APIVersion: workload_k8s.GroupVersion.String(),
				Kind:       "Workload",
			},
			ObjectMeta: metav1.ObjectMeta{
				Namespace:       "demo",
				Name:            "backend",
				ResourceVersion: "1",
			},
			Spec: &workload_api.Workload{},
			Status: &workload_api.WorkloadStatus{
				DataplaneProxies: workload_api.DataplaneProxies{
					Connected: 5,
					Healthy:   3,
					Total:     5,
				},
			},
		}

		// when - first conversion (cache miss)
		out1 := workload_api.NewWorkloadResource()
		err := converter.ToCoreResource(k8sWorkload, out1)
		Expect(err).ToNot(HaveOccurred())

		// then - status should be populated
		Expect(out1.Status).ToNot(BeNil())
		Expect(out1.Status.DataplaneProxies.Connected).To(Equal(int32(5)))
		Expect(out1.Status.DataplaneProxies.Healthy).To(Equal(int32(3)))
		Expect(out1.Status.DataplaneProxies.Total).To(Equal(int32(5)))

		// when - second conversion (cache hit)
		out2 := workload_api.NewWorkloadResource()
		err = converter.ToCoreResource(k8sWorkload, out2)
		Expect(err).ToNot(HaveOccurred())

		// then - status should STILL be populated (verifies fix)
		Expect(out2.Status).ToNot(BeNil())
		Expect(out2.Status.DataplaneProxies.Connected).To(Equal(int32(5)))
		Expect(out2.Status.DataplaneProxies.Healthy).To(Equal(int32(3)))
		Expect(out2.Status.DataplaneProxies.Total).To(Equal(int32(5)))
	})

	It("should preserve status even when cache hit with different status values", func() {
		// setup
		converter := k8s.NewCachingConverter(5 * time.Minute)

		// given - K8s Workload object with initial status
		k8sWorkload := &workload_k8s.Workload{
			TypeMeta: metav1.TypeMeta{
				APIVersion: workload_k8s.GroupVersion.String(),
				Kind:       "Workload",
			},
			ObjectMeta: metav1.ObjectMeta{
				Namespace:       "demo",
				Name:            "backend",
				ResourceVersion: "1",
			},
			Spec: &workload_api.Workload{},
			Status: &workload_api.WorkloadStatus{
				DataplaneProxies: workload_api.DataplaneProxies{
					Connected: 5,
					Healthy:   3,
					Total:     5,
				},
			},
		}

		// when - first conversion (cache miss)
		out1 := workload_api.NewWorkloadResource()
		err := converter.ToCoreResource(k8sWorkload, out1)
		Expect(err).ToNot(HaveOccurred())

		// then - status should be populated
		Expect(out1.Status.DataplaneProxies.Connected).To(Equal(int32(5)))

		// given - update status (simulating StatusUpdater)
		k8sWorkload.Status.DataplaneProxies.Connected = 7
		k8sWorkload.Status.DataplaneProxies.Total = 7

		// when - second conversion (cache hit, same ResourceVersion)
		out2 := workload_api.NewWorkloadResource()
		err = converter.ToCoreResource(k8sWorkload, out2)
		Expect(err).ToNot(HaveOccurred())

		// then - status should reflect updated values (not cached)
		Expect(out2.Status.DataplaneProxies.Connected).To(Equal(int32(7)))
		Expect(out2.Status.DataplaneProxies.Total).To(Equal(int32(7)))
	})

	It("should memoize labels on the meta adapter across repeated GetLabels calls", func() {
		adapter := &k8s.KubernetesMetaAdapter{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "backend",
				Namespace: "demo",
				Labels:    map[string]string{"app": "backend"},
				Annotations: map[string]string{
					v1alpha1.DisplayName:        "Backend Display",
					metadata.KumaServiceAccount: "sa-backend",
				},
			},
			Mesh: "default",
		}

		first := adapter.GetLabels()
		second := adapter.GetLabels()

		// same map instance returned on repeat calls -> no per-call maps.Clone
		Expect(reflect.ValueOf(first).Pointer()).To(Equal(reflect.ValueOf(second).Pointer()))
		Expect(first).To(HaveKeyWithValue("app", "backend"))
		Expect(first).To(HaveKeyWithValue(v1alpha1.DisplayName, "Backend Display"))
		Expect(first).To(HaveKeyWithValue(metadata.KumaServiceAccount, "sa-backend"))
	})

	It("should share the same labels map across cache hits for the same resourceVersion", func() {
		converter := k8s.NewCachingConverter(5 * time.Minute)
		k8sWorkload := &workload_k8s.Workload{
			TypeMeta: metav1.TypeMeta{
				APIVersion: workload_k8s.GroupVersion.String(),
				Kind:       "Workload",
			},
			ObjectMeta: metav1.ObjectMeta{
				Namespace:       "demo",
				Name:            "backend",
				ResourceVersion: "1",
				Labels:          map[string]string{"app": "backend"},
			},
			Spec: &workload_api.Workload{},
		}

		out1 := workload_api.NewWorkloadResource()
		Expect(converter.ToCoreResource(k8sWorkload, out1)).To(Succeed())

		out2 := workload_api.NewWorkloadResource()
		Expect(converter.ToCoreResource(k8sWorkload, out2)).To(Succeed())

		labels1 := out1.GetMeta().GetLabels()
		labels2 := out2.GetMeta().GetLabels()

		// cache hit path reuses the precomputed labels map -> same pointer
		Expect(reflect.ValueOf(labels1).Pointer()).To(Equal(reflect.ValueOf(labels2).Pointer()))
		Expect(labels1).To(HaveKeyWithValue("app", "backend"))
		// display name defaults to the object name when no annotation is present
		Expect(labels1).To(HaveKeyWithValue(v1alpha1.DisplayName, "backend"))
	})

	It("should compute fresh labels when resourceVersion changes", func() {
		converter := k8s.NewCachingConverter(5 * time.Minute)
		k8sWorkload := &workload_k8s.Workload{
			TypeMeta: metav1.TypeMeta{
				APIVersion: workload_k8s.GroupVersion.String(),
				Kind:       "Workload",
			},
			ObjectMeta: metav1.ObjectMeta{
				Namespace:       "demo",
				Name:            "backend",
				ResourceVersion: "1",
				Labels:          map[string]string{"app": "backend"},
			},
			Spec: &workload_api.Workload{},
		}

		out1 := workload_api.NewWorkloadResource()
		Expect(converter.ToCoreResource(k8sWorkload, out1)).To(Succeed())
		labels1 := out1.GetMeta().GetLabels()

		// simulate an update: new ResourceVersion + mutated labels
		k8sWorkload.ResourceVersion = "2"
		k8sWorkload.Labels = map[string]string{"app": "backend", "version": "v2"}

		out2 := workload_api.NewWorkloadResource()
		Expect(converter.ToCoreResource(k8sWorkload, out2)).To(Succeed())
		labels2 := out2.GetMeta().GetLabels()

		// different resourceVersion -> different cache key -> different labels map
		Expect(reflect.ValueOf(labels1).Pointer()).ToNot(Equal(reflect.ValueOf(labels2).Pointer()))
		Expect(labels1).ToNot(HaveKey("version"))
		Expect(labels2).To(HaveKeyWithValue("version", "v2"))
	})
})
