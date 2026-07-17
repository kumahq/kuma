package server

import (
	"context"
	"fmt"
	"runtime/debug"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// recoverGRPCHandler turns a recovered panic into a gRPC error so a single
// failing call is aborted instead of taking down the whole process. The panic
// value and stack are logged for debugging.
func recoverGRPCHandler(p any) error {
	log.Error(fmt.Errorf("%v", p), "recovered from panic in gRPC handler", "stack", string(debug.Stack()))
	return status.Error(codes.Internal, "internal server error")
}

func recoveryUnaryInterceptor(
	ctx context.Context,
	req any,
	_ *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (any, error) {
	var resp any
	var err error
	func() {
		defer func() {
			if p := recover(); p != nil {
				err = recoverGRPCHandler(p)
			}
		}()
		resp, err = handler(ctx, req)
	}()
	return resp, err
}

func recoveryStreamInterceptor(
	srv any,
	ss grpc.ServerStream,
	_ *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) (err error) {
	defer func() {
		if p := recover(); p != nil {
			err = recoverGRPCHandler(p)
		}
	}()
	return handler(srv, ss)
}
