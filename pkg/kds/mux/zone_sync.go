package mux

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/envoyproxy/go-control-plane/pkg/server/delta/v3"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"github.com/sethvargo/go-retry"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	system_proto "github.com/kumahq/kuma/v2/api/system/v1alpha1"
	config_store "github.com/kumahq/kuma/v2/pkg/config/core/resources/store"
	"github.com/kumahq/kuma/v2/pkg/core"
	"github.com/kumahq/kuma/v2/pkg/core/resources/apis/system"
	core_manager "github.com/kumahq/kuma/v2/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/v2/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/v2/pkg/core/resources/store"
	"github.com/kumahq/kuma/v2/pkg/core/runtime"
	"github.com/kumahq/kuma/v2/pkg/core/user"
	"github.com/kumahq/kuma/v2/pkg/events"
	"github.com/kumahq/kuma/v2/pkg/kds"
	kds_context "github.com/kumahq/kuma/v2/pkg/kds/context"
	"github.com/kumahq/kuma/v2/pkg/kds/service"
	"github.com/kumahq/kuma/v2/pkg/kds/util"
	kds_client_v2 "github.com/kumahq/kuma/v2/pkg/kds/v2/client"
	kds_sync_store_v2 "github.com/kumahq/kuma/v2/pkg/kds/v2/store"
	"github.com/kumahq/kuma/v2/pkg/log"
	"github.com/kumahq/kuma/v2/pkg/multitenant"
	"github.com/kumahq/kuma/v2/pkg/plugins/resources/k8s"
	util_proto "github.com/kumahq/kuma/v2/pkg/util/proto"
)

var clientLog = core.Log.WithName("kds-delta-client")

type KDSSyncServiceServer struct {
	filters    []kds_context.FilterV2
	extensions context.Context
	eventBus   events.EventBus
	mesh_proto.UnimplementedKDSSyncServiceServer
	context                  context.Context
	resManager               core_manager.ResourceManager
	upsertCfg                config_store.UpsertConfig
	instanceID               string
	createZoneOnFirstConnect bool
	deltaServer              delta.Server
	typesSentByZone          []core_model.ResourceType
	resourceSyncer           kds_sync_store_v2.ResourceSyncer
	k8sStore                 bool
	systemNamespace          string
	responseBackoff          time.Duration
	grpcStop                 func()
}

func NewKDSSyncServiceServer(
	rt runtime.Runtime,
	deltaServer delta.Server,
	resourceSyncer kds_sync_store_v2.ResourceSyncer,
) *KDSSyncServiceServer {
	return &KDSSyncServiceServer{
		context:                  rt.AppContext(),
		filters:                  rt.KDSContext().GlobalServerFiltersV2,
		extensions:               rt.Extensions(),
		eventBus:                 rt.EventBus(),
		resManager:               rt.ResourceManager(),
		upsertCfg:                rt.Config().Store.Upsert,
		instanceID:               rt.GetInstanceId(),
		deltaServer:              deltaServer,
		createZoneOnFirstConnect: rt.KDSContext().CreateZoneOnFirstConnect,
		typesSentByZone:          rt.KDSContext().TypesSentByZone,
		resourceSyncer:           resourceSyncer,
		k8sStore:                 rt.Config().Store.Type == config_store.KubernetesStore,
		systemNamespace:          rt.Config().Store.Kubernetes.SystemNamespace,
		responseBackoff:          rt.Config().Multizone.Global.KDS.ResponseBackoff.Duration,
	}
}

var _ mesh_proto.KDSSyncServiceServer = &KDSSyncServiceServer{}

func createZoneIfAbsent(ctx context.Context, log logr.Logger, name string, resManager core_manager.ResourceManager, createZoneOnConnect bool) error {
	ctx = user.Ctx(ctx, user.ControlPlane)
	if err := resManager.Get(ctx, system.NewZoneResource(), core_store.GetByKey(name, core_model.NoMesh)); err != nil {
		if !core_store.IsNotFound(err) || !createZoneOnConnect {
			return err
		}
		log.Info("creating Zone", "name", name)
		zone := &system.ZoneResource{
			Spec: &system_proto.Zone{
				Enabled: util_proto.Bool(true),
			},
		}
		if err := resManager.Create(ctx, zone, core_store.CreateByKey(name, core_model.NoMesh)); err != nil {
			return err
		}
	}
	return nil
}

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
		logger.Info("Global To Zone new session created")
		if err := createZoneIfAbsent(stream.Context(), logger, zone, g.resManager, g.createZoneOnFirstConnect); err != nil {
			if errors.Is(err, context.Canceled) {
				processingErrorsCh <- nil
				return
			}
			processingErrorsCh <- errors.Wrap(err, "Global CP could not create a zone")
			return
		}
		errorStream := NewErrorRecorderStream(stream)
		err := g.deltaServer.DeltaStreamHandler(errorStream, "")
		if err == nil {
			err = errorStream.Err()
		}

		if err != nil && (status.Code(err) != codes.Canceled && !errors.Is(err, context.Canceled)) {
			processingErrorsCh <- err
			return
		}

		logger.V(1).Info("GlobalToZoneSync finished gracefully")
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
	logger.V(1).Info("Zone To Global new session created")
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
	group, ctx := errgroup.WithContext(stream.Context())
	go func() {
		<-ctx.Done()
		if g.grpcStop != nil {
			logger.Info("received context done, stopping grpc server")
			g.grpcStop()
		}
	}()
	go func() {
		kdsStream := kds_client_v2.NewDeltaKDSStream(stream, zone, g.instanceID, "")
		sink := kds_client_v2.NewKDSSyncClient(
			logger,
			g.typesSentByZone,
			kdsStream,
			kds_sync_store_v2.GlobalSyncCallback(
				stream.Context(),
				g.resourceSyncer,
				g.k8sStore,
				k8s.NewSimpleKubeFactory(),
				g.systemNamespace,
			),
			g.responseBackoff,
		)

		if err := sink.Receive(ctx, group); err != nil && (status.Code(err) != codes.Canceled && !errors.Is(err, context.Canceled)) {
			processingErrorsCh <- errors.Wrap(err, "KDSSyncClient finished with an error")
			return
		}

		logger.V(1).Info("KDSSyncClient finished gracefully")
		processingErrorsCh <- nil
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
	key := core_model.ResourceKey{Name: zone}

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
	return core_manager.Upsert(ctx, g.resManager, key, zoneInsight, func(resource core_model.Resource) error {
		if zoneInsight.Spec.KdsStreams == nil {
			zoneInsight.Spec.KdsStreams = &system_proto.KDSStreams{}
		}
		stream := zoneInsight.Spec.GetKDSStream(string(typ))
		if stream == nil {
			stream = &system_proto.KDSStream{}
		}
		if stream.GetConnectTime() == nil || util_proto.MustTimestampFromProto(stream.ConnectTime).Before(connectTime) {
			stream.GlobalInstanceId = g.instanceID
			stream.ConnectTime = util_proto.MustTimestampProto(connectTime)
		}
		switch typ {
		case service.GlobalToZone:
			zoneInsight.Spec.KdsStreams.GlobalToZone = stream
		case service.ZoneToGlobal:
			zoneInsight.Spec.KdsStreams.ZoneToGlobal = stream
		}
		return nil
	}, core_manager.WithConflictRetry(g.upsertCfg.ConflictRetryBaseBackoff.Duration, g.upsertCfg.ConflictRetryMaxTimes, g.upsertCfg.ConflictRetryJitterPercent)) // we need retry because zone sink or other RPC may also update the insight.
}

func (g *KDSSyncServiceServer) SetGrpcStop(stop func()) error {
	if stop == nil {
		return fmt.Errorf("stop func cannot be nil")
	}

	g.grpcStop = stop
	return nil
}
