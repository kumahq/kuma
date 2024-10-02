package mesh_test

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/config/manager"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/dns/vips"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	test_metrics "github.com/kumahq/kuma/pkg/test/metrics"
	"github.com/kumahq/kuma/pkg/test/resources/model"
	"github.com/kumahq/kuma/pkg/test/resources/samples"
	"github.com/kumahq/kuma/pkg/xds/cache/mesh"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
)

type countingResourcesManager struct {
	store       core_store.ResourceStore
	err         error
	getQueries  int
	listQueries map[core_model.ResourceType]int
}

var _ core_manager.ReadOnlyResourceManager = &countingResourcesManager{}

func (c *countingResourcesManager) totalQueries() int {
	r := 0
	for _, n := range c.listQueries {
		r += n
	}
	return r + c.getQueries
}

func (c *countingResourcesManager) Get(ctx context.Context, res core_model.Resource, fn ...core_store.GetOptionsFunc) error {
	c.getQueries++
	if c.err != nil {
		return c.err
	}
	return c.store.Get(ctx, res, fn...)
}

func (c *countingResourcesManager) List(ctx context.Context, list core_model.ResourceList, fn ...core_store.ListOptionsFunc) error {
	if c.listQueries == nil {
		c.listQueries = map[core_model.ResourceType]int{}
	}
	c.listQueries[list.GetItemType()]++
	if c.err != nil {
		return c.err
	}
	return c.store.List(ctx, list, fn...)
}

func (c *countingResourcesManager) reset() {
	c.err = nil
	c.listQueries = map[core_model.ResourceType]int{}
	c.getQueries = 0
}

var _ = Describe("MeshSnapshot Cache", func() {
	testDataplaneResources := func(n int, mesh, version, address string) []*core_mesh.DataplaneResource {
		resources := []*core_mesh.DataplaneResource{}
		for i := 0; i < n; i++ {
			resources = append(resources, samples.DataplaneBackendBuilder().
				WithName(fmt.Sprintf("dp-%d", i)).
				WithMesh(mesh).
				WithVersion(version).
				WithAddress(address).
				Build(),
			)
		}
		return resources
	}
	testTrafficRouteResources := func(n int, mesh, version string) []*core_mesh.TrafficRouteResource {
		resources := []*core_mesh.TrafficRouteResource{}
		for i := 0; i < n; i++ {
			resources = append(resources, &core_mesh.TrafficRouteResource{
				Meta: &model.ResourceMeta{Mesh: mesh, Name: fmt.Sprintf("tr-%d", i), Version: version},
				Spec: &mesh_proto.TrafficRoute{},
			})
		}
		return resources
	}

	const baseLen = 100
	var s core_store.ResourceStore
	var countingManager *countingResourcesManager
	var meshCache *mesh.Cache
	var metrics core_metrics.Metrics

	expiration := 2 * time.Second

	BeforeEach(func() {
		s = memory.NewStore()
		countingManager = &countingResourcesManager{store: s}
		var err error
		metrics, err = core_metrics.NewMetrics("Zone")
		Expect(err).ToNot(HaveOccurred())

		lookupIPFunc := func(s string) ([]net.IP, error) {
			return []net.IP{net.ParseIP(s)}, nil
		}
		meshContextBuilder := xds_context.NewMeshContextBuilder(
			countingManager,
			[]core_model.ResourceType{core_mesh.DataplaneType, core_mesh.TrafficRouteType, core_mesh.ZoneIngressType},
			lookupIPFunc,
			"zone-1",
			vips.NewPersistence(core_manager.NewResourceManager(s), manager.NewConfigManager(s), false),
			"mesh",
			80,
			xds_context.AnyToAnyReachableServicesGraphBuilder,
		)
		meshCache, err = mesh.NewCache(
			expiration,
			meshContextBuilder,
			metrics,
		)
		Expect(err).ToNot(HaveOccurred())
	})

	BeforeEach(func() {
		for i := 0; i < 3; i++ {
			mesh := fmt.Sprintf("mesh-%d", i)
			err := s.Create(context.Background(), core_mesh.NewMeshResource(), core_store.CreateByKey(mesh, core_model.NoMesh))
			Expect(err).ToNot(HaveOccurred())

			for _, dp := range testDataplaneResources(baseLen, mesh, "v1", "192.168.0.1") {
				err := s.Create(context.Background(), dp, core_store.CreateBy(core_model.MetaToResourceKey(dp.GetMeta())))
				Expect(err).ToNot(HaveOccurred())
			}
			for _, tr := range testTrafficRouteResources(baseLen, mesh, "v2") {
				err := s.Create(context.Background(), tr, core_store.CreateBy(core_model.MetaToResourceKey(tr.GetMeta())))
				Expect(err).ToNot(HaveOccurred())
			}
		}
	})

	It("should cache a hash of mesh", func() {
		By("getting Hash for the first time")
		meshCtx, err := meshCache.GetMeshContext(context.Background(), "mesh-0")
		Expect(err).ToNot(HaveOccurred())
		expectedHash := "JLYLIWtgziYPEjOSc6/i/w=="
		Expect(meshCtx.Hash).To(Equal(expectedHash))
		Expect(countingManager.getQueries).To(Equal(1)) // one Get to obtain Mesh
		Expect(countingManager.listQueries).To(MatchAllKeys(Keys{
			core_mesh.DataplaneType:    Equal(1),
			core_mesh.TrafficRouteType: Equal(1),
			core_mesh.ZoneIngressType:  Equal(1),
		}))

		By("getting cached Hash")
		_, err = meshCache.GetMeshContext(context.Background(), "mesh-0")
		Expect(err).ToNot(HaveOccurred())
		Expect(countingManager.getQueries).To(Equal(1))           // should be the same
		Expect(countingManager.listQueries).To(MatchAllKeys(Keys{ // same as above
			core_mesh.DataplaneType:    Equal(1),
			core_mesh.TrafficRouteType: Equal(1),
			core_mesh.ZoneIngressType:  Equal(1),
		}))

		By("updating Dataplane in store and waiting until cache invalidation")
		dp := core_mesh.NewDataplaneResource()
		err = s.Get(context.Background(), dp, core_store.GetByKey("dp-1", "mesh-0"))
		Expect(err).ToNot(HaveOccurred())
		dp.Spec.Networking.Address = "1.1.1.1"
		err = s.Update(context.Background(), dp)
		Expect(err).ToNot(HaveOccurred())

		<-time.After(expiration)

		meshCtx, err = meshCache.GetMeshContext(context.Background(), "mesh-0")
		Expect(err).ToNot(HaveOccurred())
		expectedHash = "KazI2baK7/QqpBu1OOajkA=="
		Expect(meshCtx.Hash).To(Equal(expectedHash))
		Expect(countingManager.getQueries).To(Equal(2))
		Expect(countingManager.listQueries).To(MatchAllKeys(Keys{
			core_mesh.DataplaneType:    Equal(2),
			core_mesh.TrafficRouteType: Equal(2),
			core_mesh.ZoneIngressType:  Equal(2),
		}))
	})

	It("should count hashes independently for each mesh", func() {
		meshCtx0, err := meshCache.GetMeshContext(context.Background(), "mesh-0")
		Expect(err).ToNot(HaveOccurred())
		hash0 := meshCtx0.Hash

		meshCtx1, err := meshCache.GetMeshContext(context.Background(), "mesh-1")
		Expect(err).ToNot(HaveOccurred())
		hash1 := meshCtx1.Hash

		meshCtx2, err := meshCache.GetMeshContext(context.Background(), "mesh-2")
		Expect(err).ToNot(HaveOccurred())
		hash2 := meshCtx2.Hash

		<-time.After(expiration)

		// Computing one meshcontext shouldn't cause us to recompute other
		// meshcontexts
		nextMeshCtx0, err := meshCache.GetMeshContext(context.Background(), "mesh-0")
		Expect(err).ToNot(HaveOccurred())
		// If the meshcontxt hasn't been recomputed, its fields will be identical
		Expect(nextMeshCtx0.DataSourceLoader).To(BeIdenticalTo(meshCtx0.DataSourceLoader))

		dp := core_mesh.NewDataplaneResource()
		err = s.Get(context.Background(), dp, core_store.GetByKey("dp-1", "mesh-0"))
		Expect(err).ToNot(HaveOccurred())
		dp.Spec.Networking.Address = "1.1.1.1"
		err = s.Update(context.Background(), dp)
		Expect(err).ToNot(HaveOccurred())

		<-time.After(expiration)

		meshCtx0, err = meshCache.GetMeshContext(context.Background(), "mesh-0")
		Expect(err).ToNot(HaveOccurred())
		updHash0 := meshCtx0.Hash

		meshCtx1, err = meshCache.GetMeshContext(context.Background(), "mesh-1")
		Expect(err).ToNot(HaveOccurred())
		updHash1 := meshCtx1.Hash

		meshCtx2, err = meshCache.GetMeshContext(context.Background(), "mesh-2")
		Expect(err).ToNot(HaveOccurred())
		updHash2 := meshCtx2.Hash

		Expect(hash0).ToNot(Equal(updHash0))
		Expect(hash1).To(Equal(updHash1))
		Expect(hash2).To(Equal(updHash2))
	})

	It("should cache concurrent Get() requests", func() {
		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				ctx, err := meshCache.GetMeshContext(context.Background(), "mesh-0")
				Expect(err).ToNot(HaveOccurred())
				Expect(ctx).NotTo(BeNil())
				wg.Done()
			}()
		}
		wg.Wait()

		Expect(countingManager.getQueries).To(Equal(1))
		Expect(test_metrics.FindMetric(metrics, "mesh_cache", "operation", "get", "result", "miss").Gauge.GetValue()).To(Equal(1.0))
		hitWaits := 0.0
		if hw := test_metrics.FindMetric(metrics, "mesh_cache", "operation", "get", "result", "hit-wait"); hw != nil {
			hitWaits = hw.Gauge.GetValue()
		}
		hits := 0.0
		if h := test_metrics.FindMetric(metrics, "mesh_cache", "operation", "get", "result", "hit"); h != nil {
			hits = h.Gauge.GetValue()
		}
		Expect(hitWaits + hits + 1).To(Equal(100.0))
	})

	It("should retry previously failed Get() requests", func() {
		countingManager.err = errors.New("I want to fail")
		By("getting MeshCtx for the first time")
		meshCtx, err := meshCache.GetMeshContext(context.Background(), "mesh-0")
		Expect(countingManager.totalQueries()).To(Equal(1)) // Fail on the first query
		Expect(err).To(HaveOccurred())
		Expect(meshCtx).To(Equal(xds_context.MeshContext{}))

		By("getting Hash calls again")
		_, err = meshCache.GetMeshContext(context.Background(), "mesh-0")
		Expect(err).To(HaveOccurred())
		Expect(countingManager.totalQueries()).To(Equal(2)) // should be increased by one (errors are not cached)
		Expect(err).To(HaveOccurred())

		By("Getting the hash once manager is fixed")
		countingManager.reset()
		meshCtx, err = meshCache.GetMeshContext(context.Background(), "mesh-0")
		Expect(err).ToNot(HaveOccurred())
		expectedHash := "JLYLIWtgziYPEjOSc6/i/w=="
		Expect(meshCtx.Hash).To(Equal(expectedHash))
		Expect(countingManager.getQueries).To(Equal(1)) // one Get to obtain Mesh
		Expect(countingManager.listQueries).To(MatchAllKeys(Keys{
			core_mesh.DataplaneType:    Equal(1),
			core_mesh.TrafficRouteType: Equal(1),
			core_mesh.ZoneIngressType:  Equal(1),
		}))

		By("Now it should cache the hash once manager is fixed")
		countingManager.err = nil
		meshCtx, err = meshCache.GetMeshContext(context.Background(), "mesh-0")
		Expect(err).ToNot(HaveOccurred())
		Expect(meshCtx.Hash).To(Equal(expectedHash))
		Expect(countingManager.getQueries).To(Equal(1))           // same as above
		Expect(countingManager.listQueries).To(MatchAllKeys(Keys{ // same as above
			core_mesh.DataplaneType:    Equal(1),
			core_mesh.TrafficRouteType: Equal(1),
			core_mesh.ZoneIngressType:  Equal(1),
		}))
	})
})
