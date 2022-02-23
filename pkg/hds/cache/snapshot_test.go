package cache_test

import (
	"time"

	envoy_service_health_v3 "github.com/envoyproxy/go-control-plane/envoy/service/health/v3"
	envoy_types "github.com/envoyproxy/go-control-plane/pkg/cache/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/hds/cache"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

var _ = Describe("Snapshot", func() {

	expectedType := "envoy.service.health.v3.HealthCheckSpecifier"

	Describe("GetSupportedTypes()", func() {
		It("should always return ['envoy.service.health.v3.HealthCheckSpecifier']", func() {
			// when
			var snapshot *cache.Snapshot
			// then
			Expect(snapshot.GetSupportedTypes()).To(Equal([]string{expectedType}))

			// when
			snapshot = &cache.Snapshot{}
			// then
			Expect(snapshot.GetSupportedTypes()).To(Equal([]string{expectedType}))
		})
	})

	Describe("Consistent()", func() {
		It("should handle `nil`", func() {
			// when
			var snapshot *cache.Snapshot
			// then
			Expect(snapshot.Consistent()).To(MatchError("nil Snapshot"))
		})

		It("non-`nil` Snapshot should be always consistent", func() {
			// when
			snapshot := cache.NewSnapshot("v1", nil)
			// then
			Expect(snapshot.Consistent()).To(BeNil())

			// when
			snapshot = cache.NewSnapshot("v2", &envoy_service_health_v3.HealthCheckSpecifier{})
			// then
			Expect(snapshot.Consistent()).To(BeNil())
		})
	})

	Describe("GetResources()", func() {
		It("should handle `nil`", func() {
			// when
			var snapshot *cache.Snapshot
			// then
			Expect(snapshot.GetResources(expectedType)).To(BeNil())
		})

		It("should return HealthCheckSpecifier", func() {
			// given
			hcs := &envoy_service_health_v3.HealthCheckSpecifier{
				Interval: util_proto.Duration(12 * time.Second),
				ClusterHealthChecks: []*envoy_service_health_v3.ClusterHealthCheck{
					{ClusterName: "localhost:80"},
					{ClusterName: "localhost:9080"},
				},
			}
			// when
			snapshot := cache.NewSnapshot("v1", hcs)
			// then
			Expect(snapshot.GetResources(expectedType)).To(Equal(map[string]envoy_types.Resource{
				"hcs": hcs,
			}))
		})

		It("should return `nil` for unsupported resource types", func() {
			// given
			hcs := &envoy_service_health_v3.HealthCheckSpecifier{
				Interval: util_proto.Duration(12 * time.Second),
				ClusterHealthChecks: []*envoy_service_health_v3.ClusterHealthCheck{
					{ClusterName: "localhost:80"},
					{ClusterName: "localhost:9080"},
				},
			}
			// when
			snapshot := cache.NewSnapshot("v1", hcs)
			// then
			Expect(snapshot.GetResources("unsupported type")).To(BeNil())
		})
	})

	Describe("GetVersion()", func() {
		It("should handle `nil`", func() {
			// when
			var snapshot *cache.Snapshot
			// then
			Expect(snapshot.GetVersion(expectedType)).To(Equal(""))
		})

		It("should return proper version for a supported resource type", func() {
			// given
			hcs := &envoy_service_health_v3.HealthCheckSpecifier{
				Interval: util_proto.Duration(12 * time.Second),
				ClusterHealthChecks: []*envoy_service_health_v3.ClusterHealthCheck{
					{ClusterName: "localhost:80"},
					{ClusterName: "localhost:9080"},
				},
			}
			// when
			snapshot := cache.NewSnapshot("v1", hcs)
			// then
			Expect(snapshot.GetVersion(expectedType)).To(Equal("v1"))
		})

		It("should return an empty string for unsupported resource type", func() {
			// given
			hcs := &envoy_service_health_v3.HealthCheckSpecifier{
				Interval: util_proto.Duration(12 * time.Second),
				ClusterHealthChecks: []*envoy_service_health_v3.ClusterHealthCheck{
					{ClusterName: "localhost:80"},
					{ClusterName: "localhost:9080"},
				},
			}
			// when
			snapshot := cache.NewSnapshot("v1", hcs)
			// then
			Expect(snapshot.GetVersion("unsupported type")).To(Equal(""))
		})
	})

	Describe("WithVersion()", func() {
		It("should handle `nil`", func() {
			// given
			var snapshot *cache.Snapshot
			// when
			actual := snapshot.WithVersion(expectedType, "v1")
			// then
			Expect(actual).To(BeNil())
		})

		It("should return a new Snapshot if version has changed", func() {
			// given
			hcs := &envoy_service_health_v3.HealthCheckSpecifier{
				Interval: util_proto.Duration(12 * time.Second),
				ClusterHealthChecks: []*envoy_service_health_v3.ClusterHealthCheck{
					{ClusterName: "localhost:80"},
					{ClusterName: "localhost:9080"},
				},
			}
			snapshot := cache.NewSnapshot("v1", hcs)
			// when
			actual := snapshot.WithVersion(expectedType, "v2")
			// then
			Expect(actual.GetVersion(expectedType)).To(Equal("v2"))
			// and
			Expect(actual).To(Equal(cache.NewSnapshot("v2", hcs)))
		})

		It("should return the same Snapshot if version has not changed", func() {
			// given
			hcs := &envoy_service_health_v3.HealthCheckSpecifier{
				Interval: util_proto.Duration(12 * time.Second),
				ClusterHealthChecks: []*envoy_service_health_v3.ClusterHealthCheck{
					{ClusterName: "localhost:80"},
					{ClusterName: "localhost:9080"},
				},
			}
			snapshot := cache.NewSnapshot("v1", hcs)
			// when
			actual := snapshot.WithVersion(expectedType, "v1")
			// then
			Expect(actual.GetVersion(expectedType)).To(Equal("v1"))
			// and
			Expect(actual).To(BeIdenticalTo(snapshot))
		})

		It("should return the same Snapshot if resource type is not supported", func() {
			// given
			hcs := &envoy_service_health_v3.HealthCheckSpecifier{
				Interval: util_proto.Duration(12 * time.Second),
				ClusterHealthChecks: []*envoy_service_health_v3.ClusterHealthCheck{
					{ClusterName: "localhost:80"},
					{ClusterName: "localhost:9080"},
				},
			}
			snapshot := cache.NewSnapshot("v1", hcs)
			// when
			actual := snapshot.WithVersion("unsupported type", "v2")
			// then
			Expect(actual.GetVersion(expectedType)).To(Equal("v1"))
			// and
			Expect(actual).To(BeIdenticalTo(snapshot))
		})
	})
})
