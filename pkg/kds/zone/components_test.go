package zone_test

import (
	"context"
	"sync"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	common_api "github.com/kumahq/kuma/v3/api/common/v1alpha1"
	"github.com/kumahq/kuma/v3/api/system/v1alpha1"
	kuma_cp "github.com/kumahq/kuma/v3/pkg/config/app/kuma-cp"
	"github.com/kumahq/kuma/v3/pkg/core"
	config_manager "github.com/kumahq/kuma/v3/pkg/core/config/manager"
	hostnamegenerator_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/hostnamegenerator/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/v3/pkg/core/resources/apis/system"
	workload_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/workload/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/resources/manager"
	"github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/core/resources/registry"
	"github.com/kumahq/kuma/v3/pkg/core/resources/store"
	kds_client "github.com/kumahq/kuma/v3/pkg/kds/client"
	kds_context "github.com/kumahq/kuma/v3/pkg/kds/context"
	"github.com/kumahq/kuma/v3/pkg/kds/mux"
	kds_sync_store "github.com/kumahq/kuma/v3/pkg/kds/store"
	core_metrics "github.com/kumahq/kuma/v3/pkg/metrics"
	meshtrafficpermission_api "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshtrafficpermission/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/v3/pkg/test/grpc"
	"github.com/kumahq/kuma/v3/pkg/test/kds/samples"
	"github.com/kumahq/kuma/v3/pkg/test/kds/setup"
)

var _ = Describe("Zone Sync", func() {
	var zoneStore store.ResourceStore
	var globalStore store.ResourceStore
	var closeFunc func()
	zoneName := "zone-1"

	VerifySyncResourcesFromGlobalToLocal := func() {
		err := globalStore.Create(context.Background(), &mesh.MeshResource{Spec: samples.Mesh1}, store.CreateByKey("mesh-1", model.NoMesh))
		Expect(err).ToNot(HaveOccurred())

		Eventually(func(g Gomega) {
			actual := mesh.MeshResourceList{}
			err := zoneStore.List(context.Background(), &actual)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(actual.Items).To(HaveLen(1))
		}, "5s", "100ms").Should(Succeed())

		actual := mesh.MeshResourceList{}
		err = zoneStore.List(context.Background(), &actual)
		Expect(err).ToNot(HaveOccurred())

		Expect(actual.Items[0].Spec).To(Equal(samples.Mesh1))
	}

	VerifyUpToDateListOfConsumedTypes := func() {
		excludeTypes := map[model.ResourceType]bool{
			mesh.DataplaneInsightType:  true,
			mesh.DataplaneOverviewType: true,
			mesh.DataplaneType:         true,
			workload_api.WorkloadType:  true,
		}

		// take all mesh-scoped types and exclude types that won't be synced
		actualConsumedTypes := registry.Global().ObjectTypes(model.HasScope(model.ScopeMesh), model.TypeFilterFn(func(descriptor model.ResourceTypeDescriptor) bool {
			return !excludeTypes[descriptor.Name]
		}))

		// plus the global-scope types
		extraTypes := []model.ResourceType{
			mesh.MeshType,
			system.ConfigType,
			system.GlobalSecretType,
			hostnamegenerator_api.HostnameGeneratorType,
		}

		actualConsumedTypes = append(actualConsumedTypes, extraTypes...)
		Expect(actualConsumedTypes).To(ConsistOf(registry.Global().ObjectTypes(model.SentFromGlobalToZone())))
	}

	VerifySyncDoesntDeletePredefinedConfigMaps := func() {
		// create kuma-cluster-id ConfigMap in Global
		err := globalStore.Create(context.Background(), &system.ConfigResource{Spec: &v1alpha1.Config{Config: "cluster-id"}},
			store.CreateByKey(config_manager.ClusterIdConfigKey, model.NoMesh))
		Expect(err).ToNot(HaveOccurred())

		// create kuma-cp-leader ConfigMap in Zone
		err = zoneStore.Create(context.Background(), &system.ConfigResource{Spec: &v1alpha1.Config{Config: "leader"}},
			store.CreateByKey("kuma-cp-leader", model.NoMesh))
		Expect(err).ToNot(HaveOccurred())

		// create kuma-control-plane-config ConfigMap in Zone
		err = zoneStore.Create(context.Background(), &system.ConfigResource{Spec: &v1alpha1.Config{Config: "kuma-cp config"}},
			store.CreateByKey("kuma-control-plane-config", model.NoMesh))
		Expect(err).ToNot(HaveOccurred())

		Eventually(func(g Gomega) {
			actual := system.ConfigResourceList{}
			err := zoneStore.List(context.Background(), &actual)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(actual.Items).To(WithTransform(func([]*system.ConfigResource) []string {
				var names []string
				for _, item := range actual.Items {
					names = append(names, item.GetMeta().GetName())
				}
				return names
			}, ConsistOf("kuma-cluster-id", "kuma-cp-leader", "kuma-control-plane-config")))
		}, "5s", "100ms").Should(Succeed())

		actual := system.ConfigResourceList{}
		err = zoneStore.List(context.Background(), &actual)
		Expect(err).ToNot(HaveOccurred())

		actualNames := []string{}
		for _, a := range actual.Items {
			actualNames = append(actualNames, a.GetMeta().GetName())
		}
		expectedNames := []string{
			"kuma-cp-leader",
			"kuma-control-plane-config",
			"kuma-cluster-id",
		}
		Expect(actualNames).To(ConsistOf(expectedNames))
	}

	Context("GlobalToZone", func() {
		var zoneSyncer kds_sync_store.ResourceSyncer

		BeforeEach(func() {
			globalStore = memory.NewStore()
			wg := &sync.WaitGroup{}

			cfg := kuma_cp.DefaultConfig()
			cfg.Multizone.Zone.Name = "global"

			kdsCtx := kds_context.DefaultContext(context.Background(), manager.NewResourceManager(globalStore), cfg)
			srv, err := setup.NewKdsServerBuilder(globalStore).
				WithKdsContext(kdsCtx).
				Delta()
			Expect(err).ToNot(HaveOccurred())
			serverStream := grpc.NewMockDeltaServerStream()
			wg.Add(1)
			go func() {
				defer func() {
					wg.Done()
					GinkgoRecover()
				}()
				errorStream := mux.NewErrorRecorderStream(serverStream)
				Expect(srv.DeltaStreamHandler(errorStream, "")).To(Succeed())
				Expect(errorStream.Err()).ToNot(HaveOccurred())
			}()

			stop := make(chan struct{})
			clientStream := serverStream.ClientStream(stop)

			zoneStore = memory.NewStore()
			metrics, err := core_metrics.NewMetrics("")
			Expect(err).ToNot(HaveOccurred())
			zoneSyncer, err = kds_sync_store.NewResourceSyncer(core.Log.WithName("kds-syncer"), zoneStore, store.NoTransactions{}, metrics, context.Background())
			Expect(err).ToNot(HaveOccurred())

			wg.Add(1)
			go func() {
				defer GinkgoRecover()
				defer wg.Done()
				kdsStream := kds_client.NewDeltaKDSStream(clientStream, zoneName, "global-inst", "", len(kdsCtx.TypesSentByGlobal))
				syncClient := kds_client.NewKDSSyncClient(
					core.Log.WithName("kds-sink"),
					kdsCtx.TypesSentByGlobal,
					kdsStream,
					kds_sync_store.ZoneSyncCallback(context.Background(), zoneSyncer, false, nil, "kuma-system"),
					kds_client.SyncClientConfig{},
				)
				_ = syncClient.Receive()
			}()
			closeFunc = func() {
				defer GinkgoRecover()
				Expect(clientStream.CloseSend()).To(Succeed())
				close(stop)
				wg.Wait()
			}
		})

		AfterEach(func() {
			closeFunc()
		})

		It("should sync policies from global store to the local", func() {
			VerifySyncResourcesFromGlobalToLocal()
		})

		It("should sync policies from global store to the local after resource is valid", func() {
			// a policy without rules does not pass validation, so the zone NACKs it
			invalidPolicy := &meshtrafficpermission_api.MeshTrafficPermission{
				TargetRef: &common_api.TopLevelTargetRef{Kind: common_api.TopLevelTargetRefKindMesh},
			}
			err := globalStore.Create(
				context.Background(),
				&meshtrafficpermission_api.MeshTrafficPermissionResource{Spec: invalidPolicy},
				store.CreateByKey("mtp-1", model.DefaultMesh),
			)
			Expect(err).ToNot(HaveOccurred())

			// should not be synchronized
			Consistently(func(g Gomega) {
				actual := meshtrafficpermission_api.MeshTrafficPermissionResourceList{}
				err := zoneStore.List(context.Background(), &actual)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(actual.Items).To(BeEmpty())
			}, "1s", "100ms").Should(Succeed())

			policy := meshtrafficpermission_api.NewMeshTrafficPermissionResource()
			err = globalStore.Get(context.Background(), policy, store.GetByKey("mtp-1", model.DefaultMesh))
			Expect(err).ToNot(HaveOccurred())

			// when the policy becomes a valid resource
			policy.Spec = samples.MeshTrafficPermission
			err = globalStore.Update(context.Background(), policy)
			Expect(err).ToNot(HaveOccurred())

			// should be synchronized
			Eventually(func(g Gomega) {
				actual := meshtrafficpermission_api.MeshTrafficPermissionResourceList{}
				err := zoneStore.List(context.Background(), &actual)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(actual.Items).To(HaveLen(1))
			}, "5s", "100ms").Should(Succeed())

			actual := meshtrafficpermission_api.MeshTrafficPermissionResourceList{}
			err = zoneStore.List(context.Background(), &actual)
			Expect(err).ToNot(HaveOccurred())

			Expect(actual.Items[0].Spec).To(Equal(samples.MeshTrafficPermission))
		})

		It("should have up to date list of consumed types", func() {
			VerifyUpToDateListOfConsumedTypes()
		})

		It("should not delete predefined ConfigMaps in the Zone cluster", func() {
			VerifySyncDoesntDeletePredefinedConfigMaps()
		})
	})
})
