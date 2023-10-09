package mux

import (
	"context"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/events"
	"github.com/kumahq/kuma/pkg/kds/service"
	"github.com/kumahq/kuma/pkg/kds/util"
	"github.com/kumahq/kuma/pkg/log"
	"github.com/kumahq/kuma/pkg/multitenant"
)

type FilterV2 interface {
	InterceptServerStream(stream grpc.ServerStream) error
	InterceptClientStream(stream grpc.ClientStream) error
}

type OnGlobalToZoneSyncConnectFunc func(stream mesh_proto.KDSSyncService_GlobalToZoneSyncServer, errorCh chan error)

func (f OnGlobalToZoneSyncConnectFunc) OnGlobalToZoneSyncConnect(stream mesh_proto.KDSSyncService_GlobalToZoneSyncServer, errorCh chan error) {
	f(stream, errorCh)
}

type OnZoneToGlobalSyncConnectFunc func(stream mesh_proto.KDSSyncService_ZoneToGlobalSyncServer, errorCh chan error)

func (f OnZoneToGlobalSyncConnectFunc) OnZoneToGlobalSyncConnect(stream mesh_proto.KDSSyncService_ZoneToGlobalSyncServer, errorCh chan error) {
	f(stream, errorCh)
}

var clientLog = core.Log.WithName("kds-delta-client")

type KDSSyncServiceServer struct {
	globalToZoneCb OnGlobalToZoneSyncConnectFunc
	zoneToGlobalCb OnZoneToGlobalSyncConnectFunc
	filters        []FilterV2
	extensions     context.Context
	eventBus       events.EventBus
	mesh_proto.UnimplementedKDSSyncServiceServer
}

func NewKDSSyncServiceServer(
	globalToZoneCb OnGlobalToZoneSyncConnectFunc,
	zoneToGlobalCb OnZoneToGlobalSyncConnectFunc,
	filters []FilterV2,
	extensions context.Context,
	eventBus events.EventBus,
) *KDSSyncServiceServer {
	return &KDSSyncServiceServer{
		globalToZoneCb: globalToZoneCb,
		zoneToGlobalCb: zoneToGlobalCb,
		filters:        filters,
		extensions:     extensions,
		eventBus:       eventBus,
	}
}

var _ mesh_proto.KDSSyncServiceServer = &KDSSyncServiceServer{}

func (g *KDSSyncServiceServer) GlobalToZoneSync(stream mesh_proto.KDSSyncService_GlobalToZoneSyncServer) error {
	logger := log.AddFieldsFromCtx(clientLog, stream.Context(), g.extensions)
	zone, err := util.ClientIDFromIncomingCtx(stream.Context())
	if err != nil {
		return err
	}
	logger = logger.WithValues("clientID", zone)
	for _, filter := range g.filters {
		if err := filter.InterceptServerStream(stream); err != nil {
			return errors.Wrap(err, "closing KDS stream following a callback error")
		}
	}

	tenantID, _ := multitenant.TenantFromCtx(stream.Context())
	shouldDisconnectStream := g.watchZoneHealthCheck(tenantID, zone)
	defer shouldDisconnectStream.Close()

	processingErrorsCh := make(chan error)
	go g.globalToZoneCb.OnGlobalToZoneSyncConnect(stream, processingErrorsCh)
	select {
	case <-shouldDisconnectStream.Recv():
		logger.Info("ending stream, zone health check failed")
		return nil
	case <-stream.Context().Done():
		logger.Info("GlobalToZoneSync rpc stream stopped")
		return nil
	case err := <-processingErrorsCh:
		if status.Code(err) == codes.Unimplemented {
			return errors.Wrap(err, "GlobalToZoneSync rpc stream failed, because Global CP does not implement this rpc. Upgrade Global CP.")
		}
		logger.Error(err, "GlobalToZoneSync rpc stream failed prematurely, will restart in background")
		return status.Error(codes.Internal, "stream failed")
	}
}

func (g *KDSSyncServiceServer) ZoneToGlobalSync(stream mesh_proto.KDSSyncService_ZoneToGlobalSyncServer) error {
	logger := log.AddFieldsFromCtx(clientLog, stream.Context(), g.extensions)
	zone, err := util.ClientIDFromIncomingCtx(stream.Context())
	if err != nil {
		return err
	}
	logger = logger.WithValues("clientID", zone)
	for _, filter := range g.filters {
		if err := filter.InterceptServerStream(stream); err != nil {
			return errors.Wrap(err, "closing KDS stream following a callback error")
		}
	}

	tenantID, _ := multitenant.TenantFromCtx(stream.Context())
	shouldDisconnectStream := g.watchZoneHealthCheck(tenantID, zone)
	defer shouldDisconnectStream.Close()

	processingErrorsCh := make(chan error)
	go g.zoneToGlobalCb.OnZoneToGlobalSyncConnect(stream, processingErrorsCh)
	select {
	case <-shouldDisconnectStream.Recv():
		logger.Info("ending stream, zone health check failed")
		return nil
	case <-stream.Context().Done():
		logger.Info("ZoneToGlobalSync rpc stream stopped")
		return nil
	case err := <-processingErrorsCh:
		if status.Code(err) == codes.Unimplemented {
			return errors.Wrap(err, "ZoneToGlobalSync rpc stream failed, because Global CP does not implement this rpc. Upgrade Global CP.")
		}
		logger.Error(err, "ZoneToGlobalSync rpc stream failed prematurely, will restart in background")
		return status.Error(codes.Internal, "stream failed")
	}
}

func (g *KDSSyncServiceServer) watchZoneHealthCheck(tenantID, zone string) events.Listener {
	shouldDisconnectStream := g.eventBus.Subscribe(func(e events.Event) bool {
		disconnectEvent, ok := e.(service.ZoneWentOffline)
		return ok && disconnectEvent.TenantID == tenantID && disconnectEvent.Zone == zone
	})
	g.eventBus.Send(service.ZoneOpenedStream{Zone: zone, TenantID: tenantID})

	return shouldDisconnectStream
}
