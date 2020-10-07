package cla_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/pkg/xds/cache/cla"
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
	var countingRm *countingResourcesManager
	var claCache *cla.Cache
	expiration := 500 * time.Millisecond

	BeforeEach(func() {
		s = memory.NewStore()
		countingRm = &countingResourcesManager{store: s}
		claCache = cla.NewCache(countingRm, "", expiration, func(s string) ([]net.IP, error) {
			return []net.IP{net.ParseIP(s)}, nil
		})
	})

	BeforeEach(func() {
		for i := 0; i < 1; i++ {
			mesh := fmt.Sprintf("mesh-%d", i)
			err := s.Create(context.Background(), &core_mesh.MeshResource{}, core_store.CreateByKey(mesh, mesh))
			Expect(err).ToNot(HaveOccurred())

			err = s.Create(context.Background(), &core_mesh.DataplaneResource{
				Spec: mesh_proto.Dataplane{Networking: &mesh_proto.Dataplane_Networking{
					Address: "192.168.0.1",
					Inbound: []*mesh_proto.Dataplane_Networking_Inbound{{
						Port: 1010, ServicePort: 2020, Tags: map[string]string{"kuma.io/service": "backend"},
					}},
				}},
			}, core_store.CreateByKey("dp1", mesh))
			Expect(err).ToNot(HaveOccurred())

			err = s.Create(context.Background(), &core_mesh.DataplaneResource{
				Spec: mesh_proto.Dataplane{Networking: &mesh_proto.Dataplane_Networking{
					Address: "192.168.0.2",
					Inbound: []*mesh_proto.Dataplane_Networking_Inbound{{
						Port: 1011, ServicePort: 2021, Tags: map[string]string{"kuma.io/service": "backend"},
					}},
				}},
			}, core_store.CreateByKey("dp2", mesh))
			Expect(err).ToNot(HaveOccurred())
		}
	})

	It("should cache Get() queries", func() {
		By("getting CLA for the first time")
		cla, err := claCache.GetCLA(context.Background(), "mesh-0", "backend")
		Expect(err).ToNot(HaveOccurred())
		Expect(countingRm.getQueries).To(Equal(1))
		Expect(countingRm.listQueries).To(Equal(1))

		expected, err := ioutil.ReadFile(filepath.Join("testdata", "cla.get.0.json"))
		Expect(err).ToNot(HaveOccurred())

		js, err := json.Marshal(cla)
		Expect(err).ToNot(HaveOccurred())
		Expect(js).To(MatchJSON(string(expected)))

		By("getting cached CLA")
		_, err = claCache.GetCLA(context.Background(), "mesh-0", "backend")
		Expect(err).ToNot(HaveOccurred())
		Expect(countingRm.getQueries).To(Equal(1))
		Expect(countingRm.listQueries).To(Equal(1))

		By("updating Dataplane in store and waiting until cache invalidation")
		dp := &core_mesh.DataplaneResource{}
		err = s.Get(context.Background(), dp, core_store.GetByKey("dp2", "mesh-0"))
		Expect(err).ToNot(HaveOccurred())
		dp.Spec.Networking.Address = "1.1.1.1"
		err = s.Update(context.Background(), dp)
		Expect(err).ToNot(HaveOccurred())

		<-time.After(2 * time.Second)

		cla, err = claCache.GetCLA(context.Background(), "mesh-0", "backend")
		Expect(err).ToNot(HaveOccurred())
		Expect(countingRm.getQueries).To(Equal(2))
		Expect(countingRm.listQueries).To(Equal(2))

		expected, err = ioutil.ReadFile(filepath.Join("testdata", "cla.get.1.json"))
		Expect(err).ToNot(HaveOccurred())
		js, err = json.Marshal(cla)
		Expect(err).ToNot(HaveOccurred())
		Expect(js).To(MatchJSON(expected))
	})
})
