package service

import (
	observability_v1 "github.com/kumahq/kuma/api/observability/v1"

	"google.golang.org/grpc"
)

func (s *service) RegisterWithGrpcServer(server *grpc.Server) {
	observability_v1.RegisterMonitoringAssignmentDiscoveryServiceServer(server, s.server)
}
