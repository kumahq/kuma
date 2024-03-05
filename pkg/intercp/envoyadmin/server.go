package envoyadmin

import (
	"context"

	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/envoy/admin"
)

var serverLog = core.Log.WithName("intercp").WithName("catalog").WithName("server")

type server struct {
	adminClient admin.EnvoyAdminClient
	resManager  manager.ReadOnlyResourceManager
	mesh_proto.UnimplementedInterCPEnvoyAdminForwardServiceServer
}

var _ mesh_proto.InterCPEnvoyAdminForwardServiceServer = &server{}

func NewServer(adminClient admin.EnvoyAdminClient, resManager manager.ReadOnlyResourceManager) mesh_proto.InterCPEnvoyAdminForwardServiceServer {
	return &server{
		adminClient: adminClient,
		resManager:  resManager,
	}
}

func (s *server) XDSConfig(ctx context.Context, req *mesh_proto.XDSConfigRequest) (*mesh_proto.XDSConfigResponse, error) {
	ctx = extractTenantMetadata(ctx)
	serverLog.V(1).Info("received forwarded request", "operation", "XDSConfig", "request", req)
	resWithAddr, err := s.resWithAddress(ctx, req.ResourceType, req.ResourceName, req.ResourceMesh)
	if err != nil {
		return nil, err
	}
	configDump, err := s.adminClient.ConfigDump(ctx, resWithAddr)
	if err != nil {
		if errors.Is(err, &admin.KDSTransportError{}) {
			return &mesh_proto.XDSConfigResponse{
				Result: &mesh_proto.XDSConfigResponse_Error{
					Error: err.Error(),
				},
			}, nil
		}
		return nil, err
	}
	return &mesh_proto.XDSConfigResponse{
		Result: &mesh_proto.XDSConfigResponse_Config{
			Config: configDump,
		},
	}, nil
}

func (s *server) Stats(ctx context.Context, req *mesh_proto.StatsRequest) (*mesh_proto.StatsResponse, error) {
	ctx = extractTenantMetadata(ctx)
	serverLog.V(1).Info("received forwarded request", "operation", "Stats", "request", req)
	resWithAddr, err := s.resWithAddress(ctx, req.ResourceType, req.ResourceName, req.ResourceMesh)
	if err != nil {
		return nil, err
	}
	stats, err := s.adminClient.Stats(ctx, resWithAddr)
	if err != nil {
		if errors.Is(err, &admin.KDSTransportError{}) {
			return &mesh_proto.StatsResponse{
				Result: &mesh_proto.StatsResponse_Error{
					Error: err.Error(),
				},
			}, nil
		}
		return nil, err
	}
	return &mesh_proto.StatsResponse{
		Result: &mesh_proto.StatsResponse_Stats{
			Stats: stats,
		},
	}, nil
}

func (s *server) Clusters(ctx context.Context, req *mesh_proto.ClustersRequest) (*mesh_proto.ClustersResponse, error) {
	ctx = extractTenantMetadata(ctx)
	serverLog.V(1).Info("received forwarded request", "operation", "Clusters", "request", req)
	resWithAddr, err := s.resWithAddress(ctx, req.ResourceType, req.ResourceName, req.ResourceMesh)
	if err != nil {
		return nil, err
	}
	clusters, err := s.adminClient.Clusters(ctx, resWithAddr)
	if err != nil {
		if errors.Is(err, &admin.KDSTransportError{}) {
			return &mesh_proto.ClustersResponse{
				Result: &mesh_proto.ClustersResponse_Error{
					Error: err.Error(),
				},
			}, nil
		}
		return nil, err
	}
	return &mesh_proto.ClustersResponse{
		Result: &mesh_proto.ClustersResponse_Clusters{
			Clusters: clusters,
		},
	}, nil
}

func (s *server) resWithAddress(ctx context.Context, typ, name, mesh string) (model.ResourceWithAddress, error) {
	obj, err := registry.Global().NewObject(model.ResourceType(typ))
	if err != nil {
		return nil, err
	}
	if err := s.resManager.Get(ctx, obj, core_store.GetByKey(name, mesh)); err != nil {
		return nil, err
	}
	resourceWithAddr, ok := obj.(model.ResourceWithAddress)
	if !ok {
		return nil, errors.New("invalid resource type")
	}
	return resourceWithAddr, nil
}
