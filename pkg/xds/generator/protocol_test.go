package generator_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	mesh_core "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	. "github.com/kumahq/kuma/pkg/xds/generator"
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
		Entry("one-item list: no `kuma.io/protocol` tag", testCase{
			endpoints: []core_xds.Endpoint{
				{Tags: map[string]string{"kuma.io/service": "backend"}},
			},
			expected: mesh_core.ProtocolUnknown,
		}),
		Entry("one-item list: `kuma.io/protocol: http`", testCase{
			endpoints: []core_xds.Endpoint{
				{Tags: map[string]string{"kuma.io/service": "backend", "kuma.io/protocol": "http"}},
			},
			expected: mesh_core.ProtocolHTTP,
		}),
		Entry("one-item list: `kuma.io/protocol: http2`", testCase{
			endpoints: []core_xds.Endpoint{
				{Tags: map[string]string{"kuma.io/service": "backend", "kuma.io/protocol": "http2"}},
			},
			expected: mesh_core.ProtocolHTTP2,
		}),
		Entry("one-item list: `kuma.io/protocol: kafka`", testCase{
			endpoints: []core_xds.Endpoint{
				{Tags: map[string]string{"kuma.io/service": "kafka-broker", "kuma.io/protocol": "kafka"}},
			},
			expected: mesh_core.ProtocolKafka,
		}),
		Entry("one-item list: `kuma.io/protocol: tcp`", testCase{
			endpoints: []core_xds.Endpoint{
				{Tags: map[string]string{"kuma.io/service": "backend", "kuma.io/protocol": "tcp"}},
			},
			expected: mesh_core.ProtocolTCP,
		}),
		Entry("one-item list: `kuma.io/protocol: grpc`", testCase{
			endpoints: []core_xds.Endpoint{
				{Tags: map[string]string{"kuma.io/service": "backend", "kuma.io/protocol": "grpc"}},
			},
			expected: mesh_core.ProtocolGRPC,
		}),
		Entry("one-item list: `kuma.io/protocol: not-yet-supported-protocol`", testCase{
			endpoints: []core_xds.Endpoint{
				{Tags: map[string]string{"kuma.io/service": "backend", "kuma.io/protocol": "not-yet-supported-protocol"}},
			},
			expected: mesh_core.ProtocolUnknown,
		}),
		Entry("multi-item list: no `protocol` tag", testCase{
			endpoints: []core_xds.Endpoint{
				{Tags: map[string]string{"kuma.io/service": "backend", "region": "us"}},
				{Tags: map[string]string{"kuma.io/service": "backend", "region": "eu"}},
			},
			expected: mesh_core.ProtocolUnknown,
		}),
		Entry("multi-item list: no `protocol` tag on some endpoints", testCase{
			endpoints: []core_xds.Endpoint{
				{Tags: map[string]string{"kuma.io/service": "backend", "region": "us", "kuma.io/protocol": "http"}},
				{Tags: map[string]string{"kuma.io/service": "backend", "region": "eu"}},
			},
			expected: mesh_core.ProtocolUnknown,
		}),
		Entry("multi-item list: `kuma.io/protocol: http` on every endpoint", testCase{
			endpoints: []core_xds.Endpoint{
				{Tags: map[string]string{"kuma.io/service": "backend", "region": "us", "kuma.io/protocol": "http"}},
				{Tags: map[string]string{"kuma.io/service": "backend", "region": "eu", "kuma.io/protocol": "http"}},
			},
			expected: mesh_core.ProtocolHTTP,
		}),
		Entry("multi-item list: `kuma.io/protocol: http` on every endpoint", testCase{
			endpoints: []core_xds.Endpoint{
				{Tags: map[string]string{"kuma.io/service": "backend", "region": "us", "kuma.io/protocol": "http2"}},
				{Tags: map[string]string{"kuma.io/service": "backend", "region": "eu", "kuma.io/protocol": "http2"}},
			},
			expected: mesh_core.ProtocolHTTP2,
		}),
		Entry("multi-item list: `kuma.io/protocol: tcp` on every endpoint", testCase{
			endpoints: []core_xds.Endpoint{
				{Tags: map[string]string{"kuma.io/service": "backend", "region": "us", "kuma.io/protocol": "tcp"}},
				{Tags: map[string]string{"kuma.io/service": "backend", "region": "eu", "kuma.io/protocol": "tcp"}},
			},
			expected: mesh_core.ProtocolTCP,
		}),
		Entry("multi-item list: `kuma.io/protocol: tcp` and `kuma.io/protocol: http`", testCase{
			endpoints: []core_xds.Endpoint{
				{Tags: map[string]string{"kuma.io/service": "backend", "region": "us", "kuma.io/protocol": "tcp"}},
				{Tags: map[string]string{"kuma.io/service": "backend", "region": "eu", "kuma.io/protocol": "http"}},
			},
			expected: mesh_core.ProtocolTCP,
		}),
		Entry("multi-item list: `kuma.io/protocol: grpc` on every endpoint", testCase{
			endpoints: []core_xds.Endpoint{
				{Tags: map[string]string{"kuma.io/service": "backend", "region": "us", "kuma.io/protocol": "grpc"}},
				{Tags: map[string]string{"kuma.io/service": "backend", "region": "eu", "kuma.io/protocol": "grpc"}},
			},
			expected: mesh_core.ProtocolGRPC,
		}),
		Entry("multi-item list: `kuma.io/protocol: grpc` and `kuma.io/protocol: http`", testCase{
			endpoints: []core_xds.Endpoint{
				{Tags: map[string]string{"kuma.io/service": "backend", "region": "us", "kuma.io/protocol": "grpc"}},
				{Tags: map[string]string{"kuma.io/service": "backend", "region": "eu", "kuma.io/protocol": "http"}},
			},
			expected: mesh_core.ProtocolTCP,
		}),
		Entry("multi-item list: `kuma.io/protocol: grpc` and `kuma.io/protocol: http2`", testCase{
			endpoints: []core_xds.Endpoint{
				{Tags: map[string]string{"kuma.io/service": "backend", "region": "us", "kuma.io/protocol": "grpc"}},
				{Tags: map[string]string{"kuma.io/service": "backend", "region": "eu", "kuma.io/protocol": "http2"}},
			},
			expected: mesh_core.ProtocolHTTP2,
		}),
		Entry("multi-item list: `kuma.io/protocol: grpc` and `kuma.io/protocol: tcp`", testCase{
			endpoints: []core_xds.Endpoint{
				{Tags: map[string]string{"kuma.io/service": "backend", "region": "us", "kuma.io/protocol": "grpc"}},
				{Tags: map[string]string{"kuma.io/service": "backend", "region": "eu", "kuma.io/protocol": "tcp"}},
			},
			expected: mesh_core.ProtocolTCP,
		}),
	)
})
