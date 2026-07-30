package store_test

import (
	"context"
	"errors"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core"
	core_meta "github.com/kumahq/kuma/v3/pkg/core/metadata"
	"github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	meshservice_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshservice/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/core/resources/registry"
	"github.com/kumahq/kuma/v3/pkg/core/resources/store"
	"github.com/kumahq/kuma/v3/pkg/kds/util"
	client_v2 "github.com/kumahq/kuma/v3/pkg/kds/v2/client"
	sync_store "github.com/kumahq/kuma/v3/pkg/kds/v2/store"
	core_metrics "github.com/kumahq/kuma/v3/pkg/metrics"
	"github.com/kumahq/kuma/v3/pkg/plugins/resources/memory"
	. "github.com/kumahq/kuma/v3/pkg/test/matchers"
	"github.com/kumahq/kuma/v3/pkg/test/resources/builders"
	model2 "github.com/kumahq/kuma/v3/pkg/test/resources/model"
	test_store "github.com/kumahq/kuma/v3/pkg/test/store"
)

var meshBuilder = func(idx int) *mesh.MeshResource {
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

var _ = Describe("SyncResourceStoreDelta", func() {
	var syncer sync_store.ResourceSyncer
	var resourceStore store.ResourceStore

	BeforeEach(func() {
		resourceStore = memory.NewStore()
		metrics, err := core_metrics.NewMetrics("")
		Expect(err).ToNot(HaveOccurred())
		syncer, err = sync_store.NewResourceSyncer(core.Log, resourceStore, store.NoTransactions{}, metrics, context.Background())
		Expect(err).ToNot(HaveOccurred())
	})

	It("should create new resources in empty store", func() {
		upstreamResponse := client_v2.UpstreamResponse{}
		upstream := &mesh.MeshResourceList{}
		idxs := []int{1, 2, 3, 4}
		for _, i := range idxs {
			m := meshBuilder(i)
			err := upstream.AddItem(m)
			Expect(err).ToNot(HaveOccurred())
		}
		upstreamResponse.Type = upstream.GetItemType()
		upstreamResponse.AddedResources = upstream

		err, nackError := syncer.Sync(context.Background(), upstreamResponse)
		Expect(err).ToNot(HaveOccurred())
		Expect(nackError).ToNot(HaveOccurred())

		actual := &mesh.MeshResourceList{}
		err = resourceStore.List(context.Background(), actual)
		Expect(err).ToNot(HaveOccurred())
		Expect(actual.Items).To(Equal(upstream.Items))
	})

	It("should delete all resources", func() {
		upstreamResponse := client_v2.UpstreamResponse{}
		removedResources := []model.ResourceKey{}
		for i := range 10 {
			m := meshBuilder(i)
			removedResources = append(removedResources, model.WithoutMesh(fmt.Sprintf("mesh-%d", i)))
			err := resourceStore.Create(context.Background(), m, store.CreateBy(model.MetaToResourceKey(m.GetMeta())))
			Expect(err).ToNot(HaveOccurred())
		}
		upstream := &mesh.MeshResourceList{}
		upstreamResponse.Type = upstream.GetItemType()
		upstreamResponse.AddedResources = upstream
		upstreamResponse.RemovedResourcesKey = removedResources

		err, nackError := syncer.Sync(context.Background(), upstreamResponse)
		Expect(err).ToNot(HaveOccurred())
		Expect(nackError).ToNot(HaveOccurred())

		actual := &mesh.MeshResourceList{}
		err = resourceStore.List(context.Background(), actual)
		Expect(err).ToNot(HaveOccurred())
		Expect(actual.Items).To(BeEmpty())
	})

	It("should delete resources which are not represented in upstream and create new", func() {
		for i := range 10 {
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
		upstreamResponse := client_v2.UpstreamResponse{}
		upstreamResponse.Type = upstream.GetItemType()
		upstreamResponse.AddedResources = upstream
		upstreamResponse.RemovedResourcesKey = []model.ResourceKey{
			model.WithoutMesh("mesh-0"),
			model.WithoutMesh("mesh-3"),
			model.WithoutMesh("mesh-4"),
			model.WithoutMesh("mesh-5"),
			model.WithoutMesh("mesh-6"),
			model.WithoutMesh("mesh-8"),
			model.WithoutMesh("mesh-9"),
			model.WithoutMesh("mesh-10"),
		}

		err, nackError := syncer.Sync(context.Background(), upstreamResponse)
		Expect(err).ToNot(HaveOccurred())
		Expect(nackError).ToNot(HaveOccurred())

		actual := &mesh.MeshResourceList{}
		err = resourceStore.List(context.Background(), actual)
		Expect(err).ToNot(HaveOccurred())
		Expect(actual.Items).To(HaveLen(len(upstream.Items)))
		for i, item := range actual.Items {
			Expect(item.Spec).To(MatchProto(upstream.Items[i].Spec))
		}
	})

	It("should delete resources which are not represented in upstream and create new when is an initial request", func() {
		for i := range 10 {
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
		upstreamResponse := client_v2.UpstreamResponse{}
		upstreamResponse.Type = upstream.GetItemType()
		upstreamResponse.AddedResources = upstream
		upstreamResponse.IsInitialRequest = true

		err, nackError := syncer.Sync(context.Background(), upstreamResponse)
		Expect(err).ToNot(HaveOccurred())
		Expect(nackError).ToNot(HaveOccurred())

		actual := &mesh.MeshResourceList{}
		err = resourceStore.List(context.Background(), actual)
		Expect(err).ToNot(HaveOccurred())
		Expect(actual.Items).To(HaveLen(len(upstream.Items)))
		for i, item := range actual.Items {
			Expect(item.Spec).To(MatchProto(upstream.Items[i].Spec))
		}
	})

	It("should ignore resources from upstream that it does not support", func() {
		// given
		upstream := &mesh.MeshResourceList{}
		Expect(upstream.AddItem(meshBuilder(1))).To(Succeed())
		upstreamResponse := client_v2.UpstreamResponse{}
		upstreamResponse.Type = upstream.GetItemType()
		upstreamResponse.AddedResources = upstream

		// when
		err, nackError := syncer.Sync(context.Background(), upstreamResponse, sync_store.PrefilterBy(func(r model.Resource) bool {
			return r.GetMeta().GetName() != "mesh-1"
		}))

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(nackError).ToNot(HaveOccurred())
		actual := &mesh.MeshResourceList{}
		Expect(resourceStore.List(context.Background(), actual)).To(Succeed())
		Expect(actual.GetItems()).To(BeEmpty())
	})

	It("should add all resources and skip the conflict one", func() {
		mesh1 := meshBuilder(1)
		mesh2 := meshBuilder(2)
		mesh3 := meshBuilder(3)
		Expect(resourceStore.Create(
			context.Background(),
			mesh2,
			store.CreateBy(model.MetaToResourceKey(mesh2.GetMeta())),
			store.CreateWithLabels(map[string]string{mesh_proto.ResourceOriginLabel: "zone"}),
		)).ToNot(HaveOccurred())

		// given
		upstream := &mesh.MeshResourceList{}

		// try to add resource without the label
		mesh2 = meshBuilder(2)
		mesh2.Spec.Mtls.EnabledBackend = "modified-ca"
		Expect(upstream.AddItem(mesh1)).To(Succeed())
		Expect(upstream.AddItem(mesh2)).To(Succeed())
		Expect(upstream.AddItem(mesh3)).To(Succeed())
		upstreamResponse := client_v2.UpstreamResponse{}
		upstreamResponse.Type = upstream.GetItemType()
		upstreamResponse.AddedResources = upstream

		// when
		err, nackError := syncer.Sync(context.Background(), upstreamResponse, sync_store.PrefilterBy(func(r model.Resource) bool {
			return r.GetMeta().GetLabels()[mesh_proto.ResourceOriginLabel] != "zone"
		}), sync_store.SkipConflictResource())

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(nackError).To(HaveOccurred())

		actual := &mesh.MeshResourceList{}
		Expect(resourceStore.List(context.Background(), actual)).To(Succeed())
		Expect(actual.GetItems()).To(HaveLen(3))
		Expect(actual.GetItems()[0].GetSpec()).To(MatchProto(meshBuilder(2).GetSpec()))
		Expect(actual.GetItems()[1].GetSpec()).To(MatchProto(mesh1.GetSpec()))
		// should not update resource since mesh-2 already exists
		Expect(actual.GetItems()[2].GetSpec()).To(MatchProto(mesh3.GetSpec()))
	})

	It("should ignore invalid resource from upstream and add only valid", func() {
		// given
		upstream := &mesh.MeshResourceList{}
		mesh1 := meshBuilder(1)
		mesh2 := meshBuilder(2)
		mesh3 := meshBuilder(3)
		Expect(upstream.AddItem(mesh1)).To(Succeed())
		Expect(upstream.AddItem(mesh2)).To(Succeed())
		Expect(upstream.AddItem(mesh3)).To(Succeed())
		upstreamResponse := client_v2.UpstreamResponse{}
		upstreamResponse.Type = upstream.GetItemType()
		upstreamResponse.AddedResources = upstream
		upstreamResponse.InvalidResourcesKey = []model.ResourceKey{model.MetaToResourceKey(mesh2.GetMeta())}

		// when
		err, nackError := syncer.Sync(context.Background(), upstreamResponse)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(nackError).ToNot(HaveOccurred())
		actual := &mesh.MeshResourceList{}
		Expect(resourceStore.List(context.Background(), actual)).To(Succeed())
		Expect(actual.GetItems()).To(HaveLen(2))
		Expect(actual.GetItems()[0].GetSpec()).To(MatchProto(mesh1.GetSpec()))
		Expect(actual.GetItems()[1].GetSpec()).To(MatchProto(mesh3.GetSpec()))
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

		upstreamResponse := client_v2.UpstreamResponse{}
		upstreamResponse.Type = upstream.GetItemType()
		upstreamResponse.AddedResources = upstream

		Expect(syncer.Sync(context.Background(), upstreamResponse)).To(Succeed())

		// then resource's version is the same
		actual := mesh.NewMeshResource()
		Expect(resourceStore.Get(context.Background(), actual, store.GetBy(key))).To(Succeed())
		Expect(actual.GetMeta().GetVersion()).To(Equal(existing.GetMeta().GetVersion()))
	})
})

var _ = Describe("SyncResourceStoreDelta errors", func() {
	var syncer sync_store.ResourceSyncer
	var resourceStore store.ResourceStore

	BeforeEach(func() {
		resourceStore = &test_store.FailingStore{CreateErr: errors.Join(store.ErrorResourceAlreadyExists(system.GlobalSecretType, "zone-token-signing-public-key-1", ""))}
		metrics, err := core_metrics.NewMetrics("")
		Expect(err).ToNot(HaveOccurred())
		syncer, err = sync_store.NewResourceSyncer(core.Log, resourceStore, store.NoTransactions{}, metrics, context.Background())
		Expect(err).ToNot(HaveOccurred())
	})

	It("should correctly recognize user errors", func() {
		upstreamResponse := client_v2.UpstreamResponse{}
		upstream := &mesh.MeshResourceList{}
		m := meshBuilder(1)
		err := upstream.AddItem(m)
		Expect(err).ToNot(HaveOccurred())
		upstreamResponse.Type = upstream.GetItemType()
		upstreamResponse.AddedResources = upstream

		err, nackError := syncer.Sync(context.Background(), upstreamResponse, sync_store.SkipConflictResource())
		Expect(err).ToNot(HaveOccurred())
		Expect(util.IsUserError(nackError)).To(BeTrue())
		Expect(util.IsUserErrorMessage(nackError.Error())).To(BeTrue())
		Expect(nackError).To(MatchError(`user error
resource already exists: type="GlobalSecret" name="zone-token-signing-public-key-1" mesh=""`))
	})
})

// conflictingStore simulates another writer winning the race on the first
// 'conflicts' updates: the update is rejected and the stored record is bumped to
// a new version, so a retry that doesn't rebase on a fresh copy keeps losing.
type conflictingStore struct {
	store.ResourceStore
	conflicts int
	updates   int
	// mutate changes the record the winning writer leaves behind, on top of the
	// version bump.
	mutate func(model.Resource)
}

func (c *conflictingStore) Update(ctx context.Context, r model.Resource, fs ...store.UpdateOptionsFunc) error {
	c.updates++
	if c.updates <= c.conflicts {
		current, err := registry.Global().NewObject(r.Descriptor().Name)
		if err != nil {
			return err
		}
		key := model.MetaToResourceKey(r.GetMeta())
		if err := c.Get(ctx, current, store.GetBy(key)); err != nil {
			return err
		}
		if c.mutate != nil {
			c.mutate(current)
		}
		if err := c.ResourceStore.Update(ctx, current); err != nil {
			return err
		}
		return store.ErrorResourceConflict(r.Descriptor().Name, key.Name, key.Mesh)
	}
	return c.ResourceStore.Update(ctx, r, fs...)
}

var _ = Describe("SyncResourceStoreDelta write conflicts", func() {
	var resourceStore *conflictingStore
	var syncer sync_store.ResourceSyncer
	var key model.ResourceKey

	// meshBuilder(1) with a different spec, so syncing it against the stored copy
	// produces an update rather than a create.
	changedMesh := func() *mesh.MeshResource {
		m := meshBuilder(1)
		m.Spec.Mtls.EnabledBackend = "ca-changed"
		m.Spec.Mtls.Backends[0].Name = "ca-changed"
		return m
	}

	syncChangedMesh := func() (error, error) {
		upstream := &mesh.MeshResourceList{}
		Expect(upstream.AddItem(changedMesh())).To(Succeed())
		return syncer.Sync(context.Background(), client_v2.UpstreamResponse{
			Type:           upstream.GetItemType(),
			AddedResources: upstream,
		})
	}

	BeforeEach(func() {
		resourceStore = &conflictingStore{ResourceStore: memory.NewStore()}
		metrics, err := core_metrics.NewMetrics("")
		Expect(err).ToNot(HaveOccurred())
		syncer, err = sync_store.NewResourceSyncer(core.Log, resourceStore, store.NoTransactions{}, metrics, context.Background())
		Expect(err).ToNot(HaveOccurred())

		res := meshBuilder(1)
		key = model.MetaToResourceKey(res.GetMeta())
		Expect(resourceStore.Create(context.Background(), res, store.CreateBy(key))).To(Succeed())
	})

	It("should rebase the update on the winner's version and apply the change", func() {
		resourceStore.conflicts = 1

		err, nackError := syncChangedMesh()
		Expect(err).ToNot(HaveOccurred())
		Expect(nackError).ToNot(HaveOccurred())

		actual := mesh.NewMeshResource()
		Expect(resourceStore.Get(context.Background(), actual, store.GetBy(key))).To(Succeed())
		Expect(actual.Spec.Mtls.EnabledBackend).To(Equal("ca-changed"))
		Expect(resourceStore.updates).To(Equal(2))
		// create, then the concurrent writer, then the retried sync: without a
		// rebase on the fresh copy the retry would carry version 1 and conflict again
		Expect(actual.GetMeta().GetVersion()).To(Equal("3"))
	})

	It("should give up once the retries are exhausted and return the error", func() {
		resourceStore.conflicts = 100

		err, nackError := syncChangedMesh()
		Expect(store.IsConflict(err)).To(BeTrue())
		Expect(nackError).ToNot(HaveOccurred())
		// the update in the transaction, then 3 attempts: immediate, then 2 backed off
		Expect(resourceStore.updates).To(Equal(4))

		actual := mesh.NewMeshResource()
		Expect(resourceStore.Get(context.Background(), actual, store.GetBy(key))).To(Succeed())
		Expect(actual.Spec.Mtls.EnabledBackend).To(Equal("ca-1"))
	})

	It("should apply the rest of the batch when one resource conflicts", func() {
		for i := 2; i <= 3; i++ {
			res := meshBuilder(i)
			Expect(resourceStore.Create(context.Background(), res, store.CreateBy(model.MetaToResourceKey(res.GetMeta())))).To(Succeed())
		}
		// only the first update of the batch loses the race
		resourceStore.conflicts = 1

		upstream := &mesh.MeshResourceList{}
		for i := 1; i <= 3; i++ {
			m := meshBuilder(i)
			m.Spec.Mtls.EnabledBackend = "ca-changed"
			m.Spec.Mtls.Backends[0].Name = "ca-changed"
			Expect(upstream.AddItem(m)).To(Succeed())
		}
		err, nackError := syncer.Sync(context.Background(), client_v2.UpstreamResponse{
			Type:           upstream.GetItemType(),
			AddedResources: upstream,
		})
		Expect(err).ToNot(HaveOccurred())
		Expect(nackError).ToNot(HaveOccurred())

		actual := &mesh.MeshResourceList{}
		Expect(resourceStore.List(context.Background(), actual)).To(Succeed())
		Expect(actual.Items).To(HaveLen(3))
		for _, item := range actual.Items {
			Expect(item.Spec.Mtls.EnabledBackend).To(Equal("ca-changed"))
		}
		// 3 updates in the transaction, 1 retried after it
		Expect(resourceStore.updates).To(Equal(4))
	})
})

var _ = Describe("SyncResourceStoreDelta write conflicts on zone-owned status", func() {
	var resourceStore *conflictingStore
	var syncer sync_store.ResourceSyncer

	BeforeEach(func() {
		resourceStore = &conflictingStore{ResourceStore: memory.NewStore()}
		metrics, err := core_metrics.NewMetrics("")
		Expect(err).ToNot(HaveOccurred())
		syncer, err = sync_store.NewResourceSyncer(core.Log, resourceStore, store.NoTransactions{}, metrics, context.Background())
		Expect(err).ToNot(HaveOccurred())

		Expect(builders.MeshService().
			AddIntPort(80, 8080, core_meta.ProtocolHTTP).
			WithKumaVIP("10.0.0.1").
			Create(resourceStore)).To(Succeed())
	})

	// The rebase is only safe because Global owns the spec and the Zone owns the
	// status. The VIP allocator writes a status while Sync is applying a spec, and
	// the retry has to carry the allocator's status, not the one read before it ran.
	It("should keep the status written by the writer that won the race", func() {
		resourceStore.conflicts = 1
		resourceStore.mutate = func(r model.Resource) {
			Expect(r.SetStatus(&meshservice_api.MeshServiceStatus{
				VIPs: []meshservice_api.VIP{{IP: "10.0.0.2"}},
			})).To(Succeed())
		}

		// Global syncs the spec down with the status stripped
		upstream := &meshservice_api.MeshServiceResourceList{}
		Expect(upstream.AddItem(builders.MeshService().
			AddIntPort(90, 9090, core_meta.ProtocolHTTP).
			WithoutVIP().
			Build())).To(Succeed())

		err, nackError := syncer.Sync(context.Background(), client_v2.UpstreamResponse{
			Type:           upstream.GetItemType(),
			AddedResources: upstream,
		}, sync_store.IgnoreStatusChange())
		Expect(err).ToNot(HaveOccurred())
		Expect(nackError).ToNot(HaveOccurred())

		actual := meshservice_api.NewMeshServiceResource()
		Expect(resourceStore.Get(context.Background(), actual, store.GetBy(builders.MeshService().Key()))).To(Succeed())
		// upstream owns the spec
		Expect(actual.Spec.Ports).To(HaveLen(1))
		Expect(actual.Spec.Ports[0].Port).To(Equal(int32(90)))
		// the zone owns the status, and the retry must not roll it back to the copy
		// read before the allocator wrote
		Expect(actual.Status.VIPs).To(Equal([]meshservice_api.VIP{{IP: "10.0.0.2"}}))
	})
})
