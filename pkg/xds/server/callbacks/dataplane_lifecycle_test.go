package callbacks_test

import (
	"context"
	"fmt"
	"sync"
	"time"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_sd "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	envoy_server "github.com/envoyproxy/go-control-plane/pkg/server/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/structpb"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/pkg/test/resources/model"
	"github.com/kumahq/kuma/pkg/util/proto"
	util_xds_v3 "github.com/kumahq/kuma/pkg/util/xds/v3"
	xds_auth "github.com/kumahq/kuma/pkg/xds/auth"
	. "github.com/kumahq/kuma/pkg/xds/server/callbacks"
)

type staticAuthenticator struct {
	err error
}

func (s *staticAuthenticator) Authenticate(ctx context.Context, resource core_model.Resource, credential xds_auth.Credential) error {
	return s.err
}

var _ xds_auth.Authenticator = &staticAuthenticator{}

var _ = Describe("Dataplane Lifecycle", func() {
	const cpInstanceID = "xyz"

	var authenticator *staticAuthenticator
	var resManager core_manager.ResourceManager
	var callbacks envoy_server.Callbacks
	var cancel func()
	var ctx context.Context

	BeforeEach(func() {
		authenticator = &staticAuthenticator{}
		store := memory.NewStore()
		resManager = core_manager.NewResourceManager(store)
		ctx, cancel = context.WithCancel(context.Background())

		dpLifecycle := NewDataplaneLifecycle(ctx, resManager, authenticator, 0*time.Second, cpInstanceID, 0*time.Second)
		callbacks = util_xds_v3.AdaptCallbacks(DataplaneCallbacksToXdsCallbacks(dpLifecycle))

		err := resManager.Create(context.Background(), core_mesh.NewMeshResource(), core_store.CreateByKey(core_model.DefaultMesh, core_model.NoMesh))
		Expect(err).ToNot(HaveOccurred())
	})

	It("should create a DP on the first DiscoveryRequest when it is carried with metadata and delete on stream close", func() {
		// given
		node := &envoy_core.Node{
			Id: "default.backend-01",
			Metadata: &structpb.Struct{
				Fields: map[string]*structpb.Value{
					"dataplane.resource": {
						Kind: &structpb.Value_StringValue{
							StringValue: `
                                {
                                  "type": "Dataplane",
                                  "mesh": "default",
                                  "name": "backend-01",
                                  "networking": {
                                    "address": "127.0.0.1",
                                    "inbound": [
                                      {
                                        "port": 22022,
                                        "servicePort": 8443,
                                        "tags": {
                                          "kuma.io/service": "backend"
                                        }
                                      },
                                    ]
                                  }
                                }
                                `,
						},
					},
				},
			},
		}
		req := envoy_sd.DiscoveryRequest{
			Node: node,
		}
		const streamId = 123
		ctx := metadata.NewIncomingContext(context.Background(), map[string][]string{
			"authorization": {"token"},
		})
		Expect(callbacks.OnStreamOpen(ctx, streamId, "")).To(Succeed())

		// when
		err := callbacks.OnStreamRequest(streamId, &req)

		// then dp is created
		Expect(err).ToNot(HaveOccurred())
		err = resManager.Get(context.Background(), core_mesh.NewDataplaneResource(), core_store.GetByKey("backend-01", "default"))
		Expect(err).ToNot(HaveOccurred())

		// when
		callbacks.OnStreamClosed(streamId, node)

		// then dataplane should be deleted
		err = resManager.Get(context.Background(), core_mesh.NewDataplaneResource(), core_store.GetByKey("backend-01", "default"))
		Expect(core_store.IsResourceNotFound(err)).To(BeTrue())
	})

	It("should not override extisting DP with different service", func() {
		// given already created DP
		dp := &core_mesh.DataplaneResource{
			Meta: &model.ResourceMeta{
				Mesh: "default",
				Name: "dp-01",
			},
			Spec: &mesh_proto.Dataplane{
				Networking: &mesh_proto.Dataplane_Networking{
					Address: "192.168.0.1",
					Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
						{
							Port:        8080,
							ServicePort: 8081,
							Tags: map[string]string{
								"kuma.io/service": "backend",
							},
						},
					},
				},
			},
		}
		err := resManager.Create(context.Background(), dp, core_store.CreateByKey("dp-01", "default"))
		Expect(err).ToNot(HaveOccurred())

		// when
		authenticator.err = errors.New("token rejected")
		req := envoy_sd.DiscoveryRequest{
			Node: &envoy_core.Node{
				Id: "default.dp-01",
				Metadata: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"dataplane.resource": {
							Kind: &structpb.Value_StringValue{
								StringValue: `
                                {
                                  "type": "Dataplane",
                                  "mesh": "default",
                                  "name": "dp-01",
                                  "networking": {
                                    "address": "127.0.0.1",
                                    "inbound": [
                                      {
                                        "port": 22022,
                                        "servicePort": 8443,
                                        "tags": {
                                          "kuma.io/service": "web"
                                        }
                                      },
                                    ]
                                  }
                                }
                                `,
							},
						},
					},
				},
			},
		}
		const streamId = 123
		ctx := metadata.NewIncomingContext(context.Background(), map[string][]string{
			"authorization": {"token"},
		})
		Expect(callbacks.OnStreamOpen(ctx, streamId, "")).To(Succeed())
		err = callbacks.OnStreamRequest(streamId, &req)

		// then
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("token rejected"))
	})

	It("should not delete DP when it is not carried in metadata", func() {
		// given already created DP
		dp := &core_mesh.DataplaneResource{
			Meta: &model.ResourceMeta{
				Mesh: "default",
				Name: "backend-01",
			},
			Spec: &mesh_proto.Dataplane{
				Networking: &mesh_proto.Dataplane_Networking{
					Address: "192.168.0.1",
					Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
						{
							Port:        8080,
							ServicePort: 8081,
							Tags: map[string]string{
								"kuma.io/service": "backend",
							},
						},
					},
				},
			},
		}
		err := resManager.Create(context.Background(), dp, core_store.CreateByKey("backend-01", "default"))
		Expect(err).ToNot(HaveOccurred())

		node := &envoy_core.Node{
			Id: "default.backend-01",
		}
		req := envoy_sd.DiscoveryRequest{
			Node: node,
		}
		const streamId = 123

		// when
		err = callbacks.OnStreamRequest(streamId, &req)

		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		callbacks.OnStreamClosed(streamId, node)

		// then DP is not deleted because it was not carried in metadata
		err = resManager.Get(context.Background(), core_mesh.NewDataplaneResource(), core_store.GetByKey("backend-01", "default"))
		Expect(err).ToNot(HaveOccurred())
	})

	It("should not delete DP when Kuma CP is shutting down", func() {
		// given
		node := &envoy_core.Node{
			Id: "default.backend-01",
			Metadata: &structpb.Struct{
				Fields: map[string]*structpb.Value{
					"dataplane.resource": {
						Kind: &structpb.Value_StringValue{
							StringValue: `
                                {
                                  "type": "Dataplane",
                                  "mesh": "default",
                                  "name": "backend-01",
                                  "networking": {
                                    "address": "127.0.0.1",
                                    "inbound": [
                                      {
                                        "port": 22022,
                                        "servicePort": 8443,
                                        "tags": {
                                          "kuma.io/service": "backend"
                                        }
                                      },
                                    ]
                                  }
                                }
                                `,
						},
					},
				},
			},
		}
		req := envoy_sd.DiscoveryRequest{
			Node: node,
		}

		dp := &core_mesh.DataplaneResource{
			Meta: &model.ResourceMeta{
				Mesh: "default",
				Name: "dp-01",
			},
			Spec: &mesh_proto.Dataplane{
				Networking: &mesh_proto.Dataplane_Networking{
					Address: "192.168.0.1",
					Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
						{
							Port:        8080,
							ServicePort: 8081,
							Tags: map[string]string{
								"kuma.io/service": "backend",
							},
						},
					},
				},
			},
		}
		err := resManager.Create(context.Background(), dp, core_store.CreateByKey("dp-01", "default"))
		Expect(err).ToNot(HaveOccurred())

		const streamId = 123

		// when
		ctx := metadata.NewIncomingContext(context.Background(), map[string][]string{
			"authorization": {"token"},
		})
		Expect(callbacks.OnStreamOpen(ctx, streamId, "")).To(Succeed())

		err = callbacks.OnStreamRequest(streamId, &req)
		Expect(err).ToNot(HaveOccurred())

		cancel()
		// when
		callbacks.OnStreamClosed(streamId, node)

		// then DP is not deleted because Kuma CP was shutting down
		err = resManager.Get(context.Background(), core_mesh.NewDataplaneResource(), core_store.GetByKey("backend-01", "default"))
		Expect(err).ToNot(HaveOccurred())
	})

	It("should not race when registering concurrently", func() {
		const streamID = 123

		wg := sync.WaitGroup{}

		for i := 0; i < 10; i++ {
			wg.Add(1)

			go func(num int) {
				defer GinkgoRecover()

				streamID := int64(streamID + num)
				name := fmt.Sprintf("backend-%d", num)
				nodeID := fmt.Sprintf("default.%s", name)

				// given
				node := &envoy_core.Node{
					Id: nodeID,
					Metadata: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"dataplane.resource": {
								Kind: &structpb.Value_StringValue{
									StringValue: fmt.Sprintf(`
                                {
                                  "type": "Dataplane",
                                  "mesh": "default",
                                  "name": "%s",
                                  "networking": {
                                    "address": "127.0.0.%d",
                                    "inbound": [
                                      {
                                        "port": 22022,
                                        "servicePort": 8443,
                                        "tags": {
                                          "kuma.io/service": "backend"
                                        }
                                      },
                                    ]
                                  }
                                }
                                `, name, num),
								},
							},
						},
					},
				}
				req := envoy_sd.DiscoveryRequest{
					Node: node,
				}
				ctx := metadata.NewIncomingContext(context.Background(), map[string][]string{
					"authorization": {"token"},
				})
				Expect(callbacks.OnStreamOpen(ctx, streamID, "")).To(Succeed())

				// when
				err := callbacks.OnStreamRequest(streamID, &req)
				Expect(err).ToNot(HaveOccurred())

				// then dp is created
				err = resManager.Get(context.Background(), core_mesh.NewDataplaneResource(), core_store.GetByKey(name, "default"))
				Expect(err).ToNot(HaveOccurred())

				// when
				callbacks.OnStreamClosed(streamID, node)

				// then dataplane should be deleted
				err = resManager.Get(context.Background(), core_mesh.NewDataplaneResource(), core_store.GetByKey("backend-01", "default"))
				Expect(core_store.IsResourceNotFound(err)).To(BeTrue())

				wg.Done()
			}(i)

		}

		wg.Wait()
	})

	It("should not unregister proxy when it is connected to other instances", func() {
		// given a DP registered by callbacks
		node := &envoy_core.Node{
			Id: "default.backend-01",
			Metadata: &structpb.Struct{
				Fields: map[string]*structpb.Value{
					"dataplane.resource": {
						Kind: &structpb.Value_StringValue{
							StringValue: `
                                {
                                  "type": "Dataplane",
                                  "mesh": "default",
                                  "name": "backend-01",
                                  "networking": {
                                    "address": "127.0.0.1",
                                    "inbound": [
                                      {
                                        "port": 22022,
                                        "servicePort": 8443,
                                        "tags": {
                                          "kuma.io/service": "backend"
                                        }
                                      },
                                    ]
                                  }
                                }
                                `,
						},
					},
				},
			},
		}
		req := envoy_sd.DiscoveryRequest{
			Node: node,
		}

		key := core_model.ResourceKey{
			Mesh: "default",
			Name: "backend-01",
		}

		const streamId = 123
		ctx := metadata.NewIncomingContext(context.Background(), map[string][]string{
			"authorization": {"token"},
		})
		Expect(callbacks.OnStreamOpen(ctx, streamId, "")).To(Succeed())
		err := callbacks.OnStreamRequest(streamId, &req)
		Expect(err).ToNot(HaveOccurred())

		// and insight that indicates that DP has connected to another instance of CP. For example
		// 1) Envoy dropped the XDS connection to instance-1
		// 2) Envoy reconnected to instance-2
		// 3) instance-1 noticed that the connection is dropped.
		insight := core_mesh.NewDataplaneInsightResource()
		insight.Spec.Subscriptions = append(insight.Spec.Subscriptions, &mesh_proto.DiscoverySubscription{
			Id:                     "",
			ControlPlaneInstanceId: "not-a" + cpInstanceID,
			ConnectTime:            proto.MustTimestampProto(time.Now()),
		})
		Expect(resManager.Create(context.Background(), insight, core_store.CreateBy(key))).To(Succeed())

		// when
		callbacks.OnStreamClosed(streamId, node)

		// then DP is not deleted because Kuma DP is connected to another instance
		err = resManager.Get(context.Background(), core_mesh.NewDataplaneResource(), core_store.GetBy(key))
		Expect(err).ToNot(HaveOccurred())
	})
})
