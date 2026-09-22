package middleware_test

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"

	"github.com/kumahq/kuma/v3/pkg/kds/middleware"
)

var _ = Describe("ClientCert interceptor", func() {
	type testCase struct {
		requireClientCert bool
		authInfo          credentials.AuthInfo
		clientID          string
		expectedCode      codes.Code
	}

	tlsInfo := func(leaf *x509.Certificate) credentials.TLSInfo {
		state := tls.ConnectionState{}
		if leaf != nil {
			state.VerifiedChains = [][]*x509.Certificate{{leaf}}
		}
		return credentials.TLSInfo{State: state}
	}

	call := func(given testCase) error {
		ctx := peer.NewContext(context.Background(), &peer.Peer{AuthInfo: given.authInfo})
		if given.clientID != "" {
			ctx = metadata.NewIncomingContext(ctx, metadata.Pairs("client-id", given.clientID))
		}
		interceptor := middleware.ClientCertUnaryInterceptor(given.requireClientCert)
		_, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{}, func(context.Context, any) (any, error) {
			return nil, nil
		})
		return err
	}

	DescribeTable("should verify the client certificate against the zone",
		func(given testCase) {
			err := call(given)
			if given.expectedCode == codes.OK {
				Expect(err).ToNot(HaveOccurred())
			} else {
				Expect(status.Code(err)).To(Equal(given.expectedCode))
			}
		},
		Entry("cert with zone in DNS SAN", testCase{
			requireClientCert: true,
			authInfo:          tlsInfo(&x509.Certificate{DNSNames: []string{"zone-1"}}),
			clientID:          "zone-1",
			expectedCode:      codes.OK,
		}),
		Entry("cert with zone in CN", testCase{
			requireClientCert: true,
			authInfo:          tlsInfo(&x509.Certificate{Subject: pkix.Name{CommonName: "zone-1"}}),
			clientID:          "zone-1",
			expectedCode:      codes.OK,
		}),
		Entry("cert issued for another zone", testCase{
			requireClientCert: true,
			authInfo:          tlsInfo(&x509.Certificate{DNSNames: []string{"zone-2"}}),
			clientID:          "zone-1",
			expectedCode:      codes.PermissionDenied,
		}),
		Entry("cert issued for another zone when cert is optional", testCase{
			requireClientCert: false,
			authInfo:          tlsInfo(&x509.Certificate{DNSNames: []string{"zone-2"}}),
			clientID:          "zone-1",
			expectedCode:      codes.PermissionDenied,
		}),
		Entry("cert without client-id", testCase{
			requireClientCert: true,
			authInfo:          tlsInfo(&x509.Certificate{DNSNames: []string{"zone-1"}}),
			expectedCode:      codes.InvalidArgument,
		}),
		Entry("no cert when cert is required", testCase{
			requireClientCert: true,
			authInfo:          tlsInfo(nil),
			clientID:          "zone-1",
			expectedCode:      codes.Unauthenticated,
		}),
		Entry("no cert when cert is optional", testCase{
			requireClientCert: false,
			authInfo:          tlsInfo(nil),
			clientID:          "zone-1",
			expectedCode:      codes.OK,
		}),
		Entry("plaintext connection when cert is required", testCase{
			requireClientCert: true,
			clientID:          "zone-1",
			expectedCode:      codes.Unauthenticated,
		}),
	)

	It("should reject a stream with a cert issued for another zone", func() {
		ctx := peer.NewContext(context.Background(), &peer.Peer{AuthInfo: tlsInfo(&x509.Certificate{DNSNames: []string{"zone-2"}})})
		ctx = metadata.NewIncomingContext(ctx, metadata.Pairs("client-id", "zone-1"))
		interceptor := middleware.ClientCertStreamInterceptor(true)
		called := false

		err := interceptor(nil, &fakeServerStream{ctx: ctx}, &grpc.StreamServerInfo{}, func(any, grpc.ServerStream) error {
			called = true
			return nil
		})

		Expect(status.Code(err)).To(Equal(codes.PermissionDenied))
		Expect(called).To(BeFalse())
	})
})

type fakeServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (f *fakeServerStream) Context() context.Context {
	return f.ctx
}
