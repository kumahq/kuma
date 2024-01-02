package zone

import (
	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/config"
	"github.com/kumahq/kuma/pkg/config/core/resources/store"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	core_runtime "github.com/kumahq/kuma/pkg/core/runtime"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	kds_client "github.com/kumahq/kuma/pkg/kds/client"
	"github.com/kumahq/kuma/pkg/kds/mux"
	kds_server "github.com/kumahq/kuma/pkg/kds/server"
	"github.com/kumahq/kuma/pkg/kds/service"
	sync_store "github.com/kumahq/kuma/pkg/kds/store"
	"github.com/kumahq/kuma/pkg/kds/util"
	kds_client_v2 "github.com/kumahq/kuma/pkg/kds/v2/client"
	kds_server_v2 "github.com/kumahq/kuma/pkg/kds/v2/server"
	kds_sync_store_v2 "github.com/kumahq/kuma/pkg/kds/v2/store"
	resources_k8s "github.com/kumahq/kuma/pkg/plugins/resources/k8s"
	k8s_model "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/model"
	zone_tokens "github.com/kumahq/kuma/pkg/tokens/builtin/zone"
	"github.com/kumahq/kuma/pkg/tokens/builtin/zoneingress"
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
	kdsServer, err := kds_server.New(
		kdsZoneLog,
		rt,
		reg.ObjectTypes(model.HasKDSFlag(model.ProvidedByZone)),
		zone,
		rt.Config().Multizone.Zone.KDS.RefreshInterval.Duration,
		kdsCtx.ZoneProvidedFilter,
		kdsCtx.ZoneResourceMapper,
		rt.Config().Multizone.Zone.KDS.NackBackoff.Duration,
	)
	if err != nil {
		return err
	}

	kdsServerV2, err := kds_server_v2.New(
		kdsZoneLog,
		rt,
		reg.ObjectTypes(model.HasKDSFlag(model.ProvidedByZone)),
		zone,
		rt.Config().Multizone.Zone.KDS.RefreshInterval.Duration,
		kdsCtx.ZoneProvidedFilter,
		kdsCtx.ZoneResourceMapper,
		rt.Config().Multizone.Zone.KDS.NackBackoff.Duration,
	)
	if err != nil {
		return err
	}
	resourceSyncer := sync_store.NewResourceSyncer(kdsZoneLog, rt.ResourceStore())
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
	onSessionStarted := mux.OnSessionStartedFunc(func(session mux.Session) error {
		log := kdsZoneLog.WithValues("peer-id", session.PeerID())
		log.Info("new session created")
		go func() {
			if err := kdsServer.StreamKumaResources(session.ServerStream()); err != nil {
				session.SetError(errors.Wrap(err, "StreamKumaResources finished with an error"))
			} else {
				log.V(1).Info("StreamKumaResources finished gracefully")
			}
		}()
		sink := kds_client.NewKDSSink(log, reg.ObjectTypes(model.HasKDSFlag(model.ConsumedByZone)), kds_client.NewKDSStream(session.ClientStream(), zone, string(cfgJson)),
			Callbacks(
				rt.KDSContext().Configs,
				resourceSyncer,
				rt.Config().Store.Type == store.KubernetesStore,
				zone,
				kubeFactory,
				rt.Config().Store.Kubernetes.SystemNamespace,
			),
		)
		go func() {
			if err := sink.Receive(); err != nil {
				session.SetError(errors.Wrap(err, "KDSSink finished with an error"))
			} else {
				log.V(1).Info("KDSSink finished gracefully")
			}
		}()
		return nil
	})

	onGlobalToZoneSyncStarted := mux.OnGlobalToZoneSyncStartedFunc(func(stream mesh_proto.KDSSyncService_GlobalToZoneSyncClient, errChan chan error) {
		log := kdsDeltaZoneLog.WithValues("kds-version", "v2")
		syncClient := kds_client_v2.NewKDSSyncClient(
			log,
			reg.ObjectTypes(model.HasKDSFlag(model.ConsumedByZone)),
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
			if err := syncClient.Receive(); err != nil {
				errChan <- errors.Wrap(err, "GlobalToZoneSyncClient finished with an error")
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
			if err := kdsServerV2.ZoneToGlobal(session); err != nil {
				errChan <- errors.Wrap(err, "ZoneToGlobalSync finished with an error")
			} else {
				log.V(1).Info("ZoneToGlobalSync finished gracefully")
			}
		}()
	})

	muxClient := mux.NewClient(
		rt.KDSContext().ZoneClientCtx,
		rt.Config().Multizone.Zone.GlobalAddress,
		zone,
		onSessionStarted,
		onGlobalToZoneSyncStarted,
		onZoneToGlobalSyncStarted,
		*rt.Config().Multizone.Zone.KDS,
		rt.Config().Experimental,
		rt.Metrics(),
		service.NewEnvoyAdminProcessor(
			rt.ReadOnlyResourceManager(),
			rt.EnvoyAdminClient().ConfigDump,
			rt.EnvoyAdminClient().Stats,
			rt.EnvoyAdminClient().Clusters,
		),
	)
	return rt.Add(component.NewResilientComponent(kdsZoneLog.WithName("kds-mux-client"), muxClient))
}

func Callbacks(
	configToSync map[string]bool,
	syncer sync_store.ResourceSyncer,
	k8sStore bool,
	localZone string,
	kubeFactory resources_k8s.KubeFactory,
	systemNamespace string,
) *kds_client.Callbacks {
	return &kds_client.Callbacks{
		OnResourcesReceived: func(clusterID string, rs model.ResourceList) error {
			if k8sStore && rs.GetItemType() != system.ConfigType && rs.GetItemType() != system.SecretType && rs.GetItemType() != system.GlobalSecretType {
				// if type of Store is Kubernetes then we want to store upstream resources in dedicated Namespace.
				// KubernetesStore parses Name and considers substring after the last dot as a Namespace's Name.
				// System resources are not in the kubeFactory therefore we need explicit ifs for them
				kubeObject, err := kubeFactory.NewObject(rs.NewItem())
				if err != nil {
					return errors.Wrap(err, "could not convert object")
				}
				desc := rs.NewItem().Descriptor()
				if kubeObject.Scope() == k8s_model.ScopeNamespace {
					if desc.IsPluginOriginated {
						util.AddSuffixToNames(rs.GetItems(), systemNamespace)
					} else {
						util.AddSuffixToNames(rs.GetItems(), "default")
					}
				}
			}
			if rs.GetItemType() == mesh.ZoneIngressType {
				return syncer.Sync(rs, sync_store.PrefilterBy(func(r model.Resource) bool {
					return r.(*mesh.ZoneIngressResource).IsRemoteIngress(localZone)
				}))
			}
			if rs.GetItemType() == system.ConfigType {
				return syncer.Sync(rs, sync_store.PrefilterBy(func(r model.Resource) bool {
					return configToSync[r.GetMeta().GetName()]
				}))
			}
			if rs.GetItemType() == system.GlobalSecretType {
				return syncer.Sync(rs, sync_store.PrefilterBy(func(r model.Resource) bool {
					return util.ResourceNameHasAtLeastOneOfPrefixes(
						r.GetMeta().GetName(),
						zoneingress.ZoneIngressSigningKeyPrefix,
						zone_tokens.SigningPublicKeyPrefix,
					)
				}))
			}
			return syncer.Sync(rs)
		},
	}
}
