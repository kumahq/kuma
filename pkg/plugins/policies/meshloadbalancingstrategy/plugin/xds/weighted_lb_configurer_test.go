package xds_test

import (
	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/plugins/policies/meshloadbalancingstrategy/plugin/xds"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	"github.com/kumahq/kuma/pkg/xds/envoy/clusters"
)

var _ = Describe("WeightedLBConfigurer", func() {
	type testCase struct {
		expected string
	}

	DescribeTable("should generate proper envoy config",
		func(given testCase) {
			// given
			configurer := &xds.LocalityWeightedLbConfigurer{}
			cluster := clusters.NewClusterBuilder(envoy.APIV3, "test").
				MustBuild()

			// when
			err := configurer.Configure(cluster.(*envoy_cluster.Cluster))
			Expect(err).ToNot(HaveOccurred())

			// then
			actual, err := util_proto.ToYAML(cluster)
			Expect(err).ToNot(HaveOccurred())
			Expect(actual).To(MatchYAML(given.expected))
		},
		Entry("default", testCase{
			expected: `
name: test
commonLbConfig:
  localityWeightedLbConfig: {}
`,
		}),
	)
})
