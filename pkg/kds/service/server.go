package service

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"slices"
	"time"

	"github.com/pkg/errors"
	"github.com/sethvargo/go-retry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/api/system/v1alpha1"
	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	config_store "github.com/kumahq/kuma/pkg/config/core/resources/store"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/events"
	"github.com/kumahq/kuma/pkg/kds"
	"github.com/kumahq/kuma/pkg/kds/util"
	kuma_log "github.com/kumahq/kuma/pkg/log"
	"github.com/kumahq/kuma/pkg/multitenant"
	util_grpc "github.com/kumahq/kuma/pkg/util/grpc"
)

var log = core.Log.WithName("kds-service")

type StreamInterceptor interface {
	InterceptServerStream(stream grpc.ServerStream) error
}

type GlobalKDSServiceServer struct {
	envoyAdminRPCs          EnvoyAdminRPCs
	resManager              manager.ResourceManager
	instanceID              string
	filters                 []StreamInterceptor
	extensions              context.Context
	upsertCfg               config_store.UpsertConfig
	eventBus                events.EventBus
	zoneHealthCheckInterval time.Duration
	mesh_proto.UnimplementedGlobalKDSServiceServer
	context context.Context
}

func NewGlobalKDSServiceServer(ctx context.Context, envoyAdminRPCs EnvoyAdminRPCs, resManager manager.ResourceManager, instanceID string, filters []StreamInterceptor, extensions context.Context, upsertCfg config_store.UpsertConfig, eventBus events.EventBus, zoneHealthCheckInterval time.Duration) *GlobalKDSServiceServer {
	return &GlobalKDSServiceServer{
		context:                 ctx,
		envoyAdminRPCs:          envoyAdminRPCs,
		resManager:              resManager,
		instanceID:              instanceID,
		filters:                 filters,
		extensions:              extensions,
		upsertCfg:               upsertCfg,
		eventBus:                eventBus,
		zoneHealthCheckInterval: zoneHealthCheckInterval,
	}
}

var _ mesh_proto.GlobalKDSServiceServer = &GlobalKDSServiceServer{}

func (g *GlobalKDSServiceServer) StreamXDSConfigs(stream mesh_proto.GlobalKDSService_StreamXDSConfigsServer) error {
	return g.streamEnvoyAdminRPC(ConfigDumpRPC, g.envoyAdminRPCs.XDSConfigDump, stream, func() (util_grpc.ReverseUnaryMessage, error) {
		return stream.Recv()
	})
}

func (g *GlobalKDSServiceServer) StreamStats(stream mesh_proto.GlobalKDSService_StreamStatsServer) error {
	return g.streamEnvoyAdminRPC(StatsRPC, g.envoyAdminRPCs.Stats, stream, func() (util_grpc.ReverseUnaryMessage, error) {
		return stream.Recv()
	})
}

func (g *GlobalKDSServiceServer) StreamClusters(stream mesh_proto.GlobalKDSService_StreamClustersServer) error {
	return g.streamEnvoyAdminRPC(ClustersRPC, g.envoyAdminRPCs.Clusters, stream, func() (util_grpc.ReverseUnaryMessage, error) {
		return stream.Recv()
	})
}

func (g *GlobalKDSServiceServer) HealthCheck(ctx context.Context, _ *mesh_proto.ZoneHealthCheckRequest) (*mesh_proto.ZoneHealthCheckResponse, error) {
	zone, err := util.ClientIDFromIncomingCtx(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	tenantZoneID := TenantZoneClientIDFromCtx(ctx, zone)
	log := log.WithValues("clientID", tenantZoneID.String())

	insight := system.NewZoneInsightResource()
	if err := manager.Upsert(ctx, g.resManager, model.ResourceKey{Name: zone, Mesh: model.NoMesh}, insight, func(resource model.Resource) error {
		if insight.Spec.HealthCheck == nil {
			insight.Spec.HealthCheck = &system_proto.HealthCheck{}
		}

		insight.Spec.HealthCheck.Time = timestamppb.Now()
		return nil
	}, manager.WithConflictRetry(
		g.upsertCfg.ConflictRetryBaseBackoff.Duration, g.upsertCfg.ConflictRetryMaxTimes, g.upsertCfg.ConflictRetryJitterPercent,
	)); err != nil && !errors.Is(err, context.Canceled) {
		log.Error(err, "couldn't update zone insight", "zone", zone)
	}

	return &mesh_proto.ZoneHealthCheckResponse{
		Interval: durationpb.New(g.zoneHealthCheckInterval),
	}, nil
}

type ZoneWentOffline struct {
	TenantID string
	Zone     string
}
type ZoneOpenedStream struct {
	TenantID string
	Zone     string
}

func (g *GlobalKDSServiceServer) streamEnvoyAdminRPC(
	rpcName string,
	rpc util_grpc.ReverseUnaryRPCs,
	stream grpc.ServerStream,
	recv func() (util_grpc.ReverseUnaryMessage, error),
) error {
	zone, err := util.ClientIDFromIncomingCtx(stream.Context())
	if err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}
	tenantZoneID := TenantZoneClientIDFromCtx(stream.Context(), zone)

	logger := log.WithValues("rpc", rpcName, "clientID", tenantZoneID.String())
	logger = kuma_log.AddFieldsFromCtx(logger, stream.Context(), g.extensions)
	for _, filter := range g.filters {
		if err := filter.InterceptServerStream(stream); err != nil {
			switch status.Code(err) {
			case codes.InvalidArgument, codes.Unauthenticated, codes.PermissionDenied:
				logger.Info("stream interceptor terminating the stream", "cause", err)
			default:
				logger.Error(err, "stream interceptor terminating the stream")
			}
			return err
		}
	}
	shouldDisconnectStream := events.NewNeverListener()
	md, _ := metadata.FromIncomingContext(stream.Context())
	features := md.Get(kds.FeaturesMetadataKey)

	if slices.Contains(features, kds.FeatureZonePingHealth) {
		shouldDisconnectStream = g.eventBus.Subscribe(func(e events.Event) bool {
			disconnectEvent, ok := e.(ZoneWentOffline)
			return ok && disconnectEvent.TenantID == tenantZoneID.TenantID && disconnectEvent.Zone == zone
		})
		g.eventBus.Send(ZoneOpenedStream{Zone: zone, TenantID: tenantZoneID.TenantID})
	}
	defer shouldDisconnectStream.Close()

	logger.Info("Envoy Admin RPC stream started")
	rpc.ClientConnected(tenantZoneID.String(), stream)
	if err := g.storeStreamConnection(stream.Context(), zone, rpcName, g.instanceID); err != nil {
		if errors.Is(err, context.Canceled) && errors.Is(stream.Context().Err(), context.Canceled) {
			return status.Error(codes.Canceled, "stream was cancelled")
		}
		logger.Error(err, "could not store stream connection")
		return status.Error(codes.Internal, "could not store stream connection")
	}
	logger.Info("stored stream connection")
	streamResult := make(chan error, 1)
	go func() {
		for {
			resp, err := recv()
			if err == io.EOF {
				logger.Info("stream stopped")
				streamResult <- nil
				return
			}
			if status.Code(err) == codes.Canceled {
				logger.Info("stream cancelled")
				streamResult <- nil
				return
			}
			if err != nil {
				logger.Error(err, "could not receive a message")
				streamResult <- status.Error(codes.Internal, "could not receive a message")
				return
			}
			logger.V(1).Info("Envoy Admin RPC response received", "requestId", resp.GetRequestId())
			if err := rpc.ResponseReceived(tenantZoneID.String(), resp); err != nil {
				logger.Error(err, "could not mark the response as received")
				streamResult <- status.Error(codes.InvalidArgument, "could not mark the response as received")
				return
			}
		}
	}()
	select {
	case <-g.context.Done():
		logger.Info("app context done")
		return status.Error(codes.Unavailable, "stream unavailable")
	case <-shouldDisconnectStream.Recv():
		logger.Info("ending stream, zone health check failed")
		return status.Error(codes.Canceled, "stream canceled")
	case res := <-streamResult:
		return res
	}
}

func (g *GlobalKDSServiceServer) storeStreamConnection(ctx context.Context, zone string, rpcName string, instance string) error {
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
		if zoneInsight.Spec.EnvoyAdminStreams == nil {
			zoneInsight.Spec.EnvoyAdminStreams = &v1alpha1.EnvoyAdminStreams{}
		}
		switch rpcName {
		case ConfigDumpRPC:
			zoneInsight.Spec.EnvoyAdminStreams.ConfigDumpGlobalInstanceId = instance
		case StatsRPC:
			zoneInsight.Spec.EnvoyAdminStreams.StatsGlobalInstanceId = instance
		case ClustersRPC:
			zoneInsight.Spec.EnvoyAdminStreams.ClustersGlobalInstanceId = instance
		}
		return nil
	}, manager.WithConflictRetry(g.upsertCfg.ConflictRetryBaseBackoff.Duration, g.upsertCfg.ConflictRetryMaxTimes, g.upsertCfg.ConflictRetryJitterPercent)) // we need retry because zone sink or other RPC may also update the insight.
}

type TenantZoneClientID struct {
	TenantID string
	Zone     string
}

func (id TenantZoneClientID) String() string {
	if id.TenantID == "" {
		return id.Zone
	}
	return fmt.Sprintf("%s:%s", id.Zone, id.TenantID)
}

func TenantZoneClientIDFromCtx(ctx context.Context, zone string) TenantZoneClientID {
	id := TenantZoneClientID{
		Zone: zone,
	}
	tenantID, ok := multitenant.TenantFromCtx(ctx)
	if ok {
		id.TenantID = tenantID
	}
	return id
}
