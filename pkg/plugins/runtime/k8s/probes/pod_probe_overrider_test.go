package probes_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/probes"
)

var _ = Describe("ApplicationProbeProxyPort", func() {
	DescribeTable("GetApplicationProbeProxyPort",
		func(annotations map[string]string, defaultPort, expected int, expectedErr string) {
			port, err := probes.GetApplicationProbeProxyPort(annotations, uint32(defaultPort))

			if expectedErr != "" {
				Expect(err).To(MatchError(expectedErr))
			} else {
				Expect(err).ToNot(HaveOccurred())
				Expect(port).To(Equal(uint32(expected)))
			}
		},
		Entry("gateway mode with proxy port set", map[string]string{
			"kuma.io/application-probe-proxy-port": "9000",
			"kuma.io/gateway":                      "enabled",
		}, 10001, 0, "application probe proxies probes can't be enabled in gateway mode"),

		Entry("gateway mode without proxy", map[string]string{
			"kuma.io/gateway": "enabled",
		}, 10001, 0, ""),

		Entry("virtual probes disabled and no proxy port", map[string]string{
			"kuma.io/virtual-probes": "false",
		}, 10001, 0, ""),

		Entry("proxy port set", map[string]string{
			"kuma.io/application-probe-proxy-port": "9001",
		}, 10001, 9001, ""),

		Entry("default port fallback", map[string]string{}, 10001, 10001, ""),
	)
})
