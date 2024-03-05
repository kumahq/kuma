package v3_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	. "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
)

var _ = Describe("PipeListenerConfigurer", func() {
	It("should correctly set up pipe address for listener", func() {
		// when
		listener, err := NewListenerBuilder(envoy.APIV3, "test:listener").
			Configure(PipeListener("test.sock")).
			Build()

		// then
		Expect(err).ToNot(HaveOccurred())
		actual, err := util_proto.ToYAML(listener)
		Expect(err).ToNot(HaveOccurred())
		// and
		Expect(actual).To(MatchYAML(`
address:
  pipe:
    path: test.sock
name: test:listener
`))
	})
})
