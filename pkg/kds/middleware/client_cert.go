package middleware

import (
	"context"
	"slices"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"

	"github.com/kumahq/kuma/v3/pkg/kds/util"
)

func ClientCertStreamInterceptor(requireClientCert bool) grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if err := verifyClientCert(ss.Context(), requireClientCert); err != nil {
			return err
		}
		return handler(srv, ss)
	}
}

func ClientCertUnaryInterceptor(requireClientCert bool) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if err := verifyClientCert(ctx, requireClientCert); err != nil {
			return nil, err
		}
		return handler(ctx, req)
	}
}

// verifyClientCert binds the client-id declared in metadata to the verified
// client certificate, so a zone holding a valid cert can't act as another zone.
func verifyClientCert(ctx context.Context, requireClientCert bool) error {
	p, ok := peer.FromContext(ctx)
	if !ok {
		return status.Error(codes.Unauthenticated, "peer info is not available")
	}
	tlsInfo, ok := p.AuthInfo.(credentials.TLSInfo)
	if !ok || len(tlsInfo.State.VerifiedChains) == 0 || len(tlsInfo.State.VerifiedChains[0]) == 0 {
		if requireClientCert {
			return status.Error(codes.Unauthenticated, "client certificate is required")
		}
		return nil
	}
	zone, err := util.ClientIDFromIncomingCtx(ctx)
	if err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}
	leaf := tlsInfo.State.VerifiedChains[0][0]
	if !slices.Contains(leaf.DNSNames, zone) && leaf.Subject.CommonName != zone {
		return status.Errorf(codes.PermissionDenied, "client certificate is not issued for zone %q", zone)
	}
	return nil
}
