package service

import (
	observability_v1alpha1 "github.com/kumahq/kuma/api/observability/v1alpha1"

	"google.golang.org/grpc"
)

func (s *service) RegisterWithGrpcServer(server *grpc.Server) {
	observability_v1alpha1.RegisterMonitoringAssignmentDiscoveryServiceServer(server, s.server)
}
