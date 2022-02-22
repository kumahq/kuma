package generator

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
)

var _ = Describe("getCommonProtocol()", func() {
	type testCase struct {
		one      core_mesh.Protocol
		another  core_mesh.Protocol
		expected core_mesh.Protocol
	}

	DescribeTable("should correctly determine common protocol",
		func(given testCase) {
			// when
			actual := getCommonProtocol(given.one, given.another)
			// then
			Expect(actual).To(Equal(given.expected))
		},
		Entry("`unknown` and `unknown`", testCase{
			one:      core_mesh.ProtocolUnknown,
			another:  core_mesh.ProtocolUnknown,
			expected: core_mesh.ProtocolUnknown,
		}),
		Entry("`unknown` and `http`", testCase{
			one:      core_mesh.ProtocolUnknown,
			another:  core_mesh.ProtocolHTTP,
			expected: core_mesh.ProtocolUnknown,
		}),
		Entry("`http` and `unknown`", testCase{
			one:      core_mesh.ProtocolHTTP,
			another:  core_mesh.ProtocolUnknown,
			expected: core_mesh.ProtocolUnknown,
		}),
		Entry("`unknown` and `tcp`", testCase{
			one:      core_mesh.ProtocolUnknown,
			another:  core_mesh.ProtocolTCP,
			expected: core_mesh.ProtocolUnknown,
		}),
		Entry("`tcp` and `unknown`", testCase{
			one:      core_mesh.ProtocolTCP,
			another:  core_mesh.ProtocolUnknown,
			expected: core_mesh.ProtocolUnknown,
		}),
		Entry("`http` and `tcp`", testCase{
			one:      core_mesh.ProtocolHTTP,
			another:  core_mesh.ProtocolTCP,
			expected: core_mesh.ProtocolTCP,
		}),
		Entry("`tcp` and `http`", testCase{
			one:      core_mesh.ProtocolTCP,
			another:  core_mesh.ProtocolHTTP,
			expected: core_mesh.ProtocolTCP,
		}),
		Entry("`http` and `http`", testCase{
			one:      core_mesh.ProtocolHTTP,
			another:  core_mesh.ProtocolHTTP,
			expected: core_mesh.ProtocolHTTP,
		}),
		Entry("`tcp` and `tcp`", testCase{
			one:      core_mesh.ProtocolTCP,
			another:  core_mesh.ProtocolTCP,
			expected: core_mesh.ProtocolTCP,
		}),
		Entry("`http2` and `http2`", testCase{
			one:      core_mesh.ProtocolHTTP2,
			another:  core_mesh.ProtocolHTTP2,
			expected: core_mesh.ProtocolHTTP2,
		}),
		Entry("`http2` and `http`", testCase{
			one:      core_mesh.ProtocolHTTP2,
			another:  core_mesh.ProtocolHTTP,
			expected: core_mesh.ProtocolTCP,
		}),
		Entry("`http2` and `tcp`", testCase{
			one:      core_mesh.ProtocolHTTP2,
			another:  core_mesh.ProtocolTCP,
			expected: core_mesh.ProtocolTCP,
		}),
		Entry("`grpc` and `grpc`", testCase{
			one:      core_mesh.ProtocolGRPC,
			another:  core_mesh.ProtocolGRPC,
			expected: core_mesh.ProtocolGRPC,
		}),
		Entry("`grpc` and `http`", testCase{
			one:      core_mesh.ProtocolGRPC,
			another:  core_mesh.ProtocolHTTP,
			expected: core_mesh.ProtocolTCP,
		}),
		Entry("`grpc` and `http2`", testCase{
			one:      core_mesh.ProtocolGRPC,
			another:  core_mesh.ProtocolHTTP2,
			expected: core_mesh.ProtocolHTTP2,
		}),
		Entry("`grpc` and `tcp`", testCase{
			one:      core_mesh.ProtocolGRPC,
			another:  core_mesh.ProtocolTCP,
			expected: core_mesh.ProtocolTCP,
		}),
		Entry("`kafka` and `tcp`", testCase{
			one:      core_mesh.ProtocolKafka,
			another:  core_mesh.ProtocolTCP,
			expected: core_mesh.ProtocolTCP,
		}),
	)
})
