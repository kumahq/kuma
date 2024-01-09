package v3

import (
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	. "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("PipeListenerConfigurer", func() {
	It("", func() {
		// given
		listener, err := NewListenerBuilder(envoy.APIV3, "").
			Configure(PipeListener("file.sock")).Build()

		// when
		// listener, err := listenerBuilder.Build()

		// then
		Expect(err).ToNot(HaveOccurred())
		actual, err := util_proto.ToYAML(listener)
		Expect(err).ToNot(HaveOccurred())
		// and
		Expect(actual).To(MatchYAML(`
`))

	})
})
