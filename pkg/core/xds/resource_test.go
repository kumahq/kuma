package xds_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	. "github.com/Kong/kuma/pkg/core/xds"

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
						Resource: &envoy.Cluster{
							Name: "backend",
						},
					},
					{
						Name: "backend",
						Resource: &envoy.Listener{
							Name: "backend",
						},
					},
					{
						Name: "web",
						Resource: &envoy.Cluster{
							Name: "web",
						},
					},
				},
				expected: map[string]ResourcePayload{
					"backend": &envoy.Listener{
						Name: "backend",
					},
					"web": &envoy.Cluster{
						Name: "web",
					},
				},
			}),
		)
	})
})
