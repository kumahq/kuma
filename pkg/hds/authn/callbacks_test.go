package authn_test

import (
	"context"
	"errors"

	envoy_config_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_service_health_v3 "github.com/envoyproxy/go-control-plane/envoy/service/health/v3"
	"google.golang.org/grpc/metadata"

	"github.com/kumahq/kuma/pkg/hds/authn"
	hds_callbacks "github.com/kumahq/kuma/pkg/hds/callbacks"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	"github.com/kumahq/kuma/pkg/xds/auth"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type testAuthenticator struct {
	callCounter int
}

var _ auth.Authenticator = &testAuthenticator{}

func (t *testAuthenticator) Authenticate(ctx context.Context, dataplane *core_mesh.DataplaneResource, credential auth.Credential) error {
	t.callCounter++
	if credential == "pass" {
		return nil
	}
	return errors.New("invalid credential")
}

var _ = Describe("Authn Callbacks", func() {

	var testAuth *testAuthenticator
	var resManager core_manager.ResourceManager
	var callbacks hds_callbacks.Callbacks

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

	BeforeEach(func() {
		memStore := memory.NewStore()
		resManager = core_manager.NewResourceManager(memStore)
		testAuth = &testAuthenticator{}
		callbacks = authn.NewCallbacks(resManager, testAuth, authn.DPNotFoundRetry{})

		err := resManager.Create(context.Background(), core_mesh.NewMeshResource(), core_store.CreateByKey(model.DefaultMesh, model.NoMesh))
		Expect(err).ToNot(HaveOccurred())
		err = resManager.Create(context.Background(), dpRes, core_store.CreateByKey("web-01", "default"))
		Expect(err).ToNot(HaveOccurred())
	})

	It("should authenticate only first request on the stream", func() {
		// given
		ctx := metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{"authorization": "pass"}))
		streamID := int64(1)

		// when
		err := callbacks.OnStreamOpen(ctx, streamID)

		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		err = callbacks.OnHealthCheckRequest(streamID, &envoy_service_health_v3.HealthCheckRequest{
			Node: &envoy_config_core_v3.Node{
				Id: "default.web-01",
			},
		})

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(testAuth.callCounter).To(Equal(1))

		// when send second request that is already authenticated
		err = callbacks.OnHealthCheckRequest(streamID, &envoy_service_health_v3.HealthCheckRequest{})

		// then auth is called only once
		Expect(err).ToNot(HaveOccurred())
		Expect(testAuth.callCounter).To(Equal(1))
	})

	It("should not authenticate when DP is absent in the CP and it's not passed through Envoy metadata", func() {
		// given
		ctx := metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{"authorization": "pass"}))
		streamID := int64(1)

		// when
		err := callbacks.OnStreamOpen(ctx, streamID)

		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		err = callbacks.OnHealthCheckRequest(streamID, &envoy_service_health_v3.HealthCheckRequest{
			Node: &envoy_config_core_v3.Node{
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
		err := callbacks.OnStreamOpen(ctx, streamID)

		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		err = callbacks.OnHealthCheckRequest(streamID, &envoy_service_health_v3.HealthCheckRequest{
			Node: &envoy_config_core_v3.Node{
				Id: "default.web-01",
			},
		})

		// then
		Expect(err).To(MatchError("authentication failed: invalid credential"))
	})
})
