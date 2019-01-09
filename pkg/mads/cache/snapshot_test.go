package cache_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	envoy_cache "github.com/envoyproxy/go-control-plane/pkg/cache"

	. "github.com/Kong/kuma/pkg/mads/cache"

	observability_proto "github.com/Kong/kuma/api/observability/v1alpha1"
)

var _ = Describe("Snapshot", func() {

	expectedType := "type.googleapis.com/kuma.observability.v1alpha1.MonitoringAssignment"

	Describe("GetSupportedTypes()", func() {
		It("should always return ['type.googleapis.com/kuma.observability.v1alpha1.MonitoringAssignment']", func() {
			// when
			var snapshot *Snapshot
			// then
			Expect(snapshot.GetSupportedTypes()).To(Equal([]string{expectedType}))

			// when
			snapshot = &Snapshot{}
			// then
			Expect(snapshot.GetSupportedTypes()).To(Equal([]string{expectedType}))
		})
	})

	Describe("Consistent()", func() {
		It("should handle `nil`", func() {
			// when
			var snapshot *Snapshot
			// then
			Expect(snapshot.Consistent()).To(MatchError("nil snapshot"))
		})

		It("non-`nil` snapshot should be always consistet", func() {
			// when
			snapshot := NewSnapshot("v1", nil)
			// then
			Expect(snapshot.Consistent()).To(BeNil())

			// when
			snapshot = NewSnapshot("v2", map[string]envoy_cache.Resource{
				"backend": &observability_proto.MonitoringAssignment{
					Name: "/meshes/default/dataplanes/backend",
				},
			})
			// then
			Expect(snapshot.Consistent()).To(BeNil())
		})
	})

	Describe("GetResources()", func() {
		It("should handle `nil`", func() {
			// when
			var snapshot *Snapshot
			// then
			Expect(snapshot.GetResources(expectedType)).To(BeNil())
		})

		It("should return MonitoringAssignments", func() {
			// given
			assignments := map[string]envoy_cache.Resource{
				"backend": &observability_proto.MonitoringAssignment{
					Name: "/meshes/default/dataplanes/backend",
				},
			}
			// when
			snapshot := NewSnapshot("v1", assignments)
			// then
			Expect(snapshot.GetResources(expectedType)).To(Equal(assignments))
		})
	})

	Describe("GetVersion()", func() {
		It("should handle `nil`", func() {
			// when
			var snapshot *Snapshot
			// then
			Expect(snapshot.GetVersion(expectedType)).To(Equal(""))
		})

		It("should return proper version", func() {
			// given
			assignments := map[string]envoy_cache.Resource{
				"backend": &observability_proto.MonitoringAssignment{
					Name: "/meshes/default/dataplanes/backend",
				},
			}
			// when
			snapshot := NewSnapshot("v1", assignments)
			// then
			Expect(snapshot.GetVersion(expectedType)).To(Equal("v1"))
		})
	})

	Describe("SetVersion()", func() {
		It("should handle `nil`", func() {
			// given
			var snapshot *Snapshot
			// when
			snapshot.SetVersion(expectedType, "v1")
			// then
			Expect(snapshot.GetVersion(expectedType)).To(Equal(""))
		})

		It("should set version properly", func() {
			// given
			assignments := map[string]envoy_cache.Resource{
				"backend": &observability_proto.MonitoringAssignment{
					Name: "/meshes/default/dataplanes/backend",
				},
			}
			snapshot := NewSnapshot("v1", assignments)
			// when
			snapshot.SetVersion(expectedType, "v2")
			// then
			Expect(snapshot.GetVersion(expectedType)).To(Equal("v2"))
		})
	})
})
