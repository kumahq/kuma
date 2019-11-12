package generator_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/Kong/kuma/pkg/xds/generator"

	envoy "github.com/envoyproxy/go-control-plane/envoy/api/v2"
)

var _ = Describe("ResourceSet", func() {

	It("empty set should return empty list", func() {
		// when
		resources := &ResourceSet{}
		// then
		Expect(resources.List()).To(BeNil())
	})

	It("set of 1 element should return a list of 1 element", func() {
		// given
		resource := &Resource{
			Name:    "backend",
			Version: "v1",
			Resource: &envoy.Cluster{
				Name: "backend",
			},
		}
		// when
		resources := &ResourceSet{}
		// and
		resources.Add(resource)

		// then
		Expect(resources.List()).To(ConsistOf(resource))
	})

	It("set of 2 elements should return a list of 2 elements", func() {
		// given
		resource1 := &Resource{
			Name:    "backend",
			Version: "v1",
			Resource: &envoy.Cluster{
				Name: "backend",
			},
		}
		resource2 := &Resource{
			Name:    "outbound:127.0.0.1:8080",
			Version: "v2",
			Resource: &envoy.Listener{
				Name: "outbound:127.0.0.1:8080",
			},
		}

		// when
		resources := &ResourceSet{}
		// and
		resources.Add(resource1)
		// and
		resources.Add(resource2)

		// then
		Expect(resources.List()).To(ConsistOf(resource1, resource2))
	})

	It("should not be possible to add 2 resources with same name and type", func() {
		// given
		resource1 := &Resource{
			Name:    "backend",
			Version: "v1",
			Resource: &envoy.Cluster{
				Name: "backend",
			},
		}
		resource2 := &Resource{
			Name:    "backend",
			Version: "v2",
			Resource: &envoy.Cluster{
				Name: "backend",
			},
		}

		// when
		resources := &ResourceSet{}
		// and
		resources.Add(resource1)
		// and
		resources.Add(resource2)

		// then
		Expect(resources.List()).To(ConsistOf(resource1))
	})

	It("should be possible to add 2 resources with same name but different types", func() {
		// given
		resource1 := &Resource{
			Name:    "backend",
			Version: "v1",
			Resource: &envoy.Cluster{
				Name: "backend",
			},
		}
		resource2 := &Resource{
			Name:    "backend",
			Version: "v2",
			Resource: &envoy.Listener{
				Name: "backend",
			},
		}

		// when
		resources := &ResourceSet{}
		// and
		resources.Add(resource1)
		// and
		resources.Add(resource2)

		// then
		Expect(resources.List()).To(ConsistOf(resource1, resource2))
	})
})
