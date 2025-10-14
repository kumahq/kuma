package api_server

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/pkg/core/kri"
)

var _ = Describe("KRI endpoint", func() {
	It("should properly generate CoreName", func() {
		endpoint := kriEndpoint{
			cpMode:          core.Zone,
			cpZone:          "kuma-1",
			environment:     core.KubernetesEnvironment,
			systemNamespace: "kuma-system",
		}

		coreName := endpoint.getCoreName(kri.MustFromString("kri_mt_producer-policy-flow_kuma-2_producer-policy-flow-ns_to-test-server_"))
		Expect(coreName).To(Equal("to-test-server-4cx47v8b5wcd4764.kuma-system"))
	})
})
