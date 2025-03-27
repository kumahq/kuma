package mux

import (
	"context"
	"math/rand"
	"time"

	"github.com/pkg/errors"
	"github.com/sethvargo/go-retry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/api/system/v1alpha1"
	config_store "github.com/kumahq/kuma/pkg/config/core/resources/store"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/events"
	"github.com/kumahq/kuma/pkg/kds"
	"github.com/kumahq/kuma/pkg/kds/service"
	"github.com/kumahq/kuma/pkg/kds/util"
	"github.com/kumahq/kuma/pkg/log"
	"github.com/kumahq/kuma/pkg/multitenant"
	"github.com/kumahq/kuma/pkg/util/proto"
)

type FilterV2 interface {
	InterceptServerStream(stream grpc.ServerStream) error
	InterceptClientStream(stream grpc.ClientStream) error
}

type OnGlobalToZoneSyncConnectFunc func(stream mesh_proto.KDSSyncService_GlobalToZoneSyncServer) error

func (f OnGlobalToZoneSyncConnectFunc) OnGlobalToZoneSyncConnect(stream mesh_proto.KDSSyncService_GlobalToZoneSyncServer) error {
	return f(stream)
}

type OnZoneToGlobalSyncConnectFunc func(stream mesh_proto.KDSSyncService_ZoneToGlobalSyncServer) error

func (f OnZoneToGlobalSyncConnectFunc) OnZoneToGlobalSyncConnect(stream mesh_proto.KDSSyncService_ZoneToGlobalSyncServer) error {
	return f(stream)
}

var clientLog = core.Log.WithName("kds-delta-client")

type KDSSyncServiceServer struct {
	globalToZoneCb OnGlobalToZoneSyncConnectFunc
	zoneToGlobalCb OnZoneToGlobalSyncConnectFunc
	filters        []FilterV2
	extensions     context.Context
	eventBus       events.EventBus
	mesh_proto.UnimplementedKDSSyncServiceServer
	context    context.Context
	resManager manager.ResourceManager
	upsertCfg  config_store.UpsertConfig
	instanceID string
}

func NewKDSSyncServiceServer(ctx context.Context, globalToZoneCb OnGlobalToZoneSyncConnectFunc, zoneToGlobalCb OnZoneToGlobalSyncConnectFunc, filters []FilterV2, extensions context.Context, eventBus events.EventBus, resManager manager.ResourceManager, upsertCfg config_store.UpsertConfig, instanceID string) *KDSSyncServiceServer {
	return &KDSSyncServiceServer{
		context:        ctx,
		globalToZoneCb: globalToZoneCb,
		zoneToGlobalCb: zoneToGlobalCb,
		filters:        filters,
		extensions:     extensions,
		eventBus:       eventBus,
		resManager:     resManager,
		upsertCfg:      upsertCfg,
		instanceID:     instanceID,
	}
}

var _ mesh_proto.KDSSyncServiceServer = &KDSSyncServiceServer{}

func (g *KDSSyncServiceServer) GlobalToZoneSync(stream mesh_proto.KDSSyncService_GlobalToZoneSyncServer) error {
	logger := log.AddFieldsFromCtx(clientLog, stream.Context(), g.extensions)
	zone, err := util.ClientIDFromIncomingCtx(stream.Context())
	if err != nil {
		return err
	}
	logger = logger.WithValues("clientID", zone, "type", "globalToZone")
	for _, filter := range g.filters {
		if err := filter.InterceptServerStream(stream); err != nil {
			return errors.Wrap(err, "closing KDS stream following a callback error")
		}
	}

	connectTime := time.Now()
	shouldDisconnectStream := g.watchZoneHealthCheck(stream.Context(), zone, service.GlobalToZone, connectTime)
	defer shouldDisconnectStream.Close()

	processingErrorsCh := make(chan error, 1)
	go func() {
		if err := g.globalToZoneCb.OnGlobalToZoneSyncConnect(stream); err != nil {
			processingErrorsCh <- err
		}
	}()
	if err := g.storeStreamConnection(stream.Context(), zone, service.GlobalToZone, connectTime); err != nil {
		if errors.Is(err, context.Canceled) && errors.Is(stream.Context().Err(), context.Canceled) {
			return status.Error(codes.Canceled, "stream was cancelled")
		}
		logger.Error(err, "could not store stream connection")
		return status.Error(codes.Internal, "could not store stream connection")
	}
	logger.Info("stored stream connection")

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
	logger = logger.WithValues("clientID", zone, "type", "zoneToGlobal")
	for _, filter := range g.filters {
		if err := filter.InterceptServerStream(stream); err != nil {
			return errors.Wrap(err, "closing KDS stream following a callback error")
		}
	}
	connectTime := time.Now()
	shouldDisconnectStream := g.watchZoneHealthCheck(stream.Context(), zone, service.ZoneToGlobal, connectTime)
	defer shouldDisconnectStream.Close()

	processingErrorsCh := make(chan error, 1)
	go func() {
		if err := g.zoneToGlobalCb.OnZoneToGlobalSyncConnect(stream); err != nil {
			processingErrorsCh <- err
		}
	}()

	if err := g.storeStreamConnection(stream.Context(), zone, service.ZoneToGlobal, connectTime); err != nil {
		if errors.Is(err, context.Canceled) && errors.Is(stream.Context().Err(), context.Canceled) {
			return status.Error(codes.Canceled, "stream was cancelled")
		}
		logger.Error(err, "could not store stream connection")
		return status.Error(codes.Internal, "could not store stream connection")
	}
	logger.Info("stored stream connection")

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

func (g *KDSSyncServiceServer) watchZoneHealthCheck(streamContext context.Context, zone string, typ service.StreamType, connectTime time.Time) events.Listener {
	tenantID, _ := multitenant.TenantFromCtx(streamContext)

	shouldDisconnectStream := events.NewNeverListener()
	if kds.ContextHasFeature(streamContext, kds.FeatureZonePingHealth) {
		shouldDisconnectStream = g.eventBus.Subscribe(func(e events.Event) bool {
			switch event := e.(type) {
			case service.ZoneWentOffline:
				return event.TenantID == tenantID && event.Zone == zone
			case service.StreamCancelled:
				return event.TenantID == tenantID && event.Zone == zone && event.Type == typ && event.ConnTime.Equal(connectTime)
			default:
				return false
			}
		})
		g.eventBus.Send(service.ZoneOpenedStream{Zone: zone, TenantID: tenantID, Type: typ, ConnTime: connectTime})
	}

	return shouldDisconnectStream
}

func (g *KDSSyncServiceServer) storeStreamConnection(ctx context.Context, zone string, typ service.StreamType, connectTime time.Time) error {
	key := model.ResourceKey{Name: zone}

	// wait for Zone to be created, only then we can create Zone Insight
	err := retry.Do(
		ctx,
		retry.WithMaxRetries(30, retry.NewConstant(1*time.Second)),
		func(ctx context.Context) error {
			return retry.RetryableError(g.resManager.Get(ctx, system.NewZoneResource(), core_store.GetBy(key)))
		},
	)
	if err != nil {
		return err
	}

	// Add delay for Upsert. If Global CP is behind an HTTP load balancer,
	// it might be the case that each Envoy Admin stream will land on separate instance.
	// In this case, all instances will try to update Zone Insight which will result in conflicts.
	// Since it's unusual to immediately execute envoy admin rpcs after zone is connected, 0-10s delay should be fine.
	// #nosec G404 - math rand is enough
	time.Sleep(time.Duration(rand.Int31n(10000)) * time.Millisecond)

	zoneInsight := system.NewZoneInsightResource()
	return manager.Upsert(ctx, g.resManager, key, zoneInsight, func(resource model.Resource) error {
		if zoneInsight.Spec.KdsStreams == nil {
			zoneInsight.Spec.KdsStreams = &v1alpha1.KDSStreams{}
		}
		stream := zoneInsight.Spec.GetKDSStream(string(typ))
		if stream == nil {
			stream = &v1alpha1.KDSStream{}
		}
		if stream.GetConnectTime() == nil || proto.MustTimestampFromProto(stream.ConnectTime).Before(connectTime) {
			stream.GlobalInstanceId = g.instanceID
			stream.ConnectTime = proto.MustTimestampProto(connectTime)
		}
		switch typ {
		case service.GlobalToZone:
			zoneInsight.Spec.KdsStreams.GlobalToZone = stream
		case service.ZoneToGlobal:
			zoneInsight.Spec.KdsStreams.ZoneToGlobal = stream
		}
		return nil
	}, manager.WithConflictRetry(g.upsertCfg.ConflictRetryBaseBackoff.Duration, g.upsertCfg.ConflictRetryMaxTimes, g.upsertCfg.ConflictRetryJitterPercent)) // we need retry because zone sink or other RPC may also update the insight.
}
