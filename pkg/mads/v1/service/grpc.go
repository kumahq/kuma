package service

import (
	"google.golang.org/grpc"

	observability_v1 "github.com/kumahq/kuma/api/observability/v1"
)

func (s *service) RegisterWithGrpcServer(server *grpc.Server) {
	observability_v1.RegisterMonitoringAssignmentDiscoveryServiceServer(server, s.server)
}
