package topology_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	memory_resources "github.com/kumahq/kuma/pkg/plugins/resources/memory"
	. "github.com/kumahq/kuma/pkg/test/matchers"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	. "github.com/kumahq/kuma/pkg/xds/topology"
)

var _ = Describe("Retry", func() {
	var ctx context.Context
	var rm core_manager.ResourceManager

	BeforeEach(func() {
		ctx = context.Background()
		rm = core_manager.NewResourceManager(memory_resources.NewStore())
	})

	Describe("GetRetries()", func() {
		It("should pick the best matching Retry for each destination service", func() {
			// given
			mesh := &core_mesh.MeshResource{ // mesh that is relevant to this test case
				Meta: &test_model.ResourceMeta{
					Name: "demo",
				},
				Spec: &mesh_proto.Mesh{},
			}
			otherMesh := &core_mesh.MeshResource{ // mesh that is irrelevant to this test case
				Meta: &test_model.ResourceMeta{
					Name: "default",
				},
				Spec: &mesh_proto.Mesh{},
			}
			backend := &core_mesh.DataplaneResource{ // dataplane that is a source of traffic
				Meta: &test_model.ResourceMeta{
					Mesh: "demo",
					Name: "backend",
				},
				Spec: &mesh_proto.Dataplane{
					Networking: &mesh_proto.Dataplane_Networking{
						Address: "192.168.0.1",
						Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
							{
								Tags: map[string]string{
									"kuma.io/service": "backend",
									"region":          "eu",
								},
								Port:        8080,
								ServicePort: 18080,
							},
							{
								Tags: map[string]string{
									"kuma.io/service": "frontend",
									"region":          "eu",
								},
								Port:        7070,
								ServicePort: 17070,
							},
						},
						Outbound: []*mesh_proto.Dataplane_Networking_Outbound{
							{
								Port: 10001,
								Tags: map[string]string{
									"kuma.io/service": "redis",
								},
							},
							{
								Port: 10002,
								Tags: map[string]string{
									"kuma.io/service": "elastic",
								},
							},
						},
					},
				},
			}
			destinations := core_xds.DestinationMap{
				"redis": core_xds.TagSelectorSet{
					mesh_proto.MatchService("redis"),
				},
				"elastic": core_xds.TagSelectorSet{
					mesh_proto.MatchService("elastic"),
				},
			}
			retryRedis := &core_mesh.RetryResource{ // retries for `redis` service
				Meta: &test_model.ResourceMeta{
					Mesh: "demo",
					Name: "retry-redis",
				},
				Spec: &mesh_proto.Retry{
					Sources: []*mesh_proto.Selector{
						{
							Match: mesh_proto.TagSelector{
								"kuma.io/service": "frontend",
							},
						},
						{
							Match: mesh_proto.TagSelector{
								"kuma.io/service": "backend",
							},
						},
					},
					Destinations: []*mesh_proto.Selector{
						{
							Match: mesh_proto.TagSelector{
								"kuma.io/service": "redis",
							},
						},
					},
					Conf: &mesh_proto.Retry_Conf{
						Http: &mesh_proto.Retry_Conf_Http{
							NumRetries:    util_proto.UInt32(3),
							PerTryTimeout: util_proto.Duration(time.Nanosecond * 200000000),
							BackOff: &mesh_proto.Retry_Conf_BackOff{
								BaseInterval: util_proto.Duration(time.Nanosecond * 10000000),
								MaxInterval:  util_proto.Duration(time.Second * 1),
							},
							RetriableStatusCodes: []uint32{500, 504},
						},
					},
				},
			}
			retryElastic := &core_mesh.RetryResource{ // retries for `elastic` service
				Meta: &test_model.ResourceMeta{
					Mesh: "demo",
					Name: "retry-elastic",
				},
				Spec: &mesh_proto.Retry{
					Sources: []*mesh_proto.Selector{
						{
							Match: mesh_proto.TagSelector{
								"kuma.io/service": "*",
							},
						},
					},
					Destinations: []*mesh_proto.Selector{
						{
							Match: mesh_proto.TagSelector{
								"kuma.io/service": "elastic",
							},
						},
					},
					Conf: &mesh_proto.Retry_Conf{
						Http: &mesh_proto.Retry_Conf_Http{
							NumRetries:    util_proto.UInt32(7),
							PerTryTimeout: util_proto.Duration(time.Nanosecond * 200000000),
							BackOff: &mesh_proto.Retry_Conf_BackOff{
								BaseInterval: util_proto.Duration(time.Nanosecond * 10000000),
								MaxInterval:  util_proto.Duration(time.Second * 3),
							},
						},
					},
				},
			}
			retryEverything := &core_mesh.RetryResource{ // retries for any service
				Meta: &test_model.ResourceMeta{
					Mesh: "default", // other mesh
					Name: "retry-everything",
				},
				Spec: &mesh_proto.Retry{
					Sources: []*mesh_proto.Selector{
						{
							Match: mesh_proto.TagSelector{
								"kuma.io/service": "*",
							},
						},
					},
					Destinations: []*mesh_proto.Selector{
						{
							Match: mesh_proto.TagSelector{
								"kuma.io/service": "*",
							},
						},
					},
					Conf: &mesh_proto.Retry_Conf{
						Http: &mesh_proto.Retry_Conf_Http{
							NumRetries: util_proto.UInt32(5),
						},
					},
				},
			}
			for _, resource := range []core_model.Resource{
				mesh,
				backend,
				retryRedis,
				retryElastic,
				otherMesh,
				retryEverything,
			} {
				// when
				err := rm.Create(
					ctx,
					resource,
					core_store.CreateBy(
						core_model.MetaToResourceKey(resource.GetMeta()),
					),
				)
				// then
				Expect(err).ToNot(HaveOccurred())
			}

			// when
			retries, err := GetRetries(ctx, backend, destinations, rm)

			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(retries).To(HaveLen(2))
			// and
			Expect(retries).To(HaveKey("redis"))
			Expect(retries["redis"].Meta.GetName()).
				To(Equal(retryRedis.Meta.GetName()))
			Expect(retries["redis"].Spec).To(MatchProto(retryRedis.Spec))
			// and
			Expect(retries).To(HaveKey("elastic"))
			Expect(retries["elastic"].Meta.GetName()).
				To(Equal(retryElastic.Meta.GetName()))
			Expect(retries["elastic"].Spec).To(MatchProto(retryElastic.Spec))
		})
	})

	Describe("BuildRetryMap()", func() {
		sameMeta := func(meta1, meta2 core_model.ResourceMeta) bool {
			return meta1.GetMesh() == meta2.GetMesh() &&
				meta1.GetName() == meta2.GetName() &&
				meta1.GetVersion() == meta2.GetVersion()
		}
		type testCase struct {
			dataplane    *core_mesh.DataplaneResource
			destinations core_xds.DestinationMap
			retries      []*core_mesh.RetryResource
			expected     core_xds.RetryMap
		}
		DescribeTable("should correctly pick a single the most specific Retry for each outbound"+
			" interface",
			func(given testCase) {
				// setup
				expectedRetries := core_xds.RetryMap{}
				for service, expectedRetry := range given.expected {
					for _, retry := range given.retries {
						if sameMeta(expectedRetry.GetMeta(), retry.GetMeta()) {
							expectedRetries[service] = retry
							break
						}
					}
					if _, ok := expectedRetries[service]; !ok {
						expectedRetries[service] = expectedRetry
					}
				}
				if len(expectedRetries) == 0 {
					expectedRetries = nil
				}
				// when
				retries := BuildRetryMap(given.dataplane, given.retries, given.destinations)
				// expect
				Expect(retries).Should(Equal(expectedRetries))
			},
			Entry("Dataplane without outbound interfaces (and therefore no destinations)", testCase{
				dataplane:    core_mesh.NewDataplaneResource(),
				destinations: nil,
				retries:      nil,
				expected:     nil,
			}),
			Entry("if a destination service has no matching Retries, none should be used", testCase{
				dataplane: &core_mesh.DataplaneResource{
					Spec: &mesh_proto.Dataplane{
						Networking: &mesh_proto.Dataplane_Networking{
							Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
								{
									Tags: map[string]string{
										"kuma.io/service": "backend",
									},
								},
							},
							Outbound: []*mesh_proto.Dataplane_Networking_Outbound{
								{
									Tags: map[string]string{
										"kuma.io/service": "redis",
									},
								},
								{
									Tags: map[string]string{
										"kuma.io/service": "elastic",
									},
								},
							},
						},
					},
				},
				destinations: core_xds.DestinationMap{
					"redis": core_xds.TagSelectorSet{
						mesh_proto.MatchService("redis"),
					},
					"elastic": core_xds.TagSelectorSet{
						mesh_proto.MatchService("elastic"),
					},
				},
				retries:  nil,
				expected: nil,
			}),
			Entry("due to TrafficRoutes, a Dataplane might have more destinations than outbound interfaces", testCase{
				dataplane: &core_mesh.DataplaneResource{
					Spec: &mesh_proto.Dataplane{
						Networking: &mesh_proto.Dataplane_Networking{
							Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
								{
									Tags: map[string]string{
										"kuma.io/service": "backend",
									},
								},
							},
							Outbound: []*mesh_proto.Dataplane_Networking_Outbound{
								{
									Tags: map[string]string{
										"kuma.io/service": "redis",
									},
								},
							},
						},
					},
				},
				destinations: core_xds.DestinationMap{
					"redis": core_xds.TagSelectorSet{
						mesh_proto.MatchService("redis"),
					},
					"elastic": core_xds.TagSelectorSet{
						mesh_proto.MatchService("elastic"),
					},
				},
				retries: []*core_mesh.RetryResource{
					{
						Meta: &test_model.ResourceMeta{
							Name: "retry-elastic",
						},
						Spec: &mesh_proto.Retry{
							Sources: []*mesh_proto.Selector{
								{
									Match: mesh_proto.TagSelector{
										"kuma.io/service": "*",
									},
								},
							},
							Destinations: []*mesh_proto.Selector{
								{
									Match: mesh_proto.TagSelector{
										"kuma.io/service": "elastic",
									},
								},
							},
							Conf: &mesh_proto.Retry_Conf{
								Http: &mesh_proto.Retry_Conf_Http{
									NumRetries:    util_proto.UInt32(7),
									PerTryTimeout: util_proto.Duration(time.Nanosecond * 200000000),
									BackOff: &mesh_proto.Retry_Conf_BackOff{
										BaseInterval: util_proto.Duration(time.Nanosecond * 10000000),
										MaxInterval:  util_proto.Duration(time.Second * 3),
									},
								},
							},
						},
					},
					{
						Meta: &test_model.ResourceMeta{
							Name: "retry-redis",
						},
						Spec: &mesh_proto.Retry{
							Sources: []*mesh_proto.Selector{
								{
									Match: mesh_proto.TagSelector{
										"kuma.io/service": "*",
									},
								},
							},
							Destinations: []*mesh_proto.Selector{
								{
									Match: mesh_proto.TagSelector{
										"kuma.io/service": "redis",
									},
								},
							},
							Conf: &mesh_proto.Retry_Conf{
								Http: &mesh_proto.Retry_Conf_Http{
									NumRetries:    util_proto.UInt32(3),
									PerTryTimeout: util_proto.Duration(time.Nanosecond * 200000000),
									BackOff: &mesh_proto.Retry_Conf_BackOff{
										BaseInterval: util_proto.Duration(time.Nanosecond * 10000000),
										MaxInterval:  util_proto.Duration(time.Second * 1),
									},
									RetriableStatusCodes: []uint32{500, 504},
								},
							},
						},
					},
				},
				expected: core_xds.RetryMap{
					"redis": &core_mesh.RetryResource{
						Meta: &test_model.ResourceMeta{
							Name: "retry-redis",
						},
					},
					"elastic": &core_mesh.RetryResource{
						Meta: &test_model.ResourceMeta{
							Name: "retry-elastic",
						},
					},
				},
			}),
			Entry("Retries should be picked by latest creation time given two equally specific"+
				" Retries", testCase{
				dataplane: &core_mesh.DataplaneResource{
					Spec: &mesh_proto.Dataplane{
						Networking: &mesh_proto.Dataplane_Networking{
							Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
								{
									Tags: map[string]string{
										"kuma.io/service": "backend",
									},
								},
							},
							Outbound: []*mesh_proto.Dataplane_Networking_Outbound{
								{
									Tags: map[string]string{
										"kuma.io/service": "redis",
									},
								},
							},
						},
					},
				},
				destinations: core_xds.DestinationMap{
					"redis": core_xds.TagSelectorSet{
						mesh_proto.MatchService("redis"),
					},
				},
				retries: []*core_mesh.RetryResource{
					{
						Meta: &test_model.ResourceMeta{
							Name:         "retry-everything-passive",
							CreationTime: time.Unix(1, 1),
						},
						Spec: &mesh_proto.Retry{
							Sources: []*mesh_proto.Selector{
								{
									Match: mesh_proto.TagSelector{
										"kuma.io/service": "*",
									},
								},
							},
							Destinations: []*mesh_proto.Selector{
								{
									Match: mesh_proto.TagSelector{
										"kuma.io/service": "*",
									},
								},
							},
							Conf: &mesh_proto.Retry_Conf{
								Http: &mesh_proto.Retry_Conf_Http{
									NumRetries:    util_proto.UInt32(3),
									PerTryTimeout: util_proto.Duration(time.Nanosecond * 200000000),
									BackOff: &mesh_proto.Retry_Conf_BackOff{
										BaseInterval: util_proto.Duration(time.Nanosecond * 10000000),
										MaxInterval:  util_proto.Duration(time.Second * 1),
									},
									RetriableStatusCodes: []uint32{500, 504},
								},
							},
						},
					},
					{
						Meta: &test_model.ResourceMeta{
							Name:         "retry-everything-active",
							CreationTime: time.Unix(0, 0),
						},
						Spec: &mesh_proto.Retry{
							Sources: []*mesh_proto.Selector{
								{
									Match: mesh_proto.TagSelector{
										"kuma.io/service": "*",
									},
								},
							},
							Destinations: []*mesh_proto.Selector{
								{
									Match: mesh_proto.TagSelector{
										"kuma.io/service": "*",
									},
								},
							},
							Conf: &mesh_proto.Retry_Conf{
								Http: &mesh_proto.Retry_Conf_Http{
									NumRetries:    util_proto.UInt32(7),
									PerTryTimeout: util_proto.Duration(time.Nanosecond * 200000000),
									BackOff: &mesh_proto.Retry_Conf_BackOff{
										BaseInterval: util_proto.Duration(time.Nanosecond * 10000000),
										MaxInterval: util_proto.Duration(time.
											Second * 3),
									},
								},
							},
						},
					},
				},
				expected: core_xds.RetryMap{
					"redis": &core_mesh.RetryResource{
						Meta: &test_model.ResourceMeta{
							Name: "retry-everything-passive",
						},
					},
				},
			}),
			Entry("Retry with a `source` selector by 2 tags should win over a Retry with a"+
				" `source` selector by 1 tag", testCase{
				dataplane: &core_mesh.DataplaneResource{
					Spec: &mesh_proto.Dataplane{
						Networking: &mesh_proto.Dataplane_Networking{
							Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
								{
									Tags: map[string]string{
										"kuma.io/service": "backend",
										"region":          "eu",
									},
								},
							},
							Outbound: []*mesh_proto.Dataplane_Networking_Outbound{
								{
									Tags: map[string]string{
										"kuma.io/service": "redis",
									},
								},
							},
						},
					},
				},
				destinations: core_xds.DestinationMap{
					"redis": core_xds.TagSelectorSet{
						mesh_proto.MatchService("redis"),
					},
				},
				retries: []*core_mesh.RetryResource{
					{
						Meta: &test_model.ResourceMeta{
							Name: "less-specific",
						},
						Spec: &mesh_proto.Retry{
							Sources: []*mesh_proto.Selector{
								{
									Match: mesh_proto.TagSelector{
										"kuma.io/service": "backend",
									},
								},
							},
							Destinations: []*mesh_proto.Selector{
								{
									Match: mesh_proto.TagSelector{
										"kuma.io/service": "*",
									},
								},
							},
							Conf: &mesh_proto.Retry_Conf{
								Http: &mesh_proto.Retry_Conf_Http{
									NumRetries:    util_proto.UInt32(3),
									PerTryTimeout: util_proto.Duration(time.Nanosecond * 200000000),
									BackOff: &mesh_proto.Retry_Conf_BackOff{
										BaseInterval: util_proto.Duration(time.Nanosecond * 10000000),
										MaxInterval:  util_proto.Duration(time.Second * 1),
									},
									RetriableStatusCodes: []uint32{500, 504},
								},
							},
						},
					},
					{
						Meta: &test_model.ResourceMeta{
							Name: "more-specific",
						},
						Spec: &mesh_proto.Retry{
							Sources: []*mesh_proto.Selector{
								{
									Match: mesh_proto.TagSelector{
										"kuma.io/service": "backend",
										"region":          "eu",
									},
								},
							},
							Destinations: []*mesh_proto.Selector{
								{
									Match: mesh_proto.TagSelector{
										"kuma.io/service": "*",
									},
								},
							},
							Conf: &mesh_proto.Retry_Conf{
								Http: &mesh_proto.Retry_Conf_Http{
									NumRetries:    util_proto.UInt32(7),
									PerTryTimeout: util_proto.Duration(time.Nanosecond * 200000000),
									BackOff: &mesh_proto.Retry_Conf_BackOff{
										BaseInterval: util_proto.Duration(time.Nanosecond * 10000000),
										MaxInterval:  util_proto.Duration(time.Second * 3),
									},
								},
							},
						},
					},
				},
				expected: core_xds.RetryMap{
					"redis": &core_mesh.RetryResource{
						Meta: &test_model.ResourceMeta{
							Name: "more-specific",
						},
					},
				},
			}),
			Entry("Retry with a `source` selector by an exact value should win over a Retry"+
				" with a `source` selector by a wildcard value", testCase{
				dataplane: &core_mesh.DataplaneResource{
					Spec: &mesh_proto.Dataplane{
						Networking: &mesh_proto.Dataplane_Networking{
							Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
								{
									Tags: map[string]string{
										"kuma.io/service": "backend",
										"region":          "eu",
									},
								},
							},
							Outbound: []*mesh_proto.Dataplane_Networking_Outbound{
								{
									Tags: map[string]string{
										"kuma.io/service": "redis",
									},
								},
							},
						},
					},
				},
				destinations: core_xds.DestinationMap{
					"redis": core_xds.TagSelectorSet{
						mesh_proto.MatchService("redis"),
					},
				},
				retries: []*core_mesh.RetryResource{
					{
						Meta: &test_model.ResourceMeta{
							Name: "less-specific",
						},
						Spec: &mesh_proto.Retry{
							Sources: []*mesh_proto.Selector{
								{
									Match: mesh_proto.TagSelector{
										"kuma.io/service": "*",
									},
								},
							},
							Destinations: []*mesh_proto.Selector{
								{
									Match: mesh_proto.TagSelector{
										"kuma.io/service": "*",
									},
								},
							},
							Conf: &mesh_proto.Retry_Conf{
								Http: &mesh_proto.Retry_Conf_Http{
									NumRetries:    util_proto.UInt32(3),
									PerTryTimeout: util_proto.Duration(time.Nanosecond * 200000000),
									BackOff: &mesh_proto.Retry_Conf_BackOff{
										BaseInterval: util_proto.Duration(time.Nanosecond * 10000000),
										MaxInterval:  util_proto.Duration(time.Second * 1),
									},
									RetriableStatusCodes: []uint32{500, 504},
								},
							},
						},
					},
					{
						Meta: &test_model.ResourceMeta{
							Name: "more-specific",
						},
						Spec: &mesh_proto.Retry{
							Sources: []*mesh_proto.Selector{
								{
									Match: mesh_proto.TagSelector{
										"kuma.io/service": "backend",
									},
								},
							},
							Destinations: []*mesh_proto.Selector{
								{
									Match: mesh_proto.TagSelector{
										"kuma.io/service": "*",
									},
								},
							},
							Conf: &mesh_proto.Retry_Conf{
								Http: &mesh_proto.Retry_Conf_Http{
									NumRetries:    util_proto.UInt32(7),
									PerTryTimeout: util_proto.Duration(time.Nanosecond * 200000000),

									BackOff: &mesh_proto.Retry_Conf_BackOff{
										BaseInterval: util_proto.Duration(time.Nanosecond * 10000000),
										MaxInterval:  util_proto.Duration(time.Second * 3),
									},
								},
							},
						},
					},
				},
				expected: core_xds.RetryMap{
					"redis": &core_mesh.RetryResource{
						Meta: &test_model.ResourceMeta{
							Name: "more-specific",
						},
					},
				},
			}),
			Entry("Retry with a `destination` selector by an exact value should win over a"+
				" Retry with a `destination` selector by a wildcard value", testCase{
				dataplane: &core_mesh.DataplaneResource{
					Spec: &mesh_proto.Dataplane{
						Networking: &mesh_proto.Dataplane_Networking{
							Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
								{
									Tags: map[string]string{
										"kuma.io/service": "backend",
										"region":          "eu",
									},
								},
							},
							Outbound: []*mesh_proto.Dataplane_Networking_Outbound{
								{
									Tags: map[string]string{
										"kuma.io/service": "redis",
									},
								},
							},
						},
					},
				},
				destinations: core_xds.DestinationMap{
					"redis": core_xds.TagSelectorSet{
						mesh_proto.MatchService("redis"),
					},
				},
				retries: []*core_mesh.RetryResource{
					{
						Meta: &test_model.ResourceMeta{
							Name: "less-specific",
						},
						Spec: &mesh_proto.Retry{
							Sources: []*mesh_proto.Selector{
								{
									Match: mesh_proto.TagSelector{
										"kuma.io/service": "*",
									},
								},
							},
							Destinations: []*mesh_proto.Selector{
								{
									Match: mesh_proto.TagSelector{
										"kuma.io/service": "*",
									},
								},
							},
							Conf: &mesh_proto.Retry_Conf{
								Http: &mesh_proto.Retry_Conf_Http{
									NumRetries:    util_proto.UInt32(3),
									PerTryTimeout: util_proto.Duration(time.Nanosecond * 200000000),
									BackOff: &mesh_proto.Retry_Conf_BackOff{
										BaseInterval: util_proto.Duration(time.Nanosecond * 10000000),
										MaxInterval:  util_proto.Duration(time.Second * 1),
									},
									RetriableStatusCodes: []uint32{500, 504},
								},
							},
						},
					},
					{
						Meta: &test_model.ResourceMeta{
							Name: "more-specific",
						},
						Spec: &mesh_proto.Retry{
							Sources: []*mesh_proto.Selector{
								{
									Match: mesh_proto.TagSelector{
										"kuma.io/service": "*",
									},
								},
							},
							Destinations: []*mesh_proto.Selector{
								{
									Match: mesh_proto.TagSelector{
										"kuma.io/service": "redis",
									},
								},
							},
							Conf: &mesh_proto.Retry_Conf{
								Http: &mesh_proto.Retry_Conf_Http{
									NumRetries:    util_proto.UInt32(7),
									PerTryTimeout: util_proto.Duration(time.Nanosecond * 200000000),
									BackOff: &mesh_proto.Retry_Conf_BackOff{
										BaseInterval: util_proto.Duration(time.Nanosecond * 10000000),
										MaxInterval:  util_proto.Duration(time.Second * 3),
									},
								},
							},
						},
					},
				},
				expected: core_xds.RetryMap{
					"redis": &core_mesh.RetryResource{
						Meta: &test_model.ResourceMeta{
							Name: "more-specific",
						},
					},
				},
			}),
			Entry("in case if Retries have equal aggregate ranks, "+
				"most specific one should be selected based on last creation time", testCase{
				dataplane: &core_mesh.DataplaneResource{
					Spec: &mesh_proto.Dataplane{
						Networking: &mesh_proto.Dataplane_Networking{
							Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
								{
									Tags: map[string]string{
										"kuma.io/service": "backend",
										"region":          "eu",
									},
								},
							},
							Outbound: []*mesh_proto.Dataplane_Networking_Outbound{
								{
									Tags: map[string]string{
										"kuma.io/service": "redis",
									},
								},
							},
						},
					},
				},
				destinations: core_xds.DestinationMap{
					"redis": core_xds.TagSelectorSet{
						mesh_proto.MatchService("redis"),
					},
				},
				retries: []*core_mesh.RetryResource{
					{
						Meta: &test_model.ResourceMeta{
							Name:         "equally-specific-2",
							CreationTime: time.Unix(1, 1),
						},
						Spec: &mesh_proto.Retry{
							Sources: []*mesh_proto.Selector{
								{
									Match: mesh_proto.TagSelector{
										"kuma.io/service": "*",
									},
								},
							},
							Destinations: []*mesh_proto.Selector{
								{
									Match: mesh_proto.TagSelector{
										"kuma.io/service": "redis",
									},
								},
							},
							Conf: &mesh_proto.Retry_Conf{
								Http: &mesh_proto.Retry_Conf_Http{
									NumRetries:    util_proto.UInt32(3),
									PerTryTimeout: util_proto.Duration(time.Nanosecond * 200000000),
									BackOff: &mesh_proto.Retry_Conf_BackOff{
										BaseInterval: util_proto.Duration(time.Nanosecond * 10000000),
										MaxInterval:  util_proto.Duration(time.Second * 1),
									},
									RetriableStatusCodes: []uint32{500, 504},
								},
							},
						},
					},
					{
						Meta: &test_model.ResourceMeta{
							Name:         "equally-specific-1",
							CreationTime: time.Unix(0, 0),
						},
						Spec: &mesh_proto.Retry{
							Sources: []*mesh_proto.Selector{
								{
									Match: mesh_proto.TagSelector{
										"kuma.io/service": "backend",
									},
								},
							},
							Destinations: []*mesh_proto.Selector{
								{
									Match: mesh_proto.TagSelector{
										"kuma.io/service": "*",
									},
								},
							},
							Conf: &mesh_proto.Retry_Conf{
								Http: &mesh_proto.Retry_Conf_Http{
									NumRetries:    util_proto.UInt32(7),
									PerTryTimeout: util_proto.Duration(time.Nanosecond * 200000000),
									BackOff: &mesh_proto.Retry_Conf_BackOff{
										BaseInterval: util_proto.Duration(time.Nanosecond * 10000000),
										MaxInterval:  util_proto.Duration(time.Second * 3),
									},
								},
							},
						},
					},
				},
				expected: core_xds.RetryMap{
					"redis": &core_mesh.RetryResource{
						Meta: &test_model.ResourceMeta{
							Name: "equally-specific-2",
						},
					},
				},
			}),
		)
	})

})
