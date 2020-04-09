package manager_test

import (
	"context"
	"time"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	core_manager "github.com/Kong/kuma/pkg/core/resources/manager"
	core_model "github.com/Kong/kuma/pkg/core/resources/model"
	core_store "github.com/Kong/kuma/pkg/core/resources/store"
	"github.com/Kong/kuma/pkg/plugins/resources/memory"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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
	c.listQueries++
	return c.store.List(ctx, list, fn...)
}

var _ core_manager.ReadOnlyResourceManager = &countingResourcesManager{}

var _ = Describe("Cached Resource Manager", func() {

	var store core_store.ResourceStore
	var cachedManager core_manager.ReadOnlyResourceManager
	var countingManager *countingResourcesManager
	var res *core_mesh.DataplaneResource
	expiration := 100 * time.Millisecond

	BeforeEach(func() {
		// given
		store = memory.NewStore()
		countingManager = &countingResourcesManager{
			store: store,
		}
		cachedManager = core_manager.NewCachedManager(countingManager, expiration)

		// and created resources
		res = &core_mesh.DataplaneResource{
			Spec: mesh_proto.Dataplane{
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
		err := store.Create(context.Background(), res, core_store.CreateByKey("dp-1", "default"))
		Expect(err).ToNot(HaveOccurred())
	})

	It("should cache Get() queries", func() {
		// when fetched resources multiple times
		fetch := func() core_mesh.DataplaneResource {
			fetched := core_mesh.DataplaneResource{}
			err := cachedManager.Get(context.Background(), &fetched, core_store.GetByKey("dp-1", "default"))
			Expect(err).ToNot(HaveOccurred())
			return fetched
		}

		for i := 0; i < 100; i++ {
			fetch()
		}

		// then real manager should be called only once
		Expect(fetch().Spec).To(Equal(res.Spec))
		Expect(countingManager.getQueries).To(Equal(1))

		// when
		time.Sleep(expiration)

		// then
		Expect(fetch().Spec).To(Equal(res.Spec))
		Expect(countingManager.getQueries).To(Equal(2))
	})

	It("should cache List() queries", func() {
		// when fetched resources multiple times
		fetch := func() core_mesh.DataplaneResourceList {
			fetched := core_mesh.DataplaneResourceList{}
			err := cachedManager.List(context.Background(), &fetched, core_store.ListByMesh("default"))
			Expect(err).ToNot(HaveOccurred())
			return fetched
		}

		for i := 0; i < 100; i++ {
			fetch()
		}

		// then real manager should be called only once
		Expect(fetch().Items).To(HaveLen(1))
		Expect(fetch().Items[0].GetSpec()).To(Equal(&res.Spec))
		Expect(countingManager.listQueries).To(Equal(1))

		// when
		time.Sleep(expiration)

		// then
		Expect(fetch().Items).To(HaveLen(1))
		Expect(fetch().Items[0].GetSpec()).To(Equal(&res.Spec))
		Expect(countingManager.listQueries).To(Equal(2))
	})
})
