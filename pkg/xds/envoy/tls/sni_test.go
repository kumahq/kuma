package tls_test

import (
	"github.com/Kong/kuma/pkg/xds/envoy/tls"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("SNI", func() {
	It("should parse tags from SNI", func() {
		actual := tls.TagsFromSNI("backend{app=backend-app,env=prod,region=eu,version=v1}")
		expected := map[string]string{
			"service": "backend",
			"version": "v1",
			"env":     "prod",
			"region":  "eu",
			"app":     "backend-app",
		}
		Expect(actual).To(Equal(expected))
	})

	It("should convert SNI to tags", func() {
		tags := map[string]string{
			"service": "backend",
			"version": "v1",
			"env":     "prod",
			"region":  "eu",
			"app":     "backend-app",
		}
		expected := "backend{app=backend-app,env=prod,region=eu,version=v1}"
		actual := tls.SNIFromTags(tags)
		Expect(actual).To(Equal(expected))
	})

	It("should convert SNI to tags with only service name", func() {
		tags := map[string]string{
			"service": "backend",
		}
		expected := "backend"
		actual := tls.SNIFromTags(tags)
		Expect(actual).To(Equal(expected))
	})
})
