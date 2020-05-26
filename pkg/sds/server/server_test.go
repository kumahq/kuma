package server_test

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	envoy_api "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_api_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	envoy_discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v2"
	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	kuma_cp "github.com/Kong/kuma/pkg/config/app/kuma-cp"
	"github.com/Kong/kuma/pkg/core"
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
	var conn *grpc.ClientConn

	var resManager core_manager.ResourceManager

	var now atomic.Value // it has to be stored as atomic to avoid race condition

	BeforeSuite(func() {
		now.Store(time.Now())
		core.Now = func() time.Time {
			return now.Load().(time.Time)
		}
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
							DpCert: &mesh_proto.CertificateAuthorityBackend_DpCert{
								Rotation: &mesh_proto.CertificateAuthorityBackend_DpCert_Rotation{
									Expiration: "1m",
								},
							},
						},
						{
							Name: "ca-2",
							Type: "builtin",
							DpCert: &mesh_proto.CertificateAuthorityBackend_DpCert{
								Rotation: &mesh_proto.CertificateAuthorityBackend_DpCert_Rotation{
									Expiration: "1m",
								},
							},
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
			c, err := grpc.Dial(fmt.Sprintf("localhost:%d", port), grpc.WithInsecure())
			conn = c
			client = envoy_discovery.NewSecretDiscoveryServiceClient(conn)
			return err
		}).ShouldNot(HaveOccurred())
	})

	AfterSuite(func() {
		if conn != nil {
			Expect(conn.Close()).To(Succeed())
		}
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

	It("should return CA and Identity cert when DP is authorized", func(done Done) {
		// given
		ctx := metadata.AppendToOutgoingContext(context.Background(), "authorization", string(dpCredential))
		stream, err := client.StreamSecrets(ctx)
		defer func() {
			if stream != nil {
				Expect(stream.CloseSend()).To(Succeed())
			}
		}()
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

		// and insight is generated
		dpInsight := mesh_core.DataplaneInsightResource{}
		err = resManager.Get(context.Background(), &dpInsight, core_store.GetByKey("backend-01", "default"))
		Expect(err).ToNot(HaveOccurred())
		Expect(dpInsight.Spec.MTLS.CertificateRegenerations).To(Equal(uint32(1)))
		expirationSeconds := now.Load().(time.Time).Add(60 * time.Second).Unix()
		Expect(dpInsight.Spec.MTLS.CertificateExpirationTime.Seconds).To(Equal(expirationSeconds))

		close(done)
	}, 10)

	Context("should return new pair of + key", func() { // we cannot use DescribeTable because it does not support timeouts

		var firstExchangeResponse *envoy_api.DiscoveryResponse
		var stream envoy_discovery.SecretDiscoveryService_StreamSecretsClient

		BeforeEach(func(done Done) {
			// given
			ctx := metadata.AppendToOutgoingContext(context.Background(), "authorization", string(dpCredential))
			s, err := client.StreamSecrets(ctx)
			stream = s
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
			firstExchangeResponse = resp
			close(done)
		}, 10)

		AfterEach(func() {
			Expect(stream.CloseSend()).To(Succeed())
		})

		It("should return pair when CA is changed", func(done Done) {
			// when
			meshRes := mesh_core.MeshResource{}
			Expect(resManager.Get(context.Background(), &meshRes, core_store.GetByKey("default", "default"))).To(Succeed())
			meshRes.Spec.Mtls.EnabledBackend = "" // we need to first disable mTLS
			Expect(resManager.Update(context.Background(), &meshRes)).To(Succeed())
			meshRes.Spec.Mtls.EnabledBackend = "ca-2"
			Expect(resManager.Update(context.Background(), &meshRes)).To(Succeed())

			// and when send a request with version previously fetched
			req := newRequestForSecrets()
			req.VersionInfo = firstExchangeResponse.VersionInfo
			req.ResponseNonce = firstExchangeResponse.Nonce
			err := stream.Send(&req)
			Expect(err).ToNot(HaveOccurred())
			resp, err := stream.Recv()
			Expect(err).ToNot(HaveOccurred())

			// then certs are different
			Expect(resp).ToNot(BeNil())
			Expect(firstExchangeResponse.Resources).ToNot(Equal(resp.Resources))

			// and insight is updated
			dpInsight := mesh_core.DataplaneInsightResource{}
			err = resManager.Get(context.Background(), &dpInsight, core_store.GetByKey("backend-01", "default"))
			Expect(err).ToNot(HaveOccurred())
			Expect(dpInsight.Spec.MTLS.CertificateRegenerations).To(Equal(uint32(2)))
			expirationSeconds := now.Load().(time.Time).Add(60 * time.Second).Unix()
			Expect(dpInsight.Spec.MTLS.CertificateExpirationTime.Seconds).To(Equal(expirationSeconds))

			close(done)
		}, 10)

		It("should return pair when cert expired", func(done Done) {
			// when time is moved 1s after 4/5 of 60s cert expiration
			shiftedTime := now.Load().(time.Time).Add(49 * time.Second)
			now.Store(shiftedTime)

			// and when send a request with version previously fetched
			req := newRequestForSecrets()
			req.VersionInfo = firstExchangeResponse.VersionInfo
			req.ResponseNonce = firstExchangeResponse.Nonce
			err := stream.Send(&req)
			Expect(err).ToNot(HaveOccurred())
			resp, err := stream.Recv()
			Expect(err).ToNot(HaveOccurred())

			// then certs are different
			Expect(resp).ToNot(BeNil())
			Expect(firstExchangeResponse.Resources).ToNot(Equal(resp.Resources))

			close(done)
		}, 10)
	})

	It("should not return certs when DP is not authorized", func(done Done) {
		// given
		stream, err := client.StreamSecrets(context.Background())
		defer func() {
			if stream != nil {
				Expect(stream.CloseSend()).To(Succeed())
			}
		}()
		Expect(err).ToNot(HaveOccurred())
		req := newRequestForSecrets()

		// when
		err = stream.Send(&req)
		Expect(err).ToNot(HaveOccurred())
		_, err = stream.Recv()

		// then
		Expect(err).To(MatchError("rpc error: code = Unknown desc = could not parse token: token contains an invalid number of segments"))

		close(done)
	}, 10)
})
