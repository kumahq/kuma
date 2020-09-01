package server_test

import (
	"context"

	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	pstruct "github.com/golang/protobuf/ptypes/struct"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/pkg/xds/server"
)

var _ = Describe("Dataplane Lifecycle", func() {

	var resManager core_manager.ResourceManager
	var dpLifecycle *server.DataplaneLifecycle

	BeforeEach(func() {
		store := memory.NewStore()
		resManager = core_manager.NewResourceManager(store)
		dpLifecycle = server.NewDataplaneLifecycle(resManager)

		err := resManager.Create(context.Background(), &core_mesh.MeshResource{}, core_store.CreateByKey("default", "default"))
		Expect(err).ToNot(HaveOccurred())
	})

	It("should create a DP on the first DiscoveryRequest when it is carried with metadata and delete on stream close", func() {
		// given
		req := v2.DiscoveryRequest{
			Node: &envoy_core.Node{
				Id: "default.backend-01",
				Metadata: &pstruct.Struct{
					Fields: map[string]*pstruct.Value{
						"dataplane.resource": &pstruct.Value{
							Kind: &pstruct.Value_StringValue{
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
		err := dpLifecycle.OnStreamRequest(streamId, &req)

		// then dp is created
		Expect(err).ToNot(HaveOccurred())
		err = resManager.Get(context.Background(), &core_mesh.DataplaneResource{}, core_store.GetByKey("backend-01", "default"))
		Expect(err).ToNot(HaveOccurred())

		// when
		dpLifecycle.OnStreamClosed(streamId)

		// then dataplane should be deleted
		err = resManager.Get(context.Background(), &core_mesh.DataplaneResource{}, core_store.GetByKey("backend-01", "default"))
		Expect(core_store.IsResourceNotFound(err)).To(BeTrue())
	})

	It("should not create DP when it is not carried in metadata", func() {
		// given
		req := v2.DiscoveryRequest{
			Node: &envoy_core.Node{
				Id: "default.backend-01",
			},
		}
		const streamId = 123

		// when
		err := dpLifecycle.OnStreamRequest(streamId, &req)

		// then dataplane is not created
		Expect(err).ToNot(HaveOccurred())
		err = resManager.Get(context.Background(), &core_mesh.DataplaneResource{}, core_store.GetByKey("backend-01", "default"))
		Expect(core_store.IsResourceNotFound(err)).To(BeTrue())
	})
})
