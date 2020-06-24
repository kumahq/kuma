package store_test

import (
	"context"
	"fmt"
	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	"github.com/Kong/kuma/pkg/core/resources/model"
	"github.com/Kong/kuma/pkg/core/resources/store"
	sync_store "github.com/Kong/kuma/pkg/kds/store"
	"github.com/Kong/kuma/pkg/plugins/resources/memory"
	model2 "github.com/Kong/kuma/pkg/test/resources/model"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("SyncResourceStore", func() {
	var s sync_store.SyncResourceStore

	meshBuilder := func(idx int) *mesh.MeshResource {
		ca := fmt.Sprintf("ca-%d", idx)
		meshName := fmt.Sprintf("mesh-%d", idx)
		return &mesh.MeshResource{
			Meta: &model2.ResourceMeta{
				Mesh: meshName,
				Name: meshName,
			},
			Spec: mesh_proto.Mesh{
				Mtls: &mesh_proto.Mesh_Mtls{
					EnabledBackend: ca,
					Backends: []*mesh_proto.CertificateAuthorityBackend{
						{
							Name: ca,
							Type: "builtin",
						},
					},
				},
			},
		}
	}

	BeforeEach(func() {
		s = sync_store.NewSyncResourceStore(memory.NewStore())
	})

	It("should create new resources in empty store", func() {
		upstream := &mesh.MeshResourceList{}
		idxs := []int{1, 2, 3, 4}
		for _, i := range idxs {
			m := meshBuilder(i)
			err := upstream.AddItem(m)
			Expect(err).ToNot(HaveOccurred())
		}

		err := s.Sync(upstream)
		Expect(err).ToNot(HaveOccurred())

		actual := &mesh.MeshResourceList{}
		err = s.List(context.Background(), actual)
		Expect(err).ToNot(HaveOccurred())
		Expect(actual.Items).To(Equal(upstream.Items))
	})

	It("should delete all resources", func() {
		for i := 0; i < 10; i++ {
			m := meshBuilder(i)
			err := s.Create(context.Background(), m, store.CreateBy(model.MetaToResourceKey(m.GetMeta())))
			Expect(err).ToNot(HaveOccurred())
		}

		upstream := &mesh.MeshResourceList{}
		err := s.Sync(upstream)
		Expect(err).ToNot(HaveOccurred())

		actual := &mesh.MeshResourceList{}
		err = s.List(context.Background(), actual)
		Expect(err).ToNot(HaveOccurred())
		Expect(actual.Items).To(BeEmpty())
	})

	It("should delete resources which are not represented in upstream and create new", func() {
		for i := 0; i < 10; i++ {
			m := meshBuilder(i)
			err := s.Create(context.Background(), m, store.CreateBy(model.MetaToResourceKey(m.GetMeta())))
			Expect(err).ToNot(HaveOccurred())
		}

		upstream := &mesh.MeshResourceList{}
		idxs := []int{1, 2, 7, 12}
		for _, i := range idxs {
			m := meshBuilder(i)
			err := upstream.AddItem(m)
			Expect(err).ToNot(HaveOccurred())
		}

		err := s.Sync(upstream)
		Expect(err).ToNot(HaveOccurred())

		actual := &mesh.MeshResourceList{}
		err = s.List(context.Background(), actual)
		Expect(err).ToNot(HaveOccurred())
		Expect(actual.Items).To(Equal(upstream.Items))
	})
})
