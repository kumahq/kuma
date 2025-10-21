package api_server

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/pkg/core/kri"
)

var _ = Describe("KRI endpoint", func() {
	It("should properly generate CoreName on resources synced from different zone", func() {
		// given
		cpZone := "kuma-1"
		policyZone := "kuma-2"
		endpoint := kriEndpoint{
			cpMode:          core.Zone,
			cpZone:          cpZone,
			environment:     core.KubernetesEnvironment,
			systemNamespace: "kuma-system",
		}

		// when
		coreName := endpoint.getCoreName(kri.MustFromString("kri_mt_producer-policy-flow_" + policyZone + "_producer-policy-flow-ns_to-test-server_"))

		// then
		Expect(coreName).To(Equal("to-test-server-4cx47v8b5wcd4764.kuma-system"))
	})

	It("should properly generate CoreName on resources synced from global to zone", func() {
		// given
		cpZone := "kuma-1"
		policyZone := "" // global does not have a name
		endpoint := kriEndpoint{
			cpMode:          core.Zone,
			cpZone:          cpZone,
			environment:     core.KubernetesEnvironment,
			systemNamespace: "kuma-system",
		}

		// when
		coreName := endpoint.getCoreName(kri.MustFromString("kri_mt_producer-policy-flow_" + policyZone + "_producer-policy-flow-ns_to-test-server_"))

		// then
		Expect(coreName).To(Equal("to-test-server-5wxzc95dxv8zb244.kuma-system"))
	})

	It("should properly generate CoreName on resources synced from zone to global", func() {
		// given
		cpZone := "" // global does not have a name
		policyZone := "zone-1"
		endpoint := kriEndpoint{
			cpMode:          core.Global,
			cpZone:          cpZone,
			environment:     core.KubernetesEnvironment,
			systemNamespace: "kuma-system",
		}

		// when
		coreName := endpoint.getCoreName(kri.MustFromString("kri_mt_producer-policy-flow_" + policyZone + "_producer-policy-flow-ns_to-test-server_"))

		// then
		Expect(coreName).To(Equal("to-test-server-xw65cwcxxw4f7fvw.kuma-system"))
	})

	It("should properly generate CoreName on resources originating on Global", func() {
		// given
		cpZone := "" // global does not have a name
		policyZone := ""
		endpoint := kriEndpoint{
			cpMode:          core.Global,
			cpZone:          cpZone,
			environment:     core.KubernetesEnvironment,
			systemNamespace: "kuma-system",
		}

		// when
		coreName := endpoint.getCoreName(kri.MustFromString("kri_mt_producer-policy-flow_" + policyZone + "_producer-policy-flow-ns_to-test-server_"))

		// then
		Expect(coreName).To(Equal("to-test-server.producer-policy-flow-ns"))
	})

	It("should properly generate CoreName on resources from the same zone", func() {
		// given
		cpZone := "kuma-1"
		policyZone := "kuma-1"
		endpoint := kriEndpoint{
			cpMode:          core.Zone,
			cpZone:          cpZone,
			environment:     core.KubernetesEnvironment,
			systemNamespace: "kuma-system",
		}

		// when
		coreName := endpoint.getCoreName(kri.MustFromString("kri_mt_producer-policy-flow_" + policyZone + "_producer-policy-flow-ns_to-test-server_"))

		// then
		Expect(coreName).To(Equal("to-test-server.producer-policy-flow-ns"))
	})
})
