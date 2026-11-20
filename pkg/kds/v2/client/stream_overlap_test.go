package client_test

import (
	"context"
	"sync"
	"time"

	envoy_sd "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	envoy_stream "github.com/envoyproxy/go-control-plane/pkg/server/stream/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	common_api "github.com/kumahq/kuma/v2/api/common/v1alpha1"
	kuma_cp "github.com/kumahq/kuma/v2/pkg/config/app/kuma-cp"
	config_core "github.com/kumahq/kuma/v2/pkg/config/core"
	"github.com/kumahq/kuma/v2/pkg/core"
	hostnamegenerator_api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/hostnamegenerator/api/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/v2/pkg/core/resources/manager"
	"github.com/kumahq/kuma/v2/pkg/core/resources/model"
	"github.com/kumahq/kuma/v2/pkg/core/resources/store"
	"github.com/kumahq/kuma/v2/pkg/events"
	"github.com/kumahq/kuma/v2/pkg/kds"
	kds_context "github.com/kumahq/kuma/v2/pkg/kds/context"
	"github.com/kumahq/kuma/v2/pkg/kds/mux"
	kds_client "github.com/kumahq/kuma/v2/pkg/kds/v2/client"
	kds_server "github.com/kumahq/kuma/v2/pkg/kds/v2/server"
	kds_sync_store "github.com/kumahq/kuma/v2/pkg/kds/v2/store"
	core_metrics "github.com/kumahq/kuma/v2/pkg/metrics"
	meshtimeout_api "github.com/kumahq/kuma/v2/pkg/plugins/policies/meshtimeout/api/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/v2/pkg/test/grpc"
	"github.com/kumahq/kuma/v2/pkg/test/kds/samples"
	"github.com/kumahq/kuma/v2/pkg/test/kds/setup"
	test_metrics "github.com/kumahq/kuma/v2/pkg/test/metrics"
	"github.com/kumahq/kuma/v2/pkg/test/resources/builders"
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

// observedStream is global's side of a zone stream. It closes meshSent once
// global sent the Meshes and meshACKed once the zone ACKed them.
type observedStream struct {
	envoy_stream.DeltaStream
	meshSent, meshACKed chan struct{}
	sentOnce, ackedOnce sync.Once
}

func (s *observedStream) Send(resp *envoy_sd.DeltaDiscoveryResponse) error {
	err := s.DeltaStream.Send(resp)
	if resp.GetTypeUrl() == string(mesh.MeshType) {
		s.sentOnce.Do(func() { close(s.meshSent) })
	}
	return err
}

func (s *observedStream) Recv() (*envoy_sd.DeltaDiscoveryRequest, error) {
	req, err := s.DeltaStream.Recv()
	if req.GetTypeUrl() == string(mesh.MeshType) && req.GetResponseNonce() != "" {
		s.ackedOnce.Do(func() { close(s.meshACKed) })
	}
	return req, err
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

type streamWrapper func(kds_client.KDSSyncServiceStream) kds_client.KDSSyncServiceStream

// During a zone leader handover two streams of the same zone can be connected to
// the same global for a while. Global must keep their snapshots apart.
var _ = Describe("Zone streams overlapping during leader handover", func() {
	const zoneName = "zone-1"

	var globalStore store.ResourceStore
	var rt interface{ EventBus() events.EventBus }
	var connect func(name string, zoneStore store.ResourceStore, wrap streamWrapper) (func(), *observedStream)
	var fullReconciles func() uint64

	BeforeEach(func() {
		globalStore = memory.NewStore()
		cfg := kuma_cp.DefaultConfig()
		cfg.Mode = config_core.Global
		cfg.Experimental.KDSEventBasedWatchdog.Enabled = true
		cfg.Experimental.KDSEventBasedWatchdog.FlushInterval.Duration = 50 * time.Millisecond
		// keep a broken snapshot in place for the whole test
		cfg.Experimental.KDSEventBasedWatchdog.FullResyncInterval.Duration = time.Hour

		testRt := setup.NewTestRuntime(context.Background(), cfg, globalStore)
		rt = testRt
		kdsCtx := kds_context.DefaultContext(context.Background(), manager.NewResourceManager(globalStore), cfg)
		deltaServer, err := kds_server.New(core.Log.WithName("global-kds"), testRt, kdsCtx.TypesSentByGlobal, "global", 100*time.Millisecond, kdsCtx.GlobalProvidedFilter, kdsCtx.GlobalResourceMapper, time.Second)
		Expect(err).ToNot(HaveOccurred())
		// each stream's watchdog does one full reconcile when it starts
		fullReconciles = func() uint64 {
			var count uint64
			for _, result := range []string{kds_server.ResultChanged, kds_server.ResultNoChanges} {
				if m := test_metrics.FindMetric(testRt.Metrics(), "kds_delta_generation", "reason", kds_server.ReasonResync, "result", result); m != nil {
					count += m.GetSummary().GetSampleCount()
				}
			}
			return count
		}

		connect = func(name string, zoneStore store.ResourceStore, wrap streamWrapper) (func(), *observedStream) {
			metrics, err := core_metrics.NewMetrics("")
			Expect(err).ToNot(HaveOccurred())
			syncer, err := kds_sync_store.NewResourceSyncer(core.Log.WithName("zone-syncer").WithName(name), zoneStore, store.NoTransactions{}, metrics, context.Background())
			Expect(err).ToNot(HaveOccurred())

			stop := make(chan struct{})
			wg := &sync.WaitGroup{}
			serverStream := grpc.NewMockDeltaServerStream()
			observed := &observedStream{DeltaStream: serverStream, meshSent: make(chan struct{}), meshACKed: make(chan struct{})}
			wg.Go(func() {
				_ = deltaServer.DeltaStreamHandler(mux.NewErrorRecorderStream(observed), "")
			})
			cs := serverStream.ClientStream(stop)
			var clientStream kds_client.KDSSyncServiceStream = cs
			if wrap != nil {
				clientStream = wrap(cs)
			}
			wg.Go(func() {
				kdsStream := kds_client.NewDeltaKDSStream(clientStream, zoneName, name, "", len(kdsCtx.TypesSentByGlobal))
				_ = kds_client.NewKDSSyncClient(
					core.Log.WithName("zone").WithName(name),
					kdsCtx.TypesSentByGlobal,
					kdsStream,
					kds_sync_store.ZoneSyncCallback(context.Background(), syncer, false, nil, "kuma-system"),
					0,
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
			return closeFn, observed
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
			closeB, _ := connect("replica-B", zoneStore, nil)
			Eventually(meshNames(zoneStore), "5s", "50ms").Should(ConsistOf("default", "service-mesh-dev"))

			throttled := make(chan struct{})
			replicaAStore := zoneStore
			if newLeaderThrottled {
				replicaAStore = &gatedStore{ResourceStore: zoneStore, gate: throttled}
			}
			reconciles := fullReconciles()
			_, replicaA := connect("replica-A", replicaAStore, nil)
			// When replica-B's stream closes, replica-A must be past its full reconcile
			// and hold the Meshes (throttled) or a Mesh watch (not throttled), or the
			// specs pass without hitting the bug.
			Eventually(fullReconciles, "5s", "10ms").Should(BeNumerically(">", reconciles))
			if newLeaderThrottled {
				Eventually(replicaA.meshSent, "5s", "10ms").Should(BeClosed())
			} else {
				Eventually(replicaA.meshACKed, "5s", "10ms").Should(BeClosed())
			}
			closeB()
			// replica-B's watchdog clears its cache entry after the stream ends
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
		connect("replica-old", oldStore, func(s kds_client.KDSSyncServiceStream) kds_client.KDSSyncServiceStream {
			return &withoutFeature{KDSSyncServiceStream: s, feature: kds.FeatureHostnameGeneratorMzSelector}
		})
		Consistently(generators(oldStore), "500ms", "50ms").Should(Equal(0))
		connect("replica-new", newStore, nil)
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
