package tls_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/xds/envoy/tls"
)

var _ = Describe("SNI", func() {
	It("should convert SNI to tags", func() {
		tags := map[string]string{
			"kuma.io/service": "backend",
			"version":         "v1",
			"env":             "prod",
			"region":          "eu",
			"app":             "backend-app",
		}
		expected := "backend{app=backend-app,env=prod,region=eu,version=v1}"
		actual := tls.SNIFromTags(tags)
		Expect(actual).To(Equal(expected))
	})

	It("should convert SNI to tags with only service name", func() {
		tags := map[string]string{
			"kuma.io/service": "backend",
		}
		expected := "backend"
		actual := tls.SNIFromTags(tags)
		Expect(actual).To(Equal(expected))
	})
})
