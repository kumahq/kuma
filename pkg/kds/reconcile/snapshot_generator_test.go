package reconcile_test

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/types/known/structpb"

	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	core_manager "github.com/kumahq/kuma/v3/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/v3/pkg/core/resources/store"
	"github.com/kumahq/kuma/v3/pkg/kds"
	kds_cache "github.com/kumahq/kuma/v3/pkg/kds/cache"
	"github.com/kumahq/kuma/v3/pkg/kds/reconcile"
	"github.com/kumahq/kuma/v3/pkg/multitenant"
	"github.com/kumahq/kuma/v3/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/v3/pkg/test/resources/samples"
)

func nodeWithFeatures(id string, features ...string) *envoy_core.Node {
	values := make([]*structpb.Value, 0, len(features))
	for _, f := range features {
		values = append(values, structpb.NewStringValue(f))
	}
	return &envoy_core.Node{
		Id: id,
		Metadata: &structpb.Struct{
			Fields: map[string]*structpb.Value{
				kds.MetadataFeatures: structpb.NewListValue(&structpb.ListValue{Values: values}),
			},
		},
	}
}

var _ = Describe("SnapshotGenerator mapped resource sharing", func() {
	var store core_store.ResourceStore
	var resManager core_manager.ResourceManager
	var mapperCalls *atomic.Int64
	var generator reconcile.SnapshotGenerator
	var types map[core_model.ResourceType]struct{}

	var countingMapper reconcile.ResourceMapper = func(_ kds.Features, r core_model.Resource) (core_model.Resource, error) {
		mapperCalls.Add(1)
		return r, nil
	}

	generateIn := func(ctx context.Context, node *envoy_core.Node) {
		GinkgoHelper()
		builder := kds_cache.NewSnapshotBuilder([]core_model.ResourceType{core_mesh.MeshType})
		_, err := generator.GenerateSnapshot(ctx, node, builder, types)
		Expect(err).ToNot(HaveOccurred())
	}

	generate := func(node *envoy_core.Node) {
		GinkgoHelper()
		generateIn(context.Background(), node)
	}

	BeforeEach(func() {
		store = memory.NewStore()
		resManager = core_manager.NewResourceManager(store)
		mapperCalls = &atomic.Int64{}
		types = map[core_model.ResourceType]struct{}{core_mesh.MeshType: {}}
		for i := range 3 {
			Expect(resManager.Create(context.Background(), samples.MeshDefault(),
				core_store.CreateByKey(fmt.Sprintf("mesh-%d", i), core_model.NoMesh))).To(Succeed())
		}
		generator = reconcile.NewSnapshotGenerator(resManager, reconcile.Any, countingMapper)
	})

	It("maps each resource once and shares it across zones", func() {
		generate(nodeWithFeatures("zone-1"))
		Expect(mapperCalls.Load()).To(Equal(int64(3)))

		generate(nodeWithFeatures("zone-2"))
		generate(nodeWithFeatures("zone-3"))
		Expect(mapperCalls.Load()).To(Equal(int64(3)), "other zones should reuse the mapped resources")
	})

	It("does not share between zones that negotiated different features", func() {
		generate(nodeWithFeatures("zone-1", kds.FeatureHashSuffix))
		Expect(mapperCalls.Load()).To(Equal(int64(3)))

		generate(nodeWithFeatures("zone-2"))
		Expect(mapperCalls.Load()).To(Equal(int64(6)), "a different feature set must be mapped separately")

		generate(nodeWithFeatures("zone-3", kds.FeatureHashSuffix))
		Expect(mapperCalls.Load()).To(Equal(int64(6)), "the first feature set is still shared")
	})

	It("remaps after a resource changes", func() {
		generate(nodeWithFeatures("zone-1"))
		Expect(mapperCalls.Load()).To(Equal(int64(3)))

		mesh := core_mesh.NewMeshResource()
		Expect(resManager.Get(context.Background(), mesh, core_store.GetByKey("mesh-1", core_model.NoMesh))).To(Succeed())
		Expect(resManager.Update(context.Background(), mesh, core_store.UpdateWithLabels(map[string]string{"changed": "yes"}))).To(Succeed())

		generate(nodeWithFeatures("zone-1"))
		Expect(mapperCalls.Load()).To(Equal(int64(6)), "a changed resource must invalidate the shared entry")
	})

	It("does not share between tenants", func() {
		tenantA := multitenant.WithTenant(context.Background(), "tenant-a")
		tenantB := multitenant.WithTenant(context.Background(), "tenant-b")

		generateIn(tenantA, nodeWithFeatures("zone-1"))
		Expect(mapperCalls.Load()).To(Equal(int64(3)))

		generateIn(tenantB, nodeWithFeatures("zone-1"))
		Expect(mapperCalls.Load()).To(Equal(int64(6)), "another tenant must not reuse the first tenant's mapped resources")

		generateIn(tenantA, nodeWithFeatures("zone-2"))
		Expect(mapperCalls.Load()).To(Equal(int64(6)), "the first tenant's entry is still shared")
	})

	It("remaps after a resource is deleted and recreated", func() {
		generate(nodeWithFeatures("zone-1"))
		Expect(mapperCalls.Load()).To(Equal(int64(3)))

		mesh := core_mesh.NewMeshResource()
		Expect(resManager.Get(context.Background(), mesh, core_store.GetByKey("mesh-1", core_model.NoMesh))).To(Succeed())
		Expect(resManager.Delete(context.Background(), mesh, core_store.DeleteByKey("mesh-1", core_model.NoMesh))).To(Succeed())
		Expect(resManager.Create(context.Background(), samples.MeshDefault(),
			core_store.CreateByKey("mesh-1", core_model.NoMesh))).To(Succeed())

		generate(nodeWithFeatures("zone-1"))
		Expect(mapperCalls.Load()).To(Equal(int64(6)),
			"a recreated resource resets its version, so identity must not rest on version alone")
	})

	It("maps a resource once even when zones generate concurrently", func() {
		var wg sync.WaitGroup
		for i := range 20 {
			wg.Add(1)
			go func() {
				defer GinkgoRecover()
				defer wg.Done()
				generate(nodeWithFeatures(fmt.Sprintf("zone-%d", i)))
			}()
		}
		wg.Wait()

		Expect(mapperCalls.Load()).To(Equal(int64(3)),
			"concurrent zones must not each repeat the work the cache exists to share")
	})

	It("never maps a resource that every zone filters out", func() {
		rejectAll := func(context.Context, string, kds.Features, core_model.Resource) bool { return false }
		generator = reconcile.NewSnapshotGenerator(resManager, rejectAll, countingMapper)

		generate(nodeWithFeatures("zone-1"))
		Expect(mapperCalls.Load()).To(BeZero())
	})
})
