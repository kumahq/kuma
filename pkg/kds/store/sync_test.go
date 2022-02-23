package store_test

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	sync_store "github.com/kumahq/kuma/pkg/kds/store"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	. "github.com/kumahq/kuma/pkg/test/matchers"
	model2 "github.com/kumahq/kuma/pkg/test/resources/model"
)

var _ = Describe("SyncResourceStore", func() {
	var syncer sync_store.ResourceSyncer
	var resourceStore store.ResourceStore

	meshBuilder := func(idx int) *mesh.MeshResource {
		ca := fmt.Sprintf("ca-%d", idx)
		meshName := fmt.Sprintf("mesh-%d", idx)
		return &mesh.MeshResource{
			Meta: &model2.ResourceMeta{
				Name: meshName,
			},
			Spec: &mesh_proto.Mesh{
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
		resourceStore = memory.NewStore()
		syncer = sync_store.NewResourceSyncer(core.Log, resourceStore)
	})

	It("should create new resources in empty store", func() {
		upstream := &mesh.MeshResourceList{}
		idxs := []int{1, 2, 3, 4}
		for _, i := range idxs {
			m := meshBuilder(i)
			err := upstream.AddItem(m)
			Expect(err).ToNot(HaveOccurred())
		}

		err := syncer.Sync(upstream)
		Expect(err).ToNot(HaveOccurred())

		actual := &mesh.MeshResourceList{}
		err = resourceStore.List(context.Background(), actual)
		Expect(err).ToNot(HaveOccurred())
		Expect(actual.Items).To(Equal(upstream.Items))
	})

	It("should delete all resources", func() {
		for i := 0; i < 10; i++ {
			m := meshBuilder(i)
			err := resourceStore.Create(context.Background(), m, store.CreateBy(model.MetaToResourceKey(m.GetMeta())))
			Expect(err).ToNot(HaveOccurred())
		}

		upstream := &mesh.MeshResourceList{}
		err := syncer.Sync(upstream)
		Expect(err).ToNot(HaveOccurred())

		actual := &mesh.MeshResourceList{}
		err = resourceStore.List(context.Background(), actual)
		Expect(err).ToNot(HaveOccurred())
		Expect(actual.Items).To(BeEmpty())
	})

	It("should delete resources which are not represented in upstream and create new", func() {
		for i := 0; i < 10; i++ {
			m := meshBuilder(i)
			err := resourceStore.Create(context.Background(), m, store.CreateBy(model.MetaToResourceKey(m.GetMeta())))
			Expect(err).ToNot(HaveOccurred())
		}

		upstream := &mesh.MeshResourceList{}
		idxs := []int{1, 2, 7, 12}
		for _, i := range idxs {
			m := meshBuilder(i)
			err := upstream.AddItem(m)
			Expect(err).ToNot(HaveOccurred())
		}

		err := syncer.Sync(upstream)
		Expect(err).ToNot(HaveOccurred())

		actual := &mesh.MeshResourceList{}
		err = resourceStore.List(context.Background(), actual)
		Expect(err).ToNot(HaveOccurred())
		Expect(len(actual.Items)).To(Equal(len(upstream.Items)))
		for i, item := range actual.Items {
			Expect(item.Spec).To(MatchProto(upstream.Items[i].Spec))
		}
	})
})
