package cla_test

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net"
	"path/filepath"
	"sync"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	test_metrics "github.com/kumahq/kuma/pkg/test/metrics"
	"github.com/kumahq/kuma/pkg/xds/cache/cla"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
)

type countingResourcesManager struct {
	store       core_store.ResourceStore
	getQueries  int
	listQueries int
}

var _ core_manager.ReadOnlyResourceManager = &countingResourcesManager{}

func (c *countingResourcesManager) Get(ctx context.Context, res core_model.Resource, fn ...core_store.GetOptionsFunc) error {
	c.getQueries++
	return c.store.Get(ctx, res, fn...)
}

func (c *countingResourcesManager) List(ctx context.Context, list core_model.ResourceList, fn ...core_store.ListOptionsFunc) error {
	c.listQueries++
	return c.store.List(ctx, list, fn...)
}

var _ = Describe("ClusterLoadAssignment Cache", func() {
	var s core_store.ResourceStore
	var countingManager *countingResourcesManager
	var claCache *cla.Cache
	var metrics core_metrics.Metrics

	expiration := 2 * time.Second

	BeforeEach(func() {
		s = memory.NewStore()
		countingManager = &countingResourcesManager{store: s}
		var err error

		metrics, err = core_metrics.NewMetrics("Standalone")
		Expect(err).ToNot(HaveOccurred())

		claCache, err = cla.NewCache(countingManager, "", expiration,
			func(s string) ([]net.IP, error) {
				return []net.IP{net.ParseIP(s)}, nil
			}, metrics)
		Expect(err).ToNot(HaveOccurred())
	})

	BeforeEach(func() {
		mesh := "mesh-0"
		err := s.Create(context.Background(), core_mesh.NewMeshResource(), core_store.CreateByKey(mesh, core_model.NoMesh))
		Expect(err).ToNot(HaveOccurred())

		err = s.Create(context.Background(), &core_mesh.DataplaneResource{
			Spec: &mesh_proto.Dataplane{Networking: &mesh_proto.Dataplane_Networking{
				Address: "192.168.0.1",
				Inbound: []*mesh_proto.Dataplane_Networking_Inbound{{
					Port: 1010, ServicePort: 2020, Tags: map[string]string{"kuma.io/service": "backend"},
				}},
			}},
		}, core_store.CreateByKey("dp1", mesh))
		Expect(err).ToNot(HaveOccurred())

		err = s.Create(context.Background(), &core_mesh.DataplaneResource{
			Spec: &mesh_proto.Dataplane{Networking: &mesh_proto.Dataplane_Networking{
				Address: "192.168.0.2",
				Inbound: []*mesh_proto.Dataplane_Networking_Inbound{{
					Port: 1011, ServicePort: 2021, Tags: map[string]string{"kuma.io/service": "backend"},
				}},
			}},
		}, core_store.CreateByKey("dp2", mesh))
		Expect(err).ToNot(HaveOccurred())
	})

	It("should cache Get() queries", func() {
		By("getting CLA for the first time")
		cla, err := claCache.GetCLA(context.Background(), "mesh-0", "", envoy_common.NewCluster(envoy_common.WithService("backend")), envoy_common.APIV3)
		Expect(err).ToNot(HaveOccurred())
		// 1 Get request:
		// - GetMesh
		Expect(countingManager.getQueries).To(Equal(1))
		// 2 List request:
		// - GetDataplanes
		// - GetZoneIngresses
		Expect(countingManager.listQueries).To(Equal(2))

		expected, err := ioutil.ReadFile(filepath.Join("testdata", "cla.get.0.json"))
		Expect(err).ToNot(HaveOccurred())

		js, err := json.Marshal(cla)
		Expect(err).ToNot(HaveOccurred())
		Expect(js).To(MatchJSON(string(expected)))

		By("getting cached CLA")
		_, err = claCache.GetCLA(context.Background(), "mesh-0", "", envoy_common.NewCluster(envoy_common.WithService("backend")), envoy_common.APIV3)
		Expect(err).ToNot(HaveOccurred())
		Expect(countingManager.getQueries).To(Equal(1))
		Expect(countingManager.listQueries).To(Equal(2))

		By("updating Dataplane in store and waiting until cache invalidation")
		dp := core_mesh.NewDataplaneResource()
		err = s.Get(context.Background(), dp, core_store.GetByKey("dp2", "mesh-0"))
		Expect(err).ToNot(HaveOccurred())
		dp.Spec.Networking.Address = "1.1.1.1"
		err = s.Update(context.Background(), dp)
		Expect(err).ToNot(HaveOccurred())

		<-time.After(2 * time.Second)

		cla, err = claCache.GetCLA(context.Background(), "mesh-0", "", envoy_common.NewCluster(envoy_common.WithService("backend")), envoy_common.APIV3)
		Expect(err).ToNot(HaveOccurred())
		Expect(countingManager.getQueries).To(Equal(2))
		Expect(countingManager.listQueries).To(Equal(4))

		expected, err = ioutil.ReadFile(filepath.Join("testdata", "cla.get.1.json"))
		Expect(err).ToNot(HaveOccurred())
		js, err = json.Marshal(cla)
		Expect(err).ToNot(HaveOccurred())
		Expect(js).To(MatchJSON(expected))
	})

	It("should cache concurrent Get() requests", func() {
		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				cla, err := claCache.GetCLA(context.Background(), "mesh-0", "", envoy_common.NewCluster(envoy_common.WithService("backend")), envoy_common.APIV3)
				Expect(err).ToNot(HaveOccurred())

				marshalled, err := json.Marshal(cla) // to imitate Read access to 'cla'
				Expect(err).ToNot(HaveOccurred())
				Expect(len(marshalled) > 0).To(BeTrue())
				wg.Done()
			}()
		}
		wg.Wait()

		Expect(countingManager.getQueries).To(Equal(1))
		Expect(test_metrics.FindMetric(metrics, "cla_cache", "operation", "get", "result", "miss").Gauge.GetValue()).To(Equal(1.0))
		hitWaits := 0.0
		if hw := test_metrics.FindMetric(metrics, "cla_cache", "operation", "get", "result", "hit-wait"); hw != nil {
			hitWaits = hw.Gauge.GetValue()
		}
		hits := 0.0
		if h := test_metrics.FindMetric(metrics, "cla_cache", "operation", "get", "result", "hit"); h != nil {
			hits = h.Gauge.GetValue()
		}
		Expect(hitWaits + hits + 1).To(Equal(100.0))
	})

	It("should support clusters with the same names but different tags", func() {
		err := s.Create(context.Background(), &core_mesh.DataplaneResource{
			Spec: &mesh_proto.Dataplane{Networking: &mesh_proto.Dataplane_Networking{
				Address: "192.168.0.3",
				Inbound: []*mesh_proto.Dataplane_Networking_Inbound{{
					Port: 1011, ServicePort: 2021,
					Tags: map[string]string{
						"kuma.io/service": "backend",
						"version":         "v1",
					},
				}},
			}},
		}, core_store.CreateByKey("dp3", "mesh-0"))
		Expect(err).ToNot(HaveOccurred())

		err = s.Create(context.Background(), &core_mesh.DataplaneResource{
			Spec: &mesh_proto.Dataplane{Networking: &mesh_proto.Dataplane_Networking{
				Address: "192.168.0.4",
				Inbound: []*mesh_proto.Dataplane_Networking_Inbound{{
					Port: 1011, ServicePort: 2021,
					Tags: map[string]string{
						"kuma.io/service": "backend",
						"version":         "v2",
					},
				}},
			}},
		}, core_store.CreateByKey("dp4", "mesh-0"))
		Expect(err).ToNot(HaveOccurred())

		clusterV1 := envoy_common.NewCluster(
			envoy_common.WithService("backend"),
			envoy_common.WithName("backend-_0_"),
			envoy_common.WithTags(map[string]string{
				"kuma.io/service": "backend",
				"version":         "v1",
			}),
		)

		clusterV2 := envoy_common.NewCluster(
			envoy_common.WithService("backend"),
			envoy_common.WithName("backend-_0_"),
			envoy_common.WithTags(map[string]string{
				"kuma.io/service": "backend",
				"version":         "v2",
			}),
		)

		cla1, err := claCache.GetCLA(context.Background(), "mesh-0", "", clusterV1, envoy_common.APIV3)
		Expect(err).ToNot(HaveOccurred())

		cla2, err := claCache.GetCLA(context.Background(), "mesh-0", "", clusterV2, envoy_common.APIV3)
		Expect(err).ToNot(HaveOccurred())

		Expect(cla1).ToNot(Equal(cla2))

	})
})
