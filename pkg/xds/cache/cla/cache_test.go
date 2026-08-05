package cla_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/v3/pkg/core/xds"
	core_metrics "github.com/kumahq/kuma/v3/pkg/metrics"
	plugins_xds "github.com/kumahq/kuma/v3/pkg/plugins/policies/core/xds"
	"github.com/kumahq/kuma/v3/pkg/test/matchers"
	"github.com/kumahq/kuma/v3/pkg/xds/cache/cla"
	envoy_common "github.com/kumahq/kuma/v3/pkg/xds/envoy"
	envoy_endpoints "github.com/kumahq/kuma/v3/pkg/xds/envoy/endpoints/v3"
	"github.com/kumahq/kuma/v3/pkg/xds/envoy/tags"
)

var _ = Describe("ClusterLoadAssignment CachedRetriever", func() {
	var claCache *cla.CachedRetriever
	var metrics core_metrics.Metrics

	expiration := 2 * time.Second

	BeforeEach(func() {
		var err error
		metrics, err = core_metrics.NewMetrics("Zone")
		Expect(err).ToNot(HaveOccurred())

		claCache, err = cla.NewCache(expiration, metrics)
		Expect(err).ToNot(HaveOccurred())
	})

	It("should cache ClusterLoadAssignment", func() {
		// given
		endpointMap := xds.EndpointMap{
			"backend": []xds.Endpoint{
				{
					Target: "192.168.0.1",
					Port:   uint32(1000),
				},
			},
		}
		cla1, err := claCache.GetCLA(context.Background(), "mesh-0", "", plugins_xds.NewClusterBuilder().WithService("backend").Build(), envoy_common.APIV3, endpointMap)
		Expect(err).ToNot(HaveOccurred())

		cla2, err := claCache.GetCLA(context.Background(), "mesh-0", "", plugins_xds.NewClusterBuilder().WithService("backend").Build(), envoy_common.APIV3, endpointMap)
		Expect(err).ToNot(HaveOccurred())

		Expect(cla1).To(BeIdenticalTo(cla2))
	})

	It("should pick endpoints for proper service", func() {
		// given
		endpointMap := xds.EndpointMap{
			"backend": []xds.Endpoint{
				{
					Target: "192.168.0.1",
					Port:   uint32(1000),
				},
			},
			"web": []xds.Endpoint{
				{
					Target: "192.168.0.2",
					Port:   uint32(1000),
				},
			},
		}

		// when
		clusterBackend := plugins_xds.NewClusterBuilder().WithService("backend").Build()
		claBackend, err := claCache.GetCLA(context.Background(), "mesh-0", "", clusterBackend, envoy_common.APIV3, endpointMap)

		// then
		Expect(err).ToNot(HaveOccurred())
		expectedCla := envoy_endpoints.CreateClusterLoadAssignment("backend", endpointMap["backend"])
		Expect(claBackend).To(matchers.MatchProto(expectedCla))

		// when
		clusterWeb := plugins_xds.NewClusterBuilder().WithService("web").Build()
		claWeb, err := claCache.GetCLA(context.Background(), "mesh-0", "", clusterWeb, envoy_common.APIV3, endpointMap)

		// then
		Expect(err).ToNot(HaveOccurred())
		expectedCla = envoy_endpoints.CreateClusterLoadAssignment("web", endpointMap["web"])
		Expect(claWeb).To(matchers.MatchProto(expectedCla))
	})

	It("should pick only endpoints with selected tags", func() {
		// given
		endpointMap := xds.EndpointMap{
			"backend": []xds.Endpoint{
				{
					Target: "192.168.0.1",
					Port:   uint32(1000),
					Tags: map[string]string{
						"version": "v1",
					},
				},
				{
					Target: "192.168.0.2",
					Port:   uint32(1000),
					Tags: map[string]string{
						"version": "v2",
					},
				},
			},
		}

		// when
		clusterV1 := plugins_xds.NewClusterBuilder().WithService("backend").WithTags(tags.Tags{}.WithTags("version", "v1")).Build()
		claV1, err := claCache.GetCLA(context.Background(), "mesh-0", "", clusterV1, envoy_common.APIV3, endpointMap)

		// then
		Expect(err).ToNot(HaveOccurred())
		expectedCla := envoy_endpoints.CreateClusterLoadAssignment("backend", []xds.Endpoint{endpointMap["backend"][0]})
		Expect(claV1).To(matchers.MatchProto(expectedCla))

		// when
		clusterV2 := plugins_xds.NewClusterBuilder().WithService("backend").WithTags(tags.Tags{}.WithTags("version", "v2")).Build()
		claV2, err := claCache.GetCLA(context.Background(), "mesh-0", "", clusterV2, envoy_common.APIV3, endpointMap)

		// then
		Expect(err).ToNot(HaveOccurred())
		expectedCla = envoy_endpoints.CreateClusterLoadAssignment("backend", []xds.Endpoint{endpointMap["backend"][1]})
		Expect(claV2).To(matchers.MatchProto(expectedCla))
	})
})
