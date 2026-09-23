package global

import (
	"slices"
	"time"

	"github.com/pkg/errors"

	config_core "github.com/kumahq/kuma/v3/pkg/config/core"
	"github.com/kumahq/kuma/v3/pkg/core"
	"github.com/kumahq/kuma/v3/pkg/core/runtime"
	"github.com/kumahq/kuma/v3/pkg/core/runtime/component"
	kds_auth "github.com/kumahq/kuma/v3/pkg/kds/auth"
	"github.com/kumahq/kuma/v3/pkg/kds/mux"
	kds_server "github.com/kumahq/kuma/v3/pkg/kds/server"
	"github.com/kumahq/kuma/v3/pkg/kds/service"
	kds_sync_store "github.com/kumahq/kuma/v3/pkg/kds/store"
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

	deltaServer, kdsMetrics, err := kds_server.New(
		kdsDeltaGlobalLog,
		rt,
		rt.KDSContext().TypesSentByGlobal,
		"global",
		rt.KDSContext().GlobalProvidedFilter,
		rt.KDSContext().GlobalResourceMapper,
		rt.Config().Multizone.Global.KDS.NackBackoff.Duration,
		rt.Config().Multizone.Global.KDS.EventBasedWatchdog.AsRuntimeConfig(),
	)
	if err != nil {
		return err
	}

	resourceSyncer, err := kds_sync_store.NewResourceSyncer(kdsDeltaGlobalLog, rt.ResourceStore(), rt.Transactions(), rt.Metrics(), rt.Extensions())
	if err != nil {
		return err
	}

	var streamInterceptors []service.StreamInterceptor
	for _, filter := range rt.KDSContext().GlobalServerFilters {
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
	authStreamInterceptors, authUnaryInterceptors, err := kds_auth.ServerInterceptors(
		rt.Config().Multizone.Global.KDS.Auth.Type,
		rt.KDSContext().ZoneAuthenticators,
	)
	if err != nil {
		return err
	}
	if len(authStreamInterceptors) == 0 {
		kdsGlobalLog.Info("authentication of Zone CPs is disabled")
	}
	grpcStreamInterceptors := append(slices.Clone(rt.KDSContext().ServerStreamInterceptors), authStreamInterceptors...)
	grpcUnaryInterceptors := append(slices.Clone(rt.KDSContext().ServerUnaryInterceptor), authUnaryInterceptors...)
	kdsSyncServer := mux.NewKDSSyncServiceServer(rt, deltaServer, resourceSyncer, kdsMetrics)
	return rt.Add(component.NewResilientComponent(kdsGlobalLog.WithName("kds-mux-client"), mux.NewServer(
		grpcStreamInterceptors,
		grpcUnaryInterceptors,
		*rt.Config().Multizone.Global.KDS,
		rt.CertWatchers(),
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
		kdsSyncServer,
	),
		rt.Config().General.ResilientComponentBaseBackoff.Duration,
		rt.Config().General.ResilientComponentMaxBackoff.Duration),
	)
}
