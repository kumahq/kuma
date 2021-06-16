package proto_test

import (
	"github.com/golang/protobuf/proto"

	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	envoy_metadata "github.com/kumahq/kuma/pkg/xds/envoy/metadata/v3"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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
			Expect(proto.Equal(any1, any2)).To(BeTrue())
		}
	})
})
