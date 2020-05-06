package server_test

import (
	"context"
	"fmt"
	"time"

	envoy_api "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_api_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	envoy_discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v2"
	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	kuma_cp "github.com/Kong/kuma/pkg/config/app/kuma-cp"
	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	core_manager "github.com/Kong/kuma/pkg/core/resources/manager"
	core_model "github.com/Kong/kuma/pkg/core/resources/model"
	core_store "github.com/Kong/kuma/pkg/core/resources/store"
	core_xds "github.com/Kong/kuma/pkg/core/xds"
	sds_auth "github.com/Kong/kuma/pkg/sds/auth"
	"github.com/Kong/kuma/pkg/sds/server"
	"github.com/Kong/kuma/pkg/test"
	"github.com/Kong/kuma/pkg/test/runtime"
	tokens_builtin "github.com/Kong/kuma/pkg/tokens/builtin"
	tokens_issuer "github.com/Kong/kuma/pkg/tokens/builtin/issuer"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("SDS Server", func() {

	var dpCredential sds_auth.Credential
	var stop chan struct{}
	var client envoy_discovery.SecretDiscoveryServiceClient

	var resManager core_manager.ResourceManager

	BeforeSuite(func() {
		// setup runtime with SDS
		cfg := kuma_cp.DefaultConfig()
		cfg.SdsServer.DataplaneConfigurationRefreshInterval = 100 * time.Millisecond
		port, err := test.GetFreePort()
		Expect(err).ToNot(HaveOccurred())
		cfg.SdsServer.GrpcPort = port

		runtime, err := runtime.BuilderFor(cfg).Build()
		Expect(err).ToNot(HaveOccurred())
		resManager = runtime.ResourceManager()

		// setup default mesh with active mTLS and 2 CA
		meshRes := mesh_core.MeshResource{
			Spec: mesh_proto.Mesh{
				Mtls: &mesh_proto.Mesh_Mtls{
					EnabledBackend: "ca-1",
					Backends: []*mesh_proto.CertificateAuthorityBackend{
						{
							Name: "ca-1",
							Type: "builtin",
						},
						{
							Name: "ca-2",
							Type: "builtin",
						},
					},
				},
			},
		}
		err = resManager.Create(context.Background(), &meshRes, core_store.CreateByKey("default", "default"))
		Expect(err).ToNot(HaveOccurred())

		// setup backend dataplane
		dpRes := mesh_core.DataplaneResource{
			Spec: mesh_proto.Dataplane{
				Networking: &mesh_proto.Dataplane_Networking{
					Address: "192.168.0.1",
					Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
						{
							Port: 1234,
							Tags: map[string]string{
								"service": "backend",
							},
						},
					},
				},
			},
		}
		err = resManager.Create(context.Background(), &dpRes, core_store.CreateByKey("backend-01", "default"))
		Expect(err).ToNot(HaveOccurred())

		// setup Auth with Dataplane Token
		err = tokens_issuer.CreateDefaultSigningKey(runtime.SecretManager())
		Expect(err).ToNot(HaveOccurred())

		// retrieve example DP token
		tokenIssuer, err := tokens_builtin.NewDataplaneTokenIssuer(runtime)
		Expect(err).ToNot(HaveOccurred())
		dpCredential, err = tokenIssuer.Generate(core_xds.FromResourceKey(core_model.MetaToResourceKey(dpRes.GetMeta())))
		Expect(err).ToNot(HaveOccurred())

		// start the runtime
		Expect(server.SetupServer(runtime)).To(Succeed())
		stop = make(chan struct{})
		go func() {
			defer GinkgoRecover()
			err := runtime.Start(stop)
			Expect(err).ToNot(HaveOccurred())
		}()

		// wait for SDS server
		Eventually(func() error {
			conn, err := grpc.Dial(fmt.Sprintf("localhost:%d", port), grpc.WithInsecure())
			client = envoy_discovery.NewSecretDiscoveryServiceClient(conn)
			return err
		}).ShouldNot(HaveOccurred())
	})

	AfterSuite(func() {
		close(stop)
	})

	newRequestForSecrets := func() envoy_api.DiscoveryRequest {
		return envoy_api.DiscoveryRequest{
			Node: &envoy_api_core.Node{
				Id: "default.backend-01",
			},
			ResourceNames: []string{server.MeshCaResource, server.IdentityCertResource},
			TypeUrl:       envoy_resource.SecretType,
		}
	}

	It("should return CA and Identity cert when DP is authorized", func() {
		// given
		ctx := metadata.AppendToOutgoingContext(context.Background(), "authorization", string(dpCredential))
		stream, err := client.StreamSecrets(ctx)
		Expect(err).ToNot(HaveOccurred())
		req := newRequestForSecrets()

		// when
		err = stream.Send(&req)
		Expect(err).ToNot(HaveOccurred())
		resp, err := stream.Recv()
		Expect(err).ToNot(HaveOccurred())

		// then
		Expect(resp).ToNot(BeNil())
		Expect(resp.Resources).To(HaveLen(2))
	})

	It("should return new pair if CA changes", func() {
		By("first exchange of secrets")
		// given
		ctx := metadata.AppendToOutgoingContext(context.Background(), "authorization", string(dpCredential))
		stream, err := client.StreamSecrets(ctx)
		Expect(err).ToNot(HaveOccurred())
		req := newRequestForSecrets()

		// when
		err = stream.Send(&req)
		Expect(err).ToNot(HaveOccurred())
		resp, err := stream.Recv()
		Expect(err).ToNot(HaveOccurred())

		// then
		Expect(resp).ToNot(BeNil())
		Expect(resp.Resources).To(HaveLen(2))

		By("second exchange of secrets")
		// when CA changes
		meshRes := mesh_core.MeshResource{}
		Expect(resManager.Get(context.Background(), &meshRes, core_store.GetByKey("default", "default"))).To(Succeed())
		meshRes.Spec.Mtls.EnabledBackend = "" // we need to first disable mTLS
		Expect(resManager.Update(context.Background(), &meshRes)).To(Succeed())
		meshRes.Spec.Mtls.EnabledBackend = "ca-2"
		Expect(resManager.Update(context.Background(), &meshRes)).To(Succeed())

		// and when send a request with version previously fetched
		req = newRequestForSecrets()
		req.VersionInfo = resp.VersionInfo
		req.ResponseNonce = resp.Nonce
		err = stream.Send(&req)
		Expect(err).ToNot(HaveOccurred())
		resp2, err := stream.Recv()
		Expect(err).ToNot(HaveOccurred())

		// then certs are different
		Expect(resp2).ToNot(BeNil())
		Expect(resp.Resources).ToNot(Equal(resp2.Resources))
	})

	It("should not return certs when DP is not authorized", func() {
		// given
		stream, err := client.StreamSecrets(context.Background())
		Expect(err).ToNot(HaveOccurred())
		req := newRequestForSecrets()

		// when
		err = stream.Send(&req)
		Expect(err).ToNot(HaveOccurred())
		_, err = stream.Recv()

		// then
		Expect(err).To(MatchError("rpc error: code = Unknown desc = could not parse token: token contains an invalid number of segments"))
	})
})
