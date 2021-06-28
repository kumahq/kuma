package mesh_test

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	core_metrics "github.com/kumahq/kuma/pkg/metrics"
	test_metrics "github.com/kumahq/kuma/pkg/test/metrics"
	"github.com/kumahq/kuma/pkg/xds/cache/sha256"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	mesh_core "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/pkg/test/resources/model"
	"github.com/kumahq/kuma/pkg/xds/cache/mesh"
)

type countingResourcesManager struct {
	store       core_store.ResourceStore
	err         error
	getQueries  int
	listQueries int
}

var _ core_manager.ReadOnlyResourceManager = &countingResourcesManager{}

func (c *countingResourcesManager) Get(ctx context.Context, res core_model.Resource, fn ...core_store.GetOptionsFunc) error {
	c.getQueries++
	if c.err != nil {
		return c.err
	}
	return c.store.Get(ctx, res, fn...)
}

func (c *countingResourcesManager) List(ctx context.Context, list core_model.ResourceList, fn ...core_store.ListOptionsFunc) error {
	c.listQueries++
	if c.err != nil {
		return c.err
	}
	return c.store.List(ctx, list, fn...)
}

var _ = Describe("MeshSnapshot Cache", func() {
	testDataplaneResources := func(n int, mesh, version, address string) []*mesh_core.DataplaneResource {
		resources := []*mesh_core.DataplaneResource{}
		for i := 0; i < n; i++ {
			resources = append(resources, &mesh_core.DataplaneResource{
				Meta: &model.ResourceMeta{Mesh: mesh, Name: fmt.Sprintf("dp-%d", i), Version: version},
				Spec: &mesh_proto.Dataplane{
					Networking: &mesh_proto.Dataplane_Networking{
						Address: address,
					},
				},
			})
		}
		return resources
	}
	testTrafficRouteResources := func(n int, mesh, version string) []*mesh_core.TrafficRouteResource {
		resources := []*mesh_core.TrafficRouteResource{}
		for i := 0; i < n; i++ {
			resources = append(resources, &mesh_core.TrafficRouteResource{
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
		metrics, err = core_metrics.NewMetrics("Standalone")
		Expect(err).ToNot(HaveOccurred())

		meshCache, err = mesh.NewCache(countingManager, expiration,
			[]core_model.ResourceType{mesh_core.DataplaneType, mesh_core.TrafficRouteType},
			func(s string) ([]net.IP, error) {
				return []net.IP{net.ParseIP(s)}, nil
			}, metrics)
		Expect(err).ToNot(HaveOccurred())
	})

	BeforeEach(func() {
		for i := 0; i < 3; i++ {
			mesh := fmt.Sprintf("mesh-%d", i)
			err := s.Create(context.Background(), mesh_core.NewMeshResource(), core_store.CreateByKey(mesh, core_model.NoMesh))
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
		hash, err := meshCache.GetHash(context.Background(), "mesh-0")
		Expect(err).ToNot(HaveOccurred())
		expectedHash := sha256.Hash(`Dataplane:mesh-0:dp-0:1:192.168.0.1:,Dataplane:mesh-0:dp-10:1:192.168.0.1:,Dataplane:mesh-0:dp-11:1:192.168.0.1:,Dataplane:mesh-0:dp-12:1:192.168.0.1:,Dataplane:mesh-0:dp-13:1:192.168.0.1:,Dataplane:mesh-0:dp-14:1:192.168.0.1:,Dataplane:mesh-0:dp-15:1:192.168.0.1:,Dataplane:mesh-0:dp-16:1:192.168.0.1:,Dataplane:mesh-0:dp-17:1:192.168.0.1:,Dataplane:mesh-0:dp-18:1:192.168.0.1:,Dataplane:mesh-0:dp-19:1:192.168.0.1:,Dataplane:mesh-0:dp-1:1:192.168.0.1:,Dataplane:mesh-0:dp-20:1:192.168.0.1:,Dataplane:mesh-0:dp-21:1:192.168.0.1:,Dataplane:mesh-0:dp-22:1:192.168.0.1:,Dataplane:mesh-0:dp-23:1:192.168.0.1:,Dataplane:mesh-0:dp-24:1:192.168.0.1:,Dataplane:mesh-0:dp-25:1:192.168.0.1:,Dataplane:mesh-0:dp-26:1:192.168.0.1:,Dataplane:mesh-0:dp-27:1:192.168.0.1:,Dataplane:mesh-0:dp-28:1:192.168.0.1:,Dataplane:mesh-0:dp-29:1:192.168.0.1:,Dataplane:mesh-0:dp-2:1:192.168.0.1:,Dataplane:mesh-0:dp-30:1:192.168.0.1:,Dataplane:mesh-0:dp-31:1:192.168.0.1:,Dataplane:mesh-0:dp-32:1:192.168.0.1:,Dataplane:mesh-0:dp-33:1:192.168.0.1:,Dataplane:mesh-0:dp-34:1:192.168.0.1:,Dataplane:mesh-0:dp-35:1:192.168.0.1:,Dataplane:mesh-0:dp-36:1:192.168.0.1:,Dataplane:mesh-0:dp-37:1:192.168.0.1:,Dataplane:mesh-0:dp-38:1:192.168.0.1:,Dataplane:mesh-0:dp-39:1:192.168.0.1:,Dataplane:mesh-0:dp-3:1:192.168.0.1:,Dataplane:mesh-0:dp-40:1:192.168.0.1:,Dataplane:mesh-0:dp-41:1:192.168.0.1:,Dataplane:mesh-0:dp-42:1:192.168.0.1:,Dataplane:mesh-0:dp-43:1:192.168.0.1:,Dataplane:mesh-0:dp-44:1:192.168.0.1:,Dataplane:mesh-0:dp-45:1:192.168.0.1:,Dataplane:mesh-0:dp-46:1:192.168.0.1:,Dataplane:mesh-0:dp-47:1:192.168.0.1:,Dataplane:mesh-0:dp-48:1:192.168.0.1:,Dataplane:mesh-0:dp-49:1:192.168.0.1:,Dataplane:mesh-0:dp-4:1:192.168.0.1:,Dataplane:mesh-0:dp-50:1:192.168.0.1:,Dataplane:mesh-0:dp-51:1:192.168.0.1:,Dataplane:mesh-0:dp-52:1:192.168.0.1:,Dataplane:mesh-0:dp-53:1:192.168.0.1:,Dataplane:mesh-0:dp-54:1:192.168.0.1:,Dataplane:mesh-0:dp-55:1:192.168.0.1:,Dataplane:mesh-0:dp-56:1:192.168.0.1:,Dataplane:mesh-0:dp-57:1:192.168.0.1:,Dataplane:mesh-0:dp-58:1:192.168.0.1:,Dataplane:mesh-0:dp-59:1:192.168.0.1:,Dataplane:mesh-0:dp-5:1:192.168.0.1:,Dataplane:mesh-0:dp-60:1:192.168.0.1:,Dataplane:mesh-0:dp-61:1:192.168.0.1:,Dataplane:mesh-0:dp-62:1:192.168.0.1:,Dataplane:mesh-0:dp-63:1:192.168.0.1:,Dataplane:mesh-0:dp-64:1:192.168.0.1:,Dataplane:mesh-0:dp-65:1:192.168.0.1:,Dataplane:mesh-0:dp-66:1:192.168.0.1:,Dataplane:mesh-0:dp-67:1:192.168.0.1:,Dataplane:mesh-0:dp-68:1:192.168.0.1:,Dataplane:mesh-0:dp-69:1:192.168.0.1:,Dataplane:mesh-0:dp-6:1:192.168.0.1:,Dataplane:mesh-0:dp-70:1:192.168.0.1:,Dataplane:mesh-0:dp-71:1:192.168.0.1:,Dataplane:mesh-0:dp-72:1:192.168.0.1:,Dataplane:mesh-0:dp-73:1:192.168.0.1:,Dataplane:mesh-0:dp-74:1:192.168.0.1:,Dataplane:mesh-0:dp-75:1:192.168.0.1:,Dataplane:mesh-0:dp-76:1:192.168.0.1:,Dataplane:mesh-0:dp-77:1:192.168.0.1:,Dataplane:mesh-0:dp-78:1:192.168.0.1:,Dataplane:mesh-0:dp-79:1:192.168.0.1:,Dataplane:mesh-0:dp-7:1:192.168.0.1:,Dataplane:mesh-0:dp-80:1:192.168.0.1:,Dataplane:mesh-0:dp-81:1:192.168.0.1:,Dataplane:mesh-0:dp-82:1:192.168.0.1:,Dataplane:mesh-0:dp-83:1:192.168.0.1:,Dataplane:mesh-0:dp-84:1:192.168.0.1:,Dataplane:mesh-0:dp-85:1:192.168.0.1:,Dataplane:mesh-0:dp-86:1:192.168.0.1:,Dataplane:mesh-0:dp-87:1:192.168.0.1:,Dataplane:mesh-0:dp-88:1:192.168.0.1:,Dataplane:mesh-0:dp-89:1:192.168.0.1:,Dataplane:mesh-0:dp-8:1:192.168.0.1:,Dataplane:mesh-0:dp-90:1:192.168.0.1:,Dataplane:mesh-0:dp-91:1:192.168.0.1:,Dataplane:mesh-0:dp-92:1:192.168.0.1:,Dataplane:mesh-0:dp-93:1:192.168.0.1:,Dataplane:mesh-0:dp-94:1:192.168.0.1:,Dataplane:mesh-0:dp-95:1:192.168.0.1:,Dataplane:mesh-0:dp-96:1:192.168.0.1:,Dataplane:mesh-0:dp-97:1:192.168.0.1:,Dataplane:mesh-0:dp-98:1:192.168.0.1:,Dataplane:mesh-0:dp-99:1:192.168.0.1:,Dataplane:mesh-0:dp-9:1:192.168.0.1:,Mesh::mesh-0:1,TrafficRoute:mesh-0:tr-0:1,TrafficRoute:mesh-0:tr-10:1,TrafficRoute:mesh-0:tr-11:1,TrafficRoute:mesh-0:tr-12:1,TrafficRoute:mesh-0:tr-13:1,TrafficRoute:mesh-0:tr-14:1,TrafficRoute:mesh-0:tr-15:1,TrafficRoute:mesh-0:tr-16:1,TrafficRoute:mesh-0:tr-17:1,TrafficRoute:mesh-0:tr-18:1,TrafficRoute:mesh-0:tr-19:1,TrafficRoute:mesh-0:tr-1:1,TrafficRoute:mesh-0:tr-20:1,TrafficRoute:mesh-0:tr-21:1,TrafficRoute:mesh-0:tr-22:1,TrafficRoute:mesh-0:tr-23:1,TrafficRoute:mesh-0:tr-24:1,TrafficRoute:mesh-0:tr-25:1,TrafficRoute:mesh-0:tr-26:1,TrafficRoute:mesh-0:tr-27:1,TrafficRoute:mesh-0:tr-28:1,TrafficRoute:mesh-0:tr-29:1,TrafficRoute:mesh-0:tr-2:1,TrafficRoute:mesh-0:tr-30:1,TrafficRoute:mesh-0:tr-31:1,TrafficRoute:mesh-0:tr-32:1,TrafficRoute:mesh-0:tr-33:1,TrafficRoute:mesh-0:tr-34:1,TrafficRoute:mesh-0:tr-35:1,TrafficRoute:mesh-0:tr-36:1,TrafficRoute:mesh-0:tr-37:1,TrafficRoute:mesh-0:tr-38:1,TrafficRoute:mesh-0:tr-39:1,TrafficRoute:mesh-0:tr-3:1,TrafficRoute:mesh-0:tr-40:1,TrafficRoute:mesh-0:tr-41:1,TrafficRoute:mesh-0:tr-42:1,TrafficRoute:mesh-0:tr-43:1,TrafficRoute:mesh-0:tr-44:1,TrafficRoute:mesh-0:tr-45:1,TrafficRoute:mesh-0:tr-46:1,TrafficRoute:mesh-0:tr-47:1,TrafficRoute:mesh-0:tr-48:1,TrafficRoute:mesh-0:tr-49:1,TrafficRoute:mesh-0:tr-4:1,TrafficRoute:mesh-0:tr-50:1,TrafficRoute:mesh-0:tr-51:1,TrafficRoute:mesh-0:tr-52:1,TrafficRoute:mesh-0:tr-53:1,TrafficRoute:mesh-0:tr-54:1,TrafficRoute:mesh-0:tr-55:1,TrafficRoute:mesh-0:tr-56:1,TrafficRoute:mesh-0:tr-57:1,TrafficRoute:mesh-0:tr-58:1,TrafficRoute:mesh-0:tr-59:1,TrafficRoute:mesh-0:tr-5:1,TrafficRoute:mesh-0:tr-60:1,TrafficRoute:mesh-0:tr-61:1,TrafficRoute:mesh-0:tr-62:1,TrafficRoute:mesh-0:tr-63:1,TrafficRoute:mesh-0:tr-64:1,TrafficRoute:mesh-0:tr-65:1,TrafficRoute:mesh-0:tr-66:1,TrafficRoute:mesh-0:tr-67:1,TrafficRoute:mesh-0:tr-68:1,TrafficRoute:mesh-0:tr-69:1,TrafficRoute:mesh-0:tr-6:1,TrafficRoute:mesh-0:tr-70:1,TrafficRoute:mesh-0:tr-71:1,TrafficRoute:mesh-0:tr-72:1,TrafficRoute:mesh-0:tr-73:1,TrafficRoute:mesh-0:tr-74:1,TrafficRoute:mesh-0:tr-75:1,TrafficRoute:mesh-0:tr-76:1,TrafficRoute:mesh-0:tr-77:1,TrafficRoute:mesh-0:tr-78:1,TrafficRoute:mesh-0:tr-79:1,TrafficRoute:mesh-0:tr-7:1,TrafficRoute:mesh-0:tr-80:1,TrafficRoute:mesh-0:tr-81:1,TrafficRoute:mesh-0:tr-82:1,TrafficRoute:mesh-0:tr-83:1,TrafficRoute:mesh-0:tr-84:1,TrafficRoute:mesh-0:tr-85:1,TrafficRoute:mesh-0:tr-86:1,TrafficRoute:mesh-0:tr-87:1,TrafficRoute:mesh-0:tr-88:1,TrafficRoute:mesh-0:tr-89:1,TrafficRoute:mesh-0:tr-8:1,TrafficRoute:mesh-0:tr-90:1,TrafficRoute:mesh-0:tr-91:1,TrafficRoute:mesh-0:tr-92:1,TrafficRoute:mesh-0:tr-93:1,TrafficRoute:mesh-0:tr-94:1,TrafficRoute:mesh-0:tr-95:1,TrafficRoute:mesh-0:tr-96:1,TrafficRoute:mesh-0:tr-97:1,TrafficRoute:mesh-0:tr-98:1,TrafficRoute:mesh-0:tr-99:1,TrafficRoute:mesh-0:tr-9:1`)
		Expect(hash).To(Equal(expectedHash))
		Expect(countingManager.getQueries).To(Equal(1))  // one Get to obtain Mesh
		Expect(countingManager.listQueries).To(Equal(2)) // 2 List to fetch Dataplanes and TrafficRoutes

		By("getting cached Hash")
		_, err = meshCache.GetHash(context.Background(), "mesh-0")
		Expect(err).ToNot(HaveOccurred())
		Expect(countingManager.getQueries).To(Equal(1))  // should be the same
		Expect(countingManager.listQueries).To(Equal(2)) // should be the same

		By("updating Dataplane in store and waiting until cache invalidation")
		dp := mesh_core.NewDataplaneResource()
		err = s.Get(context.Background(), dp, core_store.GetByKey("dp-1", "mesh-0"))
		Expect(err).ToNot(HaveOccurred())
		dp.Spec.Networking.Address = "1.1.1.1"
		err = s.Update(context.Background(), dp)
		Expect(err).ToNot(HaveOccurred())

		<-time.After(expiration)

		hash, err = meshCache.GetHash(context.Background(), "mesh-0")
		Expect(err).ToNot(HaveOccurred())
		expectedHash = sha256.Hash(`Dataplane:mesh-0:dp-0:1:192.168.0.1:,Dataplane:mesh-0:dp-10:1:192.168.0.1:,Dataplane:mesh-0:dp-11:1:192.168.0.1:,Dataplane:mesh-0:dp-12:1:192.168.0.1:,Dataplane:mesh-0:dp-13:1:192.168.0.1:,Dataplane:mesh-0:dp-14:1:192.168.0.1:,Dataplane:mesh-0:dp-15:1:192.168.0.1:,Dataplane:mesh-0:dp-16:1:192.168.0.1:,Dataplane:mesh-0:dp-17:1:192.168.0.1:,Dataplane:mesh-0:dp-18:1:192.168.0.1:,Dataplane:mesh-0:dp-19:1:192.168.0.1:,Dataplane:mesh-0:dp-1:2:1.1.1.1:,Dataplane:mesh-0:dp-20:1:192.168.0.1:,Dataplane:mesh-0:dp-21:1:192.168.0.1:,Dataplane:mesh-0:dp-22:1:192.168.0.1:,Dataplane:mesh-0:dp-23:1:192.168.0.1:,Dataplane:mesh-0:dp-24:1:192.168.0.1:,Dataplane:mesh-0:dp-25:1:192.168.0.1:,Dataplane:mesh-0:dp-26:1:192.168.0.1:,Dataplane:mesh-0:dp-27:1:192.168.0.1:,Dataplane:mesh-0:dp-28:1:192.168.0.1:,Dataplane:mesh-0:dp-29:1:192.168.0.1:,Dataplane:mesh-0:dp-2:1:192.168.0.1:,Dataplane:mesh-0:dp-30:1:192.168.0.1:,Dataplane:mesh-0:dp-31:1:192.168.0.1:,Dataplane:mesh-0:dp-32:1:192.168.0.1:,Dataplane:mesh-0:dp-33:1:192.168.0.1:,Dataplane:mesh-0:dp-34:1:192.168.0.1:,Dataplane:mesh-0:dp-35:1:192.168.0.1:,Dataplane:mesh-0:dp-36:1:192.168.0.1:,Dataplane:mesh-0:dp-37:1:192.168.0.1:,Dataplane:mesh-0:dp-38:1:192.168.0.1:,Dataplane:mesh-0:dp-39:1:192.168.0.1:,Dataplane:mesh-0:dp-3:1:192.168.0.1:,Dataplane:mesh-0:dp-40:1:192.168.0.1:,Dataplane:mesh-0:dp-41:1:192.168.0.1:,Dataplane:mesh-0:dp-42:1:192.168.0.1:,Dataplane:mesh-0:dp-43:1:192.168.0.1:,Dataplane:mesh-0:dp-44:1:192.168.0.1:,Dataplane:mesh-0:dp-45:1:192.168.0.1:,Dataplane:mesh-0:dp-46:1:192.168.0.1:,Dataplane:mesh-0:dp-47:1:192.168.0.1:,Dataplane:mesh-0:dp-48:1:192.168.0.1:,Dataplane:mesh-0:dp-49:1:192.168.0.1:,Dataplane:mesh-0:dp-4:1:192.168.0.1:,Dataplane:mesh-0:dp-50:1:192.168.0.1:,Dataplane:mesh-0:dp-51:1:192.168.0.1:,Dataplane:mesh-0:dp-52:1:192.168.0.1:,Dataplane:mesh-0:dp-53:1:192.168.0.1:,Dataplane:mesh-0:dp-54:1:192.168.0.1:,Dataplane:mesh-0:dp-55:1:192.168.0.1:,Dataplane:mesh-0:dp-56:1:192.168.0.1:,Dataplane:mesh-0:dp-57:1:192.168.0.1:,Dataplane:mesh-0:dp-58:1:192.168.0.1:,Dataplane:mesh-0:dp-59:1:192.168.0.1:,Dataplane:mesh-0:dp-5:1:192.168.0.1:,Dataplane:mesh-0:dp-60:1:192.168.0.1:,Dataplane:mesh-0:dp-61:1:192.168.0.1:,Dataplane:mesh-0:dp-62:1:192.168.0.1:,Dataplane:mesh-0:dp-63:1:192.168.0.1:,Dataplane:mesh-0:dp-64:1:192.168.0.1:,Dataplane:mesh-0:dp-65:1:192.168.0.1:,Dataplane:mesh-0:dp-66:1:192.168.0.1:,Dataplane:mesh-0:dp-67:1:192.168.0.1:,Dataplane:mesh-0:dp-68:1:192.168.0.1:,Dataplane:mesh-0:dp-69:1:192.168.0.1:,Dataplane:mesh-0:dp-6:1:192.168.0.1:,Dataplane:mesh-0:dp-70:1:192.168.0.1:,Dataplane:mesh-0:dp-71:1:192.168.0.1:,Dataplane:mesh-0:dp-72:1:192.168.0.1:,Dataplane:mesh-0:dp-73:1:192.168.0.1:,Dataplane:mesh-0:dp-74:1:192.168.0.1:,Dataplane:mesh-0:dp-75:1:192.168.0.1:,Dataplane:mesh-0:dp-76:1:192.168.0.1:,Dataplane:mesh-0:dp-77:1:192.168.0.1:,Dataplane:mesh-0:dp-78:1:192.168.0.1:,Dataplane:mesh-0:dp-79:1:192.168.0.1:,Dataplane:mesh-0:dp-7:1:192.168.0.1:,Dataplane:mesh-0:dp-80:1:192.168.0.1:,Dataplane:mesh-0:dp-81:1:192.168.0.1:,Dataplane:mesh-0:dp-82:1:192.168.0.1:,Dataplane:mesh-0:dp-83:1:192.168.0.1:,Dataplane:mesh-0:dp-84:1:192.168.0.1:,Dataplane:mesh-0:dp-85:1:192.168.0.1:,Dataplane:mesh-0:dp-86:1:192.168.0.1:,Dataplane:mesh-0:dp-87:1:192.168.0.1:,Dataplane:mesh-0:dp-88:1:192.168.0.1:,Dataplane:mesh-0:dp-89:1:192.168.0.1:,Dataplane:mesh-0:dp-8:1:192.168.0.1:,Dataplane:mesh-0:dp-90:1:192.168.0.1:,Dataplane:mesh-0:dp-91:1:192.168.0.1:,Dataplane:mesh-0:dp-92:1:192.168.0.1:,Dataplane:mesh-0:dp-93:1:192.168.0.1:,Dataplane:mesh-0:dp-94:1:192.168.0.1:,Dataplane:mesh-0:dp-95:1:192.168.0.1:,Dataplane:mesh-0:dp-96:1:192.168.0.1:,Dataplane:mesh-0:dp-97:1:192.168.0.1:,Dataplane:mesh-0:dp-98:1:192.168.0.1:,Dataplane:mesh-0:dp-99:1:192.168.0.1:,Dataplane:mesh-0:dp-9:1:192.168.0.1:,Mesh::mesh-0:1,TrafficRoute:mesh-0:tr-0:1,TrafficRoute:mesh-0:tr-10:1,TrafficRoute:mesh-0:tr-11:1,TrafficRoute:mesh-0:tr-12:1,TrafficRoute:mesh-0:tr-13:1,TrafficRoute:mesh-0:tr-14:1,TrafficRoute:mesh-0:tr-15:1,TrafficRoute:mesh-0:tr-16:1,TrafficRoute:mesh-0:tr-17:1,TrafficRoute:mesh-0:tr-18:1,TrafficRoute:mesh-0:tr-19:1,TrafficRoute:mesh-0:tr-1:1,TrafficRoute:mesh-0:tr-20:1,TrafficRoute:mesh-0:tr-21:1,TrafficRoute:mesh-0:tr-22:1,TrafficRoute:mesh-0:tr-23:1,TrafficRoute:mesh-0:tr-24:1,TrafficRoute:mesh-0:tr-25:1,TrafficRoute:mesh-0:tr-26:1,TrafficRoute:mesh-0:tr-27:1,TrafficRoute:mesh-0:tr-28:1,TrafficRoute:mesh-0:tr-29:1,TrafficRoute:mesh-0:tr-2:1,TrafficRoute:mesh-0:tr-30:1,TrafficRoute:mesh-0:tr-31:1,TrafficRoute:mesh-0:tr-32:1,TrafficRoute:mesh-0:tr-33:1,TrafficRoute:mesh-0:tr-34:1,TrafficRoute:mesh-0:tr-35:1,TrafficRoute:mesh-0:tr-36:1,TrafficRoute:mesh-0:tr-37:1,TrafficRoute:mesh-0:tr-38:1,TrafficRoute:mesh-0:tr-39:1,TrafficRoute:mesh-0:tr-3:1,TrafficRoute:mesh-0:tr-40:1,TrafficRoute:mesh-0:tr-41:1,TrafficRoute:mesh-0:tr-42:1,TrafficRoute:mesh-0:tr-43:1,TrafficRoute:mesh-0:tr-44:1,TrafficRoute:mesh-0:tr-45:1,TrafficRoute:mesh-0:tr-46:1,TrafficRoute:mesh-0:tr-47:1,TrafficRoute:mesh-0:tr-48:1,TrafficRoute:mesh-0:tr-49:1,TrafficRoute:mesh-0:tr-4:1,TrafficRoute:mesh-0:tr-50:1,TrafficRoute:mesh-0:tr-51:1,TrafficRoute:mesh-0:tr-52:1,TrafficRoute:mesh-0:tr-53:1,TrafficRoute:mesh-0:tr-54:1,TrafficRoute:mesh-0:tr-55:1,TrafficRoute:mesh-0:tr-56:1,TrafficRoute:mesh-0:tr-57:1,TrafficRoute:mesh-0:tr-58:1,TrafficRoute:mesh-0:tr-59:1,TrafficRoute:mesh-0:tr-5:1,TrafficRoute:mesh-0:tr-60:1,TrafficRoute:mesh-0:tr-61:1,TrafficRoute:mesh-0:tr-62:1,TrafficRoute:mesh-0:tr-63:1,TrafficRoute:mesh-0:tr-64:1,TrafficRoute:mesh-0:tr-65:1,TrafficRoute:mesh-0:tr-66:1,TrafficRoute:mesh-0:tr-67:1,TrafficRoute:mesh-0:tr-68:1,TrafficRoute:mesh-0:tr-69:1,TrafficRoute:mesh-0:tr-6:1,TrafficRoute:mesh-0:tr-70:1,TrafficRoute:mesh-0:tr-71:1,TrafficRoute:mesh-0:tr-72:1,TrafficRoute:mesh-0:tr-73:1,TrafficRoute:mesh-0:tr-74:1,TrafficRoute:mesh-0:tr-75:1,TrafficRoute:mesh-0:tr-76:1,TrafficRoute:mesh-0:tr-77:1,TrafficRoute:mesh-0:tr-78:1,TrafficRoute:mesh-0:tr-79:1,TrafficRoute:mesh-0:tr-7:1,TrafficRoute:mesh-0:tr-80:1,TrafficRoute:mesh-0:tr-81:1,TrafficRoute:mesh-0:tr-82:1,TrafficRoute:mesh-0:tr-83:1,TrafficRoute:mesh-0:tr-84:1,TrafficRoute:mesh-0:tr-85:1,TrafficRoute:mesh-0:tr-86:1,TrafficRoute:mesh-0:tr-87:1,TrafficRoute:mesh-0:tr-88:1,TrafficRoute:mesh-0:tr-89:1,TrafficRoute:mesh-0:tr-8:1,TrafficRoute:mesh-0:tr-90:1,TrafficRoute:mesh-0:tr-91:1,TrafficRoute:mesh-0:tr-92:1,TrafficRoute:mesh-0:tr-93:1,TrafficRoute:mesh-0:tr-94:1,TrafficRoute:mesh-0:tr-95:1,TrafficRoute:mesh-0:tr-96:1,TrafficRoute:mesh-0:tr-97:1,TrafficRoute:mesh-0:tr-98:1,TrafficRoute:mesh-0:tr-99:1,TrafficRoute:mesh-0:tr-9:1`)
		Expect(hash).To(Equal(expectedHash))
		Expect(countingManager.getQueries).To(Equal(2))
		Expect(countingManager.listQueries).To(Equal(4))
	})

	It("should count hashes independently for each mesh", func() {
		hash0, err := meshCache.GetHash(context.Background(), "mesh-0")
		Expect(err).ToNot(HaveOccurred())

		hash1, err := meshCache.GetHash(context.Background(), "mesh-1")
		Expect(err).ToNot(HaveOccurred())

		hash2, err := meshCache.GetHash(context.Background(), "mesh-2")
		Expect(err).ToNot(HaveOccurred())

		dp := mesh_core.NewDataplaneResource()
		err = s.Get(context.Background(), dp, core_store.GetByKey("dp-1", "mesh-0"))
		Expect(err).ToNot(HaveOccurred())
		dp.Spec.Networking.Address = "1.1.1.1"
		err = s.Update(context.Background(), dp)
		Expect(err).ToNot(HaveOccurred())

		<-time.After(expiration)

		updHash0, err := meshCache.GetHash(context.Background(), "mesh-0")
		Expect(err).ToNot(HaveOccurred())

		updHash1, err := meshCache.GetHash(context.Background(), "mesh-1")
		Expect(err).ToNot(HaveOccurred())

		updHash2, err := meshCache.GetHash(context.Background(), "mesh-2")
		Expect(err).ToNot(HaveOccurred())

		Expect(hash0).ToNot(Equal(updHash0))
		Expect(hash1).To(Equal(updHash1))
		Expect(hash2).To(Equal(updHash2))
	})

	It("should cache concurrent Get() requests", func() {
		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				s, err := meshCache.GetHash(context.Background(), "mesh-0")
				Expect(len(s) > 0).To(BeTrue())
				Expect(err).ToNot(HaveOccurred())
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
		By("getting Hash for the first time")
		hash, err := meshCache.GetHash(context.Background(), "mesh-0")
		Expect(countingManager.getQueries).To(Equal(1)) // should be the same
		Expect(err).To(HaveOccurred())
		Expect(hash).To(BeEmpty())

		By("getting Hash calls again")
		_, err = meshCache.GetHash(context.Background(), "mesh-0")
		Expect(err).To(HaveOccurred())
		Expect(countingManager.getQueries).To(Equal(2)) // should be increased by one (errors are not cached)
		Expect(err).To(HaveOccurred())
		Expect(hash).To(BeEmpty())

		By("Getting the hash once manager is fixed")
		countingManager.err = nil
		hash, err = meshCache.GetHash(context.Background(), "mesh-0")
		Expect(err).ToNot(HaveOccurred())
		expectedHash := sha256.Hash(`Dataplane:mesh-0:dp-0:1:192.168.0.1:,Dataplane:mesh-0:dp-10:1:192.168.0.1:,Dataplane:mesh-0:dp-11:1:192.168.0.1:,Dataplane:mesh-0:dp-12:1:192.168.0.1:,Dataplane:mesh-0:dp-13:1:192.168.0.1:,Dataplane:mesh-0:dp-14:1:192.168.0.1:,Dataplane:mesh-0:dp-15:1:192.168.0.1:,Dataplane:mesh-0:dp-16:1:192.168.0.1:,Dataplane:mesh-0:dp-17:1:192.168.0.1:,Dataplane:mesh-0:dp-18:1:192.168.0.1:,Dataplane:mesh-0:dp-19:1:192.168.0.1:,Dataplane:mesh-0:dp-1:1:192.168.0.1:,Dataplane:mesh-0:dp-20:1:192.168.0.1:,Dataplane:mesh-0:dp-21:1:192.168.0.1:,Dataplane:mesh-0:dp-22:1:192.168.0.1:,Dataplane:mesh-0:dp-23:1:192.168.0.1:,Dataplane:mesh-0:dp-24:1:192.168.0.1:,Dataplane:mesh-0:dp-25:1:192.168.0.1:,Dataplane:mesh-0:dp-26:1:192.168.0.1:,Dataplane:mesh-0:dp-27:1:192.168.0.1:,Dataplane:mesh-0:dp-28:1:192.168.0.1:,Dataplane:mesh-0:dp-29:1:192.168.0.1:,Dataplane:mesh-0:dp-2:1:192.168.0.1:,Dataplane:mesh-0:dp-30:1:192.168.0.1:,Dataplane:mesh-0:dp-31:1:192.168.0.1:,Dataplane:mesh-0:dp-32:1:192.168.0.1:,Dataplane:mesh-0:dp-33:1:192.168.0.1:,Dataplane:mesh-0:dp-34:1:192.168.0.1:,Dataplane:mesh-0:dp-35:1:192.168.0.1:,Dataplane:mesh-0:dp-36:1:192.168.0.1:,Dataplane:mesh-0:dp-37:1:192.168.0.1:,Dataplane:mesh-0:dp-38:1:192.168.0.1:,Dataplane:mesh-0:dp-39:1:192.168.0.1:,Dataplane:mesh-0:dp-3:1:192.168.0.1:,Dataplane:mesh-0:dp-40:1:192.168.0.1:,Dataplane:mesh-0:dp-41:1:192.168.0.1:,Dataplane:mesh-0:dp-42:1:192.168.0.1:,Dataplane:mesh-0:dp-43:1:192.168.0.1:,Dataplane:mesh-0:dp-44:1:192.168.0.1:,Dataplane:mesh-0:dp-45:1:192.168.0.1:,Dataplane:mesh-0:dp-46:1:192.168.0.1:,Dataplane:mesh-0:dp-47:1:192.168.0.1:,Dataplane:mesh-0:dp-48:1:192.168.0.1:,Dataplane:mesh-0:dp-49:1:192.168.0.1:,Dataplane:mesh-0:dp-4:1:192.168.0.1:,Dataplane:mesh-0:dp-50:1:192.168.0.1:,Dataplane:mesh-0:dp-51:1:192.168.0.1:,Dataplane:mesh-0:dp-52:1:192.168.0.1:,Dataplane:mesh-0:dp-53:1:192.168.0.1:,Dataplane:mesh-0:dp-54:1:192.168.0.1:,Dataplane:mesh-0:dp-55:1:192.168.0.1:,Dataplane:mesh-0:dp-56:1:192.168.0.1:,Dataplane:mesh-0:dp-57:1:192.168.0.1:,Dataplane:mesh-0:dp-58:1:192.168.0.1:,Dataplane:mesh-0:dp-59:1:192.168.0.1:,Dataplane:mesh-0:dp-5:1:192.168.0.1:,Dataplane:mesh-0:dp-60:1:192.168.0.1:,Dataplane:mesh-0:dp-61:1:192.168.0.1:,Dataplane:mesh-0:dp-62:1:192.168.0.1:,Dataplane:mesh-0:dp-63:1:192.168.0.1:,Dataplane:mesh-0:dp-64:1:192.168.0.1:,Dataplane:mesh-0:dp-65:1:192.168.0.1:,Dataplane:mesh-0:dp-66:1:192.168.0.1:,Dataplane:mesh-0:dp-67:1:192.168.0.1:,Dataplane:mesh-0:dp-68:1:192.168.0.1:,Dataplane:mesh-0:dp-69:1:192.168.0.1:,Dataplane:mesh-0:dp-6:1:192.168.0.1:,Dataplane:mesh-0:dp-70:1:192.168.0.1:,Dataplane:mesh-0:dp-71:1:192.168.0.1:,Dataplane:mesh-0:dp-72:1:192.168.0.1:,Dataplane:mesh-0:dp-73:1:192.168.0.1:,Dataplane:mesh-0:dp-74:1:192.168.0.1:,Dataplane:mesh-0:dp-75:1:192.168.0.1:,Dataplane:mesh-0:dp-76:1:192.168.0.1:,Dataplane:mesh-0:dp-77:1:192.168.0.1:,Dataplane:mesh-0:dp-78:1:192.168.0.1:,Dataplane:mesh-0:dp-79:1:192.168.0.1:,Dataplane:mesh-0:dp-7:1:192.168.0.1:,Dataplane:mesh-0:dp-80:1:192.168.0.1:,Dataplane:mesh-0:dp-81:1:192.168.0.1:,Dataplane:mesh-0:dp-82:1:192.168.0.1:,Dataplane:mesh-0:dp-83:1:192.168.0.1:,Dataplane:mesh-0:dp-84:1:192.168.0.1:,Dataplane:mesh-0:dp-85:1:192.168.0.1:,Dataplane:mesh-0:dp-86:1:192.168.0.1:,Dataplane:mesh-0:dp-87:1:192.168.0.1:,Dataplane:mesh-0:dp-88:1:192.168.0.1:,Dataplane:mesh-0:dp-89:1:192.168.0.1:,Dataplane:mesh-0:dp-8:1:192.168.0.1:,Dataplane:mesh-0:dp-90:1:192.168.0.1:,Dataplane:mesh-0:dp-91:1:192.168.0.1:,Dataplane:mesh-0:dp-92:1:192.168.0.1:,Dataplane:mesh-0:dp-93:1:192.168.0.1:,Dataplane:mesh-0:dp-94:1:192.168.0.1:,Dataplane:mesh-0:dp-95:1:192.168.0.1:,Dataplane:mesh-0:dp-96:1:192.168.0.1:,Dataplane:mesh-0:dp-97:1:192.168.0.1:,Dataplane:mesh-0:dp-98:1:192.168.0.1:,Dataplane:mesh-0:dp-99:1:192.168.0.1:,Dataplane:mesh-0:dp-9:1:192.168.0.1:,Mesh::mesh-0:1,TrafficRoute:mesh-0:tr-0:1,TrafficRoute:mesh-0:tr-10:1,TrafficRoute:mesh-0:tr-11:1,TrafficRoute:mesh-0:tr-12:1,TrafficRoute:mesh-0:tr-13:1,TrafficRoute:mesh-0:tr-14:1,TrafficRoute:mesh-0:tr-15:1,TrafficRoute:mesh-0:tr-16:1,TrafficRoute:mesh-0:tr-17:1,TrafficRoute:mesh-0:tr-18:1,TrafficRoute:mesh-0:tr-19:1,TrafficRoute:mesh-0:tr-1:1,TrafficRoute:mesh-0:tr-20:1,TrafficRoute:mesh-0:tr-21:1,TrafficRoute:mesh-0:tr-22:1,TrafficRoute:mesh-0:tr-23:1,TrafficRoute:mesh-0:tr-24:1,TrafficRoute:mesh-0:tr-25:1,TrafficRoute:mesh-0:tr-26:1,TrafficRoute:mesh-0:tr-27:1,TrafficRoute:mesh-0:tr-28:1,TrafficRoute:mesh-0:tr-29:1,TrafficRoute:mesh-0:tr-2:1,TrafficRoute:mesh-0:tr-30:1,TrafficRoute:mesh-0:tr-31:1,TrafficRoute:mesh-0:tr-32:1,TrafficRoute:mesh-0:tr-33:1,TrafficRoute:mesh-0:tr-34:1,TrafficRoute:mesh-0:tr-35:1,TrafficRoute:mesh-0:tr-36:1,TrafficRoute:mesh-0:tr-37:1,TrafficRoute:mesh-0:tr-38:1,TrafficRoute:mesh-0:tr-39:1,TrafficRoute:mesh-0:tr-3:1,TrafficRoute:mesh-0:tr-40:1,TrafficRoute:mesh-0:tr-41:1,TrafficRoute:mesh-0:tr-42:1,TrafficRoute:mesh-0:tr-43:1,TrafficRoute:mesh-0:tr-44:1,TrafficRoute:mesh-0:tr-45:1,TrafficRoute:mesh-0:tr-46:1,TrafficRoute:mesh-0:tr-47:1,TrafficRoute:mesh-0:tr-48:1,TrafficRoute:mesh-0:tr-49:1,TrafficRoute:mesh-0:tr-4:1,TrafficRoute:mesh-0:tr-50:1,TrafficRoute:mesh-0:tr-51:1,TrafficRoute:mesh-0:tr-52:1,TrafficRoute:mesh-0:tr-53:1,TrafficRoute:mesh-0:tr-54:1,TrafficRoute:mesh-0:tr-55:1,TrafficRoute:mesh-0:tr-56:1,TrafficRoute:mesh-0:tr-57:1,TrafficRoute:mesh-0:tr-58:1,TrafficRoute:mesh-0:tr-59:1,TrafficRoute:mesh-0:tr-5:1,TrafficRoute:mesh-0:tr-60:1,TrafficRoute:mesh-0:tr-61:1,TrafficRoute:mesh-0:tr-62:1,TrafficRoute:mesh-0:tr-63:1,TrafficRoute:mesh-0:tr-64:1,TrafficRoute:mesh-0:tr-65:1,TrafficRoute:mesh-0:tr-66:1,TrafficRoute:mesh-0:tr-67:1,TrafficRoute:mesh-0:tr-68:1,TrafficRoute:mesh-0:tr-69:1,TrafficRoute:mesh-0:tr-6:1,TrafficRoute:mesh-0:tr-70:1,TrafficRoute:mesh-0:tr-71:1,TrafficRoute:mesh-0:tr-72:1,TrafficRoute:mesh-0:tr-73:1,TrafficRoute:mesh-0:tr-74:1,TrafficRoute:mesh-0:tr-75:1,TrafficRoute:mesh-0:tr-76:1,TrafficRoute:mesh-0:tr-77:1,TrafficRoute:mesh-0:tr-78:1,TrafficRoute:mesh-0:tr-79:1,TrafficRoute:mesh-0:tr-7:1,TrafficRoute:mesh-0:tr-80:1,TrafficRoute:mesh-0:tr-81:1,TrafficRoute:mesh-0:tr-82:1,TrafficRoute:mesh-0:tr-83:1,TrafficRoute:mesh-0:tr-84:1,TrafficRoute:mesh-0:tr-85:1,TrafficRoute:mesh-0:tr-86:1,TrafficRoute:mesh-0:tr-87:1,TrafficRoute:mesh-0:tr-88:1,TrafficRoute:mesh-0:tr-89:1,TrafficRoute:mesh-0:tr-8:1,TrafficRoute:mesh-0:tr-90:1,TrafficRoute:mesh-0:tr-91:1,TrafficRoute:mesh-0:tr-92:1,TrafficRoute:mesh-0:tr-93:1,TrafficRoute:mesh-0:tr-94:1,TrafficRoute:mesh-0:tr-95:1,TrafficRoute:mesh-0:tr-96:1,TrafficRoute:mesh-0:tr-97:1,TrafficRoute:mesh-0:tr-98:1,TrafficRoute:mesh-0:tr-99:1,TrafficRoute:mesh-0:tr-9:1`)
		Expect(hash).To(Equal(expectedHash))
		Expect(countingManager.getQueries).To(Equal(3))  // one Get to obtain Mesh
		Expect(countingManager.listQueries).To(Equal(2)) // 2 List to fetch Dataplanes and TrafficRoutes

		By("Now it should cache the hash once manager is fixed")
		countingManager.err = nil
		hash, err = meshCache.GetHash(context.Background(), "mesh-0")
		Expect(err).ToNot(HaveOccurred())
		Expect(hash).To(Equal(expectedHash))
		Expect(countingManager.getQueries).To(Equal(3))  // one Get to obtain Mesh
		Expect(countingManager.listQueries).To(Equal(2)) // 2 List to fetch Dataplanes and TrafficRoutes
	})
})
