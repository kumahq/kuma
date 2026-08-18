package service

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	core_manager "github.com/kumahq/kuma/v3/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/core/resources/registry"
	core_store "github.com/kumahq/kuma/v3/pkg/core/resources/store"
	"github.com/kumahq/kuma/v3/pkg/envoy/admin"
)

type EnvoyAdminProcessor interface {
	StartProcessingXDSConfigs(stream mesh_proto.GlobalKDSService_StreamXDSConfigsClient, errorCh chan error)
	StartProcessingStats(stream mesh_proto.GlobalKDSService_StreamStatsClient, errorCh chan error)
	StartProcessingClusters(stream mesh_proto.GlobalKDSService_StreamClustersClient, errorCh chan error)
}

type envoyAdminProcessor struct {
	resManager  core_manager.ReadOnlyResourceManager
	adminClient admin.EnvoyAdminClient
	maxMsgSize  int
}

var _ EnvoyAdminProcessor = &envoyAdminProcessor{}

func NewEnvoyAdminProcessor(
	resManager core_manager.ReadOnlyResourceManager,
	adminClient admin.EnvoyAdminClient,
	maxMsgSize uint32,
) EnvoyAdminProcessor {
	return &envoyAdminProcessor{
		resManager:  resManager,
		adminClient: adminClient,
		maxMsgSize:  int(maxMsgSize),
	}
}

// tooLargeError explains that a response can't be delivered over KDS. Sending it
// would exceed the message size limit, which aborts the whole RPC stream and
// takes down every other KDS stream multiplexed on the same connection, so we
// answer with an error instead. The receiver enforces the limit on the
// uncompressed message, which is what proto.Size reports.
func tooLargeError(rpcName string, size int, maxMsgSize int) string {
	return fmt.Sprintf(
		"%s response is %d bytes which exceeds the maximum KDS message size of %d bytes, increase maxMsgSize on both the zone CP and the global CP",
		rpcName, size, maxMsgSize,
	)
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
			config, err := s.executeAdminFn(stream.Context(), req.ResourceType, req.ResourceName, req.ResourceMesh, func(ctx context.Context, proxy core_model.ResourceWithAddress) ([]byte, error) {
				return s.adminClient.ConfigDump(ctx, proxy, req.IncludeEds)
			})

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
			if size := proto.Size(resp); size > s.maxMsgSize {
				resp.Result = &mesh_proto.XDSConfigResponse_Error{
					Error: tooLargeError(ConfigDumpRPC, size, s.maxMsgSize),
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
			stats, err := s.executeAdminFn(stream.Context(), req.ResourceType, req.ResourceName, req.ResourceMesh, func(ctx context.Context, proxy core_model.ResourceWithAddress) ([]byte, error) {
				return s.adminClient.Stats(ctx, proxy, req.Format, req.UsedOnly)
			})

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
			if size := proto.Size(resp); size > s.maxMsgSize {
				resp.Result = &mesh_proto.StatsResponse_Error{
					Error: tooLargeError(StatsRPC, size, s.maxMsgSize),
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
			clusters, err := s.executeAdminFn(stream.Context(), req.ResourceType, req.ResourceName, req.ResourceMesh, func(ctx context.Context, proxy core_model.ResourceWithAddress) ([]byte, error) {
				return s.adminClient.Clusters(ctx, proxy, req.Format)
			})

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
			if size := proto.Size(resp); size > s.maxMsgSize {
				resp.Result = &mesh_proto.ClustersResponse_Error{
					Error: tooLargeError(ClustersRPC, size, s.maxMsgSize),
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
	adminFn func(ctx context.Context, proxy core_model.ResourceWithAddress) ([]byte, error),
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
