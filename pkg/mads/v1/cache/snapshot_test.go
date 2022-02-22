package cache_test

import (
	envoy_types "github.com/envoyproxy/go-control-plane/pkg/cache/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	observability_proto "github.com/kumahq/kuma/api/observability/v1"
	. "github.com/kumahq/kuma/pkg/mads/v1/cache"
)

var _ = Describe("Snapshot", func() {

	expectedType := "type.googleapis.com/kuma.observability.v1.MonitoringAssignment"

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
			snapshot = NewSnapshot("v2", map[string]envoy_types.Resource{
				"backend": &observability_proto.MonitoringAssignment{
					Mesh:    "default",
					Service: "backend",
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
			assignments := map[string]envoy_types.Resource{
				"backend": &observability_proto.MonitoringAssignment{
					Mesh:    "default",
					Service: "backend",
				},
			}
			// when
			snapshot := NewSnapshot("v1", assignments)
			// then
			Expect(snapshot.GetResources(expectedType)).To(Equal(assignments))
		})

		It("should return `nil` for unsupported resource types", func() {
			// given
			assignments := map[string]envoy_types.Resource{
				"backend": &observability_proto.MonitoringAssignment{
					Mesh:    "default",
					Service: "backend",
				},
			}
			// when
			snapshot := NewSnapshot("v1", assignments)
			// then
			Expect(snapshot.GetResources("unsupported type")).To(BeNil())
		})
	})

	Describe("GetVersion()", func() {
		It("should handle `nil`", func() {
			// when
			var snapshot *Snapshot
			// then
			Expect(snapshot.GetVersion(expectedType)).To(Equal(""))
		})

		It("should return proper version for a supported resource type", func() {
			// given
			assignments := map[string]envoy_types.Resource{
				"backend": &observability_proto.MonitoringAssignment{
					Mesh:    "default",
					Service: "backend",
				},
			}
			// when
			snapshot := NewSnapshot("v1", assignments)
			// then
			Expect(snapshot.GetVersion(expectedType)).To(Equal("v1"))
		})

		It("should return an empty string for unsupported resource type", func() {
			// given
			assignments := map[string]envoy_types.Resource{
				"backend": &observability_proto.MonitoringAssignment{
					Mesh:    "default",
					Service: "backend",
				},
			}
			// when
			snapshot := NewSnapshot("v1", assignments)
			// then
			Expect(snapshot.GetVersion("unsupported type")).To(Equal(""))
		})
	})

	Describe("WithVersion()", func() {
		It("should handle `nil`", func() {
			// given
			var snapshot *Snapshot
			// when
			actual := snapshot.WithVersion(expectedType, "v1")
			// then
			Expect(actual).To(BeNil())
		})

		It("should return a new snapshot if version has changed", func() {
			// given
			assignments := map[string]envoy_types.Resource{
				"backend": &observability_proto.MonitoringAssignment{
					Mesh:    "default",
					Service: "backend",
				},
			}
			snapshot := NewSnapshot("v1", assignments)
			// when
			actual := snapshot.WithVersion(expectedType, "v2")
			// then
			Expect(actual.GetVersion(expectedType)).To(Equal("v2"))
			// and
			Expect(actual).To(Equal(NewSnapshot("v2", assignments)))
		})

		It("should return the same snapshot if version has not changed", func() {
			// given
			assignments := map[string]envoy_types.Resource{
				"backend": &observability_proto.MonitoringAssignment{
					Mesh:    "default",
					Service: "backend",
				},
			}
			snapshot := NewSnapshot("v1", assignments)
			// when
			actual := snapshot.WithVersion(expectedType, "v1")
			// then
			Expect(actual.GetVersion(expectedType)).To(Equal("v1"))
			// and
			Expect(actual).To(BeIdenticalTo(snapshot))
		})

		It("should return the same snapshot if resource type is not supported", func() {
			// given
			assignments := map[string]envoy_types.Resource{
				"backend": &observability_proto.MonitoringAssignment{
					Mesh:    "default",
					Service: "backend",
				},
			}
			snapshot := NewSnapshot("v1", assignments)
			// when
			actual := snapshot.WithVersion("unsupported type", "v2")
			// then
			Expect(actual.GetVersion(expectedType)).To(Equal("v1"))
			// and
			Expect(actual).To(BeIdenticalTo(snapshot))
		})
	})
})
