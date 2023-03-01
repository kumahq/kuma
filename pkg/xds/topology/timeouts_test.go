package topology_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/pkg/test/resources/builders"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	"github.com/kumahq/kuma/pkg/xds/topology"
)

var _ = Describe("Timeout", func() {
	var ctx context.Context
	var rm core_manager.ResourceManager
	var dataplane *core_mesh.DataplaneResource

	BeforeEach(func() {
		rm = core_manager.NewResourceManager(memory.NewStore())

		err := rm.Create(ctx, core_mesh.NewMeshResource(), store.CreateByKey("mesh-1", model.NoMesh))
		Expect(err).ToNot(HaveOccurred())

		dataplane = builders.Dataplane().
			WithAddress("192.168.0.1").
			WithMesh("mesh-1").
			WithInboundOfTags(mesh_proto.ServiceTag, "frontend", "version", "v1").
			AddOutboundsToServices("backend", "web", "redis", "db").
			Build()

		err = rm.Create(ctx, dataplane, store.CreateBy(model.MetaToResourceKey(dataplane.GetMeta())))
		Expect(err).ToNot(HaveOccurred())
	})

	Context("GetTimeouts()", func() {
		It("should pick Timeout which matches sources and apply to right outbound", func() {
			timeoutsFrontendV1ToRedis := &core_mesh.TimeoutResource{
				Meta: &test_model.ResourceMeta{Mesh: "mesh-1", Name: "timeouts-redis"},
				Spec: &mesh_proto.Timeout{
					Sources:      []*mesh_proto.Selector{{Match: map[string]string{mesh_proto.ServiceTag: "frontend", "version": "v1"}}},
					Destinations: []*mesh_proto.Selector{{Match: map[string]string{mesh_proto.ServiceTag: "redis"}}},
					Conf:         &mesh_proto.Timeout_Conf{ConnectTimeout: util_proto.Duration(10 * time.Second)},
				},
			}

			err := rm.Create(ctx, timeoutsFrontendV1ToRedis, store.CreateBy(model.MetaToResourceKey(timeoutsFrontendV1ToRedis.GetMeta())))
			Expect(err).ToNot(HaveOccurred())

			timeoutMap, err := topology.GetTimeouts(ctx, dataplane, rm)
			Expect(err).ToNot(HaveOccurred())
			Expect(timeoutMap).To(HaveLen(1))
			Expect(timeoutMap).To(HaveKeyWithValue(
				mesh_proto.OutboundInterface{DataplaneIP: "127.0.0.1", DataplanePort: builders.FirstOutboundPort + 2},
				timeoutsFrontendV1ToRedis,
			))
		})

		It("should pick Timeout which matches for all sources and all destinations", func() {
			timeoutsAllToAll := &core_mesh.TimeoutResource{
				Meta: &test_model.ResourceMeta{Mesh: "mesh-1", Name: "timeouts-redis"},
				Spec: &mesh_proto.Timeout{
					Sources:      []*mesh_proto.Selector{{Match: map[string]string{mesh_proto.ServiceTag: "*"}}},
					Destinations: []*mesh_proto.Selector{{Match: map[string]string{mesh_proto.ServiceTag: "*"}}},
					Conf:         &mesh_proto.Timeout_Conf{ConnectTimeout: util_proto.Duration(10 * time.Second)},
				},
			}

			err := rm.Create(ctx, timeoutsAllToAll, store.CreateBy(model.MetaToResourceKey(timeoutsAllToAll.GetMeta())))
			Expect(err).ToNot(HaveOccurred())

			timeoutMap, err := topology.GetTimeouts(ctx, dataplane, rm)
			Expect(err).ToNot(HaveOccurred())
			Expect(timeoutMap).To(HaveLen(4))

			Expect(timeoutMap).To(HaveKeyWithValue(
				mesh_proto.OutboundInterface{DataplaneIP: "127.0.0.1", DataplanePort: builders.FirstOutboundPort},
				timeoutsAllToAll,
			))
			Expect(timeoutMap).To(HaveKeyWithValue(
				mesh_proto.OutboundInterface{DataplaneIP: "127.0.0.1", DataplanePort: builders.FirstOutboundPort + 1},
				timeoutsAllToAll,
			))
			Expect(timeoutMap).To(HaveKeyWithValue(
				mesh_proto.OutboundInterface{DataplaneIP: "127.0.0.1", DataplanePort: builders.FirstOutboundPort + 2},
				timeoutsAllToAll,
			))
			Expect(timeoutMap).To(HaveKeyWithValue(
				mesh_proto.OutboundInterface{DataplaneIP: "127.0.0.1", DataplanePort: builders.FirstOutboundPort + 3},
				timeoutsAllToAll,
			))
		})
	})
})
