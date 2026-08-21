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

// tooLargeError reports that a response can't be delivered over KDS. Receiving a
// message past the size limit makes gRPC abort the whole RPC stream, which takes
// down every other KDS stream multiplexed on the same connection, so the
// processor answers with this error instead.
//
// The limit that actually binds is the global CP's receive limit, which gRPC
// applies to the decompressed message, so proto.Size is the size to check. That
// limit is configured on the global CP and is not advertised over KDS, so this
// check uses the zone's own maxMsgSize as the only locally known proxy for it.
// The two match by default (10MiB on both sides) and are expected to be kept
// aligned; a zone configured below the global's limit rejects responses the
// global would have accepted.
//
// The advice is deliberately limited to reachableBackends. Raising maxMsgSize is
// not suggested: it is the global side that has to be raised, which a zone
// operator may not control, and raising only the zone side lets messages past
// this check that the global CP then rejects, dropping the stream.
func (s *envoyAdminProcessor) tooLargeError(rpcName string, resourceName string, size int) string {
	log.Info("Envoy admin response exceeds the maximum KDS message size, replying with an error instead of sending it",
		"rpc", rpcName,
		"resource", resourceName,
		"size", size,
		"maxMsgSize", s.maxMsgSize,
		"hint", "trim the proxy config with reachableBackends",
	)
	return fmt.Sprintf(
		"the original %s response is %d bytes which exceeds the maximum KDS message size of %d bytes. A response this large usually means the proxy is configured for every service in the mesh, consider trimming its config with reachableBackends",
		rpcName, size, s.maxMsgSize,
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
					Error: s.tooLargeError(ConfigDumpRPC, req.ResourceName, size),
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
					Error: s.tooLargeError(StatsRPC, req.ResourceName, size),
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
					Error: s.tooLargeError(ClustersRPC, req.ResourceName, size),
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
