package middleware

import (
	"context"
	"sync/atomic"

	"google.golang.org/grpc"
)

func StreamIDStreamInterceptor(streamCounter *int64) grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		return handler(srv, &streamIDStream{ServerStream: ss, streamID: atomic.AddInt64(streamCounter, 1)})
	}
}

type streamIDStream struct {
	streamID int64
	grpc.ServerStream
}

func (s *streamIDStream) Context() context.Context {
	return WithStreamID(s.ServerStream.Context(), s.streamID)
}

func StreamIDUnaryInterceptor(streamCounter *int64) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		return handler(WithStreamID(ctx, atomic.AddInt64(streamCounter, 1)), req)
	}
}

func WithStreamID(ctx context.Context, streamID int64) context.Context {
	return context.WithValue(ctx, streamIDCtx{}, streamID)
}

func StreamIDFromCtx(ctx context.Context) (int64, bool) {
	value, ok := ctx.Value(streamIDCtx{}).(int64)
	return value, ok
}

type streamIDCtx struct{}
