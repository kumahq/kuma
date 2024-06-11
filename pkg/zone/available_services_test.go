package zone_test

import (
	"context"
	"net"
	"time"

	"github.com/google/go-cmp/cmp"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/proto"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	config_manager "github.com/kumahq/kuma/pkg/core/config/manager"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/dns/vips"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	cache_mesh "github.com/kumahq/kuma/pkg/xds/cache/mesh"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	"github.com/kumahq/kuma/pkg/xds/server"
	"github.com/kumahq/kuma/pkg/zone"
)

var _ = Describe("AvailableServices", func() {
	var resManager manager.ResourceManager
	var meshContextBuilder xds_context.MeshContextBuilder
	var metrics core_metrics.Metrics
	var meshCache *cache_mesh.Cache

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
	})
	AfterEach(func() {
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
		mesh := core_mesh.NewMeshResource()
		mesh.Spec = &mesh_proto.Mesh{
			Mtls: &mesh_proto.Mesh_Mtls{
				EnabledBackend: "builtin",
				Backends: []*mesh_proto.CertificateAuthorityBackend{
					{
						Name: "builtin",
						Type: "builtin",
					},
				},
			},
			Routing: &mesh_proto.Routing{
				ZoneEgress: true,
			},
		}
		Expect(resManager.Create(context.Background(), mesh, store.CreateByKey("mesh1", ""))).To(Succeed())
		externalService := &core_mesh.ExternalServiceResource{
			Meta: &test_model.ResourceMeta{
				Mesh: "mesh1",
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
		Expect(resManager.Create(context.Background(), externalService, store.CreateByKey("es-1", "mesh1"))).To(Succeed())
		dp := &core_mesh.DataplaneResource{
			Spec: &mesh_proto.Dataplane{
				Networking: &mesh_proto.Dataplane_Networking{
					Address: "127.0.0.1",
					Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
						{
							Port: 8000,
							Tags: map[string]string{
								"kuma.io/service": "backend",
								"version":         "v1",
								"region":          "eu",
							},
						},
					},
				},
			},
		}
		Expect(resManager.Create(context.Background(), dp, store.CreateByKey("dp-1", "mesh1"))).To(Succeed())
		dp = &core_mesh.DataplaneResource{
			Spec: &mesh_proto.Dataplane{
				Networking: &mesh_proto.Dataplane_Networking{
					Address: "127.0.0.1",
					Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
						{
							Port: 8000,
							Tags: map[string]string{
								"kuma.io/service": "web",
								"version":         "v2",
								"region":          "us",
							},
						},
					},
				},
			},
		}
		Expect(resManager.Create(context.Background(), dp, store.CreateByKey("dp-2", "mesh1"))).To(Succeed())
		dp = &core_mesh.DataplaneResource{
			Spec: &mesh_proto.Dataplane{
				Networking: &mesh_proto.Dataplane_Networking{
					Address: "127.0.0.1",
					Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
						{
							Port: 8000,
							Tags: map[string]string{
								"kuma.io/service": "web",
								"version":         "v2",
								"region":          "us",
							},
						},
					},
				},
			},
		}
		Expect(resManager.Create(context.Background(), dp, store.CreateByKey("dp-3", "mesh1"))).To(Succeed())

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

		stop := make(chan struct{})
		done = make(chan struct{})
		go func() {
			defer GinkgoRecover()
			Expect(tracker.Start(stop)).To(Succeed())
			close(done)
		}()
		defer close(stop)

		expected := []*mesh_proto.ZoneIngress_AvailableService{
			{
				Instances: 1,
				Tags: map[string]string{
					"kuma.io/service": "backend",
					"version":         "v1",
					"region":          "eu",
				},
				Mesh: "mesh1",
			},
			{
				Instances: 2,
				Tags: map[string]string{
					"kuma.io/service": "web",
					"version":         "v2",
					"region":          "us",
				},
				Mesh: "mesh1",
			},
			{
				Instances: 1,
				Tags: map[string]string{
					"kuma.io/service":  "httpbin",
					"version":          "v1",
					mesh_proto.ZoneTag: "zone",
				},
				Mesh:            "mesh1",
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
