package auth_test

import (
	"context"
	"encoding/json"

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
	"github.com/kumahq/kuma/pkg/core/resources/model/rest"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	v3 "github.com/kumahq/kuma/pkg/util/xds/v3"
	"github.com/kumahq/kuma/pkg/xds/auth"
)

type testAuthenticator struct {
	callCounter     int
	zoneCallCounter int
}

var _ auth.Authenticator = &testAuthenticator{}

func (t *testAuthenticator) Authenticate(_ context.Context, resource core_model.Resource, credential auth.Credential) error {
	switch resource := resource.(type) {
	case *core_mesh.DataplaneResource:
		t.callCounter++
		if credential == "pass" {
			return nil
		}
	case *core_mesh.ZoneIngressResource:
		t.zoneCallCounter++
		if credential == "zone pass" {
			return nil
		}
	case *core_mesh.ZoneEgressResource:
		t.zoneCallCounter++
		if credential == "zone pass" {
			return nil
		}
	default:
		return errors.Errorf("no matching authenticator for %s resource", resource.Descriptor().Name)
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

	zoneEgress := &core_mesh.ZoneEgressResource{
		Meta: &test_model.ResourceMeta{
			Name: "egress",
			Mesh: core_model.NoMesh,
		},
		Spec: &mesh_proto.ZoneEgress{
			Networking: &mesh_proto.ZoneEgress_Networking{
				Address: "1.1.1.1",
				Port:    10002,
			},
		},
	}

	BeforeEach(func() {
		memStore := memory.NewStore()
		resManager = core_manager.NewResourceManager(memStore)
		testAuth = &testAuthenticator{}
		callbacks = v3.AdaptCallbacks(auth.NewCallbacks(resManager, testAuth, auth.DPNotFoundRetry{}))

		err := resManager.Create(context.Background(), core_mesh.NewMeshResource(), core_store.CreateByKey(core_model.DefaultMesh, core_model.NoMesh))
		Expect(err).ToNot(HaveOccurred())
		err = resManager.Create(context.Background(), dpRes, core_store.CreateByKey("web-01", "default"))
		Expect(err).ToNot(HaveOccurred())
		err = resManager.Create(context.Background(), zoneIngress, core_store.CreateBy(core_model.MetaToResourceKey(zoneIngress.GetMeta())))
		Expect(err).ToNot(HaveOccurred())
		err = resManager.Create(context.Background(), zoneEgress, core_store.CreateBy(core_model.MetaToResourceKey(zoneEgress.GetMeta())))
		Expect(err).ToNot(HaveOccurred())
	})

	It("should successfully authenticate", func() {
		// given
		ctx := metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{"authorization": "pass"}))
		streamID := int64(1)

		// when
		err := callbacks.OnStreamOpen(ctx, streamID, "")

		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		err = callbacks.OnStreamRequest(streamID, &envoy_sd.DiscoveryRequest{
			Node: &envoy_core.Node{
				Id: "default.web-01",
			},
		})

		// then
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
		res := rest.From.Resource(dpRes)
		jsonRes, err := json.Marshal(res)
		Expect(err).ToNot(HaveOccurred())
		err = callbacks.OnStreamRequest(streamID, &envoy_sd.DiscoveryRequest{
			Node: &envoy_core.Node{
				Id: "default.web-01",
				Metadata: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"dataplane.resource": {
							Kind: &structpb.Value_StringValue{
								StringValue: string(jsonRes),
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
		err = callbacks.OnStreamRequest(streamID, &envoy_sd.DiscoveryRequest{
			Node: &envoy_core.Node{
				Id: "default.web-01",
			},
		})

		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		err = callbacks.OnStreamRequest(streamID, &envoy_sd.DiscoveryRequest{
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
		err = callbacks.OnStreamRequest(streamID, &envoy_sd.DiscoveryRequest{
			Node: &envoy_core.Node{
				Id: "default.web-02",
			},
		})

		// then
		Expect(err.Error()).To(ContainSubstring("not found"))
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
		err = callbacks.OnStreamRequest(streamID, &envoy_sd.DiscoveryRequest{
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
		err = callbacks.OnStreamRequest(streamID, &envoy_sd.DiscoveryRequest{
			Node: &envoy_core.Node{
				Id: ".ingress",
				Metadata: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"dataplane.proxyType": {
							Kind: &structpb.Value_StringValue{
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
	})

	It("should authenticate egress", func() {
		// given
		ctx := metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{"authorization": "zone pass"}))
		streamID := int64(1)

		// when
		err := callbacks.OnStreamOpen(ctx, streamID, "")

		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		err = callbacks.OnStreamRequest(streamID, &envoy_sd.DiscoveryRequest{
			Node: &envoy_core.Node{
				Id: ".egress",
				Metadata: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"dataplane.proxyType": {
							Kind: &structpb.Value_StringValue{
								StringValue: "egress",
							},
						},
					},
				},
			},
		})

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(testAuth.zoneCallCounter).To(Equal(1))
	})
})
