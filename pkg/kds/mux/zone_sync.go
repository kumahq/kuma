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
	"github.com/kumahq/kuma/pkg/kds"
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
	context context.Context
}

func NewKDSSyncServiceServer(ctx context.Context, globalToZoneCb OnGlobalToZoneSyncConnectFunc, zoneToGlobalCb OnZoneToGlobalSyncConnectFunc, filters []FilterV2, extensions context.Context, eventBus events.EventBus) *KDSSyncServiceServer {
	return &KDSSyncServiceServer{
		context:        ctx,
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

	shouldDisconnectStream := g.watchZoneHealthCheck(stream.Context(), zone)
	defer shouldDisconnectStream.Close()

	processingErrorsCh := make(chan error, 1)
	go g.globalToZoneCb.OnGlobalToZoneSyncConnect(stream, processingErrorsCh)
	select {
	case <-shouldDisconnectStream.Recv():
		logger.Info("ending stream, zone health check failed")
		return status.Error(codes.Canceled, "stream canceled - zone hc failed")
	case <-stream.Context().Done():
		logger.Info("GlobalToZoneSync rpc stream stopped")
		return status.Error(codes.Canceled, "stream canceled - stream stopped")
	case <-g.context.Done():
		logger.Info("app context done")
		return status.Error(codes.Unavailable, "stream unavailable")
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

	shouldDisconnectStream := g.watchZoneHealthCheck(stream.Context(), zone)
	defer shouldDisconnectStream.Close()

	processingErrorsCh := make(chan error, 1)
	go g.zoneToGlobalCb.OnZoneToGlobalSyncConnect(stream, processingErrorsCh)
	select {
	case <-shouldDisconnectStream.Recv():
		logger.Info("ending stream, zone health check failed")
		return status.Error(codes.Canceled, "stream canceled - zone hc failed")
	case <-stream.Context().Done():
		logger.Info("ZoneToGlobalSync rpc stream stopped")
		return status.Error(codes.Canceled, "stream canceled - stream stopped")
	case <-g.context.Done():
		logger.Info("app context done")
		return status.Error(codes.Unavailable, "stream unavailable")
	case err := <-processingErrorsCh:
		if status.Code(err) == codes.Unimplemented {
			return errors.Wrap(err, "ZoneToGlobalSync rpc stream failed, because Global CP does not implement this rpc. Upgrade Global CP.")
		}
		logger.Error(err, "ZoneToGlobalSync rpc stream failed prematurely, will restart in background")
		return status.Error(codes.Internal, "stream failed")
	}
}

func (g *KDSSyncServiceServer) watchZoneHealthCheck(streamContext context.Context, zone string) events.Listener {
	tenantID, _ := multitenant.TenantFromCtx(streamContext)

	shouldDisconnectStream := events.NewNeverListener()

	if kds.ContextHasFeature(streamContext, kds.FeatureZonePingHealth) {
		shouldDisconnectStream = g.eventBus.Subscribe(func(e events.Event) bool {
			disconnectEvent, ok := e.(service.ZoneWentOffline)
			return ok && disconnectEvent.TenantID == tenantID && disconnectEvent.Zone == zone
		})
		g.eventBus.Send(service.ZoneOpenedStream{Zone: zone, TenantID: tenantID})
	}

	return shouldDisconnectStream
}
