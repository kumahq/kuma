package gatewayapi_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	gatewayapi "sigs.k8s.io/gateway-api/apis/v1beta1"

	k8s_gatewayapi "github.com/kumahq/kuma/pkg/plugins/runtime/k8s/controllers/gatewayapi"
)

var _ = Describe("ValidateListeners", func() {
	It("works with one simple listener", func() {
		same := gatewayapi.NamespacesFromSame
		listeners := []gatewayapi.Listener{
			{
				Name:     gatewayapi.SectionName("prod"),
				Protocol: gatewayapi.HTTPProtocolType,
				Port:     gatewayapi.PortNumber(80),
				AllowedRoutes: &gatewayapi.AllowedRoutes{
					Namespaces: &gatewayapi.RouteNamespaces{
						From: &same,
					},
				},
			},
		}
		valids, conditions := k8s_gatewayapi.ValidateListeners(false, listeners)
		Expect(valids).To(ConsistOf(
			HaveField("Name", gatewayapi.SectionName("prod")),
		))
		Expect(conditions).To(BeEmpty())
	})
	It("works with conflicting protocols", func() {
		same := gatewayapi.NamespacesFromSame
		listeners := []gatewayapi.Listener{
			{
				Name:     gatewayapi.SectionName("prod-1"),
				Protocol: gatewayapi.HTTPProtocolType,
				Port:     gatewayapi.PortNumber(80),
				AllowedRoutes: &gatewayapi.AllowedRoutes{
					Namespaces: &gatewayapi.RouteNamespaces{
						From: &same,
					},
				},
			},
			{
				Name:     gatewayapi.SectionName("prod-2"),
				Protocol: gatewayapi.HTTPSProtocolType,
				Port:     gatewayapi.PortNumber(80),
				AllowedRoutes: &gatewayapi.AllowedRoutes{
					Namespaces: &gatewayapi.RouteNamespaces{
						From: &same,
					},
				},
			},
		}
		valids, conditions := k8s_gatewayapi.ValidateListeners(false, listeners)

		Expect(valids).To(BeEmpty())

		protocolConflicted := ContainElements(
			MatchFields(IgnoreExtras, Fields{
				"Type":   Equal(string(gatewayapi.ListenerConditionConflicted)),
				"Status": Equal(kube_meta.ConditionTrue),
				"Reason": Equal(string(gatewayapi.ListenerReasonProtocolConflict)),
			}),
			MatchFields(IgnoreExtras, Fields{
				"Type":   Equal(string(gatewayapi.ListenerConditionProgrammed)),
				"Status": Equal(kube_meta.ConditionFalse),
			}),
		)
		Expect(conditions).To(
			MatchAllKeys(Keys{
				gatewayapi.SectionName("prod-1"): protocolConflicted,
				gatewayapi.SectionName("prod-2"): protocolConflicted,
			}),
		)
	})
	It("works with non-conflicting differing hostnames", func() {
		same := gatewayapi.NamespacesFromSame
		foo := gatewayapi.Hostname("foo.com")
		bar := gatewayapi.Hostname("bar.com")
		listeners := []gatewayapi.Listener{
			{
				Name:     gatewayapi.SectionName("prod-1"),
				Protocol: gatewayapi.HTTPProtocolType,
				Hostname: &foo,
				Port:     gatewayapi.PortNumber(80),
				AllowedRoutes: &gatewayapi.AllowedRoutes{
					Namespaces: &gatewayapi.RouteNamespaces{
						From: &same,
					},
				},
			},
			{
				Name:     gatewayapi.SectionName("prod-2"),
				Protocol: gatewayapi.HTTPProtocolType,
				Hostname: &bar,
				Port:     gatewayapi.PortNumber(80),
				AllowedRoutes: &gatewayapi.AllowedRoutes{
					Namespaces: &gatewayapi.RouteNamespaces{
						From: &same,
					},
				},
			},
		}
		valids, conditions := k8s_gatewayapi.ValidateListeners(false, listeners)
		Expect(valids).To(ConsistOf(
			HaveField("Name", gatewayapi.SectionName("prod-1")),
			HaveField("Name", gatewayapi.SectionName("prod-2")),
		))
		Expect(conditions).To(BeEmpty())
	})
	It("enforces HTTP and cross-mesh", func() {
		same := gatewayapi.NamespacesFromSame
		listeners := []gatewayapi.Listener{
			{
				Name:     gatewayapi.SectionName("valid-mesh"),
				Protocol: gatewayapi.HTTPProtocolType,
				Port:     gatewayapi.PortNumber(80),
				AllowedRoutes: &gatewayapi.AllowedRoutes{
					Namespaces: &gatewayapi.RouteNamespaces{
						From: &same,
					},
				},
			},
			{
				Name:     gatewayapi.SectionName("invalid-mesh"),
				Protocol: gatewayapi.HTTPSProtocolType,
				Port:     gatewayapi.PortNumber(80),
				AllowedRoutes: &gatewayapi.AllowedRoutes{
					Namespaces: &gatewayapi.RouteNamespaces{
						From: &same,
					},
				},
			},
		}

		valids, conditions := k8s_gatewayapi.ValidateListeners(true, listeners)
		Expect(valids).To(ConsistOf(
			HaveField("Name", gatewayapi.SectionName("valid-mesh")),
		))

		invalid := ContainElements(
			MatchFields(IgnoreExtras, Fields{
				"Type":   Equal(string(gatewayapi.ListenerConditionAccepted)),
				"Status": Equal(kube_meta.ConditionFalse),
				"Reason": Equal(string(gatewayapi.ListenerReasonUnsupportedProtocol)),
			}),
			MatchFields(IgnoreExtras, Fields{
				"Type":   Equal(string(gatewayapi.ListenerConditionProgrammed)),
				"Status": Equal(kube_meta.ConditionFalse),
			}),
		)
		Expect(conditions).To(
			MatchAllKeys(Keys{
				gatewayapi.SectionName("invalid-mesh"): invalid,
			}),
		)
	})
	It("works with multiple listeners for same hostname:port conflict", func() {
		same := gatewayapi.NamespacesFromSame
		foo := gatewayapi.Hostname("foo.com")
		listeners := []gatewayapi.Listener{
			{
				Name:     gatewayapi.SectionName("prod-1"),
				Protocol: gatewayapi.HTTPProtocolType,
				Hostname: &foo,
				Port:     gatewayapi.PortNumber(80),
				AllowedRoutes: &gatewayapi.AllowedRoutes{
					Namespaces: &gatewayapi.RouteNamespaces{
						From: &same,
					},
				},
			},
			{
				Name:     gatewayapi.SectionName("prod-2"),
				Protocol: gatewayapi.HTTPProtocolType,
				Hostname: &foo,
				Port:     gatewayapi.PortNumber(80),
				AllowedRoutes: &gatewayapi.AllowedRoutes{
					Namespaces: &gatewayapi.RouteNamespaces{
						From: &same,
					},
				},
			},
		}
		valids, conditions := k8s_gatewayapi.ValidateListeners(false, listeners)

		Expect(valids).To(BeEmpty())

		hostnameConflicted := ContainElements(
			MatchFields(IgnoreExtras, Fields{
				"Type":   Equal(string(gatewayapi.ListenerConditionConflicted)),
				"Status": Equal(kube_meta.ConditionTrue),
				"Reason": Equal(string(gatewayapi.ListenerReasonHostnameConflict)),
			}),
			MatchFields(IgnoreExtras, Fields{
				"Type":   Equal(string(gatewayapi.ListenerConditionProgrammed)),
				"Status": Equal(kube_meta.ConditionFalse),
			}),
		)
		Expect(conditions).To(
			MatchAllKeys(Keys{
				gatewayapi.SectionName("prod-1"): hostnameConflicted,
				gatewayapi.SectionName("prod-2"): hostnameConflicted,
			}),
		)
	})
})
