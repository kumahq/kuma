package api_server

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/v2/pkg/config/core"
	"github.com/kumahq/kuma/v2/pkg/core/kri"
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
		coreNames := endpoint.getCoreNames(kri.MustFromString("kri_mt_producer-policy-flow_" + policyZone + "_producer-policy-flow-ns_to-test-server_"))

		// then
		Expect(coreNames).To(Equal([]string{"to-test-server-4cx47v8b5wcd4764.kuma-system"}))
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
		coreNames := endpoint.getCoreNames(kri.MustFromString("kri_mt_producer-policy-flow_" + policyZone + "_producer-policy-flow-ns_to-test-server_"))

		// then (local name tried first, hash is fallback)
		Expect(coreNames).To(Equal([]string{"to-test-server.producer-policy-flow-ns", "to-test-server-5wxzc95dxv8zb244.kuma-system"}))
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
		coreNames := endpoint.getCoreNames(kri.MustFromString("kri_mt_producer-policy-flow_" + policyZone + "_producer-policy-flow-ns_to-test-server_"))

		// then
		Expect(coreNames).To(Equal([]string{"to-test-server-xw65cwcxxw4f7fvw.kuma-system"}))
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
		coreNames := endpoint.getCoreNames(kri.MustFromString("kri_mt_producer-policy-flow_" + policyZone + "_producer-policy-flow-ns_to-test-server_"))

		// then
		Expect(coreNames).To(Equal([]string{"to-test-server.producer-policy-flow-ns"}))
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
		coreNames := endpoint.getCoreNames(kri.MustFromString("kri_mt_producer-policy-flow_" + policyZone + "_producer-policy-flow-ns_to-test-server_"))

		// then
		Expect(coreNames).To(Equal([]string{"to-test-server.producer-policy-flow-ns"}))
	})

	It("should resolve locally created Universal resource without zone label", func() {
		// given: Zone CP, resource created without zone/namespace labels
		cpZone := "default"
		policyZone := ""
		endpoint := kriEndpoint{
			cpMode:          core.Zone,
			cpZone:          cpZone,
			environment:     core.UniversalEnvironment,
			systemNamespace: "kuma-system",
		}

		// when: KRI has empty zone and namespace slots
		coreNames := endpoint.getCoreNames(kri.MustFromString("kri_mtp_default_" + policyZone + "__without-namespace_"))

		// then: local name tried first (fixes issue #15544), hash as fallback
		Expect(coreNames[0]).To(Equal("without-namespace"))
	})
})
