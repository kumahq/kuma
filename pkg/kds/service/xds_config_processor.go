package service

import (
	"context"
	"fmt"
	"time"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
)

type XDSConfigProcessor interface {
	StartProcessing(stream mesh_proto.GlobalKDSService_StreamXDSConfigsClient, errorCh chan error)
}

type ConfigDumpFn = func(ctx context.Context, proxy core_model.ResourceWithAddress) ([]byte, error)

type xdsConfigDumpProcessor struct {
	resManager   core_manager.ReadOnlyResourceManager
	configDumpFn ConfigDumpFn
}

var _ XDSConfigProcessor = &xdsConfigDumpProcessor{}

func NewXDSConfigProcessor(
	resManager core_manager.ReadOnlyResourceManager,
	configDumpFn ConfigDumpFn,
) XDSConfigProcessor {
	return &xdsConfigDumpProcessor{
		resManager:   resManager,
		configDumpFn: configDumpFn,
	}
}

func (s *xdsConfigDumpProcessor) StartProcessing(
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
			config, err := s.executeConfigDump(stream.Context(), req)

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

func (s *xdsConfigDumpProcessor) executeConfigDump(ctx context.Context, req *mesh_proto.XDSConfigRequest) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	res, err := registry.Global().NewObject(core_model.ResourceType(req.ResourceType))
	if err != nil {
		return nil, err
	}
	if err := s.resManager.Get(ctx, res, core_store.GetByKey(req.ResourceName, req.ResourceMesh)); err != nil {
		return nil, err
	}

	resWithAddr, ok := res.(core_model.ResourceWithAddress)
	if !ok {
		return nil, fmt.Errorf("invalid type %T", resWithAddr)
	}

	return s.configDumpFn(ctx, resWithAddr)
}
