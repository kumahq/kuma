package zone_test

import (
	"context"
	"net"
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
	test_model "github.com/kumahq/kuma/v2/pkg/test/resources/model"
	"github.com/kumahq/kuma/v2/pkg/test/resources/samples"
	cache_mesh "github.com/kumahq/kuma/v2/pkg/xds/cache/mesh"
	xds_context "github.com/kumahq/kuma/v2/pkg/xds/context"
	"github.com/kumahq/kuma/v2/pkg/xds/server"
	"github.com/kumahq/kuma/v2/pkg/zone"
)

var _ = Describe("AvailableServices", func() {
	var resManager manager.ResourceManager
	var meshContextBuilder xds_context.MeshContextBuilder
	var metrics core_metrics.Metrics
	var meshCache *cache_mesh.Cache

	var stop chan struct{}
	var done chan struct{}
	BeforeEach(func() {
		resourceStore := memory.NewStore()
		resManager = manager.NewResourceManager(resourceStore)
		meshContextBuilder = xds_context.NewMeshContextBuilder(
			resManager,
			server.MeshResourceTypes(),
			net.LookupIP,
			"zone",
			vips.NewPersistence(resManager, config_manager.NewConfigManager(resourceStore), false),
			".mesh",
			80,
			xds_context.AnyToAnyReachableServicesGraphBuilder,
			false,
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
		ingress := &core_mesh.ZoneIngressResource{
			Spec: &mesh_proto.ZoneIngress{
				Networking: &mesh_proto.ZoneIngress_Networking{
					Port:    10000,
					Address: "127.0.0.1",
				},
			},
		}
		Expect(resManager.Create(context.Background(), ingress, store.CreateByKey("ingress-1", ""))).To(Succeed())
		Expect(samples.MeshMTLSBuilder().WithEgressRoutingEnabled().Create(resManager)).To(Succeed())
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
			g.Expect(resManager.Get(context.Background(), zi, store.GetByKey("ingress-1", ""))).To(Succeed())
			g.Expect(zi.Spec.AvailableServices).To(BeComparableTo(expected, cmp.Comparer(proto.Equal)))
		}).Should(Succeed())
	})
})
