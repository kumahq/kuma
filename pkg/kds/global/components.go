package global

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	config_core "github.com/kumahq/kuma/pkg/config/core"
	store_config "github.com/kumahq/kuma/pkg/config/core/resources/store"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/runtime"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	"github.com/kumahq/kuma/pkg/core/user"
	"github.com/kumahq/kuma/pkg/kds/mux"
	"github.com/kumahq/kuma/pkg/kds/service"
	"github.com/kumahq/kuma/pkg/kds/util"
	kds_client_v2 "github.com/kumahq/kuma/pkg/kds/v2/client"
	kds_server_v2 "github.com/kumahq/kuma/pkg/kds/v2/server"
	kds_sync_store_v2 "github.com/kumahq/kuma/pkg/kds/v2/store"
	kuma_log "github.com/kumahq/kuma/pkg/log"
	resources_k8s "github.com/kumahq/kuma/pkg/plugins/resources/k8s"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

var (
	kdsGlobalLog      = core.Log.WithName("kds-global")
	kdsDeltaGlobalLog = core.Log.WithName("kds-delta-global")
)

func Setup(rt runtime.Runtime) error {
	if rt.Config().Mode != config_core.Global {
		// Only run on global
		return nil
	}

	kdsServerV2, err := kds_server_v2.New(
		kdsDeltaGlobalLog,
		rt,
		rt.KDSContext().TypesSentByGlobal,
		"global",
		rt.Config().Multizone.Global.KDS.RefreshInterval.Duration,
		rt.KDSContext().GlobalProvidedFilter,
		rt.KDSContext().GlobalResourceMapper,
		rt.Config().Multizone.Global.KDS.NackBackoff.Duration,
	)
	if err != nil {
		return err
	}

	resourceSyncerV2, err := kds_sync_store_v2.NewResourceSyncer(kdsDeltaGlobalLog, rt.ResourceStore(), rt.Transactions(), rt.Metrics(), rt.Extensions())
	if err != nil {
		return err
	}
	kubeFactory := resources_k8s.NewSimpleKubeFactory()

	onGlobalToZoneSyncConnect := mux.OnGlobalToZoneSyncConnectFunc(func(stream mesh_proto.KDSSyncService_GlobalToZoneSyncServer, errCh chan error) {
		zoneID, err := util.ClientIDFromIncomingCtx(stream.Context())
		if err != nil {
			errCh <- errors.Wrap(err, "failed to extract Zone ID from context on GlobalToZoneSyncConnect")
			return
		}

		log := kdsDeltaGlobalLog.WithValues("peer-id", zoneID)
		log = kuma_log.AddFieldsFromCtx(log, stream.Context(), rt.Extensions())
		log.Info("Global To Zone new session created")
		if err := createZoneIfAbsent(stream.Context(), log, zoneID, rt.ResourceManager(), rt.KDSContext().CreateZoneOnFirstConnect); err != nil {
			if errors.Is(err, context.Canceled) {
				log.Info("failed to create a zone on context canceled")
			} else {
				errCh <- errors.Wrap(err, "Global CP could not create a zone")
			}
			return
		}

		if err := kdsServerV2.GlobalToZoneSync(stream); err != nil && (status.Code(err) != codes.Canceled && !errors.Is(err, context.Canceled)) {
			errCh <- errors.Wrap(err, " GlobalToZoneSync finished with an error")
			return
		}

		log.V(1).Info("GlobalToZoneSync finished gracefully")
	})

	onZoneToGlobalSyncConnect := mux.OnZoneToGlobalSyncConnectFunc(func(stream mesh_proto.KDSSyncService_ZoneToGlobalSyncServer, errCh chan error) {
		zoneID, err := util.ClientIDFromIncomingCtx(stream.Context())
		if err != nil {
			errCh <- errors.Wrap(err, "failed to extract Zone ID from context on ZoneToGlobalSyncConnect")
			return
		}

		log := kdsDeltaGlobalLog.WithValues("peer-id", zoneID)
		log = kuma_log.AddFieldsFromCtx(log, stream.Context(), rt.Extensions())
		kdsStream := kds_client_v2.NewDeltaKDSStream(stream, zoneID, rt, "")
		sink := kds_client_v2.NewKDSSyncClient(
			log,
			rt.KDSContext().TypesSentByZone,
			kdsStream,
			kds_sync_store_v2.GlobalSyncCallback(stream.Context(), resourceSyncerV2, rt.Config().Store.Type == store_config.KubernetesStore, kubeFactory, rt.Config().Store.Kubernetes.SystemNamespace),
			rt.Config().Multizone.Global.KDS.ResponseBackoff.Duration,
		)

		go func() {
			if err := sink.SendReq(); err != nil {
				err = errors.Wrap(err, "ZoneToGlobalSyncClient send request finished with an error")
				log.Error(err, "failed to send discovery requests")
				errCh <- err
			} else {
				log.V(1).Info("all discovery requests sent")
			}
		}()
		go func() {
			if err := sink.ReceiveResp(); err != nil && (status.Code(err) != codes.Canceled && !errors.Is(err, context.Canceled)) {
				err = errors.Wrap(err, "ZoneToGlobalSyncClient finished with an error")
				log.Error(err, " failed to receive discovery responses")
				errCh <- err
			} else {
				log.V(1).Info("ZoneToGlobalSyncClient finished gracefully")
			}
		}()
	})

	var streamInterceptors []service.StreamInterceptor
	for _, filter := range rt.KDSContext().GlobalServerFiltersV2 {
		streamInterceptors = append(streamInterceptors, filter)
	}

	if rt.Config().Multizone.Global.KDS.ZoneHealthCheck.Timeout.Duration > time.Duration(0) {
		zwLog := kdsGlobalLog.WithName("zone-watch")
		zw, err := mux.NewZoneWatch(
			zwLog,
			rt.Config().Multizone.Global.KDS.ZoneHealthCheck,
			rt.Metrics(),
			rt.EventBus(),
			rt.ReadOnlyResourceManager(),
			rt.Extensions(),
		)
		if err != nil {
			return errors.Wrap(err, "couldn't create ZoneWatch")
		}
		if err := rt.Add(component.NewResilientComponent(zwLog, zw, rt.Config().General.ResilientComponentBaseBackoff.Duration, rt.Config().General.ResilientComponentMaxBackoff.Duration)); err != nil {
			return err
		}
	}
	return rt.Add(component.NewResilientComponent(kdsGlobalLog.WithName("kds-mux-client"), mux.NewServer(
		rt.KDSContext().ServerStreamInterceptors,
		rt.KDSContext().ServerUnaryInterceptor,
		*rt.Config().Multizone.Global.KDS,
		rt.Metrics(),
		service.NewGlobalKDSServiceServer(
			rt.AppContext(),
			rt.KDSContext().EnvoyAdminRPCs,
			rt.ResourceManager(),
			rt.GetInstanceId(),
			streamInterceptors,
			rt.Extensions(),
			rt.Config().Store.Upsert,
			rt.EventBus(),
			rt.Config().Multizone.Global.KDS.ZoneHealthCheck.PollInterval.Duration,
		),
		mux.NewKDSSyncServiceServer(
			rt.AppContext(),
			onGlobalToZoneSyncConnect,
			onZoneToGlobalSyncConnect,
			rt.KDSContext().GlobalServerFiltersV2,
			rt.Extensions(),
			rt.EventBus(),
			rt.ResourceManager(),
			rt.Config().Store.Upsert,
			rt.GetInstanceId(),
		),
	),
		rt.Config().General.ResilientComponentBaseBackoff.Duration,
		rt.Config().General.ResilientComponentMaxBackoff.Duration),
	)
}

func createZoneIfAbsent(ctx context.Context, log logr.Logger, name string, resManager core_manager.ResourceManager, createZoneOnConnect bool) error {
	ctx = user.Ctx(ctx, user.ControlPlane)
	if err := resManager.Get(ctx, system.NewZoneResource(), store.GetByKey(name, model.NoMesh)); err != nil {
		if !store.IsNotFound(err) || !createZoneOnConnect {
			return err
		}
		log.Info("creating Zone", "name", name)
		zone := &system.ZoneResource{
			Spec: &system_proto.Zone{
				Enabled: util_proto.Bool(true),
			},
		}
		if err := resManager.Create(ctx, zone, store.CreateByKey(name, model.NoMesh)); err != nil {
			return err
		}
	}
	return nil
}
