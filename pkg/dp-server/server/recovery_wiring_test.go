package server

import (
	"context"
	"net"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/emptypb"

	dp_server "github.com/kumahq/kuma/v3/pkg/config/dp-server"
	"github.com/kumahq/kuma/v3/pkg/metrics"
)

// panicServiceDesc registers a unary and a stream method that always panic, so
// a call routed through the real DpServer.GrpcServer() exercises the recovery
// interceptors wired in NewDpServer. It covers the wiring itself: dropping the
// ChainUnary/StreamInterceptor options makes these expectations fail.
var panicServiceDesc = grpc.ServiceDesc{
	ServiceName: "kuma.test.PanicService",
	HandlerType: (*any)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Unary",
			Handler: func(srv any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error) {
				in := new(emptypb.Empty)
				if err := dec(in); err != nil {
					return nil, err
				}
				handler := func(context.Context, any) (any, error) {
					panic("boom")
				}
				if interceptor == nil {
					return handler(ctx, in)
				}
				info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/kuma.test.PanicService/Unary"}
				return interceptor(ctx, in, info, handler)
			},
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName: "Stream",
			Handler: func(any, grpc.ServerStream) error {
				panic("boom")
			},
			ServerStreams: true,
			ClientStreams: true,
		},
	},
}

// panicService is the registered implementation. The handlers in
// panicServiceDesc do not use it, but RegisterService wants a non-nil value.
type panicService struct{}

var _ = Describe("DpServer gRPC recovery wiring", func() {
	var lis *bufconn.Listener
	var conn *grpc.ClientConn
	var grpcServer *grpc.Server

	BeforeEach(func() {
		m, err := metrics.NewMetrics("")
		Expect(err).ToNot(HaveOccurred())
		dpServer, err := NewDpServer(
			*dp_server.DefaultDpServerConfig(),
			m,
			func(http.ResponseWriter, *http.Request) bool { return true },
		)
		Expect(err).ToNot(HaveOccurred())

		grpcServer = dpServer.GrpcServer()
		grpcServer.RegisterService(&panicServiceDesc, &panicService{})

		lis = bufconn.Listen(1024 * 1024)
		go func() {
			defer GinkgoRecover()
			_ = grpcServer.Serve(lis)
		}()

		conn, err = grpc.NewClient(
			"passthrough:///bufnet",
			grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
				return lis.DialContext(ctx)
			}),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		if conn != nil {
			Expect(conn.Close()).To(Succeed())
		}
		if grpcServer != nil {
			grpcServer.Stop()
		}
	})

	It("returns Internal when a unary handler panics", func() {
		err := conn.Invoke(context.Background(), "/kuma.test.PanicService/Unary", &emptypb.Empty{}, &emptypb.Empty{})
		Expect(err).To(HaveOccurred())
		Expect(status.Code(err)).To(Equal(codes.Internal))
		Expect(status.Convert(err).Message()).To(Equal("internal server error"))
	})

	It("returns Internal when a stream handler panics", func() {
		stream, err := conn.NewStream(context.Background(), &panicServiceDesc.Streams[0], "/kuma.test.PanicService/Stream")
		Expect(err).ToNot(HaveOccurred())

		err = stream.RecvMsg(&emptypb.Empty{})
		Expect(err).To(HaveOccurred())
		Expect(status.Code(err)).To(Equal(codes.Internal))
		Expect(status.Convert(err).Message()).To(Equal("internal server error"))
	})
})
