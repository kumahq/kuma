package server_test

import (
	"context"
	"fmt"
	"net"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	mesh_core "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/pkg/test/resources/model"
	"github.com/kumahq/kuma/pkg/xds/server"
)

var _ = Describe("MeshSnapshot", func() {
	testDataplaneResources := func(n int, version, address string) []*mesh_core.DataplaneResource {
		rv := []*mesh_core.DataplaneResource{}
		for i := 0; i < n; i++ {
			rv = append(rv, &mesh_core.DataplaneResource{
				Meta: &model.ResourceMeta{Mesh: "default", Name: fmt.Sprintf("dp-%d", i), Version: version},
				Spec: mesh_proto.Dataplane{
					Networking: &mesh_proto.Dataplane_Networking{
						Address: address,
					},
				},
			})
		}
		return rv
	}
	testTrafficRouteResources := func(n int, version string) []*mesh_core.TrafficRouteResource {
		rv := []*mesh_core.TrafficRouteResource{}
		for i := 0; i < n; i++ {
			rv = append(rv, &mesh_core.TrafficRouteResource{
				Meta: &model.ResourceMeta{Mesh: "default", Name: fmt.Sprintf("tr-%d", i), Version: version},
			})
		}
		return rv
	}

	Context("Hash()", func() {
		It("should have the same hash for the same snapshots", func() {
			s1 := server.MeshSnapshot{
				Mesh: &mesh_core.MeshResource{Meta: &model.ResourceMeta{Mesh: "default", Name: "default"}},
				Resources: map[core_model.ResourceType]core_model.ResourceList{
					mesh_core.DataplaneType:    &mesh_core.DataplaneResourceList{Items: testDataplaneResources(10000, "version-3", "localhost")},
					mesh_core.TrafficRouteType: &mesh_core.TrafficRouteResourceList{Items: testTrafficRouteResources(10000, "version-1")},
				},
			}
			s2 := server.MeshSnapshot{
				Mesh: &mesh_core.MeshResource{Meta: &model.ResourceMeta{Mesh: "default", Name: "default"}},
				Resources: map[core_model.ResourceType]core_model.ResourceList{
					mesh_core.DataplaneType:    &mesh_core.DataplaneResourceList{Items: testDataplaneResources(10000, "version-3", "localhost")},
					mesh_core.TrafficRouteType: &mesh_core.TrafficRouteResourceList{Items: testTrafficRouteResources(10000, "version-1")},
				},
			}

			Expect(s1.Hash()).To(Equal(s2.Hash()))
		})
	})

	Context("List()", func() {
		const baseLen = 10000
		var snapshot server.MeshSnapshot

		BeforeEach(func() {
			snapshot = server.MeshSnapshot{
				Mesh: &mesh_core.MeshResource{Meta: &model.ResourceMeta{Mesh: "default", Name: "default"}},
				Resources: map[core_model.ResourceType]core_model.ResourceList{
					mesh_core.DataplaneType:    &mesh_core.DataplaneResourceList{Items: testDataplaneResources(baseLen, "version-3", "localhost")},
					mesh_core.TrafficRouteType: &mesh_core.TrafficRouteResourceList{Items: testTrafficRouteResources(baseLen, "version-1")},
				},
			}
		})
		It("should return ResourceList of given type", func() {
			trList, exists := snapshot.Resources[mesh_core.TrafficRouteType]
			Expect(exists).To(BeTrue())
			Expect(trList.GetItems()).To(HaveLen(baseLen))
		})
	})

	Context("GetMeshSnapshot", func() {
		const baseLen = 100
		var s store.ResourceStore
		var expectedSnapshot server.MeshSnapshot

		BeforeEach(func() {
			s = memory.NewStore()
			err := s.Create(context.Background(), &mesh_core.MeshResource{}, store.CreateByKey("default", "default"))
			Expect(err).ToNot(HaveOccurred())

			for _, dp := range testDataplaneResources(baseLen, "v1", "service.test") {
				err := s.Create(context.Background(), dp, store.CreateBy(core_model.MetaToResourceKey(dp.GetMeta())))
				Expect(err).ToNot(HaveOccurred())
			}
			for _, tr := range testTrafficRouteResources(baseLen, "v2") {
				err := s.Create(context.Background(), tr, store.CreateBy(core_model.MetaToResourceKey(tr.GetMeta())))
				Expect(err).ToNot(HaveOccurred())
			}

			expectedSnapshot = server.MeshSnapshot{
				Mesh: &mesh_core.MeshResource{Meta: &model.ResourceMeta{Mesh: "default", Name: "default", Version: "1"}},
				Resources: map[core_model.ResourceType]core_model.ResourceList{
					mesh_core.DataplaneType:    &mesh_core.DataplaneResourceList{Items: testDataplaneResources(baseLen, "1", "192.168.0.1")},
					mesh_core.TrafficRouteType: &mesh_core.TrafficRouteResourceList{Items: testTrafficRouteResources(baseLen, "1")},
				},
			}
		})

		It("should create snapshot using ResourceStore which is equal to predefined snapshot", func() {
			ipFunc := func(s string) ([]net.IP, error) {
				switch s {
				case "service.test":
					return []net.IP{net.ParseIP("192.168.0.1")}, nil
				}
				return nil, nil
			}
			actualSnapshot, err := server.GetMeshSnapshot(context.Background(), "default", s,
				[]core_model.ResourceType{mesh_core.DataplaneType, mesh_core.TrafficRouteType}, ipFunc)
			Expect(err).ToNot(HaveOccurred())
			Expect(actualSnapshot.Hash()).To(Equal(expectedSnapshot.Hash()))
		})

		It("should create snapshot with another hash if DNS changed", func() {
			ipFunc := func(s string) ([]net.IP, error) {
				switch s {
				case "service.test":
					return []net.IP{net.ParseIP("1.1.1.1")}, nil
				}
				return nil, nil
			}
			actualSnapshot, err := server.GetMeshSnapshot(context.Background(), "default", s,
				[]core_model.ResourceType{mesh_core.DataplaneType, mesh_core.TrafficRouteType}, ipFunc)
			Expect(err).ToNot(HaveOccurred())
			Expect(actualSnapshot.Hash()).ToNot(Equal(expectedSnapshot.Hash()))
		})
	})
})
