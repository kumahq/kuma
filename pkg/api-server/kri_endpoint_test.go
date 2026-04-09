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

		// when: zone="" means globally originated, hash only includes non-empty labels
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

	It("should resolve globally originated resource on Zone CP without zone/ns labels", func() {
		// given: Zone CP, resource originated on Global without zone/namespace labels
		cpZone := "default"
		policyZone := ""
		endpoint := kriEndpoint{
			cpMode:          core.Zone,
			cpZone:          cpZone,
			environment:     core.UniversalEnvironment,
			systemNamespace: "kuma-system",
		}

		// when: zone="" means globally originated, hash computed with no extra values
		coreName := endpoint.getCoreName(kri.MustFromString("kri_mtp_default_" + policyZone + "__without-namespace_"))

		// then
		Expect(coreName).To(Equal("without-namespace-c5v4498z5w4x9bcx"))
	})

	It("should resolve local global-scoped resource on Zone CP", func() {
		// given: Zone CP, global-scoped resource (mesh="") created locally
		// Reproduces github.com/kumahq/kuma/issues/15803
		endpoint := kriEndpoint{
			cpMode:          core.Zone,
			cpZone:          "default",
			environment:     core.KubernetesEnvironment,
			systemNamespace: "kuma-system",
		}

		// when: HostnameGenerator KRI with empty mesh, zone matches cpZone
		coreName := endpoint.getCoreName(kri.MustFromString("kri_hg__default_kuma-system_local-mesh-service-1_"))

		// then: locally originated, uses plain name
		Expect(coreName).To(Equal("local-mesh-service-1.kuma-system"))
	})
})
