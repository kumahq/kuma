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
	sync_store "github.com/kumahq/kuma/pkg/kds/v2/store"
	zone_client "github.com/kumahq/kuma/pkg/kds/v2/zone/client"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	. "github.com/kumahq/kuma/pkg/test/matchers"
	model2 "github.com/kumahq/kuma/pkg/test/resources/model"
)

var _ = Describe("SyncResourceStoreDelta", func() {
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
		upstreamResponse := zone_client.UpstreamResponse{}
		upstream := &mesh.MeshResourceList{}
		idxs := []int{1, 2, 3, 4}
		for _, i := range idxs {
			m := meshBuilder(i)
			err := upstream.AddItem(m)
			Expect(err).ToNot(HaveOccurred())
		}
		upstreamResponse.Type = upstream.GetItemType()
		upstreamResponse.AddedResources = upstream

		err := syncer.Sync(upstreamResponse)
		Expect(err).ToNot(HaveOccurred())

		actual := &mesh.MeshResourceList{}
		err = resourceStore.List(context.Background(), actual)
		Expect(err).ToNot(HaveOccurred())
		Expect(actual.Items).To(Equal(upstream.Items))
	})

	It("should delete all resources", func() {
		upstreamResponse := zone_client.UpstreamResponse{}
		removedResources := []string{}
		for i := 0; i < 10; i++ {
			m := meshBuilder(i)
			removedResources = append(removedResources, fmt.Sprintf("mesh-%d.", i))
			err := resourceStore.Create(context.Background(), m, store.CreateBy(model.MetaToResourceKey(m.GetMeta())))
			Expect(err).ToNot(HaveOccurred())
		}
		upstream := &mesh.MeshResourceList{}
		upstreamResponse.Type = upstream.GetItemType()
		upstreamResponse.AddedResources = upstream
		upstreamResponse.RemovedResourceNames = removedResources

		err := syncer.Sync(upstreamResponse)
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
		upstreamResponse := zone_client.UpstreamResponse{}
		upstreamResponse.Type = upstream.GetItemType()
		upstreamResponse.AddedResources = upstream
		upstreamResponse.RemovedResourceNames = []string{"mesh-0.", "mesh-3.", "mesh-4.", "mesh-5.", "mesh-6.", "mesh-8.", "mesh-9.", "mesh-10."}

		err := syncer.Sync(upstreamResponse)
		Expect(err).ToNot(HaveOccurred())

		actual := &mesh.MeshResourceList{}
		err = resourceStore.List(context.Background(), actual)
		Expect(err).ToNot(HaveOccurred())
		Expect(len(actual.Items)).To(Equal(len(upstream.Items)))
		for i, item := range actual.Items {
			Expect(item.Spec).To(MatchProto(upstream.Items[i].Spec))
		}
	})

	It("should delete resources which are not represented in upstream and create new when is an initial request", func() {
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
		upstreamResponse := zone_client.UpstreamResponse{}
		upstreamResponse.Type = upstream.GetItemType()
		upstreamResponse.AddedResources = upstream
		upstreamResponse.IsInitialRequest = true

		err := syncer.Sync(upstreamResponse)
		Expect(err).ToNot(HaveOccurred())

		actual := &mesh.MeshResourceList{}
		err = resourceStore.List(context.Background(), actual)
		Expect(err).ToNot(HaveOccurred())
		Expect(len(actual.Items)).To(Equal(len(upstream.Items)))
		for i, item := range actual.Items {
			Expect(item.Spec).To(MatchProto(upstream.Items[i].Spec))
		}
	})

	It("should ignore resources from upstream that it does not support", func() {
		// given
		upstream := &mesh.MeshResourceList{}
		Expect(upstream.AddItem(meshBuilder(1))).To(Succeed())
		upstreamResponse := zone_client.UpstreamResponse{}
		upstreamResponse.Type = upstream.GetItemType()
		upstreamResponse.AddedResources = upstream

		// when
		err := syncer.Sync(upstreamResponse, sync_store.PrefilterBy(func(r model.Resource) bool {
			return r.GetMeta().GetName() != "mesh-1"
		}))

		// then
		Expect(err).ToNot(HaveOccurred())
		actual := &mesh.MeshResourceList{}
		Expect(resourceStore.List(context.Background(), actual)).To(Succeed())
		Expect(actual.GetItems()).To(BeEmpty())
	})

	It("should not update resource with the equal spec", func() {
		// given resource in the store
		res := meshBuilder(1)
		key := model.MetaToResourceKey(res.GetMeta())
		Expect(resourceStore.Create(context.Background(), res, store.CreateBy(key))).To(Succeed())
		existing := mesh.NewMeshResource()
		Expect(resourceStore.Get(context.Background(), existing, store.GetBy(key))).To(Succeed())

		// when sync the resource with equal 'spec'
		upstream := &mesh.MeshResourceList{}
		Expect(upstream.AddItem(meshBuilder(1))).To(Succeed())

		upstreamResponse := zone_client.UpstreamResponse{}
		upstreamResponse.Type = upstream.GetItemType()
		upstreamResponse.AddedResources = upstream

		Expect(syncer.Sync(upstreamResponse)).To(Succeed())

		// then resource's version is the same
		actual := mesh.NewMeshResource()
		Expect(resourceStore.Get(context.Background(), actual, store.GetBy(key))).To(Succeed())
		Expect(actual.GetMeta().GetVersion()).To(Equal(existing.GetMeta().GetVersion()))
	})
})
