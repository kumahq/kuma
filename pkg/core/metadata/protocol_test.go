package metadata_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	core_meta "github.com/kumahq/kuma/pkg/core/metadata"
)

var _ = Describe("GetCommoncore_meta.Protocol()", func() {
	type testCase struct {
		one      core_meta.Protocol
		another  core_meta.Protocol
		expected core_meta.Protocol
	}

	DescribeTable("should correctly determine common protocol",
		func(given testCase) {
			// when
			actual := core_meta.GetCommonProtocol(given.one, given.another)
			// then
			Expect(actual).To(Equal(given.expected))
		},
		Entry("`unknown` and `unknown`", testCase{
			one:      core_meta.ProtocolUnknown,
			another:  core_meta.ProtocolUnknown,
			expected: core_meta.ProtocolUnknown,
		}),
		Entry("`unknown` and `http`", testCase{
			one:      core_meta.ProtocolUnknown,
			another:  core_meta.ProtocolHTTP,
			expected: core_meta.ProtocolUnknown,
		}),
		Entry("`http` and `unknown`", testCase{
			one:      core_meta.ProtocolHTTP,
			another:  core_meta.ProtocolUnknown,
			expected: core_meta.ProtocolUnknown,
		}),
		Entry("`unknown` and `tcp`", testCase{
			one:      core_meta.ProtocolUnknown,
			another:  core_meta.ProtocolTCP,
			expected: core_meta.ProtocolUnknown,
		}),
		Entry("`tcp` and `unknown`", testCase{
			one:      core_meta.ProtocolTCP,
			another:  core_meta.ProtocolUnknown,
			expected: core_meta.ProtocolUnknown,
		}),
		Entry("`http` and `tcp`", testCase{
			one:      core_meta.ProtocolHTTP,
			another:  core_meta.ProtocolTCP,
			expected: core_meta.ProtocolTCP,
		}),
		Entry("`tcp` and `http`", testCase{
			one:      core_meta.ProtocolTCP,
			another:  core_meta.ProtocolHTTP,
			expected: core_meta.ProtocolTCP,
		}),
		Entry("`http` and `http`", testCase{
			one:      core_meta.ProtocolHTTP,
			another:  core_meta.ProtocolHTTP,
			expected: core_meta.ProtocolHTTP,
		}),
		Entry("`tcp` and `tcp`", testCase{
			one:      core_meta.ProtocolTCP,
			another:  core_meta.ProtocolTCP,
			expected: core_meta.ProtocolTCP,
		}),
		Entry("`http2` and `http2`", testCase{
			one:      core_meta.ProtocolHTTP2,
			another:  core_meta.ProtocolHTTP2,
			expected: core_meta.ProtocolHTTP2,
		}),
		Entry("`http2` and `http`", testCase{
			one:      core_meta.ProtocolHTTP2,
			another:  core_meta.ProtocolHTTP,
			expected: core_meta.ProtocolTCP,
		}),
		Entry("`http2` and `tcp`", testCase{
			one:      core_meta.ProtocolHTTP2,
			another:  core_meta.ProtocolTCP,
			expected: core_meta.ProtocolTCP,
		}),
		Entry("`grpc` and `grpc`", testCase{
			one:      core_meta.ProtocolGRPC,
			another:  core_meta.ProtocolGRPC,
			expected: core_meta.ProtocolGRPC,
		}),
		Entry("`grpc` and `http`", testCase{
			one:      core_meta.ProtocolGRPC,
			another:  core_meta.ProtocolHTTP,
			expected: core_meta.ProtocolTCP,
		}),
		Entry("`grpc` and `http2`", testCase{
			one:      core_meta.ProtocolGRPC,
			another:  core_meta.ProtocolHTTP2,
			expected: core_meta.ProtocolHTTP2,
		}),
		Entry("`grpc` and `tcp`", testCase{
			one:      core_meta.ProtocolGRPC,
			another:  core_meta.ProtocolTCP,
			expected: core_meta.ProtocolTCP,
		}),
		Entry("`kafka` and `tcp`", testCase{
			one:      core_meta.ProtocolKafka,
			another:  core_meta.ProtocolTCP,
			expected: core_meta.ProtocolTCP,
		}),
	)
})
