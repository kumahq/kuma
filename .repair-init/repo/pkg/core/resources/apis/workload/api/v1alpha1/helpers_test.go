package v1alpha1_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/workload/api/v1alpha1"
	test_model "github.com/kumahq/kuma/v3/pkg/test/resources/model"
)

var _ = Describe("WorkloadResource Hash", func() {
	workload := func(mesh, name, version string, status api.WorkloadStatus) *api.WorkloadResource {
		return &api.WorkloadResource{
			Meta: &test_model.ResourceMeta{
				Mesh:    mesh,
				Name:    name,
				Version: version,
			},
			Spec:   &api.Workload{},
			Status: &status,
		}
	}

	It("ignores resource version and status", func() {
		first := workload("mesh-1", "backend", "1", api.WorkloadStatus{
			DataplaneProxies: api.DataplaneProxies{
				Total:     1,
				Connected: 1,
				Healthy:   1,
			},
		})
		second := workload("mesh-1", "backend", "2", api.WorkloadStatus{
			DataplaneProxies: api.DataplaneProxies{
				Total:     3,
				Connected: 2,
				Healthy:   1,
			},
		})

		Expect(second.Hash()).To(Equal(first.Hash()))
	})

	It("changes when the name changes", func() {
		first := workload("mesh-1", "backend", "1", api.WorkloadStatus{})
		second := workload("mesh-1", "frontend", "1", api.WorkloadStatus{})

		Expect(second.Hash()).ToNot(Equal(first.Hash()))
	})

	It("changes when the mesh changes", func() {
		first := workload("mesh-1", "backend", "1", api.WorkloadStatus{})
		second := workload("mesh-2", "backend", "1", api.WorkloadStatus{})

		Expect(second.Hash()).ToNot(Equal(first.Hash()))
	})

	It("hashes empty identity consistently", func() {
		first := workload("", "", "1", api.WorkloadStatus{})
		second := workload("", "", "2", api.WorkloadStatus{
			DataplaneProxies: api.DataplaneProxies{Total: 1},
		})

		Expect(second.Hash()).To(Equal(first.Hash()))
	})
})
