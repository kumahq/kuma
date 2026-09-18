package auth

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

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
