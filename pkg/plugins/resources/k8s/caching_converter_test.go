package k8s_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	workload_api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/workload/api/v1alpha1"
	workload_k8s "github.com/kumahq/kuma/v2/pkg/core/resources/apis/workload/k8s/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/plugins/resources/k8s"
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
})
