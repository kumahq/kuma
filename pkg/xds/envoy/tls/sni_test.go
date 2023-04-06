package tls_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	envoy_tags "github.com/kumahq/kuma/pkg/xds/envoy/tags"
	"github.com/kumahq/kuma/pkg/xds/envoy/tls"
)

var _ = Describe("SNI", func() {
	It("should convert tags to SNI", func() {
		// given
		tags := map[string]string{
			"kuma.io/service": "backend",
			"version":         "v1",
			"env":             "prod",
			"region":          "eu",
			"app":             "backend-app",
		}
		expected := "backend{app=backend-app,env=prod,region=eu,version=v1}"

		// when
		actual := tls.SNIFromTags(tags)

		// then
		Expect(actual).To(Equal(expected))
	})

	It("should convert tags to SNI with only service name", func() {
		// given
		tags := map[string]string{
			"kuma.io/service": "backend",
		}
		expected := "backend"

		// when
		actual := tls.SNIFromTags(tags)

		// then
		Expect(actual).To(Equal(expected))
	})

	It("should convert SNI to tags", func() {
		// given
		sni := "backend{app=backend-app,env=prod,region=eu,version=v1}"
		expectedTags := envoy_tags.Tags{
			"kuma.io/service": "backend",
			"version":         "v1",
			"env":             "prod",
			"region":          "eu",
			"app":             "backend-app",
		}

		// when
		tags, err := tls.TagsFromSNI(sni)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(tags).To(Equal(expectedTags))
	})

	It("should convert SNI to tags with only service name", func() {
		// given
		sni := "backend"
		expectedTags := envoy_tags.Tags{
			"kuma.io/service": "backend",
		}

		// when
		tags, err := tls.TagsFromSNI(sni)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(tags).To(Equal(expectedTags))
	})

	DescribeTable("should fail when converting SNI to tags", func(sni string, errorMessage string) {
		// when
		_, err := tls.TagsFromSNI(sni)

		// then
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal(errorMessage))
	},
		Entry("broken tags", "backend{", "invalid format of tags, pairs should be separated by , and key should be separated from value by ="),
		Entry("to many separators", "backend{mesh=default{mesh", "cannot parse tags from sni: backend{mesh=default{mesh"),
	)
})
