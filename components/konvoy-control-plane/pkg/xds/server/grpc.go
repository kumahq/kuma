package server

import (
	"context"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"

	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v2"
	xds "github.com/envoyproxy/go-control-plane/pkg/server"
)

const grpcMaxConcurrentStreams = 1000000

func RunGrpcServer(ctx context.Context, server xds.Server, port int) {
	var grpcOptions []grpc.ServerOption
	grpcOptions = append(grpcOptions, grpc.MaxConcurrentStreams(grpcMaxConcurrentStreams))
	grpcServer := grpc.NewServer(grpcOptions...)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatal(err)
	}

	// register services
	discovery.RegisterAggregatedDiscoveryServiceServer(grpcServer, server)
	v2.RegisterEndpointDiscoveryServiceServer(grpcServer, server)
	v2.RegisterClusterDiscoveryServiceServer(grpcServer, server)
	v2.RegisterRouteDiscoveryServiceServer(grpcServer, server)
	v2.RegisterListenerDiscoveryServiceServer(grpcServer, server)
	discovery.RegisterSecretDiscoveryServiceServer(grpcServer, server)

	go func() {
		if err = grpcServer.Serve(lis); err != nil {
			log.Println(err)
		}
	}()

	log.Printf("xDS gRPC server listening on %d\n", port)

	<-ctx.Done()
	grpcServer.GracefulStop()
}
