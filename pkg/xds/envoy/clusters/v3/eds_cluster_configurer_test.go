package clusters_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	"github.com/kumahq/kuma/pkg/xds/envoy/clusters"
)

var _ = Describe("EdsClusterConfigurer", func() {

	It("should generate proper Envoy config", func() {
		// given
		clusterName := "test:cluster"
		expected := `
        altStatName: test_cluster
        connectTimeout: 5s
        edsClusterConfig:
          edsConfig:
            ads: {}
            resourceApiVersion: V3
        name: test:cluster
        type: EDS`

		// when
		cluster, err := clusters.NewClusterBuilder(envoy.APIV3).
			Configure(clusters.EdsCluster(clusterName)).
			Build()

		// then
		Expect(err).ToNot(HaveOccurred())

		actual, err := util_proto.ToYAML(cluster)
		Expect(err).ToNot(HaveOccurred())
		Expect(actual).To(MatchYAML(expected))
	})
})
