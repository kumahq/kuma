package cmd

import (
	"context"
	"fmt"
	"net"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	grpchealth "google.golang.org/grpc/health/grpc_health_v1"

	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/test/server/grpc/api"
)

var grpcServerLog = grpcLog.WithName("server")

type grpcServer struct {
	api.UnimplementedGreeterServer
	grpchealth.UnimplementedHealthServer
	id string
}

func (g *grpcServer) Check(ctx context.Context, request *grpchealth.HealthCheckRequest) (*grpchealth.HealthCheckResponse, error) {
	return &grpchealth.HealthCheckResponse{Status: grpchealth.HealthCheckResponse_SERVING}, nil
}

func (g *grpcServer) Watch(request *grpchealth.HealthCheckRequest, server grpchealth.Health_WatchServer) error {
	return server.Send(&grpchealth.HealthCheckResponse{Status: grpchealth.HealthCheckResponse_SERVING})
}

func (g *grpcServer) SayHello(ctx context.Context, request *api.HelloRequest) (*api.HelloReply, error) {
	grpcServerLog.Info("received request", "name", request.GetName())
	return &api.HelloReply{Message: "Hello " + request.GetName() + " from " + g.id}, nil
}

func (g *grpcServer) SayHellos(server api.Greeter_SayHellosServer) error {
	for {
		request, err := server.Recv()
		if err != nil {
			return err
		}

		grpcServerLog.Info("received request", "name", request.GetName())

		err = server.Send(&api.HelloReply{
			Message: "Hello " + request.GetName() + " from " + g.id,
		})
		if err != nil {
			grpcServerLog.Error(err, "send failed")
			return err
		}
	}
}

func newGRPCServerCmd() *cobra.Command {
	args := struct {
		port uint32
	}{}
	cmd := &cobra.Command{
		Use:   "server",
		Short: "Run GRPC Test Server",
		Long:  "Run GRPC Test Server.",
		Example: `
# start a GRPC server on port 8080 that responds with "Hello ${request.GetName()} from ${server.id}"
test-server grpc server --port 8080
`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ln, err := net.Listen("tcp", fmt.Sprintf(":%d", args.port))
			if err != nil {
				return err
			}

			s := grpc.NewServer()
			api.RegisterGreeterServer(s, &grpcServer{
				id: core.NewUUID(),
			})
			grpchealth.RegisterHealthServer(s, &grpcServer{
				id: core.NewUUID(),
			})

			grpcServerLog.Info("server is listening", "address", ln.Addr())
			if err := s.Serve(ln); err != nil {
				return err
			}
			return nil
		},
	}
	cmd.PersistentFlags().Uint32Var(&args.port, "port", 8080, "port server is listening on")
	return cmd
}
