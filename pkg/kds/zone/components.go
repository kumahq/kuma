package zone

import (
	"github.com/pkg/errors"

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
	sync_store "github.com/kumahq/kuma/pkg/kds/store"
	"github.com/kumahq/kuma/pkg/kds/util"
	resources_k8s "github.com/kumahq/kuma/pkg/plugins/resources/k8s"
	k8s_model "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/model"
	zone_tokens "github.com/kumahq/kuma/pkg/tokens/builtin/zone"
	"github.com/kumahq/kuma/pkg/tokens/builtin/zoneingress"
)

var (
	kdsZoneLog = core.Log.WithName("kds-zone")
)

func Setup(rt core_runtime.Runtime) error {
	zone := rt.Config().Multizone.Zone.Name
	reg := registry.Global()
	kdsCtx := rt.KDSContext()
	kdsServer, err := kds_server.New(kdsZoneLog, rt, reg.ObjectTypes(model.HasKDSFlag(model.ProvidedByZone)),
		zone, rt.Config().Multizone.Zone.KDS.RefreshInterval,
		kdsCtx.ZoneProvidedFilter, kdsCtx.ZoneResourceMapper, false)
	if err != nil {
		return err
	}
	resourceSyncer := sync_store.NewResourceSyncer(kdsZoneLog, rt.ResourceStore())
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
				log.Error(err, "StreamKumaResources finished with an error")
			}
		}()
		sink := kds_client.NewKDSSink(log, reg.ObjectTypes(model.HasKDSFlag(model.ConsumedByZone)), kds_client.NewKDSStream(session.ClientStream(), zone, string(cfgJson)),
			Callbacks(rt, resourceSyncer, rt.Config().Store.Type == store.KubernetesStore, zone, kubeFactory),
		)
		go func() {
			if err := sink.Receive(); err != nil {
				log.Error(err, "KDSSink finished with an error")
			}
		}()
		return nil
	})
	muxClient := mux.NewClient(
		rt.Config().Multizone.Zone.GlobalAddress,
		zone,
		onSessionStarted,
		*rt.Config().Multizone.Zone.KDS,
		rt.Metrics(),
		rt.KDSContext().ZoneClientCtx,
	)
	return rt.Add(component.NewResilientComponent(kdsZoneLog.WithName("kds-mux-client"), muxClient))
}

func Callbacks(rt core_runtime.Runtime, syncer sync_store.ResourceSyncer, k8sStore bool, localZone string, kubeFactory resources_k8s.KubeFactory) *kds_client.Callbacks {
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
				if kubeObject.Scope() == k8s_model.ScopeNamespace {
					util.AddSuffixToNames(rs.GetItems(), "default")
				}
			}
			if rs.GetItemType() == mesh.ZoneIngressType {
				return syncer.Sync(rs, sync_store.PrefilterBy(func(r model.Resource) bool {
					return r.(*mesh.ZoneIngressResource).IsRemoteIngress(localZone)
				}))
			}
			if rs.GetItemType() == system.ConfigType {
				return syncer.Sync(rs, sync_store.PrefilterBy(func(r model.Resource) bool {
					return rt.KDSContext().Configs[r.GetMeta().GetName()]
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
