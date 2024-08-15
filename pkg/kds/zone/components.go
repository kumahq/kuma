package zone

import (
	"context"

	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/config"
	"github.com/kumahq/kuma/pkg/config/core/resources/store"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	core_runtime "github.com/kumahq/kuma/pkg/core/runtime"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	"github.com/kumahq/kuma/pkg/kds/mux"
	"github.com/kumahq/kuma/pkg/kds/service"
	kds_client_v2 "github.com/kumahq/kuma/pkg/kds/v2/client"
	kds_server_v2 "github.com/kumahq/kuma/pkg/kds/v2/server"
	kds_sync_store_v2 "github.com/kumahq/kuma/pkg/kds/v2/store"
	resources_k8s "github.com/kumahq/kuma/pkg/plugins/resources/k8s"
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
	reg := registry.Global()
	kdsCtx := rt.KDSContext()

	kdsServerV2, err := kds_server_v2.New(
		kdsZoneLog,
		rt,
		reg.ObjectTypes(model.HasKDSFlag(model.ZoneToGlobalFlag)),
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
	kubeFactory := resources_k8s.NewSimpleKubeFactory()
	cfg := rt.Config()
	cfgForDisplay, err := config.ConfigForDisplay(&cfg)
	if err != nil {
		return errors.Wrap(err, "could not construct config for display")
	}
	cfgJson, err := config.ToJson(cfgForDisplay)
	if err != nil {
		return errors.Wrap(err, "could not marshall config to json")
	}

	onGlobalToZoneSyncStarted := mux.OnGlobalToZoneSyncStartedFunc(func(stream mesh_proto.KDSSyncService_GlobalToZoneSyncClient, errChan chan error) {
		log := kdsDeltaZoneLog.WithValues("kds-version", "v2")
		syncClient := kds_client_v2.NewKDSSyncClient(
			log,
			reg.ObjectTypes(model.HasKDSFlag(model.GlobalToZoneSelector)),
			kds_client_v2.NewDeltaKDSStream(stream, zone, rt, string(cfgJson)),
			kds_sync_store_v2.ZoneSyncCallback(
				stream.Context(),
				rt.KDSContext().Configs,
				resourceSyncerV2,
				rt.Config().Store.Type == store.KubernetesStore,
				zone,
				kubeFactory,
				rt.Config().Store.Kubernetes.SystemNamespace,
			),
			rt.Config().Multizone.Zone.KDS.ResponseBackoff.Duration,
		)
		go func() {
			err := syncClient.Receive()
			if err != nil && !errors.Is(err, context.Canceled) {
				err = errors.Wrap(err, "GlobalToZoneSyncClient finished with an error")
				select {
				case errChan <- err:
				default:
					log.Error(err, "failed to write error to closed channel")
				}
			} else {
				log.V(1).Info("GlobalToZoneSyncClient finished gracefully")
			}
		}()
	})

	onZoneToGlobalSyncStarted := mux.OnZoneToGlobalSyncStartedFunc(func(stream mesh_proto.KDSSyncService_ZoneToGlobalSyncClient, errChan chan error) {
		log := kdsDeltaZoneLog.WithValues("kds-version", "v2", "peer-id", "global")
		log.Info("ZoneToGlobalSync new session created")
		session := kds_server_v2.NewServerStream(stream)
		go func() {
			err := kdsServerV2.ZoneToGlobal(session)
			if err != nil && !errors.Is(err, context.Canceled) {
				err = errors.Wrap(err, "ZoneToGlobalSync finished with an error")
				select {
				case errChan <- err:
				default:
					log.Error(err, "failed to write error to closed channel")
				}
			} else {
				log.V(1).Info("ZoneToGlobalSync finished gracefully")
			}
		}()
	})

	muxClient := mux.NewClient(
		rt.KDSContext().ZoneClientCtx,
		rt.Config().Multizone.Zone.GlobalAddress,
		zone,
		onGlobalToZoneSyncStarted,
		onZoneToGlobalSyncStarted,
		*rt.Config().Multizone.Zone.KDS,
		rt.Config().Experimental,
		rt.Metrics(),
		service.NewEnvoyAdminProcessor(
			rt.ReadOnlyResourceManager(),
			rt.EnvoyAdminClient(),
		),
	)
	return rt.Add(component.NewResilientComponent(
		kdsZoneLog.WithName("kds-mux-client"),
		muxClient,
		rt.Config().General.ResilientComponentBaseBackoff.Duration,
		rt.Config().General.ResilientComponentMaxBackoff.Duration,
	))
}
