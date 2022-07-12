package service

import (
	"context"
	"time"

	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
)

type EnvoyAdminProcessor interface {
	StartProcessingXDSConfigs(stream mesh_proto.GlobalKDSService_StreamXDSConfigsClient, errorCh chan error)
	StartProcessingStats(stream mesh_proto.GlobalKDSService_StreamStatsClient, errorCh chan error)
	StartProcessingClusters(stream mesh_proto.GlobalKDSService_StreamClustersClient, errorCh chan error)
}

type EnvoyAdminFn = func(ctx context.Context, proxy core_model.ResourceWithAddress) ([]byte, error)

type envoyAdminProcessor struct {
	resManager core_manager.ReadOnlyResourceManager

	configDumpFn EnvoyAdminFn
	statsFn      EnvoyAdminFn
	clustersFn   EnvoyAdminFn
}

var _ EnvoyAdminProcessor = &envoyAdminProcessor{}

func NewEnvoyAdminProcessor(
	resManager core_manager.ReadOnlyResourceManager,
	configDumpFn EnvoyAdminFn,
	statsFn EnvoyAdminFn,
	clustersFn EnvoyAdminFn,
) EnvoyAdminProcessor {
	return &envoyAdminProcessor{
		resManager:   resManager,
		configDumpFn: configDumpFn,
		statsFn:      statsFn,
		clustersFn:   clustersFn,
	}
}

func (s *envoyAdminProcessor) StartProcessingXDSConfigs(
	stream mesh_proto.GlobalKDSService_StreamXDSConfigsClient,
	errorCh chan error,
) {
	for {
		req, err := stream.Recv()
		if err != nil {
			errorCh <- err
			return
		}
		go func() { // schedule in the background to be able to quickly process more requests
			config, err := s.executeAdminFn(stream.Context(), req.ResourceType, req.ResourceName, req.ResourceMesh, s.configDumpFn)

			resp := &mesh_proto.XDSConfigResponse{
				RequestId: req.RequestId,
			}
			if len(config) > 0 {
				resp.Result = &mesh_proto.XDSConfigResponse_Config{
					Config: config,
				}
			}
			if err != nil { // send the error to the client instead of terminating stream.
				resp.Result = &mesh_proto.XDSConfigResponse_Error{
					Error: err.Error(),
				}
			}
			if err := stream.Send(resp); err != nil {
				errorCh <- err
				return
			}
		}()
	}
}

func (s *envoyAdminProcessor) StartProcessingStats(
	stream mesh_proto.GlobalKDSService_StreamStatsClient,
	errorCh chan error,
) {
	for {
		req, err := stream.Recv()
		if err != nil {
			errorCh <- err
			return
		}
		go func() { // schedule in the background to be able to quickly process more requests
			stats, err := s.executeAdminFn(stream.Context(), req.ResourceType, req.ResourceName, req.ResourceMesh, s.statsFn)

			resp := &mesh_proto.StatsResponse{
				RequestId: req.RequestId,
			}
			if len(stats) > 0 {
				resp.Result = &mesh_proto.StatsResponse_Stats{
					Stats: stats,
				}
			}
			if err != nil { // send the error to the client instead of terminating stream.
				resp.Result = &mesh_proto.StatsResponse_Error{
					Error: err.Error(),
				}
			}
			if err := stream.Send(resp); err != nil {
				errorCh <- err
				return
			}
		}()
	}
}

func (s *envoyAdminProcessor) StartProcessingClusters(
	stream mesh_proto.GlobalKDSService_StreamClustersClient,
	errorCh chan error,
) {
	for {
		req, err := stream.Recv()
		if err != nil {
			errorCh <- err
			return
		}
		go func() { // schedule in the background to be able to quickly process more requests
			clusters, err := s.executeAdminFn(stream.Context(), req.ResourceType, req.ResourceName, req.ResourceMesh, s.clustersFn)

			resp := &mesh_proto.ClustersResponse{
				RequestId: req.RequestId,
			}
			if len(clusters) > 0 {
				resp.Result = &mesh_proto.ClustersResponse_Clusters{
					Clusters: clusters,
				}
			}
			if err != nil { // send the error to the client instead of terminating stream.
				resp.Result = &mesh_proto.ClustersResponse_Error{
					Error: err.Error(),
				}
			}
			if err := stream.Send(resp); err != nil {
				errorCh <- err
				return
			}
		}()
	}
}

func (s *envoyAdminProcessor) executeAdminFn(
	ctx context.Context,
	resType string,
	resName string,
	resMesh string,
	adminFn EnvoyAdminFn,
) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	res, err := registry.Global().NewObject(core_model.ResourceType(resType))
	if err != nil {
		return nil, err
	}
	if err := s.resManager.Get(ctx, res, core_store.GetByKey(resName, resMesh)); err != nil {
		return nil, err
	}

	resWithAddr, ok := res.(core_model.ResourceWithAddress)
	if !ok {
		return nil, errors.Errorf("invalid type %T", resWithAddr)
	}

	return adminFn(ctx, resWithAddr)
}
