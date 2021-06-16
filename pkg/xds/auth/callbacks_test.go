package auth_test

import (
	"context"
	"errors"

	envoy_api "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	envoy_server "github.com/envoyproxy/go-control-plane/pkg/server/v2"
	pstruct "github.com/golang/protobuf/ptypes/struct"
	"google.golang.org/grpc/metadata"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/model/rest"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	util_xds_v2 "github.com/kumahq/kuma/pkg/util/xds/v2"
	"github.com/kumahq/kuma/pkg/xds/auth"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type testAuthenticator struct {
	callCounter     int
	zoneCallCounter int
}

var _ auth.Authenticator = &testAuthenticator{}

func (t *testAuthenticator) Authenticate(ctx context.Context, dataplane *core_mesh.DataplaneResource, credential auth.Credential) error {
	t.callCounter++
	if credential == "pass" {
		return nil
	}
	return errors.New("invalid credential")
}

func (t *testAuthenticator) AuthenticateZoneIngress(ctx context.Context, zoneIngress *core_mesh.ZoneIngressResource, credential auth.Credential) error {
	t.zoneCallCounter++
	if credential == "zone pass" {
		return nil
	}
	return errors.New("invalid credential")
}

var _ = Describe("Auth Callbacks", func() {

	var testAuth *testAuthenticator
	var resManager core_manager.ResourceManager
	var callbacks envoy_server.Callbacks

	dpRes := &core_mesh.DataplaneResource{
		Meta: &test_model.ResourceMeta{
			Name: "web-01",
			Mesh: "default",
		},
		Spec: &mesh_proto.Dataplane{
			Networking: &mesh_proto.Dataplane_Networking{
				Address: "127.0.0.1",
				Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
					{
						Port:        8080,
						ServicePort: 8081,
						Tags: map[string]string{
							"kuma.io/service":  "web",
							"kuma.io/protocol": "http",
						},
					},
				},
			},
		},
	}

	zoneIngress := &core_mesh.ZoneIngressResource{
		Meta: &test_model.ResourceMeta{
			Name: "ingress",
			Mesh: core_model.NoMesh,
		},
		Spec: &mesh_proto.ZoneIngress{
			Networking: &mesh_proto.ZoneIngress_Networking{
				Address: "1.1.1.1",
				Port:    10001,
			},
		},
	}

	BeforeEach(func() {
		memStore := memory.NewStore()
		resManager = core_manager.NewResourceManager(memStore)
		testAuth = &testAuthenticator{}
		callbacks = util_xds_v2.AdaptCallbacks(auth.NewCallbacks(resManager, testAuth, auth.DPNotFoundRetry{}))

		err := resManager.Create(context.Background(), core_mesh.NewMeshResource(), core_store.CreateByKey(core_model.DefaultMesh, core_model.NoMesh))
		Expect(err).ToNot(HaveOccurred())
		err = resManager.Create(context.Background(), dpRes, core_store.CreateByKey("web-01", "default"))
		Expect(err).ToNot(HaveOccurred())
		err = resManager.Create(context.Background(), zoneIngress, core_store.CreateBy(core_model.MetaToResourceKey(zoneIngress.GetMeta())))
		Expect(err).ToNot(HaveOccurred())
	})

	It("should authenticate only first request on the stream", func() {
		// given
		ctx := metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{"authorization": "pass"}))
		streamID := int64(1)

		// when
		err := callbacks.OnStreamOpen(ctx, streamID, "")

		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		err = callbacks.OnStreamRequest(streamID, &envoy_api.DiscoveryRequest{
			Node: &envoy_core.Node{
				Id: "default.web-01",
			},
		})

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(testAuth.callCounter).To(Equal(1))

		// when send second request that is already authenticated
		err = callbacks.OnStreamRequest(streamID, &envoy_api.DiscoveryRequest{})

		// then auth is called only once
		Expect(err).ToNot(HaveOccurred())
		Expect(testAuth.callCounter).To(Equal(1))
	})

	It("should authenticate when DP is passed through Envoy metadata", func() {
		// given mesh without web-01 dataplane
		ctx := metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{"authorization": "pass"}))
		streamID := int64(1)
		err := resManager.Delete(context.Background(), core_mesh.NewDataplaneResource(), core_store.DeleteByKey("web-01", "default"))
		Expect(err).ToNot(HaveOccurred())

		// when
		err = callbacks.OnStreamOpen(ctx, streamID, "")

		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		json, err := rest.From.Resource(dpRes).MarshalJSON()
		Expect(err).ToNot(HaveOccurred())
		err = callbacks.OnStreamRequest(streamID, &envoy_api.DiscoveryRequest{
			Node: &envoy_core.Node{
				Id: "default.web-01",
				Metadata: &pstruct.Struct{
					Fields: map[string]*pstruct.Value{
						"dataplane.resource": {
							Kind: &pstruct.Value_StringValue{
								StringValue: string(json),
							},
						},
					},
				},
			},
		})

		// then
		Expect(err).ToNot(HaveOccurred())
	})

	It("should not let change node id once it's authenticated", func() {
		// given
		ctx := metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{"authorization": "pass"}))
		streamID := int64(1)

		// when
		err := callbacks.OnStreamOpen(ctx, streamID, "")

		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		err = callbacks.OnStreamRequest(streamID, &envoy_api.DiscoveryRequest{
			Node: &envoy_core.Node{
				Id: "default.web-01",
			},
		})

		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		err = callbacks.OnStreamRequest(streamID, &envoy_api.DiscoveryRequest{
			Node: &envoy_core.Node{
				Id: "default.web-02",
			},
		})

		// then
		Expect(err).To(MatchError("stream was authenticated for ID default.web-01. Received request is for node with ID default.web-02. Node ID cannot be changed after stream is initialized"))
	})

	It("should not authenticate when DP is absent in the CP and it's not passed through Envoy metadata", func() {
		// given
		ctx := metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{"authorization": "pass"}))
		streamID := int64(1)

		// when
		err := callbacks.OnStreamOpen(ctx, streamID, "")

		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		err = callbacks.OnStreamRequest(streamID, &envoy_api.DiscoveryRequest{
			Node: &envoy_core.Node{
				Id: "default.web-02",
			},
		})

		// then
		Expect(err).To(MatchError("retryable: dataplane not found. Create Dataplane in Kuma CP first or pass it as an argument to kuma-dp"))
	})

	It("should throw an error on authentication fail", func() {
		// given
		ctx := metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{"authorization": "invalid"}))
		streamID := int64(1)

		// when
		err := callbacks.OnStreamOpen(ctx, streamID, "")

		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		err = callbacks.OnStreamRequest(streamID, &envoy_api.DiscoveryRequest{
			Node: &envoy_core.Node{
				Id: "default.web-01",
			},
		})

		// then
		Expect(err).To(MatchError("authentication failed: invalid credential"))
	})

	It("should authenticate ingress", func() {
		// given
		ctx := metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{"authorization": "zone pass"}))
		streamID := int64(1)

		// when
		err := callbacks.OnStreamOpen(ctx, streamID, "")

		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		err = callbacks.OnStreamRequest(streamID, &envoy_api.DiscoveryRequest{
			Node: &envoy_core.Node{
				Id: ".ingress",
				Metadata: &pstruct.Struct{
					Fields: map[string]*pstruct.Value{
						"dataplane.proxyType": {
							Kind: &pstruct.Value_StringValue{
								StringValue: "ingress",
							},
						},
					},
				},
			},
		})

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(testAuth.zoneCallCounter).To(Equal(1))

		// when send second request that is already authenticated
		err = callbacks.OnStreamRequest(streamID, &envoy_api.DiscoveryRequest{})

		// then auth is called only once
		Expect(err).ToNot(HaveOccurred())
		Expect(testAuth.zoneCallCounter).To(Equal(1))
	})
})
