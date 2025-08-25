package clusters_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	core_meta "github.com/kumahq/kuma/pkg/core/metadata"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	"github.com/kumahq/kuma/pkg/xds/envoy/clusters"
)

var _ = Describe("PassThroughClusterConfigurer", func() {
	It("should generate proper Envoy config", func() {
		// given
		clusterName := "test:cluster"
		expected := `
        altStatName: test_cluster
        connectTimeout: 5s
        lbPolicy: CLUSTER_PROVIDED
        name: test:cluster
        type: ORIGINAL_DST`

		// when
		cluster, err := clusters.NewClusterBuilder(envoy.APIV3, clusterName).
			Configure(clusters.PassThroughCluster()).
			Configure(clusters.Timeout(DefaultTimeout(), core_meta.ProtocolTCP)).
			Build()

		// then
		Expect(err).ToNot(HaveOccurred())

		actual, err := util_proto.ToYAML(cluster)
		Expect(err).ToNot(HaveOccurred())
		Expect(actual).To(MatchYAML(expected))
	})
})
