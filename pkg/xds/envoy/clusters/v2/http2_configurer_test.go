package clusters_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	"github.com/kumahq/kuma/pkg/xds/envoy/clusters"
)

var _ = Describe("Http2Configurer", func() {

	It("should generate proper Envoy config", func() {
		// given
		expected := `http2ProtocolOptions: {}`

		// when
		cluster, err := clusters.NewClusterBuilder(envoy.APIV2).
			Configure(clusters.Http2()).
			Build()

		// then
		Expect(err).ToNot(HaveOccurred())

		actual, err := util_proto.ToYAML(cluster)
		Expect(err).ToNot(HaveOccurred())
		Expect(actual).To(MatchYAML(expected))
	})
})
