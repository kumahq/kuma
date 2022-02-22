package manager_test

import (
	"context"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/pkg/test"
	. "github.com/kumahq/kuma/pkg/test/matchers"
	test_metrics "github.com/kumahq/kuma/pkg/test/metrics"
)

type countingResourcesManager struct {
	store       core_store.ResourceStore
	getQueries  int
	listQueries int
}

func (c *countingResourcesManager) Get(ctx context.Context, res core_model.Resource, fn ...core_store.GetOptionsFunc) error {
	c.getQueries++
	return c.store.Get(ctx, res, fn...)
}

func (c *countingResourcesManager) List(ctx context.Context, list core_model.ResourceList, fn ...core_store.ListOptionsFunc) error {
	opts := core_store.NewListOptions(fn...)
	if list.GetItemType() == core_mesh.TrafficLogType && opts.Mesh == "slow" {
		time.Sleep(10 * time.Second)
	}
	c.listQueries++
	return c.store.List(ctx, list, fn...)
}

var _ core_manager.ReadOnlyResourceManager = &countingResourcesManager{}

var _ = Describe("Cached Resource Manager", func() {

	var store core_store.ResourceStore
	var cachedManager core_manager.ReadOnlyResourceManager
	var countingManager *countingResourcesManager
	var res *core_mesh.DataplaneResource
	var metrics core_metrics.Metrics
	expiration := 500 * time.Millisecond

	BeforeEach(func() {
		// given
		store = memory.NewStore()
		countingManager = &countingResourcesManager{
			store: store,
		}
		m, err := core_metrics.NewMetrics("Standalone")
		metrics = m
		Expect(err).ToNot(HaveOccurred())
		cachedManager, err = core_manager.NewCachedManager(countingManager, expiration, metrics)
		Expect(err).ToNot(HaveOccurred())

		// and created resources
		res = &core_mesh.DataplaneResource{
			Spec: &mesh_proto.Dataplane{
				Networking: &mesh_proto.Dataplane_Networking{
					Address: "127.0.0.1",
					Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
						{
							Port:        80,
							ServicePort: 8080,
						},
					},
				},
			},
		}
		err = store.Create(context.Background(), res, core_store.CreateByKey("dp-1", "default"))
		Expect(err).ToNot(HaveOccurred())
	})

	It("should cache Get() queries", func() {
		// when fetched resources multiple times
		fetch := func() *core_mesh.DataplaneResource {
			fetched := core_mesh.NewDataplaneResource()
			err := cachedManager.Get(context.Background(), fetched, core_store.GetByKey("dp-1", "default"))
			Expect(err).ToNot(HaveOccurred())
			return fetched
		}

		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				fetch()
				wg.Done()
			}()
		}
		wg.Wait()

		// then real manager should be called only once
		Expect(fetch().Spec).To(MatchProto(res.Spec))
		Expect(countingManager.getQueries).To(Equal(1))

		// when
		time.Sleep(expiration)

		// then
		Expect(fetch().Spec).To(MatchProto(res.Spec))
		Expect(countingManager.getQueries).To(Equal(2))

		// and metrics are published
		Expect(test_metrics.FindMetric(metrics, "store_cache", "operation", "get", "result", "miss").Counter.GetValue()).To(Equal(2.0))
		hits := test_metrics.FindMetric(metrics, "store_cache", "operation", "get", "result", "hit").Counter.GetValue()
		hitWaits := 0.0
		hitWaitMetric := test_metrics.FindMetric(metrics, "store_cache", "operation", "get", "result", "hit-wait")
		if hitWaitMetric != nil {
			hitWaits = hitWaitMetric.Counter.GetValue()
		}
		Expect(hits + hitWaits).To(Equal(100.0))
	})

	It("should not cache Get() not found", func() {
		// when fetched resources multiple times
		fetch := func() {
			_ = cachedManager.Get(context.Background(), core_mesh.NewDataplaneResource(), core_store.GetByKey("non-existing", "default"))
		}

		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				fetch()
				wg.Done()
			}()
		}
		wg.Wait()

		// then real manager should be called every time
		Expect(countingManager.getQueries).To(Equal(100))
	})

	It("should cache List() queries", func() {
		// when fetched resources multiple times
		fetch := func() core_mesh.DataplaneResourceList {
			fetched := core_mesh.DataplaneResourceList{}
			err := cachedManager.List(context.Background(), &fetched, core_store.ListByMesh("default"))
			Expect(err).ToNot(HaveOccurred())
			return fetched
		}

		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				fetch()
				wg.Done()
			}()
		}
		wg.Wait()

		// then real manager should be called only once
		list := fetch()
		Expect(list.Items).To(HaveLen(1))
		Expect(list.Items[0].GetSpec()).To(MatchProto(res.Spec))
		Expect(countingManager.listQueries).To(Equal(1))

		// when
		time.Sleep(expiration)

		// then
		list = fetch()
		Expect(list.Items).To(HaveLen(1))
		Expect(list.Items[0].GetSpec()).To(MatchProto(res.Spec))
		Expect(countingManager.listQueries).To(Equal(2))

		// and metrics are published
		Expect(test_metrics.FindMetric(metrics, "store_cache", "operation", "list", "result", "miss").Counter.GetValue()).To(Equal(2.0))
		hits := test_metrics.FindMetric(metrics, "store_cache", "operation", "list", "result", "hit").Counter.GetValue()
		hitWaits := 0.0
		hitWaitMetric := test_metrics.FindMetric(metrics, "store_cache", "operation", "list", "result", "hit-wait")
		if hitWaitMetric != nil {
			hitWaits = hitWaitMetric.Counter.GetValue()
		}
		Expect(hits + hitWaits).To(Equal(100.0))
	})

	It("should let concurrent List() queries for different types and meshes", test.Within(5*time.Second, func() {
		// given ongoing TrafficLog from mesh slow that takes a lot of time to complete
		go func() {
			fetched := core_mesh.TrafficLogResourceList{}
			err := cachedManager.List(context.Background(), &fetched, core_store.ListByMesh("slow"))
			Expect(err).ToNot(HaveOccurred())
		}()

		// when trying to fetch TrafficLog from different mesh that takes normal time to response
		fetched := core_mesh.TrafficLogResourceList{}
		err := cachedManager.List(context.Background(), &fetched, core_store.ListByMesh("default"))

		// then first request does not block request for other mesh
		Expect(err).ToNot(HaveOccurred())

		// when trying to fetch different resource type
		fetchedTp := core_mesh.TrafficPermissionResourceList{}
		err = cachedManager.List(context.Background(), &fetchedTp, core_store.ListByMesh("default"))

		// then first request does not block request for other type
		Expect(err).ToNot(HaveOccurred())
	}))
})
