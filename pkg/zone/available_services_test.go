package zone_test

import (
	"context"
	"errors"
	"net"
	"sync/atomic"
	"time"

	"github.com/google/go-cmp/cmp"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/proto"

	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/core"
	config_manager "github.com/kumahq/kuma/v2/pkg/core/config/manager"
	core_mesh "github.com/kumahq/kuma/v2/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/v2/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/v2/pkg/core/resources/model"
	"github.com/kumahq/kuma/v2/pkg/core/resources/store"
	"github.com/kumahq/kuma/v2/pkg/dns/vips"
	core_metrics "github.com/kumahq/kuma/v2/pkg/metrics"
	"github.com/kumahq/kuma/v2/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/v2/pkg/test/resources/builders"
	test_model "github.com/kumahq/kuma/v2/pkg/test/resources/model"
	"github.com/kumahq/kuma/v2/pkg/test/resources/samples"
	cache_mesh "github.com/kumahq/kuma/v2/pkg/xds/cache/mesh"
	xds_context "github.com/kumahq/kuma/v2/pkg/xds/context"
	"github.com/kumahq/kuma/v2/pkg/xds/server"
	"github.com/kumahq/kuma/v2/pkg/zone"
)

var _ = Describe("AvailableServices Tracker", func() {
	Context("Enabled Available Services", func() {
		var resManager manager.ResourceManager
		var meshContextBuilder xds_context.MeshContextBuilder
		var metrics core_metrics.Metrics
		var meshCache *cache_mesh.Cache

		var stop chan struct{}
		var done chan struct{}
		BeforeEach(func() {
			resourceStore := memory.NewStore()
			resManager = manager.NewResourceManager(resourceStore)

			Expect(samples.MeshMTLSBuilder().WithEgressRoutingEnabled().Create(resManager)).To(Succeed())

			meshContextBuilder = xds_context.NewMeshContextBuilder(
				resManager,
				server.MeshResourceTypes(),
				net.LookupIP,
				"zone",
				vips.NewPersistence(resManager, config_manager.NewConfigManager(resourceStore), false),
				".mesh",
				80,
				xds_context.AnyToAnyReachableServicesGraphBuilder,
				nil,
			)
			var err error
			metrics, err = core_metrics.NewMetrics("Zone")
			Expect(err).ToNot(HaveOccurred())

			meshCache, err = cache_mesh.NewCache(
				1*time.Second,
				meshContextBuilder,
				metrics,
			)
			Expect(err).ToNot(HaveOccurred())

			tracker, err := zone.NewZoneAvailableServicesTracker(
				core.Log.WithName("test"),
				metrics,
				resManager,
				meshCache,
				20*time.Millisecond,
				nil,
				"zone",
			)
			Expect(err).ToNot(HaveOccurred())

			stop = make(chan struct{})
			done = make(chan struct{})
			go func() {
				defer GinkgoRecover()
				Expect(tracker.Start(stop)).To(Succeed())
				close(done)
			}()
		})
		AfterEach(func() {
			close(stop)
			Eventually(done).Should(BeClosed())
		})
		It("should update all ZoneIngresses", func() {
			Expect(builders.ZoneIngress().Create(resManager)).To(Succeed())
			externalService := &core_mesh.ExternalServiceResource{
				Meta: &test_model.ResourceMeta{
					Mesh: "default",
					Name: "es-1",
				},
				Spec: &mesh_proto.ExternalService{
					Networking: &mesh_proto.ExternalService_Networking{
						Address: "127.0.0.1:80",
					},
					Tags: map[string]string{
						"kuma.io/service":  "httpbin",
						"version":          "v1",
						mesh_proto.ZoneTag: "zone",
					},
				},
			}
			Expect(resManager.Create(context.Background(), externalService, store.CreateByKey("es-1", core_model.DefaultMesh))).To(Succeed())
			Expect(samples.DataplaneBackendBuilder().Create(resManager)).To(Succeed())
			Expect(samples.DataplaneWebBuilder().Create(resManager)).To(Succeed())

			expected := []*mesh_proto.ZoneIngress_AvailableService{
				{
					Instances: 1,
					Tags: map[string]string{
						"kuma.io/service":  "web",
						"kuma.io/protocol": "http",
					},
					Mesh: core_model.DefaultMesh,
				},
				{
					Instances: 1,
					Tags: map[string]string{
						"kuma.io/service": "backend",
					},
					Mesh: core_model.DefaultMesh,
				},
				{
					Instances: 1,
					Tags: map[string]string{
						"kuma.io/service":  "httpbin",
						"version":          "v1",
						mesh_proto.ZoneTag: "zone",
					},
					Mesh:            core_model.DefaultMesh,
					ExternalService: true,
				},
			}
			Eventually(func(g Gomega) {
				zi := core_mesh.NewZoneIngressResource()
				g.Expect(resManager.Get(context.Background(), zi, store.GetByKey("zoneingress-1", ""))).To(Succeed())
				g.Expect(zi.Spec.AvailableServices).To(BeComparableTo(expected, cmp.Comparer(proto.Equal)))
			}).Should(Succeed())
		})
	})

	Context("Disabled Available Services", func() {
		var resManager manager.ResourceManager
		var stop chan struct{}

		BeforeEach(func() {
			resourceStore := memory.NewStore()
			resManager = manager.NewResourceManager(resourceStore)
			var err error
			metrics, err := core_metrics.NewMetrics("Zone")
			Expect(err).ToNot(HaveOccurred())

			Expect(samples.MeshMTLSBuilder().WithEgressRoutingEnabled().Create(resManager)).To(Succeed())

			meshContextBuilder := xds_context.NewMeshContextBuilder(
				resManager,
				server.MeshResourceTypes(),
				net.LookupIP,
				"zone",
				vips.NewPersistence(resManager, config_manager.NewConfigManager(resourceStore), false),
				".mesh",
				80,
				xds_context.AnyToAnyReachableServicesGraphBuilder,
				nil,
			)
			meshCache, err := cache_mesh.NewCache(
				1*time.Second,
				meshContextBuilder,
				metrics,
			)
			Expect(err).ToNot(HaveOccurred())

			tracker, err := zone.NewZoneAvailableServicesTracker(
				core.Log.WithName("test"),
				metrics,
				resManager,
				meshCache, // not used when available services are disabled
				20*time.Millisecond,
				nil,
				"zone",
			)
			Expect(err).ToNot(HaveOccurred())

			stop = make(chan struct{})
			go func() {
				defer GinkgoRecover()
				Expect(tracker.Start(stop)).To(Succeed())
			}()
		})

		AfterEach(func() {
			close(stop)
		})

		It("should clear available services list when available services are disabled", func() {
			Expect(builders.ZoneIngress().Create(resManager)).To(Succeed())
			externalService := &core_mesh.ExternalServiceResource{
				Meta: &test_model.ResourceMeta{
					Mesh: "default",
					Name: "es-1",
				},
				Spec: &mesh_proto.ExternalService{
					Networking: &mesh_proto.ExternalService_Networking{
						Address: "127.0.0.1:80",
					},
					Tags: map[string]string{
						"kuma.io/service":  "httpbin",
						"version":          "v1",
						mesh_proto.ZoneTag: "zone",
					},
				},
			}
			Expect(resManager.Create(context.Background(), externalService, store.CreateByKey("es-1", core_model.DefaultMesh))).To(Succeed())
			Expect(samples.DataplaneBackendBuilder().Create(resManager)).To(Succeed())
			Expect(samples.DataplaneWebBuilder().Create(resManager)).To(Succeed())

			Eventually(func(g Gomega) {
				zi := core_mesh.NewZoneIngressResource()
				g.Expect(resManager.Get(context.Background(), zi, store.GetByKey("zoneingress-1", ""))).To(Succeed())
				g.Expect(zi.Spec.AvailableServices).To(BeEmpty())
			}).Should(Succeed())
		})
	})

	It("should not modify a listed ZoneIngress spec when the update is rejected", func() {
		resourceStore := memory.NewStore()
		resManager := manager.NewResourceManager(resourceStore)
		Expect(samples.MeshDefaultBuilder().Create(resManager)).To(Succeed())
		Expect(samples.DataplaneBackendBuilder().Create(resManager)).To(Succeed())
		metrics, err := core_metrics.NewMetrics("Zone")
		Expect(err).ToNot(HaveOccurred())
		meshCache, err := cache_mesh.NewCache(1*time.Second, xds_context.NewMeshContextBuilder(
			resManager,
			server.MeshResourceTypes(),
			net.LookupIP,
			"zone",
			vips.NewPersistence(resManager, config_manager.NewConfigManager(resourceStore), false),
			".mesh",
			80,
			xds_context.AnyToAnyReachableServicesGraphBuilder,
			nil,
		), metrics)
		Expect(err).ToNot(HaveOccurred())
		rejecting := &rejectingManager{ResourceManager: resManager, spec: builders.ZoneIngress().Build().Spec}
		tracker, err := zone.NewZoneAvailableServicesTracker(core.Log.WithName("test"), metrics, rejecting, meshCache, 20*time.Millisecond, nil, "zone")
		Expect(err).ToNot(HaveOccurred())

		stop := make(chan struct{})
		done := make(chan struct{})
		go func() {
			defer GinkgoRecover()
			Expect(tracker.Start(stop)).To(Succeed())
			close(done)
		}()
		defer func() {
			close(stop)
			Eventually(done).Should(BeClosed())
		}()

		Eventually(rejecting.updates.Load).Should(BeNumerically(">", 1))
		Expect(rejecting.spec.AvailableServices).To(BeEmpty())
	})
})

// rejectingManager mimics the Kubernetes store's conversion cache: every List
// returns a new ZoneIngress around the same spec, and every Update is rejected
type rejectingManager struct {
	manager.ResourceManager
	spec    *mesh_proto.ZoneIngress
	updates atomic.Int32
}

func (m *rejectingManager) List(ctx context.Context, list core_model.ResourceList, fs ...store.ListOptionsFunc) error {
	zis, ok := list.(*core_mesh.ZoneIngressResourceList)
	if !ok {
		return m.ResourceManager.List(ctx, list, fs...)
	}
	zi := core_mesh.NewZoneIngressResource()
	zi.SetMeta(&test_model.ResourceMeta{Name: "zi-1"})
	zi.Spec = m.spec
	return zis.AddItem(zi)
}

func (m *rejectingManager) Update(context.Context, core_model.Resource, ...store.UpdateOptionsFunc) error {
	m.updates.Add(1)
	return errors.New("rejected")
}
