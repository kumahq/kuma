package zone

import (
	"github.com/kumahq/kuma/v2/pkg/core"
	core_runtime "github.com/kumahq/kuma/v2/pkg/core/runtime"
	"github.com/kumahq/kuma/v2/pkg/core/runtime/component"
	"github.com/kumahq/kuma/v2/pkg/kds/mux"
	"github.com/kumahq/kuma/v2/pkg/kds/service"
	kds_server "github.com/kumahq/kuma/v2/pkg/kds/v2/server"
	kds_sync_store_v2 "github.com/kumahq/kuma/v2/pkg/kds/v2/store"
)

var (
	kdsZoneLog      = core.Log.WithName("kds-zone")
	kdsDeltaZoneLog = core.Log.WithName("kds-delta-zone")
)

func Setup(rt core_runtime.Runtime) error {
	if !rt.Config().IsFederatedZoneCP() {
		// Only run on zone
		return nil
	}
	zone := rt.Config().Multizone.Zone.Name
	kdsCtx := rt.KDSContext()

	deltaServer, err := kds_server.New(
		kdsZoneLog,
		rt,
		kdsCtx.TypesSentByZone,
		zone,
		rt.Config().Multizone.Zone.KDS.RefreshInterval.Duration,
		kdsCtx.ZoneProvidedFilter,
		kdsCtx.ZoneResourceMapper,
		rt.Config().Multizone.Zone.KDS.NackBackoff.Duration,
	)
	if err != nil {
		return err
	}

	resourceSyncerV2, err := kds_sync_store_v2.NewResourceSyncer(kdsDeltaZoneLog, rt.ResourceStore(), rt.Transactions(), rt.Metrics(), rt.Extensions())
	if err != nil {
		return err
	}

	muxClient := mux.NewClient(
		rt.KDSContext().ZoneClientCtx,
		rt.Config().Multizone.Zone.GlobalAddress,
		zone,
		*rt.Config().Multizone.Zone.KDS,
		rt.Config().Experimental,
		rt.Metrics(),
		service.NewEnvoyAdminProcessor(
			rt.ReadOnlyResourceManager(),
			rt.EnvoyAdminClient(),
		),
		resourceSyncerV2,
		rt,
		deltaServer,
	)
	return rt.Add(component.NewResilientComponent(
		kdsZoneLog.WithName("kds-mux-client"),
		muxClient,
		rt.Config().General.ResilientComponentBaseBackoff.Duration,
		rt.Config().General.ResilientComponentMaxBackoff.Duration,
	))
}
