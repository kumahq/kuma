package clusters_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	"github.com/kumahq/kuma/pkg/xds/envoy/clusters"
)

var _ = Describe("LbSubset", func() {

	type testCase struct {
		clusterName string
		tags        envoy.TagKeysSlice
		expected    string
	}

	DescribeTable("should generate proper Envoy config",
		func(given testCase) {
			// when
			cluster, err := clusters.NewClusterBuilder(envoy.APIV3).
				Configure(clusters.EdsCluster(given.clusterName)).
				Configure(clusters.LbSubset(given.tags)).
				Configure(clusters.Timeout(DefaultTimeout(), core_mesh.ProtocolTCP)).
				Build()

			// then
			Expect(err).ToNot(HaveOccurred())

			actual, err := util_proto.ToYAML(cluster)
			Expect(err).ToNot(HaveOccurred())
			Expect(actual).To(MatchYAML(given.expected))
		},
		Entry("LbSubset is empty if there are no tags", testCase{
			clusterName: "backend",
			tags:        []envoy.TagKeys{},
			expected: `
            connectTimeout: 5s
            edsClusterConfig:
              edsConfig:
                ads: {}
                resourceApiVersion: V3
            name: backend
            type: EDS`,
		}),
		Entry("LbSubset is set when more than service tag is set", testCase{
			clusterName: "backend",
			tags: []envoy.TagKeys{
				{"version"},
				{"cluster", "version"},
			},
			expected: `
            connectTimeout: 5s
            edsClusterConfig:
              edsConfig:
                ads: {}
                resourceApiVersion: V3
            lbSubsetConfig:
              fallbackPolicy: ANY_ENDPOINT
              subsetSelectors:
              - fallbackPolicy: NO_FALLBACK
                keys:
                - version
              - fallbackPolicy: NO_FALLBACK
                keys:
                - cluster
                - version
            name: backend
            type: EDS`,
		}),
	)
})
