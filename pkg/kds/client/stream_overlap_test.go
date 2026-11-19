package client_test

import (
	"context"
	"sync"
	"time"

	envoy_sd "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	common_api "github.com/kumahq/kuma/v3/api/common/v1alpha1"
	kuma_cp "github.com/kumahq/kuma/v3/pkg/config/app/kuma-cp"
	config_core "github.com/kumahq/kuma/v3/pkg/config/core"
	"github.com/kumahq/kuma/v3/pkg/core"
	hostnamegenerator_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/hostnamegenerator/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/v3/pkg/core/resources/manager"
	"github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/core/resources/store"
	"github.com/kumahq/kuma/v3/pkg/events"
	"github.com/kumahq/kuma/v3/pkg/kds"
	kds_client "github.com/kumahq/kuma/v3/pkg/kds/client"
	kds_context "github.com/kumahq/kuma/v3/pkg/kds/context"
	"github.com/kumahq/kuma/v3/pkg/kds/mux"
	kds_server "github.com/kumahq/kuma/v3/pkg/kds/server"
	kds_sync_store "github.com/kumahq/kuma/v3/pkg/kds/store"
	core_metrics "github.com/kumahq/kuma/v3/pkg/metrics"
	meshtimeout_api "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshtimeout/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/v3/pkg/test/grpc"
	"github.com/kumahq/kuma/v3/pkg/test/kds/samples"
	"github.com/kumahq/kuma/v3/pkg/test/kds/setup"
	"github.com/kumahq/kuma/v3/pkg/test/resources/builders"
)

// gatedStore holds every List until the gate opens, like a zone CP whose API
// calls are throttled while it works through its initial KDS responses.
type gatedStore struct {
	store.ResourceStore
	gate <-chan struct{}
}

func (s *gatedStore) List(ctx context.Context, rl model.ResourceList, fs ...store.ListOptionsFunc) error {
	<-s.gate
	return s.ResourceStore.List(ctx, rl, fs...)
}

// withoutFeature drops one KDS feature from the node metadata the zone sends,
// like an older zone CP replica that does not support it.
type withoutFeature struct {
	kds_client.KDSSyncServiceStream
	feature string
}

func (s *withoutFeature) Send(req *envoy_sd.DeltaDiscoveryRequest) error {
	if list := req.GetNode().GetMetadata().GetFields()[kds.MetadataFeatures].GetListValue(); list != nil {
		kept := list.Values[:0]
		for _, v := range list.Values {
			if v.GetStringValue() != s.feature {
				kept = append(kept, v)
			}
		}
		list.Values = kept
	}
	return s.KDSSyncServiceStream.Send(req)
}

// During a zone leader handover two streams of the same zone can be connected to
// the same global for a while. Global must keep their snapshots apart.
var _ = Describe("Zone streams overlapping during leader handover", func() {
	const zoneName = "zone-1"

	var globalStore store.ResourceStore
	var rt interface{ EventBus() events.EventBus }
	var connect func(name string, zoneStore store.ResourceStore, gate <-chan struct{}, missingFeature string) func()

	BeforeEach(func() {
		globalStore = memory.NewStore()
		cfg := kuma_cp.DefaultConfig()
		cfg.Mode = config_core.Global
		watchdogCfg := cfg.Multizone.Global.KDS.EventBasedWatchdog.AsRuntimeConfig()
		watchdogCfg.FlushInterval.Duration = 50 * time.Millisecond
		// keep a broken snapshot in place for the whole test
		watchdogCfg.FullResyncInterval.Duration = time.Hour

		testRt := setup.NewTestRuntime(context.Background(), cfg, globalStore)
		rt = testRt
		kdsCtx := kds_context.DefaultContext(context.Background(), manager.NewResourceManager(globalStore), cfg)
		deltaServer, _, err := kds_server.New(core.Log.WithName("global-kds"), testRt, kdsCtx.TypesSentByGlobal, "global", kdsCtx.GlobalProvidedFilter, kdsCtx.GlobalResourceMapper, time.Second, watchdogCfg)
		Expect(err).ToNot(HaveOccurred())

		connect = func(name string, zoneStore store.ResourceStore, gate <-chan struct{}, missingFeature string) func() {
			metrics, err := core_metrics.NewMetrics("")
			Expect(err).ToNot(HaveOccurred())
			rs := zoneStore
			if gate != nil {
				rs = &gatedStore{ResourceStore: zoneStore, gate: gate}
			}
			syncer, err := kds_sync_store.NewResourceSyncer(core.Log.WithName("zone-syncer").WithName(name), rs, store.NoTransactions{}, metrics, context.Background())
			Expect(err).ToNot(HaveOccurred())

			stop := make(chan struct{})
			wg := &sync.WaitGroup{}
			serverStream := grpc.NewMockDeltaServerStream()
			wg.Go(func() {
				_ = deltaServer.DeltaStreamHandler(mux.NewErrorRecorderStream(serverStream), "")
			})
			cs := serverStream.ClientStream(stop)
			var clientStream kds_client.KDSSyncServiceStream = cs
			if missingFeature != "" {
				clientStream = &withoutFeature{KDSSyncServiceStream: cs, feature: missingFeature}
			}
			wg.Go(func() {
				kdsStream := kds_client.NewDeltaKDSStream(clientStream, zoneName, name, "", len(kdsCtx.TypesSentByGlobal))
				_ = kds_client.NewKDSSyncClient(
					core.Log.WithName("zone").WithName(name),
					kdsCtx.TypesSentByGlobal,
					kdsStream,
					kds_sync_store.ZoneSyncCallback(context.Background(), syncer, false, nil, "kuma-system"),
					kds_client.SyncClientConfig{},
				).Receive()
			})
			var once sync.Once
			closeFn := func() {
				once.Do(func() {
					_ = cs.CloseSend()
					close(stop)
					wg.Wait()
				})
			}
			DeferCleanup(closeFn)
			return closeFn
		}
	})

	meshNames := func(s store.ResourceStore) func(Gomega) []string {
		return func(g Gomega) []string {
			list := mesh.MeshResourceList{}
			g.Expect(s.List(context.Background(), &list)).To(Succeed())
			var out []string
			for _, m := range list.Items {
				out = append(out, m.GetMeta().GetName())
			}
			return out
		}
	}

	Context("old leader's stream closes after the new leader connected", func() {
		var zoneStore store.ResourceStore

		// handover connects replica-B, then replica-A, then closes replica-B's stream.
		// With newLeaderThrottled, replica-A has ACKed nothing when that happens.
		handover := func(newLeaderThrottled bool) {
			closeB := connect("replica-B", zoneStore, nil, "")
			Eventually(meshNames(zoneStore), "5s", "50ms").Should(ConsistOf("default", "service-mesh-dev"))

			throttled := make(chan struct{})
			var gate <-chan struct{}
			if newLeaderThrottled {
				gate = throttled
			}
			connect("replica-A", zoneStore, gate, "")
			time.Sleep(500 * time.Millisecond)
			closeB()
			time.Sleep(200 * time.Millisecond)
			close(throttled)
			time.Sleep(time.Second)
		}

		BeforeEach(func() {
			zoneStore = memory.NewStore()
			for _, name := range []string{"default", "service-mesh-dev"} {
				Expect(globalStore.Create(context.Background(), &mesh.MeshResource{Spec: samples.Mesh1}, store.CreateByKey(name, model.NoMesh))).To(Succeed())
			}
			unrelated := meshtimeout_api.NewMeshTimeoutResource()
			unrelated.Spec = builders.MeshTimeout().
				WithTargetRef(common_api.TargetRef{Kind: common_api.Mesh}).
				AddTo(common_api.TargetRef{Kind: common_api.Mesh}, meshtimeout_api.Conf{IdleTimeout: &kube_meta.Duration{Duration: time.Minute}}).
				Build().Spec
			Expect(globalStore.Create(context.Background(), unrelated, store.CreateByKey("unrelated", model.DefaultMesh))).To(Succeed())
		})

		It("should not delete resources that still exist on global", func() {
			// given
			handover(true)

			// when an unrelated resource changes on global
			rt.EventBus().Send(events.ResourceChangedEvent{Operation: events.Update, Type: meshtimeout_api.MeshTimeoutType, Key: model.ResourceKey{Name: "unrelated", Mesh: model.DefaultMesh}})

			// then
			Consistently(meshNames(zoneStore), "2s", "50ms").Should(ConsistOf("default", "service-mesh-dev"))
		})

		It("should keep delivering changes to the new leader", func() {
			// given
			handover(false)

			// when
			Expect(globalStore.Create(context.Background(), &mesh.MeshResource{Spec: samples.Mesh1}, store.CreateByKey("new-mesh", model.NoMesh))).To(Succeed())
			rt.EventBus().Send(events.ResourceChangedEvent{Operation: events.Create, Type: mesh.MeshType, Key: model.ResourceKey{Name: "new-mesh"}})

			// then
			Eventually(meshNames(zoneStore), "5s", "50ms").Should(ConsistOf("default", "service-mesh-dev", "new-mesh"))
		})
	})

	It("should filter resources by each stream's own features", func() {
		// given a HostnameGenerator selecting MeshMultiZoneServices, only sent to
		// zones that announce the hg-mz-selector feature
		hg := hostnamegenerator_api.NewHostnameGeneratorResource()
		hg.Spec = &hostnamegenerator_api.HostnameGenerator{
			Selector: hostnamegenerator_api.Selector{
				MeshMultiZoneService: &common_api.LabelSelector{MatchLabels: &map[string]string{"team": "a"}},
			},
			Template: "{{ .DisplayName }}.mzsvc.mesh.local",
		}
		Expect(globalStore.Create(context.Background(), hg, store.CreateByKey("mz-hostnames", model.NoMesh))).To(Succeed())
		generators := func(s store.ResourceStore) func(Gomega) int {
			return func(g Gomega) int {
				list := hostnamegenerator_api.HostnameGeneratorResourceList{}
				g.Expect(s.List(context.Background(), &list)).To(Succeed())
				return len(list.Items)
			}
		}

		// and an old replica without the feature, followed by an upgraded one
		oldStore, newStore := memory.NewStore(), memory.NewStore()
		connect("replica-old", oldStore, nil, kds.FeatureHostnameGeneratorMzSelector)
		Consistently(generators(oldStore), "500ms", "50ms").Should(Equal(0))
		connect("replica-new", newStore, nil, "")
		Eventually(generators(newStore), "5s", "20ms").Should(Equal(1))

		// when the generator keeps changing on global
		stop := make(chan struct{})
		defer close(stop)
		go func() {
			ticker := time.NewTicker(200 * time.Millisecond)
			defer ticker.Stop()
			for {
				select {
				case <-stop:
					return
				case <-ticker.C:
					rt.EventBus().Send(events.ResourceChangedEvent{Operation: events.Update, Type: hostnamegenerator_api.HostnameGeneratorType, Key: model.ResourceKey{Name: "mz-hostnames"}})
				}
			}
		}()

		// then each replica keeps its own view
		Consistently(generators(oldStore), "3s", "20ms").Should(Equal(0))
		Consistently(generators(newStore), "1s", "20ms").Should(Equal(1))
	})
})
