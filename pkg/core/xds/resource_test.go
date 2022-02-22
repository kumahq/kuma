package xds_test

import (
	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/pkg/core/xds"
)

var _ = Describe("ResourceSet", func() {

	It("empty set should return empty list", func() {
		// when
		resources := NewResourceSet()
		// then
		Expect(len(resources.List())).To(Equal(0))
	})

	It("set of 1 element should return a list of 1 element", func() {
		// given
		resource := &Resource{
			Name: "backend",
			Resource: &envoy_cluster.Cluster{
				Name: "backend",
			},
		}
		// when
		resources := NewResourceSet()
		// and
		resources.Add(resource)

		// then
		Expect(resources.List()).To(ConsistOf(resource))
	})

	It("set of 2 elements should return a list of 2 elements", func() {
		// given
		resource1 := &Resource{
			Name: "backend",
			Resource: &envoy_cluster.Cluster{
				Name: "backend",
			},
		}
		resource2 := &Resource{
			Name: "outbound:127.0.0.1:8080",
			Resource: &envoy_listener.Listener{
				Name: "outbound:127.0.0.1:8080",
			},
		}

		// when
		resources := NewResourceSet()
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
			Name: "backend",
			Resource: &envoy_cluster.Cluster{
				Name: "backend",
			},
		}
		resource2 := &Resource{
			Name: "backend",
			Resource: &envoy_cluster.Cluster{
				Name: "backend",
			},
		}

		// when
		resources := NewResourceSet()
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
			Name: "backend",
			Resource: &envoy_cluster.Cluster{
				Name: "backend",
			},
		}
		resource2 := &Resource{
			Name: "backend",
			Resource: &envoy_listener.Listener{
				Name: "backend",
			},
		}

		// when
		resources := NewResourceSet()
		// and
		resources.Add(resource1)
		// and
		resources.Add(resource2)

		// then
		Expect(resources.List()).To(ConsistOf(resource1, resource2))
	})
})

var _ = Describe("ResourceList", func() {

	Describe("ToIndex()", func() {

		type testCase struct {
			input    ResourceList
			expected map[string]ResourcePayload
		}

		DescribeTable("should correctly generate an index of resources",
			func(given testCase) {
				Expect(given.input.ToIndex()).To(Equal(given.expected))
			},
			Entry("nil", testCase{
				input:    nil,
				expected: nil,
			}),
			Entry("empty", testCase{
				input:    ResourceList{},
				expected: nil,
			}),
			Entry("multiple resources with the same name", testCase{
				input: ResourceList{
					{
						Name: "backend",
						Resource: &envoy_cluster.Cluster{
							Name: "backend",
						},
					},
					{
						Name: "backend",
						Resource: &envoy_listener.Listener{
							Name: "backend",
						},
					},
					{
						Name: "web",
						Resource: &envoy_cluster.Cluster{
							Name: "web",
						},
					},
				},
				expected: map[string]ResourcePayload{
					"backend": &envoy_listener.Listener{
						Name: "backend",
					},
					"web": &envoy_cluster.Cluster{
						Name: "web",
					},
				},
			}),
		)
	})
})
