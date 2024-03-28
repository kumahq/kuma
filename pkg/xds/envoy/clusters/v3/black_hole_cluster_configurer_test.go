package clusters_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	"github.com/kumahq/kuma/pkg/xds/envoy/clusters"
)

var _ = Describe("BlackHoleClusterConfigurer", func() {
	It("should generate proper Envoy config", func() {
		// given
		clusterName := "test:blackhole"
		expected := `
        connectTimeout: 5s
        name: test:blackhole
        type: STATIC`

		// when
		cluster, err := clusters.NewClusterBuilder(envoy.APIV3, clusterName).
			Configure(clusters.BlackHoleCluster()).
			Build()

		// then
		Expect(err).ToNot(HaveOccurred())

		actual, err := util_proto.ToYAML(cluster)
		Expect(err).ToNot(HaveOccurred())
		Expect(actual).To(MatchYAML(expected))
	})
})
