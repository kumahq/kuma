package reconcile_test

import (
	"context"
	"fmt"
	"sync/atomic"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	"google.golang.org/protobuf/types/known/structpb"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	core_manager "github.com/kumahq/kuma/v3/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/v3/pkg/core/resources/store"
	"github.com/kumahq/kuma/v3/pkg/kds"
	kds_cache "github.com/kumahq/kuma/v3/pkg/kds/cache"
	"github.com/kumahq/kuma/v3/pkg/kds/reconcile"
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
	var mapperCalls *atomic.Int64
	var generator reconcile.SnapshotGenerator
	var types map[core_model.ResourceType]struct{}

	countingMapper := func(_ kds.Features, r core_model.Resource) (core_model.Resource, error) {
		mapperCalls.Add(1)
		return r, nil
	}

	generate := func(node *envoy_core.Node) {
		GinkgoHelper()
		builder := kds_cache.NewSnapshotBuilder([]core_model.ResourceType{core_mesh.MeshType})
		_, err := generator.GenerateSnapshot(context.Background(), node, builder, types)
		Expect(err).ToNot(HaveOccurred())
	}

	BeforeEach(func() {
		store = memory.NewStore()
		mapperCalls = &atomic.Int64{}
		types = map[core_model.ResourceType]struct{}{core_mesh.MeshType: {}}
		for i := 0; i < 3; i++ {
			Expect(store.Create(context.Background(), samples.MeshDefault(),
				core_store.CreateByKey(fmt.Sprintf("mesh-%d", i), core_model.NoMesh))).To(Succeed())
		}
		generator = reconcile.NewSnapshotGenerator(core_manager.NewResourceManager(store), reconcile.Any, countingMapper)
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
		Expect(store.Get(context.Background(), mesh, core_store.GetByKey("mesh-1", core_model.NoMesh))).To(Succeed())
		Expect(store.Update(context.Background(), mesh, core_store.UpdateWithLabels(map[string]string{"changed": "yes"}))).To(Succeed())

		generate(nodeWithFeatures("zone-1"))
		Expect(mapperCalls.Load()).To(Equal(int64(6)), "a changed resource must invalidate the shared entry")
	})

	It("never maps a resource that every zone filters out", func() {
		rejectAll := func(context.Context, string, kds.Features, core_model.Resource) bool { return false }
		generator = reconcile.NewSnapshotGenerator(core_manager.NewResourceManager(store), rejectAll, countingMapper)

		generate(nodeWithFeatures("zone-1"))
		Expect(mapperCalls.Load()).To(BeZero())
	})
})
