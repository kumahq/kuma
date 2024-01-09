package v3

import (
	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("DirectResponseConfigurer", func() {
	// It("", func() {
	// 	// when
	// 	filterChain, err := NewFilterChainBuilder(envoy.APIV3, envoy.AnonymousResource).
	// 		Configure(DirectResponse("", []DirectResponseEndpoints{{
	// 			Path:       "/",
	// 			StatusCode: 200,
	// 			Response:   "test",
	// 		}})).
	// 		Build()
	//
	// 	// then
	// 	Expect(err).ToNot(HaveOccurred())
	//
	// 	// then
	// 	actual, err := util_proto.ToYAML(filterChain)
	// 	Expect(err).ToNot(HaveOccurred())
	// 	Expect(actual).To(MatchYAML(``))
	// })
})
