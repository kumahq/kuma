package xds

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	core_meta "github.com/kumahq/kuma/pkg/core/metadata"
	"github.com/kumahq/kuma/pkg/plugins/policies/meshhealthcheck/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/util/pointer"
)

var _ = Describe("MeshHealthCheck configurer", func() {
	type testCase struct {
		protocol       core_meta.Protocol
		grpc           *v1alpha1.GrpcHealthCheck
		http           *v1alpha1.HttpHealthCheck
		tcp            *v1alpha1.TcpHealthCheck
		expectedHcType HCProtocol
	}

	DescribeTable("should select the correct protocol",
		func(given testCase) {
			hcType := selectHealthCheckType(given.protocol, given.tcp, given.http, given.grpc)
			Expect(hcType).To(Equal(given.expectedHcType))
		},
		// no health check
		Entry("no HC defined", testCase{
			protocol:       core_meta.ProtocolTCP,
			grpc:           nil,
			http:           nil,
			tcp:            nil,
			expectedHcType: HCNone,
		}),
		Entry("HC defined for grpc when service protocol is tcp", testCase{
			protocol:       core_meta.ProtocolTCP,
			grpc:           &v1alpha1.GrpcHealthCheck{},
			http:           nil,
			tcp:            nil,
			expectedHcType: HCNone,
		}),
		Entry("HC defined for http when service protocol is tcp", testCase{
			protocol:       core_meta.ProtocolTCP,
			grpc:           nil,
			http:           &v1alpha1.HttpHealthCheck{},
			tcp:            nil,
			expectedHcType: HCNone,
		}),
		Entry("HC defined for grpc when service protocol is http", testCase{
			protocol:       core_meta.ProtocolTCP,
			grpc:           &v1alpha1.GrpcHealthCheck{},
			http:           nil,
			tcp:            nil,
			expectedHcType: HCNone,
		}),
		Entry("HC defined for tcp when service protocol is http", testCase{
			protocol:       core_meta.ProtocolHTTP,
			grpc:           nil,
			http:           nil,
			tcp:            &v1alpha1.TcpHealthCheck{},
			expectedHcType: HCNone,
		}),
		Entry("HC defined for tcp when service protocol is grpc", testCase{
			protocol:       core_meta.ProtocolGRPC,
			grpc:           nil,
			http:           nil,
			tcp:            &v1alpha1.TcpHealthCheck{},
			expectedHcType: HCNone,
		}),
		Entry("HC defined for http when service protocol is grpc", testCase{
			protocol:       core_meta.ProtocolGRPC,
			grpc:           nil,
			http:           &v1alpha1.HttpHealthCheck{},
			tcp:            nil,
			expectedHcType: HCNone,
		}),

		// matching health check
		Entry("HC defined for grpc when service protocol is grpc", testCase{
			protocol:       core_meta.ProtocolGRPC,
			grpc:           &v1alpha1.GrpcHealthCheck{},
			http:           nil,
			tcp:            nil,
			expectedHcType: HCProtocolGRPC,
		}),
		Entry("HC defined for http when service protocol is http", testCase{
			protocol:       core_meta.ProtocolHTTP,
			grpc:           nil,
			http:           &v1alpha1.HttpHealthCheck{},
			tcp:            nil,
			expectedHcType: HCProtocolHTTP,
		}),
		Entry("HC defined for http when service protocol is http2", testCase{
			protocol:       core_meta.ProtocolHTTP2,
			grpc:           nil,
			http:           &v1alpha1.HttpHealthCheck{},
			tcp:            nil,
			expectedHcType: HCProtocolHTTP,
		}),
		Entry("HC defined for tcp when service protocol is tcp", testCase{
			protocol:       core_meta.ProtocolTCP,
			grpc:           nil,
			http:           nil,
			tcp:            &v1alpha1.TcpHealthCheck{},
			expectedHcType: HCProtocolTCP,
		}),

		// disabling matching health check
		Entry("HC defined for grpc when service protocol is grpc", testCase{
			protocol:       core_meta.ProtocolGRPC,
			grpc:           &v1alpha1.GrpcHealthCheck{Disabled: pointer.To(true)},
			http:           nil,
			tcp:            nil,
			expectedHcType: HCNone,
		}),
		Entry("HC defined for http when service protocol is http", testCase{
			protocol:       core_meta.ProtocolHTTP,
			grpc:           nil,
			http:           &v1alpha1.HttpHealthCheck{Disabled: pointer.To(true)},
			tcp:            nil,
			expectedHcType: HCNone,
		}),
		Entry("HC defined for http when service protocol is http2", testCase{
			protocol:       core_meta.ProtocolHTTP2,
			grpc:           nil,
			http:           &v1alpha1.HttpHealthCheck{Disabled: pointer.To(true)},
			tcp:            nil,
			expectedHcType: HCNone,
		}),
		Entry("HC defined for tcp when service protocol is tcp", testCase{
			protocol:       core_meta.ProtocolTCP,
			grpc:           nil,
			http:           nil,
			tcp:            &v1alpha1.TcpHealthCheck{Disabled: pointer.To(true)},
			expectedHcType: HCNone,
		}),

		// fallback HTTP to TCP
		Entry("HC defined for tcp and disabled for http when service protocol is http", testCase{
			protocol:       core_meta.ProtocolHTTP,
			grpc:           nil,
			http:           &v1alpha1.HttpHealthCheck{Disabled: pointer.To(true)},
			tcp:            &v1alpha1.TcpHealthCheck{},
			expectedHcType: HCProtocolTCP,
		}),

		// fallback GRPC to TCP
		Entry("HC defined for tcp and disabled for grpc when service protocol is grpc", testCase{
			protocol:       core_meta.ProtocolGRPC,
			grpc:           &v1alpha1.GrpcHealthCheck{Disabled: pointer.To(true)},
			http:           nil,
			tcp:            &v1alpha1.TcpHealthCheck{},
			expectedHcType: HCProtocolTCP,
		}),

		// all protocols defined
		Entry("HC defined for all protocols when service protocol is grpc", testCase{
			protocol:       core_meta.ProtocolGRPC,
			grpc:           &v1alpha1.GrpcHealthCheck{},
			http:           &v1alpha1.HttpHealthCheck{},
			tcp:            &v1alpha1.TcpHealthCheck{},
			expectedHcType: HCProtocolGRPC,
		}),

		// all protocols disabled
		Entry("HC disabled for all protocols when service protocol is http", testCase{
			protocol:       core_meta.ProtocolHTTP,
			grpc:           &v1alpha1.GrpcHealthCheck{Disabled: pointer.To(true)},
			http:           &v1alpha1.HttpHealthCheck{Disabled: pointer.To(true)},
			tcp:            &v1alpha1.TcpHealthCheck{Disabled: pointer.To(true)},
			expectedHcType: HCNone,
		}),
	)
})
