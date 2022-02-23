package generator_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	. "github.com/kumahq/kuma/pkg/xds/generator"
)

var _ = Describe("InferServiceProtocol()", func() {

	type testCase struct {
		endpoints []core_xds.Endpoint
		expected  core_mesh.Protocol
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
			expected:  core_mesh.ProtocolUnknown,
		}),
		Entry("one-item list: no `kuma.io/protocol` tag", testCase{
			endpoints: []core_xds.Endpoint{
				{Tags: map[string]string{"kuma.io/service": "backend"}},
			},
			expected: core_mesh.ProtocolUnknown,
		}),
		Entry("one-item list: `kuma.io/protocol: http`", testCase{
			endpoints: []core_xds.Endpoint{
				{Tags: map[string]string{"kuma.io/service": "backend", "kuma.io/protocol": "http"}},
			},
			expected: core_mesh.ProtocolHTTP,
		}),
		Entry("one-item list: `kuma.io/protocol: http2`", testCase{
			endpoints: []core_xds.Endpoint{
				{Tags: map[string]string{"kuma.io/service": "backend", "kuma.io/protocol": "http2"}},
			},
			expected: core_mesh.ProtocolHTTP2,
		}),
		Entry("one-item list: `kuma.io/protocol: kafka`", testCase{
			endpoints: []core_xds.Endpoint{
				{Tags: map[string]string{"kuma.io/service": "kafka-broker", "kuma.io/protocol": "kafka"}},
			},
			expected: core_mesh.ProtocolKafka,
		}),
		Entry("one-item list: `kuma.io/protocol: tcp`", testCase{
			endpoints: []core_xds.Endpoint{
				{Tags: map[string]string{"kuma.io/service": "backend", "kuma.io/protocol": "tcp"}},
			},
			expected: core_mesh.ProtocolTCP,
		}),
		Entry("one-item list: `kuma.io/protocol: grpc`", testCase{
			endpoints: []core_xds.Endpoint{
				{Tags: map[string]string{"kuma.io/service": "backend", "kuma.io/protocol": "grpc"}},
			},
			expected: core_mesh.ProtocolGRPC,
		}),
		Entry("one-item list: `kuma.io/protocol: not-yet-supported-protocol`", testCase{
			endpoints: []core_xds.Endpoint{
				{Tags: map[string]string{"kuma.io/service": "backend", "kuma.io/protocol": "not-yet-supported-protocol"}},
			},
			expected: core_mesh.ProtocolUnknown,
		}),
		Entry("multi-item list: no `protocol` tag", testCase{
			endpoints: []core_xds.Endpoint{
				{Tags: map[string]string{"kuma.io/service": "backend", "region": "us"}},
				{Tags: map[string]string{"kuma.io/service": "backend", "region": "eu"}},
			},
			expected: core_mesh.ProtocolUnknown,
		}),
		Entry("multi-item list: no `protocol` tag on some endpoints", testCase{
			endpoints: []core_xds.Endpoint{
				{Tags: map[string]string{"kuma.io/service": "backend", "region": "us", "kuma.io/protocol": "http"}},
				{Tags: map[string]string{"kuma.io/service": "backend", "region": "eu"}},
			},
			expected: core_mesh.ProtocolUnknown,
		}),
		Entry("multi-item list: `kuma.io/protocol: http` on every endpoint", testCase{
			endpoints: []core_xds.Endpoint{
				{Tags: map[string]string{"kuma.io/service": "backend", "region": "us", "kuma.io/protocol": "http"}},
				{Tags: map[string]string{"kuma.io/service": "backend", "region": "eu", "kuma.io/protocol": "http"}},
			},
			expected: core_mesh.ProtocolHTTP,
		}),
		Entry("multi-item list: `kuma.io/protocol: http` on every endpoint", testCase{
			endpoints: []core_xds.Endpoint{
				{Tags: map[string]string{"kuma.io/service": "backend", "region": "us", "kuma.io/protocol": "http2"}},
				{Tags: map[string]string{"kuma.io/service": "backend", "region": "eu", "kuma.io/protocol": "http2"}},
			},
			expected: core_mesh.ProtocolHTTP2,
		}),
		Entry("multi-item list: `kuma.io/protocol: tcp` on every endpoint", testCase{
			endpoints: []core_xds.Endpoint{
				{Tags: map[string]string{"kuma.io/service": "backend", "region": "us", "kuma.io/protocol": "tcp"}},
				{Tags: map[string]string{"kuma.io/service": "backend", "region": "eu", "kuma.io/protocol": "tcp"}},
			},
			expected: core_mesh.ProtocolTCP,
		}),
		Entry("multi-item list: `kuma.io/protocol: tcp` and `kuma.io/protocol: http`", testCase{
			endpoints: []core_xds.Endpoint{
				{Tags: map[string]string{"kuma.io/service": "backend", "region": "us", "kuma.io/protocol": "tcp"}},
				{Tags: map[string]string{"kuma.io/service": "backend", "region": "eu", "kuma.io/protocol": "http"}},
			},
			expected: core_mesh.ProtocolTCP,
		}),
		Entry("multi-item list: `kuma.io/protocol: grpc` on every endpoint", testCase{
			endpoints: []core_xds.Endpoint{
				{Tags: map[string]string{"kuma.io/service": "backend", "region": "us", "kuma.io/protocol": "grpc"}},
				{Tags: map[string]string{"kuma.io/service": "backend", "region": "eu", "kuma.io/protocol": "grpc"}},
			},
			expected: core_mesh.ProtocolGRPC,
		}),
		Entry("multi-item list: `kuma.io/protocol: grpc` and `kuma.io/protocol: http`", testCase{
			endpoints: []core_xds.Endpoint{
				{Tags: map[string]string{"kuma.io/service": "backend", "region": "us", "kuma.io/protocol": "grpc"}},
				{Tags: map[string]string{"kuma.io/service": "backend", "region": "eu", "kuma.io/protocol": "http"}},
			},
			expected: core_mesh.ProtocolTCP,
		}),
		Entry("multi-item list: `kuma.io/protocol: grpc` and `kuma.io/protocol: http2`", testCase{
			endpoints: []core_xds.Endpoint{
				{Tags: map[string]string{"kuma.io/service": "backend", "region": "us", "kuma.io/protocol": "grpc"}},
				{Tags: map[string]string{"kuma.io/service": "backend", "region": "eu", "kuma.io/protocol": "http2"}},
			},
			expected: core_mesh.ProtocolHTTP2,
		}),
		Entry("multi-item list: `kuma.io/protocol: grpc` and `kuma.io/protocol: tcp`", testCase{
			endpoints: []core_xds.Endpoint{
				{Tags: map[string]string{"kuma.io/service": "backend", "region": "us", "kuma.io/protocol": "grpc"}},
				{Tags: map[string]string{"kuma.io/service": "backend", "region": "eu", "kuma.io/protocol": "tcp"}},
			},
			expected: core_mesh.ProtocolTCP,
		}),
	)
})
