package server

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ = Describe("gRPC recovery interceptors", func() {
	It("recovers from a panic in a unary handler", func() {
		handler := func(context.Context, any) (any, error) {
			panic("boom")
		}

		var err error
		Expect(func() {
			_, err = recoveryUnaryInterceptor(context.Background(), nil, &grpc.UnaryServerInfo{}, handler)
		}).ToNot(Panic())
		Expect(err).To(HaveOccurred())
		Expect(status.Code(err)).To(Equal(codes.Internal))
	})

	It("recovers from a panic in a stream handler", func() {
		handler := func(any, grpc.ServerStream) error {
			panic("boom")
		}

		var err error
		Expect(func() {
			err = recoveryStreamInterceptor(nil, nil, &grpc.StreamServerInfo{}, handler)
		}).ToNot(Panic())
		Expect(err).To(HaveOccurred())
		Expect(status.Code(err)).To(Equal(codes.Internal))
	})

	It("passes through when the handler does not panic", func() {
		unary := func(context.Context, any) (any, error) {
			return "ok", nil
		}
		resp, err := recoveryUnaryInterceptor(context.Background(), nil, &grpc.UnaryServerInfo{}, unary)
		Expect(err).ToNot(HaveOccurred())
		Expect(resp).To(Equal("ok"))

		stream := func(any, grpc.ServerStream) error {
			return nil
		}
		Expect(recoveryStreamInterceptor(nil, nil, &grpc.StreamServerInfo{}, stream)).To(Succeed())
	})
})
