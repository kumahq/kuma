package global

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"

	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	config_core "github.com/kumahq/kuma/pkg/config/core"
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
	kds_server "github.com/kumahq/kuma/pkg/kds/v2/server"
	kds_sync_store "github.com/kumahq/kuma/pkg/kds/v2/store"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	store_config "github.com/kumahq/kuma/pkg/config/core/resources/store"
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

	deltaServer, err := kds_server.New(
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

	resourceSyncer, err := kds_sync_store.NewResourceSyncer(kdsDeltaGlobalLog, rt.ResourceStore(), rt.Transactions(), rt.Metrics(), rt.Extensions())
	if err != nil {
		return err
	}

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
			rt,
			deltaServer,
			resourceSyncer,
			rt.Config().Store.Type == store_config.KubernetesStore,
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
