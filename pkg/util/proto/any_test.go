package proto_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/test/matchers"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	envoy_metadata "github.com/kumahq/kuma/pkg/xds/envoy/metadata/v3"
)

var _ = Describe("MarshalAnyDeterministic", func() {
	It("should marshal deterministically", func() {
		tags := map[string]string{
			"service": "backend",
			"version": "v1",
			"cloud":   "aws",
		}
		metadata := envoy_metadata.EndpointMetadata(tags)
		for i := 0; i < 100; i++ {
			any1, _ := util_proto.MarshalAnyDeterministic(metadata)
			any2, _ := util_proto.MarshalAnyDeterministic(metadata)
			Expect(any1).To(matchers.MatchProto(any2))
		}
	})
})
