package auth

import (
	"context"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/kumahq/kuma/v3/pkg/config/multizone"
	"github.com/kumahq/kuma/v3/pkg/kds/util"
)

func StreamServerInterceptor(authenticator Authenticator) grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, _ *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if err := authenticate(ss.Context(), authenticator); err != nil {
			return err
		}
		return handler(srv, ss)
	}
}

func UnaryServerInterceptor(authenticator Authenticator) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if err := authenticate(ctx, authenticator); err != nil {
			return nil, err
		}
		return handler(ctx, req)
	}
}

func authenticate(ctx context.Context, authenticator Authenticator) error {
	zone, err := util.ClientIDFromIncomingCtx(ctx)
	if err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}
	md, _ := metadata.FromIncomingContext(ctx)
	if err := authenticator.Authenticate(ctx, md, zone); err != nil {
		if status.Code(err) != codes.Unknown {
			return err
		}
		return status.Error(codes.Unauthenticated, err.Error())
	}
	return nil
}

// ServerInterceptors returns the interceptors authenticating Zone CPs, empty when
// authentication is disabled. The authenticator is the one registered on the KDS
// context, nil when nothing built one for authType. Kuma builds it for the
// zoneToken type, distributions add types of their own, so a type left without an
// authenticator is a configuration error rather than an open control plane.
func ServerInterceptors(authType multizone.KDSAuthType, authenticator Authenticator) ([]grpc.StreamServerInterceptor, []grpc.UnaryServerInterceptor, error) {
	switch {
	case authenticator != nil:
		return []grpc.StreamServerInterceptor{StreamServerInterceptor(authenticator)},
			[]grpc.UnaryServerInterceptor{UnaryServerInterceptor(authenticator)},
			nil
	case authType == multizone.KDSAuthNone:
		return nil, nil, nil
	default:
		return nil, nil, errors.Errorf("multizone.global.kds.auth.type %q is not supported by this control plane", authType)
	}
}
