package callbacks_test

import (
	"context"
	"fmt"
	"sync"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_sd "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	envoy_server "github.com/envoyproxy/go-control-plane/pkg/server/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/structpb"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/pkg/test/resources/model"
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

		callbacks = util_xds_v3.AdaptCallbacks(DataplaneCallbacksToXdsCallbacks(NewDataplaneLifecycle(ctx, resManager, authenticator)))

		err := resManager.Create(context.Background(), core_mesh.NewMeshResource(), core_store.CreateByKey(core_model.DefaultMesh, core_model.NoMesh))
		Expect(err).ToNot(HaveOccurred())
	})

	It("should create a DP on the first DiscoveryRequest when it is carried with metadata and delete on stream close", func() {
		// given
		req := envoy_sd.DiscoveryRequest{
			Node: &envoy_core.Node{
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
			},
		}
		const streamId = 123

		// when
		err := callbacks.OnStreamRequest(streamId, &req)

		// then dp is created
		Expect(err).ToNot(HaveOccurred())
		err = resManager.Get(context.Background(), core_mesh.NewDataplaneResource(), core_store.GetByKey("backend-01", "default"))
		Expect(err).ToNot(HaveOccurred())

		// when
		callbacks.OnStreamClosed(streamId)

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
		authenticator.err = errors.New("rejected")
		req := envoy_sd.DiscoveryRequest{
			Node: &envoy_core.Node{
				Id: "default.backend-01",
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
		Expect(callbacks.OnStreamOpen(context.Background(), streamId, "")).To(Succeed())
		err = callbacks.OnStreamRequest(streamId, &req)

		// then
		Expect(err).To(HaveOccurred())
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

		req := envoy_sd.DiscoveryRequest{
			Node: &envoy_core.Node{
				Id: "default.backend-01",
			},
		}
		const streamId = 123

		// when
		err = callbacks.OnStreamRequest(streamId, &req)

		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		callbacks.OnStreamClosed(streamId)

		// then DP is not deleted because it was not carried in metadata
		err = resManager.Get(context.Background(), core_mesh.NewDataplaneResource(), core_store.GetByKey("backend-01", "default"))
		Expect(err).ToNot(HaveOccurred())
	})

	It("should not delete DP when Kuma CP is shutting down", func() {
		// given
		req := envoy_sd.DiscoveryRequest{
			Node: &envoy_core.Node{
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
			},
		}

		const streamId = 123

		// when
		err := callbacks.OnStreamRequest(streamId, &req)
		Expect(err).ToNot(HaveOccurred())

		cancel()
		// when
		callbacks.OnStreamClosed(streamId)

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
				nodeID := fmt.Sprintf("default.backend-%d", num)

				// given
				req := envoy_sd.DiscoveryRequest{
					Node: &envoy_core.Node{
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
                                `, nodeID, num),
									},
								},
							},
						},
					},
				}

				// when
				err := callbacks.OnStreamRequest(streamID, &req)
				Expect(err).ToNot(HaveOccurred())

				// then dp is created
				err = resManager.Get(context.Background(), core_mesh.NewDataplaneResource(), core_store.GetByKey(nodeID, "default"))
				Expect(err).ToNot(HaveOccurred())

				// when
				callbacks.OnStreamClosed(streamID)

				// then dataplane should be deleted
				err = resManager.Get(context.Background(), core_mesh.NewDataplaneResource(), core_store.GetByKey("backend-01", "default"))
				Expect(core_store.IsResourceNotFound(err)).To(BeTrue())

				wg.Done()
			}(i)

		}

		wg.Wait()

	})
})
