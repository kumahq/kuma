package generator

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	mesh_core "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
)

var _ = Describe("getCommonProtocol()", func() {
	type testCase struct {
		one      mesh_core.Protocol
		another  mesh_core.Protocol
		expected mesh_core.Protocol
	}

	DescribeTable("should correctly determine common protocol",
		func(given testCase) {
			// when
			actual := getCommonProtocol(given.one, given.another)
			// then
			Expect(actual).To(Equal(given.expected))
		},
		Entry("`unknown` and `unknown`", testCase{
			one:      mesh_core.ProtocolUnknown,
			another:  mesh_core.ProtocolUnknown,
			expected: mesh_core.ProtocolUnknown,
		}),
		Entry("`unknown` and `http`", testCase{
			one:      mesh_core.ProtocolUnknown,
			another:  mesh_core.ProtocolHTTP,
			expected: mesh_core.ProtocolUnknown,
		}),
		Entry("`http` and `unknown`", testCase{
			one:      mesh_core.ProtocolHTTP,
			another:  mesh_core.ProtocolUnknown,
			expected: mesh_core.ProtocolUnknown,
		}),
		Entry("`unknown` and `tcp`", testCase{
			one:      mesh_core.ProtocolUnknown,
			another:  mesh_core.ProtocolTCP,
			expected: mesh_core.ProtocolUnknown,
		}),
		Entry("`tcp` and `unknown`", testCase{
			one:      mesh_core.ProtocolTCP,
			another:  mesh_core.ProtocolUnknown,
			expected: mesh_core.ProtocolUnknown,
		}),
		Entry("`http` and `tcp`", testCase{
			one:      mesh_core.ProtocolHTTP,
			another:  mesh_core.ProtocolTCP,
			expected: mesh_core.ProtocolTCP,
		}),
		Entry("`tcp` and `http`", testCase{
			one:      mesh_core.ProtocolTCP,
			another:  mesh_core.ProtocolHTTP,
			expected: mesh_core.ProtocolTCP,
		}),
		Entry("`http` and `http`", testCase{
			one:      mesh_core.ProtocolHTTP,
			another:  mesh_core.ProtocolHTTP,
			expected: mesh_core.ProtocolHTTP,
		}),
		Entry("`tcp` and `tcp`", testCase{
			one:      mesh_core.ProtocolTCP,
			another:  mesh_core.ProtocolTCP,
			expected: mesh_core.ProtocolTCP,
		}),
		Entry("`http2` and `http2`", testCase{
			one:      mesh_core.ProtocolHTTP2,
			another:  mesh_core.ProtocolHTTP2,
			expected: mesh_core.ProtocolHTTP2,
		}),
		Entry("`http2` and `http`", testCase{
			one:      mesh_core.ProtocolHTTP2,
			another:  mesh_core.ProtocolHTTP,
			expected: mesh_core.ProtocolTCP,
		}),
		Entry("`http2` and `tcp`", testCase{
			one:      mesh_core.ProtocolHTTP2,
			another:  mesh_core.ProtocolTCP,
			expected: mesh_core.ProtocolTCP,
		}),
		Entry("`grpc` and `grpc`", testCase{
			one:      mesh_core.ProtocolGRPC,
			another:  mesh_core.ProtocolGRPC,
			expected: mesh_core.ProtocolGRPC,
		}),
		Entry("`grpc` and `http`", testCase{
			one:      mesh_core.ProtocolGRPC,
			another:  mesh_core.ProtocolHTTP,
			expected: mesh_core.ProtocolTCP,
		}),
		Entry("`grpc` and `http2`", testCase{
			one:      mesh_core.ProtocolGRPC,
			another:  mesh_core.ProtocolHTTP2,
			expected: mesh_core.ProtocolHTTP2,
		}),
		Entry("`grpc` and `tcp`", testCase{
			one:      mesh_core.ProtocolGRPC,
			another:  mesh_core.ProtocolTCP,
			expected: mesh_core.ProtocolTCP,
		}),
		Entry("`kafka` and `tcp`", testCase{
			one:      mesh_core.ProtocolKafka,
			another:  mesh_core.ProtocolTCP,
			expected: mesh_core.ProtocolTCP,
		}),
	)
})
