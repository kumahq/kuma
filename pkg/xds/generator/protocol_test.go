package generator_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	. "github.com/Kong/kuma/pkg/xds/generator"

	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/Kong/kuma/pkg/core/xds"
)

var _ = Describe("InferServiceProtocol()", func() {

	type testCase struct {
		endpoints []core_xds.Endpoint
		expected  mesh_core.Protocol
	}

	DescribeTable("should correctly infer common protocol for a group of endpoints",
		func(given testCase) {
			// when
			actual := InferServiceProtocol(given.endpoints)
			// then
			Expect(actual).To(Equal(given.expected))
		},
		Entry("empty list", testCase{
			endpoints: nil,
			expected:  mesh_core.ProtocolUnknown,
		}),
		Entry("one-item list: no `protocol` tag", testCase{
			endpoints: []core_xds.Endpoint{
				{Tags: map[string]string{"service": "backend"}},
			},
			expected: mesh_core.ProtocolUnknown,
		}),
		Entry("one-item list: `protocol: http`", testCase{
			endpoints: []core_xds.Endpoint{
				{Tags: map[string]string{"service": "backend", "protocol": "http"}},
			},
			expected: mesh_core.ProtocolHTTP,
		}),
		Entry("one-item list: `protocol: tcp`", testCase{
			endpoints: []core_xds.Endpoint{
				{Tags: map[string]string{"service": "backend", "protocol": "tcp"}},
			},
			expected: mesh_core.ProtocolTCP,
		}),
		Entry("one-item list: `protocol: not-yet-supported-protocol`", testCase{
			endpoints: []core_xds.Endpoint{
				{Tags: map[string]string{"service": "backend", "protocol": "not-yet-supported-protocol"}},
			},
			expected: mesh_core.ProtocolUnknown,
		}),
		Entry("multi-item list: no `protocol` tag", testCase{
			endpoints: []core_xds.Endpoint{
				{Tags: map[string]string{"service": "backend", "region": "us"}},
				{Tags: map[string]string{"service": "backend", "region": "eu"}},
			},
			expected: mesh_core.ProtocolUnknown,
		}),
		Entry("multi-item list: no `protocol` tag on some endpoints", testCase{
			endpoints: []core_xds.Endpoint{
				{Tags: map[string]string{"service": "backend", "region": "us", "protocol": "http"}},
				{Tags: map[string]string{"service": "backend", "region": "eu"}},
			},
			expected: mesh_core.ProtocolUnknown,
		}),
		Entry("multi-item list: `protocol: http` on every endpoint", testCase{
			endpoints: []core_xds.Endpoint{
				{Tags: map[string]string{"service": "backend", "region": "us", "protocol": "http"}},
				{Tags: map[string]string{"service": "backend", "region": "eu", "protocol": "http"}},
			},
			expected: mesh_core.ProtocolHTTP,
		}),
		Entry("multi-item list: `protocol: tcp` on every endpoint", testCase{
			endpoints: []core_xds.Endpoint{
				{Tags: map[string]string{"service": "backend", "region": "us", "protocol": "tcp"}},
				{Tags: map[string]string{"service": "backend", "region": "eu", "protocol": "tcp"}},
			},
			expected: mesh_core.ProtocolTCP,
		}),
		Entry("multi-item list: `protocol: tcp` and `protocol: http`", testCase{
			endpoints: []core_xds.Endpoint{
				{Tags: map[string]string{"service": "backend", "region": "us", "protocol": "tcp"}},
				{Tags: map[string]string{"service": "backend", "region": "eu", "protocol": "http"}},
			},
			expected: mesh_core.ProtocolTCP,
		}),
	)
})
