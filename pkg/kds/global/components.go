package global

import (
	"time"

	"github.com/pkg/errors"

	config_core "github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/runtime"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	"github.com/kumahq/kuma/pkg/kds/mux"
	"github.com/kumahq/kuma/pkg/kds/service"
	kds_server "github.com/kumahq/kuma/pkg/kds/v2/server"
	kds_sync_store "github.com/kumahq/kuma/pkg/kds/v2/store"
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
		),
	),
		rt.Config().General.ResilientComponentBaseBackoff.Duration,
		rt.Config().General.ResilientComponentMaxBackoff.Duration),
	)
}
