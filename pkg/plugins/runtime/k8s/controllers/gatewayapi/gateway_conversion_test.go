package gatewayapi_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	gatewayapi "sigs.k8s.io/gateway-api/apis/v1alpha2"

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
		valids, conditions := k8s_gatewayapi.ValidateListeners(listeners)
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
				Protocol: gatewayapi.UDPProtocolType,
				Port:     gatewayapi.PortNumber(80),
				AllowedRoutes: &gatewayapi.AllowedRoutes{
					Namespaces: &gatewayapi.RouteNamespaces{
						From: &same,
					},
				},
			},
		}
		valids, conditions := k8s_gatewayapi.ValidateListeners(listeners)
		Expect(valids).To(ConsistOf(
			HaveField("Name", gatewayapi.SectionName("prod-1")),
		))
		Expect(conditions).To(HaveKey(gatewayapi.SectionName("prod-2")))
	})
	It("works with differing hostnames", func() {
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
		valids, conditions := k8s_gatewayapi.ValidateListeners(listeners)
		Expect(valids).To(ConsistOf(
			HaveField("Name", gatewayapi.SectionName("prod-1")),
			HaveField("Name", gatewayapi.SectionName("prod-2")),
		))
		Expect(conditions).To(BeEmpty())
	})
	It("works with multiple listeners for same hostname:port", func() {
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
		valids, conditions := k8s_gatewayapi.ValidateListeners(listeners)
		Expect(valids).To(BeEmpty())
		Expect(conditions).To(HaveKey(gatewayapi.SectionName("prod-1")))
		Expect(conditions).To(HaveKey(gatewayapi.SectionName("prod-2")))
	})
})
