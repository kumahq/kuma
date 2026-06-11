package generate_test

import (
	"context"
	"maps"
	"net"
	"sync/atomic"
	"time"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/util/intstr"

	common_api "github.com/kumahq/kuma/v2/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	kuma_cp "github.com/kumahq/kuma/v2/pkg/config/app/kuma-cp"
	config_manager "github.com/kumahq/kuma/v2/pkg/core/config/manager"
	core_meta "github.com/kumahq/kuma/v2/pkg/core/metadata"
	core_mesh "github.com/kumahq/kuma/v2/pkg/core/resources/apis/mesh"
	meshservice_api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshservice/api/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshservice/generate"
	"github.com/kumahq/kuma/v2/pkg/core/resources/manager"
	"github.com/kumahq/kuma/v2/pkg/core/resources/model"
	"github.com/kumahq/kuma/v2/pkg/core/resources/store"
	"github.com/kumahq/kuma/v2/pkg/dns/vips"
	core_metrics "github.com/kumahq/kuma/v2/pkg/metrics"
	"github.com/kumahq/kuma/v2/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/v2/pkg/plugins/runtime/k8s/metadata"
	test_metrics "github.com/kumahq/kuma/v2/pkg/test/metrics"
	"github.com/kumahq/kuma/v2/pkg/test/resources/builders"
	"github.com/kumahq/kuma/v2/pkg/test/resources/samples"
	"github.com/kumahq/kuma/v2/pkg/util/pointer"
	cache_mesh "github.com/kumahq/kuma/v2/pkg/xds/cache/mesh"
	xds_context "github.com/kumahq/kuma/v2/pkg/xds/context"
	"github.com/kumahq/kuma/v2/pkg/xds/server"
)

type countingResourceManager struct {
	manager.ResourceManager
	updates atomic.Int32
	creates atomic.Int32
}

func (c *countingResourceManager) Update(ctx context.Context, r model.Resource, fs ...store.UpdateOptionsFunc) error {
	c.updates.Add(1)
	return c.ResourceManager.Update(ctx, r, fs...)
}

func (c *countingResourceManager) Create(ctx context.Context, r model.Resource, fs ...store.CreateOptionsFunc) error {
	c.creates.Add(1)
	return c.ResourceManager.Create(ctx, r, fs...)
}

var _ = Describe("MeshService generator", func() {
	var stopCh chan struct{}
	var resManager manager.ResourceManager
	var meshContextBuilder xds_context.MeshContextBuilder
	var metrics core_metrics.Metrics

	gracePeriodInterval := 500 * time.Millisecond

	BeforeEach(func() {
		m, err := core_metrics.NewMetrics("")
		Expect(err).ToNot(HaveOccurred())
		metrics = m
		store := memory.NewStore()
		resManager = manager.NewResourceManager(store)
		meshContextBuilder = xds_context.NewMeshContextBuilder(
			resManager,
			server.MeshResourceTypes(),
			net.LookupIP,
			"zone",
			vips.NewPersistence(resManager, config_manager.NewConfigManager(store), false),
			".mesh",
			80,
			xds_context.AnyToAnyReachableServicesGraphBuilder,
			nil,
		)
		meshCache, err := cache_mesh.NewCache(
			100*time.Millisecond,
			meshContextBuilder,
			metrics,
		)
		Expect(err).ToNot(HaveOccurred())
		allocator, err := generate.New(
			logr.Discard(),
			50*time.Millisecond,
			gracePeriodInterval,
			metrics,
			resManager,
			meshCache,
			"zone",
			false,
			kuma_cp.MeshServiceLabelPropagation{},
		)

		Expect(err).ToNot(HaveOccurred())
		stopCh = make(chan struct{})
		ch := stopCh
		go func() {
			defer GinkgoRecover()
			Expect(allocator.Start(ch)).To(Succeed())
		}()

		Expect(
			samples.MeshDefaultBuilder().WithMeshServicesEnabled(mesh_proto.Mesh_MeshServices_Everywhere).Create(resManager),
		).To(Succeed())
	})

	AfterEach(func() {
		close(stopCh)
	})

	It("should generate MeshService from a single Dataplane", func() {
		err := samples.DataplaneBackendBuilder().Create(resManager)
		Expect(err).ToNot(HaveOccurred())

		Eventually(func(g Gomega) {
			ms := meshservice_api.NewMeshServiceResource()
			g.Expect(resManager.Get(context.Background(), ms, store.GetByKey("backend", model.DefaultMesh))).To(Succeed())
			g.Expect(ms.Spec.Ports).To(Equal([]meshservice_api.Port{
				{
					Name:        pointer.To("80"),
					Port:        80,
					TargetPort:  pointer.To(intstr.FromInt(80)),
					AppProtocol: core_meta.ProtocolTCP,
				},
			}))
		}, "2s", "100ms").Should(Succeed())
	})

	Context("should not generate MeshService from a Dataplane with not supported kuma.io/service", func() {
		It("kuma.io/service with underscore sign", func() {
			err := builders.Dataplane().WithAddress("192.168.0.1").WithServices("backend_svc").Create(resManager)
			Expect(err).ToNot(HaveOccurred())

			Consistently(func(g Gomega) {
				mss := &meshservice_api.MeshServiceResourceList{}
				g.Expect(resManager.List(context.Background(), mss)).To(Succeed())
				g.Expect(mss.GetItems()).To(BeEmpty())
			}, "1s", "100ms").Should(Succeed())
		})
		It("kuma.io/service with dot sign", func() {
			err := builders.Dataplane().WithAddress("192.168.0.1").WithServices("backend.svc").Create(resManager)
			Expect(err).ToNot(HaveOccurred())

			Consistently(func(g Gomega) {
				mss := &meshservice_api.MeshServiceResourceList{}
				g.Expect(resManager.List(context.Background(), mss)).To(Succeed())
				g.Expect(mss.GetItems()).To(BeEmpty())
			}, "1s", "100ms").Should(Succeed())
		})
		It("kuma.io/service with an alphanumeric started character", func() {
			err := builders.Dataplane().WithAddress("192.168.0.1").WithServices("00-backend-svc").Create(resManager)
			Expect(err).ToNot(HaveOccurred())

			Consistently(func(g Gomega) {
				mss := &meshservice_api.MeshServiceResourceList{}
				g.Expect(resManager.List(context.Background(), mss)).To(Succeed())
				g.Expect(mss.GetItems()).To(BeEmpty())
			}, "1s", "100ms").Should(Succeed())
		})
		It("kuma.io/service with too long name", func() {
			// 64 chars
			err := builders.Dataplane().WithAddress("192.168.0.1").WithServices("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa").Create(resManager)
			Expect(err).ToNot(HaveOccurred())

			Consistently(func(g Gomega) {
				mss := &meshservice_api.MeshServiceResourceList{}
				g.Expect(resManager.List(context.Background(), mss)).To(Succeed())
				g.Expect(mss.GetItems()).To(BeEmpty())
			}, "1s", "100ms").Should(Succeed())
		})
	})

	It("should generate MeshService from a single Dataplane with inbound name", func() {
		err := samples.DataplaneBackendBuilder().WithoutInbounds().
			AddInbound(
				builders.Inbound().
					WithPort(builders.FirstInboundPort).
					WithServicePort(builders.FirstInboundServicePort).
					WithName("main").
					WithTags(map[string]string{mesh_proto.ServiceTag: "backend", mesh_proto.ProtocolTag: "tcp"}),
			).Create(resManager)
		Expect(err).ToNot(HaveOccurred())

		Eventually(func(g Gomega) {
			ms := meshservice_api.NewMeshServiceResource()
			g.Expect(resManager.Get(context.Background(), ms, store.GetByKey("backend", model.DefaultMesh))).To(Succeed())
			g.Expect(ms.Spec.Ports).To(Equal([]meshservice_api.Port{
				{
					Name:        pointer.To("main"),
					Port:        80,
					TargetPort:  pointer.To(intstr.FromInt(80)),
					AppProtocol: core_meta.ProtocolTCP,
				},
			}))
		}, "2s", "100ms").Should(Succeed())
	})

	It("should not change MeshService if a conflicting Dataplanes appears", func() {
		err := samples.DataplaneBackendBuilder().Create(resManager)
		Expect(err).ToNot(HaveOccurred())

		ms := meshservice_api.NewMeshServiceResource()

		Eventually(func(g Gomega) {
			g.Expect(resManager.Get(context.Background(), ms, store.GetByKey("backend", model.DefaultMesh))).To(Succeed())
			g.Expect(ms.Spec.Ports).To(Equal([]meshservice_api.Port{
				{
					Name:        pointer.To("80"),
					Port:        80,
					TargetPort:  pointer.To(intstr.FromInt(80)),
					AppProtocol: core_meta.ProtocolTCP,
				},
			}))
		}, "2s", "100ms").Should(Succeed())

		err = samples.DataplaneBackendBuilder().WithName("dp-2").
			AddInbound(
				builders.Inbound().
					WithPort(81).
					WithServicePort(81).
					WithTags(map[string]string{
						mesh_proto.ServiceTag: "backend",
					}),
			).
			Create(resManager)
		Expect(err).ToNot(HaveOccurred())

		Consistently(func(g Gomega) {
			g.Expect(resManager.Get(context.Background(), ms, store.GetByKey("backend", model.DefaultMesh))).To(Succeed())
			g.Expect(ms.Spec.Ports).To(Equal([]meshservice_api.Port{
				{
					Name:        pointer.To("80"),
					Port:        80,
					TargetPort:  pointer.To(intstr.FromInt(80)),
					AppProtocol: core_meta.ProtocolTCP,
				},
			}))
		}, "2s", "100ms").Should(Succeed())
	})

	It("should allow MeshService to be changed if all Dataplanes change", func() {
		err := samples.DataplaneBackendBuilder().Create(resManager)
		Expect(err).ToNot(HaveOccurred())

		ms := meshservice_api.NewMeshServiceResource()

		Eventually(func(g Gomega) {
			g.Expect(resManager.Get(context.Background(), ms, store.GetByKey("backend", model.DefaultMesh))).To(Succeed())
			g.Expect(ms.Spec.Ports).To(Equal([]meshservice_api.Port{
				{
					Name:        pointer.To("80"),
					Port:        80,
					TargetPort:  pointer.To(intstr.FromInt(80)),
					AppProtocol: core_meta.ProtocolTCP,
				},
			}))
		}, "2s", "100ms").Should(Succeed())

		dp := core_mesh.NewDataplaneResource()
		Expect(resManager.Get(context.Background(), dp, store.GetByKey("dp-1", model.DefaultMesh))).To(Succeed())
		dp.Spec.Networking.Inbound[0].Port += 1
		dp.Spec.Networking.Inbound[0].ServicePort += 1
		Expect(resManager.Update(context.Background(), dp)).To(Succeed())

		Eventually(func(g Gomega) {
			g.Expect(resManager.Get(context.Background(), ms, store.GetByKey("backend", model.DefaultMesh))).To(Succeed())
			g.Expect(ms.Spec.Ports).To(Equal([]meshservice_api.Port{
				{
					Name:        pointer.To("81"),
					Port:        81,
					TargetPort:  pointer.To(intstr.FromInt(81)),
					AppProtocol: core_meta.ProtocolTCP,
				},
			}))
		}, "2s", "100ms").Should(Succeed())
	})

	It("should eventually delete MeshService if all Dataplanes disappear", func() {
		err := samples.DataplaneBackendBuilder().Create(resManager)
		Expect(err).ToNot(HaveOccurred())

		ms := meshservice_api.NewMeshServiceResource()

		Eventually(func(g Gomega) {
			g.Expect(resManager.Get(context.Background(), ms, store.GetByKey("backend", model.DefaultMesh))).To(Succeed())
			g.Expect(ms.Spec.Ports).To(Equal([]meshservice_api.Port{
				{
					Name:        pointer.To("80"),
					Port:        80,
					TargetPort:  pointer.To(intstr.FromInt(80)),
					AppProtocol: core_meta.ProtocolTCP,
				},
			}))
		}, "2s", "100ms").Should(Succeed())

		dp := core_mesh.NewDataplaneResource()
		Expect(resManager.Delete(context.Background(), dp, store.DeleteByKey("dp-1", model.DefaultMesh))).To(Succeed())

		Eventually(func(g Gomega) {
			g.Expect(resManager.Get(context.Background(), ms, store.GetByKey("backend", model.DefaultMesh))).ToNot(Succeed())
		}, "2s", "100ms").Should(Succeed())
	})

	It("should not delete MeshService not managed by the generator", func() {
		err := samples.DataplaneBackendBuilder().Create(resManager)
		Expect(err).ToNot(HaveOccurred())

		Expect(samples.MeshServiceWebBuilder().Create(resManager)).To(Succeed())

		ms := meshservice_api.NewMeshServiceResource()
		Consistently(func(g Gomega) {
			g.Expect(resManager.Get(context.Background(), ms, store.GetByKey("web", model.DefaultMesh))).To(Succeed())
		}, "2s", "100ms").Should(Succeed())
	})

	It("should not delete MeshService immediately", func() {
		err := samples.DataplaneBackendBuilder().Create(resManager)
		Expect(err).ToNot(HaveOccurred())

		ms := meshservice_api.NewMeshServiceResource()

		Eventually(func(g Gomega) {
			g.Expect(resManager.Get(context.Background(), ms, store.GetByKey("backend", model.DefaultMesh))).To(Succeed())
			g.Expect(ms.Spec.Ports).To(Equal([]meshservice_api.Port{
				{
					Name:        pointer.To("80"),
					Port:        80,
					TargetPort:  pointer.To(intstr.FromInt(80)),
					AppProtocol: core_meta.ProtocolTCP,
				},
			}))
		}, "2s", "100ms").Should(Succeed())

		dp := core_mesh.NewDataplaneResource()
		Expect(resManager.Delete(context.Background(), dp, store.DeleteByKey("dp-1", model.DefaultMesh))).To(Succeed())

		labelGracePeriodStartedAt := ""
		// Wait until the MeshService has been marked with grace period start
		Eventually(func(g Gomega) {
			g.Expect(resManager.Get(context.Background(), ms, store.GetByKey("backend", model.DefaultMesh))).To(Succeed())
			g.Expect(ms.GetMeta().GetLabels()).To(HaveKey(mesh_proto.DeletionGracePeriodStartedLabel))
			labelGracePeriodStartedAt = ms.GetMeta().GetLabels()[mesh_proto.DeletionGracePeriodStartedLabel]
		}, "2s", "100ms").Should(Succeed())

		gracePeriodStartedAt := time.Time{}
		Expect(gracePeriodStartedAt.UnmarshalText([]byte(labelGracePeriodStartedAt))).To(Succeed())

		gracePeriodEndsAt := gracePeriodStartedAt.Add(gracePeriodInterval)
		// Before the grace period it still exists and afterwards it eventually
		// disappears
		Consistently(func(g Gomega) {
			err := resManager.Get(context.Background(), ms, store.GetByKey("backend", model.DefaultMesh))
			if time.Now().Before(gracePeriodEndsAt) {
				g.Expect(err).To(Succeed())
			}
		}, time.Until(gracePeriodEndsAt.Add(-50*time.Millisecond)).String(), "50ms").Should(Succeed())
		Eventually(func(g Gomega) {
			g.Expect(resManager.Get(context.Background(), ms, store.GetByKey("backend", model.DefaultMesh))).ToNot(Succeed())
		}, "2s", "100ms").Should(Succeed())
	})

	It("should not delete MeshService synced from another zone", func() {
		// MeshService generated in another zone and synced here via KDS: it
		// keeps the generator's managed-by label but has origin=global and no
		// local Dataplane backing it. The generator must leave it untouched.
		ms := meshservice_api.NewMeshServiceResource()
		ms.Spec = &meshservice_api.MeshService{
			Selector: meshservice_api.Selector{
				DataplaneTags: &map[string]string{mesh_proto.ServiceTag: "remote-backend"},
			},
			Ports: []meshservice_api.Port{{
				Name:        pointer.To("80"),
				Port:        80,
				TargetPort:  pointer.To(intstr.FromInt(80)),
				AppProtocol: core_meta.ProtocolTCP,
			}},
		}
		Expect(resManager.Create(
			context.Background(),
			ms,
			store.CreateByKey("remote-backend", model.DefaultMesh),
			store.CreateWithLabels(map[string]string{
				mesh_proto.ManagedByLabel:      "meshservice-generator",
				mesh_proto.ResourceOriginLabel: string(mesh_proto.GlobalResourceOrigin),
				mesh_proto.ZoneTag:             "other-zone",
			}),
		)).To(Succeed())

		// It's never marked for deletion nor removed.
		Consistently(func(g Gomega) {
			synced := meshservice_api.NewMeshServiceResource()
			g.Expect(resManager.Get(context.Background(), synced, store.GetByKey("remote-backend", model.DefaultMesh))).To(Succeed())
			g.Expect(synced.GetMeta().GetLabels()).ToNot(HaveKey(mesh_proto.DeletionGracePeriodStartedLabel))
		}, "2s", "100ms").Should(Succeed())
	})

	It("should not delete MeshService if a Dataplane comes back", func() {
		err := samples.DataplaneBackendBuilder().Create(resManager)
		Expect(err).ToNot(HaveOccurred())

		ms := meshservice_api.NewMeshServiceResource()

		Eventually(func(g Gomega) {
			g.Expect(resManager.Get(context.Background(), ms, store.GetByKey("backend", model.DefaultMesh))).To(Succeed())
			g.Expect(ms.Spec.Ports).To(Equal([]meshservice_api.Port{
				{
					Name:        pointer.To("80"),
					Port:        80,
					TargetPort:  pointer.To(intstr.FromInt(80)),
					AppProtocol: core_meta.ProtocolTCP,
				},
			}))
		}, "2s", "100ms").Should(Succeed())

		dp := core_mesh.NewDataplaneResource()
		Expect(resManager.Delete(context.Background(), dp, store.DeleteByKey("dp-1", model.DefaultMesh))).To(Succeed())

		// Wait until the MeshService has been marked with grace period start
		Eventually(func(g Gomega) {
			g.Expect(resManager.Get(context.Background(), ms, store.GetByKey("backend", model.DefaultMesh))).To(Succeed())
			g.Expect(ms.GetMeta().GetLabels()).To(HaveKey("kuma.io/deletion-grace-period-started-at"))
		}, "2s", "100ms").Should(Succeed())

		Expect(
			samples.DataplaneBackendBuilder().Create(resManager),
		).To(Succeed())

		// The MeshService isn't ever deleted
		Consistently(func(g Gomega) {
			ms := meshservice_api.NewMeshServiceResource()
			g.Expect(resManager.Get(context.Background(), ms, store.GetByKey("backend", model.DefaultMesh))).To(Succeed())
		}, "2s", "100ms").Should(Succeed())
	})

	It("should preserve existing labels when updating MeshService", func() {
		err := samples.DataplaneBackendBuilder().Create(resManager)
		Expect(err).ToNot(HaveOccurred())

		ms := meshservice_api.NewMeshServiceResource()
		Eventually(func(g Gomega) {
			g.Expect(resManager.Get(context.Background(), ms, store.GetByKey("backend", model.DefaultMesh))).To(Succeed())
		}, "2s", "100ms").Should(Succeed())

		labelsWithCustom := make(map[string]string)
		maps.Copy(labelsWithCustom, ms.GetMeta().GetLabels())
		labelsWithCustom["custom.io/extra"] = "preserved"
		Expect(resManager.Update(context.Background(), ms, store.UpdateWithLabels(labelsWithCustom))).To(Succeed())

		dp := core_mesh.NewDataplaneResource()
		Expect(resManager.Get(context.Background(), dp, store.GetByKey("dp-1", model.DefaultMesh))).To(Succeed())
		dp.Spec.Networking.Inbound[0].Port += 1
		dp.Spec.Networking.Inbound[0].ServicePort += 1
		Expect(resManager.Update(context.Background(), dp)).To(Succeed())

		Eventually(func(g Gomega) {
			g.Expect(resManager.Get(context.Background(), ms, store.GetByKey("backend", model.DefaultMesh))).To(Succeed())
			g.Expect(ms.Spec.Ports[0].Port).To(Equal(int32(81)))
			g.Expect(ms.GetMeta().GetLabels()).To(HaveKeyWithValue("custom.io/extra", "preserved"))
		}, "2s", "100ms").Should(Succeed())
	})

	It("should emit metric", func() {
		Eventually(func(g Gomega) {
			g.Expect(test_metrics.FindMetric(metrics, "component_meshservice_generator")).ToNot(BeNil())
		}, "2s", "100ms").Should(Succeed())
	})

	Context("with InboundTagsDisabled", func() {
		BeforeEach(func() {
			close(stopCh)

			m, err := core_metrics.NewMetrics("")
			Expect(err).ToNot(HaveOccurred())
			metrics = m
			s := memory.NewStore()
			resManager = manager.NewResourceManager(s)
			meshContextBuilder = xds_context.NewMeshContextBuilder(
				resManager,
				server.MeshResourceTypes(),
				net.LookupIP,
				"zone",
				vips.NewPersistence(resManager, config_manager.NewConfigManager(s), false),
				".mesh",
				80,
				xds_context.AnyToAnyReachableServicesGraphBuilder,
				nil,
			)
			meshCache, err := cache_mesh.NewCache(
				100*time.Millisecond,
				meshContextBuilder,
				metrics,
			)
			Expect(err).ToNot(HaveOccurred())
			allocator, err := generate.New(
				logr.Discard(),
				50*time.Millisecond,
				gracePeriodInterval,
				metrics,
				resManager,
				meshCache,
				"zone",
				true,
				kuma_cp.MeshServiceLabelPropagation{},
			)
			Expect(err).ToNot(HaveOccurred())
			stopCh = make(chan struct{})
			innerCh := stopCh
			go func() {
				defer GinkgoRecover()
				Expect(allocator.Start(innerCh)).To(Succeed())
			}()
			Expect(
				samples.MeshDefaultBuilder().WithMeshServicesEnabled(mesh_proto.Mesh_MeshServices_Everywhere).Create(resManager),
			).To(Succeed())
		})

		createDataplane := func(name, workload string, labels map[string]string, inbounds ...*builders.InboundBuilder) {
			dp := builders.Dataplane().WithAddress("192.168.0.1").WithoutInbounds()
			for _, inbound := range inbounds {
				dp.AddInbound(inbound)
			}
			allLabels := map[string]string{}
			maps.Copy(allLabels, labels)
			if workload != "" {
				allLabels[metadata.KumaWorkload] = workload
			}
			Expect(resManager.Create(
				context.Background(),
				dp.Build(),
				store.CreateByKey(name, model.DefaultMesh),
				store.CreateWithLabels(allLabels),
			)).To(Succeed())
		}

		It("should not generate MeshService for Dataplane without kuma.io/workload label", func() {
			createDataplane("dp-1", "", nil,
				builders.Inbound().WithPort(80).WithServicePort(8080).WithName("http"),
			)

			Consistently(func(g Gomega) {
				mss := &meshservice_api.MeshServiceResourceList{}
				g.Expect(resManager.List(context.Background(), mss)).To(Succeed())
				g.Expect(mss.GetItems()).To(BeEmpty())
			}, "1s", "100ms").Should(Succeed())
		})

		It("should not generate MeshService for invalid kuma.io/workload label", func() {
			createDataplane("dp-1", "Invalid_Workload", nil,
				builders.Inbound().WithPort(80).WithServicePort(8080).WithName("http"),
			)

			Consistently(func(g Gomega) {
				mss := &meshservice_api.MeshServiceResourceList{}
				g.Expect(resManager.List(context.Background(), mss)).To(Succeed())
				g.Expect(mss.GetItems()).To(BeEmpty())
			}, "1s", "100ms").Should(Succeed())
		})

		It("should generate one MeshService per workload with a port per inbound", func() {
			createDataplane("dp-1", "backend", nil,
				builders.Inbound().WithPort(80).WithServicePort(8080).WithName("http"),
				builders.Inbound().WithPort(90).WithServicePort(9090).WithName("grpc"),
			)

			ms := meshservice_api.NewMeshServiceResource()
			Eventually(func(g Gomega) {
				g.Expect(resManager.Get(context.Background(), ms, store.GetByKey("backend", model.DefaultMesh))).To(Succeed())
				g.Expect(ms.Spec.Selector).To(Equal(meshservice_api.Selector{
					DataplaneLabels: &common_api.LabelSelector{
						MatchLabels: &map[string]string{metadata.KumaWorkload: "backend"},
					},
				}))
				g.Expect(ms.Spec.Ports).To(Equal([]meshservice_api.Port{
					{
						Name:        pointer.To("http"),
						Port:        80,
						TargetPort:  pointer.To(intstr.FromInt32(80)),
						AppProtocol: core_meta.ProtocolTCP,
					},
					{
						Name:        pointer.To("grpc"),
						Port:        90,
						TargetPort:  pointer.To(intstr.FromInt32(90)),
						AppProtocol: core_meta.ProtocolTCP,
					},
				}))
			}, "2s", "100ms").Should(Succeed())
		})

		It("should generate a single MeshService for many Dataplanes of one workload", func() {
			createDataplane("dp-1", "backend", nil,
				builders.Inbound().WithPort(80).WithServicePort(8080).WithName("http"),
			)
			createDataplane("dp-2", "backend", nil,
				builders.Inbound().WithPort(80).WithServicePort(8080).WithName("http"),
			)

			Eventually(func(g Gomega) {
				mss := &meshservice_api.MeshServiceResourceList{}
				g.Expect(resManager.List(context.Background(), mss)).To(Succeed())
				g.Expect(mss.GetItems()).To(HaveLen(1))
				g.Expect(mss.GetItems()[0].GetMeta().GetName()).To(Equal("backend"))
			}, "2s", "100ms").Should(Succeed())
		})
	})

	Context("with InboundTagsDisabled and LabelPropagation enabled", func() {
		BeforeEach(func() {
			close(stopCh)

			m, err := core_metrics.NewMetrics("")
			Expect(err).ToNot(HaveOccurred())
			metrics = m
			s := memory.NewStore()
			resManager = manager.NewResourceManager(s)
			meshContextBuilder = xds_context.NewMeshContextBuilder(
				resManager,
				server.MeshResourceTypes(),
				net.LookupIP,
				"zone",
				vips.NewPersistence(resManager, config_manager.NewConfigManager(s), false),
				".mesh",
				80,
				xds_context.AnyToAnyReachableServicesGraphBuilder,
				nil,
			)
			meshCache, err := cache_mesh.NewCache(
				100*time.Millisecond,
				meshContextBuilder,
				metrics,
			)
			Expect(err).ToNot(HaveOccurred())
			allocator, err := generate.New(
				logr.Discard(),
				50*time.Millisecond,
				gracePeriodInterval,
				metrics,
				resManager,
				meshCache,
				"zone",
				true,
				kuma_cp.MeshServiceLabelPropagation{Enabled: true},
			)
			Expect(err).ToNot(HaveOccurred())
			stopCh = make(chan struct{})
			innerCh := stopCh
			go func() {
				defer GinkgoRecover()
				Expect(allocator.Start(innerCh)).To(Succeed())
			}()
			Expect(
				samples.MeshDefaultBuilder().WithMeshServicesEnabled(mesh_proto.Mesh_MeshServices_Everywhere).Create(resManager),
			).To(Succeed())
		})

		It("propagates non-kuma Dataplane labels to the generated MeshService", func() {
			dp := builders.Dataplane().WithAddress("192.168.0.1").WithoutInbounds().
				AddInbound(builders.Inbound().WithPort(80).WithServicePort(8080).WithName("http")).
				Build()
			Expect(resManager.Create(
				context.Background(),
				dp,
				store.CreateByKey("dp-1", model.DefaultMesh),
				store.CreateWithLabels(map[string]string{
					metadata.KumaWorkload: "backend",
					"team":                "payments",
				}),
			)).To(Succeed())

			ms := meshservice_api.NewMeshServiceResource()
			Eventually(func(g Gomega) {
				g.Expect(resManager.Get(context.Background(), ms, store.GetByKey("backend", model.DefaultMesh))).To(Succeed())
				g.Expect(ms.GetMeta().GetLabels()).To(HaveKeyWithValue("team", "payments"))
			}, "2s", "100ms").Should(Succeed())
		})
	})

	Context("with LabelPropagation enabled", func() {
		var countingMgr *countingResourceManager

		BeforeEach(func() {
			close(stopCh)

			m, err := core_metrics.NewMetrics("")
			Expect(err).ToNot(HaveOccurred())
			metrics = m
			s := memory.NewStore()
			baseManager := manager.NewResourceManager(s)
			countingMgr = &countingResourceManager{ResourceManager: baseManager}
			resManager = countingMgr
			meshContextBuilder = xds_context.NewMeshContextBuilder(
				resManager,
				server.MeshResourceTypes(),
				net.LookupIP,
				"zone",
				vips.NewPersistence(resManager, config_manager.NewConfigManager(s), false),
				".mesh",
				80,
				xds_context.AnyToAnyReachableServicesGraphBuilder,
				nil,
			)
			meshCache, err := cache_mesh.NewCache(
				100*time.Millisecond,
				meshContextBuilder,
				metrics,
			)
			Expect(err).ToNot(HaveOccurred())
			allocator, err := generate.New(
				logr.Discard(),
				50*time.Millisecond,
				gracePeriodInterval,
				metrics,
				resManager,
				meshCache,
				"zone",
				false,
				kuma_cp.MeshServiceLabelPropagation{Enabled: true},
			)
			Expect(err).ToNot(HaveOccurred())
			stopCh = make(chan struct{})
			innerCh := stopCh
			go func() {
				defer GinkgoRecover()
				Expect(allocator.Start(innerCh)).To(Succeed())
			}()
			Expect(
				samples.MeshDefaultBuilder().WithMeshServicesEnabled(mesh_proto.Mesh_MeshServices_Everywhere).Create(resManager),
			).To(Succeed())
		})

		It("removes a previously-propagated label when the Dataplane stops carrying it", func() {
			err := builders.Dataplane().
				WithAddress("127.0.0.1").
				WithoutInbounds().
				AddInbound(builders.Inbound().
					WithPort(80).
					WithServicePort(8080).
					WithTags(map[string]string{mesh_proto.ServiceTag: "backend", "appci": "jeffy"}),
				).
				Create(resManager)
			Expect(err).ToNot(HaveOccurred())

			ms := meshservice_api.NewMeshServiceResource()
			Eventually(func(g Gomega) {
				g.Expect(resManager.Get(context.Background(), ms, store.GetByKey("backend", model.DefaultMesh))).To(Succeed())
				g.Expect(ms.GetMeta().GetLabels()).To(HaveKeyWithValue("appci", "jeffy"))
			}, "2s", "100ms").Should(Succeed())

			dp := core_mesh.NewDataplaneResource()
			Expect(resManager.Get(context.Background(), dp, store.GetByKey("dp-1", model.DefaultMesh))).To(Succeed())
			delete(dp.Spec.Networking.Inbound[0].Tags, "appci")
			Expect(resManager.Update(context.Background(), dp)).To(Succeed())

			// The reconciler tracks which keys were propagated via kuma.io/pkey-N
			// labels; when a key is absent from the current vote it is removed.
			Eventually(func(g Gomega) {
				g.Expect(resManager.Get(context.Background(), ms, store.GetByKey("backend", model.DefaultMesh))).To(Succeed())
				g.Expect(ms.GetMeta().GetLabels()).ToNot(HaveKey("appci"))
				g.Expect(ms.GetMeta().GetLabels()).To(HaveKeyWithValue("kuma.io/mesh", model.DefaultMesh))
			}, "2s", "100ms").Should(Succeed())
		})

		It("removes a propagated dotted-key label (qualified name with '/') when DP stops carrying it", func() {
			// Regression test: keys like "app.example.com/tier" contain "/" which is
			// invalid as a label value. The old tracking encoded the key name as a
			// label value and silently skipped such keys, leaving them stuck on the
			// MeshService forever after the carrier DP was removed.
			err := builders.Dataplane().
				WithAddress("127.0.0.1").
				WithoutInbounds().
				AddInbound(builders.Inbound().
					WithPort(80).
					WithServicePort(8080).
					WithTags(map[string]string{
						mesh_proto.ServiceTag:  "backend",
						"app.example.com/tier": "gold",
					}),
				).
				Create(resManager)
			Expect(err).ToNot(HaveOccurred())

			ms := meshservice_api.NewMeshServiceResource()
			Eventually(func(g Gomega) {
				g.Expect(resManager.Get(context.Background(), ms, store.GetByKey("backend", model.DefaultMesh))).To(Succeed())
				g.Expect(ms.GetMeta().GetLabels()).To(HaveKeyWithValue("app.example.com/tier", "gold"))
			}, "2s", "100ms").Should(Succeed())

			dp := core_mesh.NewDataplaneResource()
			Expect(resManager.Get(context.Background(), dp, store.GetByKey("dp-1", model.DefaultMesh))).To(Succeed())
			delete(dp.Spec.Networking.Inbound[0].Tags, "app.example.com/tier")
			Expect(resManager.Update(context.Background(), dp)).To(Succeed())

			Eventually(func(g Gomega) {
				g.Expect(resManager.Get(context.Background(), ms, store.GetByKey("backend", model.DefaultMesh))).To(Succeed())
				g.Expect(ms.GetMeta().GetLabels()).ToNot(HaveKey("app.example.com/tier"))
				g.Expect(ms.GetMeta().GetLabels()).To(HaveKeyWithValue("kuma.io/mesh", model.DefaultMesh))
			}, "2s", "100ms").Should(Succeed())
		})

		It("removes a propagated dotted-key Dataplane resource-label when DP stops carrying it", func() {
			// Same regression as above but exercised via Dataplane metadata labels
			// (the path reported in the linked issue) instead of inbound tags.
			dp := builders.Dataplane().
				WithAddress("127.0.0.1").
				WithoutInbounds().
				AddInbound(builders.Inbound().
					WithPort(80).
					WithServicePort(8080).
					WithTags(map[string]string{mesh_proto.ServiceTag: "backend"}),
				).
				Build()
			Expect(resManager.Create(context.Background(), dp,
				store.CreateByKey("dp-1", model.DefaultMesh),
				store.CreateWithLabels(map[string]string{"app.example.com/tier": "gold"}),
			)).To(Succeed())

			ms := meshservice_api.NewMeshServiceResource()
			Eventually(func(g Gomega) {
				g.Expect(resManager.Get(context.Background(), ms, store.GetByKey("backend", model.DefaultMesh))).To(Succeed())
				g.Expect(ms.GetMeta().GetLabels()).To(HaveKeyWithValue("app.example.com/tier", "gold"))
			}, "2s", "100ms").Should(Succeed())

			loaded := core_mesh.NewDataplaneResource()
			Expect(resManager.Get(context.Background(), loaded, store.GetByKey("dp-1", model.DefaultMesh))).To(Succeed())
			Expect(resManager.Update(context.Background(), loaded, store.UpdateWithLabels(map[string]string{}))).To(Succeed())

			Eventually(func(g Gomega) {
				g.Expect(resManager.Get(context.Background(), ms, store.GetByKey("backend", model.DefaultMesh))).To(Succeed())
				g.Expect(ms.GetMeta().GetLabels()).ToNot(HaveKey("app.example.com/tier"))
				g.Expect(ms.GetMeta().GetLabels()).To(HaveKeyWithValue("kuma.io/mesh", model.DefaultMesh))
			}, "2s", "100ms").Should(Succeed())
		})

		It("does not Update when nothing changes between reconciles", func() {
			err := samples.DataplaneBackendBuilder().Create(resManager)
			Expect(err).ToNot(HaveOccurred())

			ms := meshservice_api.NewMeshServiceResource()
			Eventually(func(g Gomega) {
				g.Expect(resManager.Get(context.Background(), ms, store.GetByKey("backend", model.DefaultMesh))).To(Succeed())
			}, "2s", "100ms").Should(Succeed())

			baseline := countingMgr.updates.Load()
			Consistently(func(g Gomega) {
				g.Expect(countingMgr.updates.Load()).To(Equal(baseline))
			}, "300ms", "50ms").Should(Succeed())
		})

		It("propagates non-reserved labels on create and excludes reserved keys", func() {
			err := builders.Dataplane().
				WithAddress("127.0.0.1").
				WithoutInbounds().
				AddInbound(builders.Inbound().
					WithPort(80).
					WithServicePort(8080).
					WithTags(map[string]string{
						mesh_proto.ServiceTag: "backend",
						"appci":               "jeffy",
						mesh_proto.ZoneTag:    "user-zone",
					}),
				).
				Create(resManager)
			Expect(err).ToNot(HaveOccurred())

			ms := meshservice_api.NewMeshServiceResource()
			Eventually(func(g Gomega) {
				g.Expect(resManager.Get(context.Background(), ms, store.GetByKey("backend", model.DefaultMesh))).To(Succeed())
				g.Expect(ms.GetMeta().GetLabels()).To(HaveKeyWithValue("appci", "jeffy"))
				// kuma.io/zone is a system label — present but set by the generator, not from the inbound tag
				g.Expect(ms.GetMeta().GetLabels()).To(HaveKeyWithValue(mesh_proto.ZoneTag, "zone"))
				g.Expect(ms.GetMeta().GetLabels()).ToNot(HaveKeyWithValue(mesh_proto.ZoneTag, "user-zone"))
			}, "2s", "100ms").Should(Succeed())
		})

		It("issues an Update when only a propagated label changed", func() {
			err := samples.DataplaneBackendBuilder().Create(resManager)
			Expect(err).ToNot(HaveOccurred())

			ms := meshservice_api.NewMeshServiceResource()
			Eventually(func(g Gomega) {
				g.Expect(resManager.Get(context.Background(), ms, store.GetByKey("backend", model.DefaultMesh))).To(Succeed())
			}, "2s", "100ms").Should(Succeed())

			baseline := countingMgr.updates.Load()

			dp := core_mesh.NewDataplaneResource()
			Expect(resManager.Get(context.Background(), dp, store.GetByKey("dp-1", model.DefaultMesh))).To(Succeed())
			dp.Spec.Networking.Inbound[0].Tags["appci"] = "jeffy"
			Expect(resManager.Update(context.Background(), dp)).To(Succeed())

			Eventually(func(g Gomega) {
				g.Expect(countingMgr.updates.Load()).To(BeNumerically(">", baseline))
				g.Expect(resManager.Get(context.Background(), ms, store.GetByKey("backend", model.DefaultMesh))).To(Succeed())
				g.Expect(ms.GetMeta().GetLabels()).To(HaveKeyWithValue("appci", "jeffy"))
			}, "2s", "100ms").Should(Succeed())
		})

		It("preserves externally-added labels on reconcile", func() {
			err := samples.DataplaneBackendBuilder().Create(resManager)
			Expect(err).ToNot(HaveOccurred())

			ms := meshservice_api.NewMeshServiceResource()
			Eventually(func(g Gomega) {
				g.Expect(resManager.Get(context.Background(), ms, store.GetByKey("backend", model.DefaultMesh))).To(Succeed())
			}, "2s", "100ms").Should(Succeed())

			newLabels := maps.Clone(ms.GetMeta().GetLabels())
			newLabels["external.io/foo"] = "bar"
			Expect(resManager.Update(context.Background(), ms, store.UpdateWithLabels(newLabels))).To(Succeed())

			Consistently(func(g Gomega) {
				g.Expect(resManager.Get(context.Background(), ms, store.GetByKey("backend", model.DefaultMesh))).To(Succeed())
				g.Expect(ms.GetMeta().GetLabels()).To(HaveKeyWithValue("external.io/foo", "bar"))
			}, "500ms", "100ms").Should(Succeed())
		})

		It("propagates all non-reserved labels when AllowedLabelKeys is empty", func() {
			err := builders.Dataplane().
				WithAddress("127.0.0.1").
				WithoutInbounds().
				AddInbound(builders.Inbound().
					WithPort(80).
					WithServicePort(8080).
					WithTags(map[string]string{
						mesh_proto.ServiceTag: "backend",
						"appci":               "jeffy",
						"team":                "blue",
					}),
				).
				Create(resManager)
			Expect(err).ToNot(HaveOccurred())

			ms := meshservice_api.NewMeshServiceResource()
			Eventually(func(g Gomega) {
				g.Expect(resManager.Get(context.Background(), ms, store.GetByKey("backend", model.DefaultMesh))).To(Succeed())
				g.Expect(ms.GetMeta().GetLabels()).To(HaveKeyWithValue("appci", "jeffy"))
				g.Expect(ms.GetMeta().GetLabels()).To(HaveKeyWithValue("team", "blue"))
			}, "2s", "100ms").Should(Succeed())
		})

		It("registers component_meshservice_generator_dropped_labels_total before the first drop and increments by reason", func() {
			for _, reason := range []string{"invalid", "inbound_conflict"} {
				m := test_metrics.FindMetric(metrics, "component_meshservice_generator_dropped_labels_total", "reason", reason)
				Expect(m).ToNot(BeNil())
				Expect(m.GetCounter().GetValue()).To(Equal(0.0))
			}

			err := builders.Dataplane().WithName("dp-conflict").WithAddress("10.0.0.1").
				WithoutInbounds().
				AddInbound(builders.Inbound().WithPort(80).WithServicePort(8080).
					WithTags(map[string]string{mesh_proto.ServiceTag: "svc-conflict", "appci": "a"})).
				AddInbound(builders.Inbound().WithPort(81).WithServicePort(8081).
					WithTags(map[string]string{mesh_proto.ServiceTag: "svc-conflict", "appci": "b"})).
				Create(resManager)
			Expect(err).ToNot(HaveOccurred())

			// Colon is valid in Kuma tag values but rejected by IsValidLabelValue,
			// triggering drop("invalid", "appci") in step 2 of dpContribution.
			err = builders.Dataplane().WithName("dp-invalid").WithAddress("10.0.0.2").
				WithoutInbounds().
				AddInbound(builders.Inbound().WithPort(80).WithServicePort(8080).
					WithTags(map[string]string{mesh_proto.ServiceTag: "svc-invalid", "appci": "colon:invalid"})).
				Create(resManager)
			Expect(err).ToNot(HaveOccurred())

			Eventually(func(g Gomega) {
				mConflict := test_metrics.FindMetric(metrics, "component_meshservice_generator_dropped_labels_total", "reason", "inbound_conflict")
				g.Expect(mConflict).ToNot(BeNil())
				g.Expect(mConflict.GetCounter().GetValue()).To(BeNumerically(">=", 1.0))

				mInvalid := test_metrics.FindMetric(metrics, "component_meshservice_generator_dropped_labels_total", "reason", "invalid")
				g.Expect(mInvalid).ToNot(BeNil())
				g.Expect(mInvalid.GetCounter().GetValue()).To(BeNumerically(">=", 1.0))
			}, "2s", "100ms").Should(Succeed())

			// Label contract: every series must carry a "reason" label.
			// The zone constant label added by metrics wrapping is also present.
			families, err := metrics.Gather()
			Expect(err).ToNot(HaveOccurred())
			for _, f := range families {
				if f.GetName() != "component_meshservice_generator_dropped_labels_total" {
					continue
				}
				for _, m := range f.GetMetric() {
					var reasonFound bool
					for _, lp := range m.GetLabel() {
						if lp.GetName() == "reason" {
							reasonFound = true
						}
					}
					Expect(reasonFound).To(BeTrue(), "every series must have a reason label")
				}
			}
		})

		It("still registers the legacy histogram after adding dropped-labels vec", func() {
			m := test_metrics.FindMetric(metrics, "component_meshservice_generator")
			Expect(m).ToNot(BeNil())
		})

		Context("with AllowedLabelKeys", func() {
			BeforeEach(func() {
				close(stopCh)

				m, err := core_metrics.NewMetrics("")
				Expect(err).ToNot(HaveOccurred())
				metrics = m
				s := memory.NewStore()
				baseManager := manager.NewResourceManager(s)
				countingMgr = &countingResourceManager{ResourceManager: baseManager}
				resManager = countingMgr
				meshContextBuilder = xds_context.NewMeshContextBuilder(
					resManager,
					server.MeshResourceTypes(),
					net.LookupIP,
					"zone",
					vips.NewPersistence(resManager, config_manager.NewConfigManager(s), false),
					".mesh",
					80,
					xds_context.AnyToAnyReachableServicesGraphBuilder,
					nil,
				)
				meshCache, err := cache_mesh.NewCache(
					100*time.Millisecond,
					meshContextBuilder,
					metrics,
				)
				Expect(err).ToNot(HaveOccurred())
				allocator, err := generate.New(
					logr.Discard(),
					50*time.Millisecond,
					gracePeriodInterval,
					metrics,
					resManager,
					meshCache,
					"zone",
					false,
					kuma_cp.MeshServiceLabelPropagation{Enabled: true, AllowedLabelKeys: []string{"appci"}},
				)
				Expect(err).ToNot(HaveOccurred())
				stopCh = make(chan struct{})
				innerCh := stopCh
				go func() {
					defer GinkgoRecover()
					Expect(allocator.Start(innerCh)).To(Succeed())
				}()
				Expect(
					samples.MeshDefaultBuilder().WithMeshServicesEnabled(mesh_proto.Mesh_MeshServices_Everywhere).Create(resManager),
				).To(Succeed())
			})

			It("propagates only allowed keys", func() {
				err := builders.Dataplane().
					WithAddress("127.0.0.1").
					WithoutInbounds().
					AddInbound(builders.Inbound().
						WithPort(80).
						WithServicePort(8080).
						WithTags(map[string]string{
							mesh_proto.ServiceTag: "backend",
							"appci":               "jeffy",
							"team":                "blue",
						}),
					).
					Create(resManager)
				Expect(err).ToNot(HaveOccurred())

				ms := meshservice_api.NewMeshServiceResource()
				Eventually(func(g Gomega) {
					g.Expect(resManager.Get(context.Background(), ms, store.GetByKey("backend", model.DefaultMesh))).To(Succeed())
					g.Expect(ms.GetMeta().GetLabels()).To(HaveKeyWithValue("appci", "jeffy"))
					g.Expect(ms.GetMeta().GetLabels()).ToNot(HaveKey("team"))
				}, "2s", "100ms").Should(Succeed())
			})
		})
	})
})
