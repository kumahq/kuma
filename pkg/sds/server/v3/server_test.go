package v3_test

import (
	"context"
	"fmt"
	"path/filepath"
	"sync/atomic"
	"time"

	envoy_api_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	envoy_service_secret "github.com/envoyproxy/go-control-plane/envoy/service/secret/v3"
	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	prometheus_client "github.com/prometheus/client_model/go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	dp_server_cfg "github.com/kumahq/kuma/pkg/config/dp-server"
	"github.com/kumahq/kuma/pkg/core"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	dp_server "github.com/kumahq/kuma/pkg/dp-server"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
	sds_server "github.com/kumahq/kuma/pkg/sds/server"
	"github.com/kumahq/kuma/pkg/test"
	test_metrics "github.com/kumahq/kuma/pkg/test/metrics"
	"github.com/kumahq/kuma/pkg/test/runtime"
	tokens_builtin "github.com/kumahq/kuma/pkg/tokens/builtin"
	tokens_issuer "github.com/kumahq/kuma/pkg/tokens/builtin/issuer"
	"github.com/kumahq/kuma/pkg/xds/envoy/tls"
)

var _ = Describe("SDS Server", func() {

	var dpCredential tokens_issuer.Token
	var stop chan struct{}
	var client envoy_service_secret.SecretDiscoveryServiceClient
	var conn *grpc.ClientConn
	var metrics core_metrics.Metrics

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
		cfg.DpServer.Port = port
		cfg.DpServer.TlsCertFile = filepath.Join("..", "..", "..", "..", "test", "certs", "server-cert.pem")
		cfg.DpServer.TlsKeyFile = filepath.Join("..", "..", "..", "..", "test", "certs", "server-key.pem")
		cfg.DpServer.Auth.Type = dp_server_cfg.DpServerAuthDpToken

		builder, err := runtime.BuilderFor(cfg)
		Expect(err).ToNot(HaveOccurred())
		runtime, err := builder.Build()
		Expect(err).ToNot(HaveOccurred())
		metrics = runtime.Metrics()
		resManager = runtime.ResourceManager()

		// setup default mesh with active mTLS and 2 CA
		meshRes := core_mesh.MeshResource{
			Spec: &mesh_proto.Mesh{
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
		err = resManager.Create(context.Background(), &meshRes, core_store.CreateByKey(model.DefaultMesh, model.NoMesh))
		Expect(err).ToNot(HaveOccurred())

		// setup backend dataplane
		dpRes := core_mesh.DataplaneResource{
			Spec: &mesh_proto.Dataplane{
				Networking: &mesh_proto.Dataplane_Networking{
					Address: "192.168.0.1",
					Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
						{
							Port: 1234,
							Tags: map[string]string{
								"kuma.io/service": "backend",
							},
						},
					},
				},
			},
		}
		err = resManager.Create(context.Background(), &dpRes, core_store.CreateByKey("backend-01", "default"))
		Expect(err).ToNot(HaveOccurred())

		// retrieve example DP token
		tokenIssuer, err := tokens_builtin.NewDataplaneTokenIssuer(runtime.ReadOnlyResourceManager())
		Expect(err).ToNot(HaveOccurred())
		dpCredential, err = tokenIssuer.Generate(tokens_issuer.DataplaneIdentity{
			Name: dpRes.GetMeta().GetName(),
			Mesh: dpRes.GetMeta().GetMesh(),
		})
		Expect(err).ToNot(HaveOccurred())

		// start the runtime
		Expect(sds_server.Setup(runtime)).To(Succeed())
		Expect(dp_server.SetupServer(runtime)).To(Succeed())
		stop = make(chan struct{})
		go func() {
			defer GinkgoRecover()
			err := runtime.Start(stop)
			Expect(err).ToNot(HaveOccurred())
		}()

		// wait for SDS server
		Eventually(func() error {
			creds, err := credentials.NewClientTLSFromFile(filepath.Join("..", "..", "..", "..", "test", "certs", "server-cert.pem"), "")
			if err != nil {
				return err
			}
			c, err := grpc.Dial(fmt.Sprintf("localhost:%d", port), grpc.WithTransportCredentials(creds))
			if err != nil {
				return err
			}
			conn = c
			client = envoy_service_secret.NewSecretDiscoveryServiceClient(conn)
			_, err = client.StreamSecrets(context.Background()) // dial is not enough, we need to double check if we can start to stream secrets
			return err
		}).ShouldNot(HaveOccurred())
	})

	AfterSuite(func() {
		if conn != nil {
			Expect(conn.Close()).To(Succeed())
		}
		close(stop)
	})

	BeforeEach(func() {
		// make sure no insight is present
		err := resManager.Delete(context.Background(), core_mesh.NewDataplaneInsightResource(), core_store.DeleteByKey("backend-01", "default"))
		if !core_store.IsResourceNotFound(err) {
			Expect(err).ToNot(HaveOccurred())
		}
		// and no watchdog is running
		Eventually(func() float64 {
			return test_metrics.FindMetric(metrics, "sds_watchdogs").GetGauge().GetValue()
		}, "30s").Should(Equal(0.0))
	})

	newRequestForSecrets := func() envoy_discovery.DiscoveryRequest {
		return envoy_discovery.DiscoveryRequest{
			Node: &envoy_api_core.Node{
				Id: "default.backend-01",
			},
			ResourceNames: []string{tls.MeshCaResource, tls.IdentityCertResource},
			TypeUrl:       envoy_resource.SecretType,
		}
	}

	It("should return CA and Identity cert when DP is authorized", test.Within(time.Minute, func() {
		// given
		ctx := metadata.AppendToOutgoingContext(context.Background(), "authorization", dpCredential)
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

		// and insight is generated (insight is updated async, to does not have to be done before response is sent)
		Eventually(func() error {
			dpInsight := core_mesh.NewDataplaneInsightResource()
			err := resManager.Get(context.Background(), dpInsight, core_store.GetByKey("backend-01", "default"))
			if err != nil {
				return err
			}
			if dpInsight.Spec.MTLS.CertificateRegenerations != 1 {
				return errors.Errorf("Certs were generated %d times. Expected 1", dpInsight.Spec.MTLS.CertificateRegenerations)
			}
			expirationSeconds := now.Load().(time.Time).Add(60 * time.Second).Unix()
			if dpInsight.Spec.MTLS.CertificateExpirationTime.Seconds != expirationSeconds {
				return errors.Errorf("Expiration time is not correct. Got %d, expected %d", dpInsight.Spec.MTLS.CertificateExpirationTime.Seconds, expirationSeconds)
			}
			return nil
		}, "30s", "1s").ShouldNot(HaveOccurred())

		// and metrics are published (metrics are published async, it does not have to be done before response is sent)
		Eventually(func() float64 {
			return test_metrics.FindMetric(metrics, "sds_cert_generation").GetCounter().GetValue()
		}, "5s").Should(Equal(1.0))
		Eventually(func() *prometheus_client.Metric {
			return test_metrics.FindMetric(metrics, "sds_generation")
		}, "5s").ShouldNot(BeNil())
		Eventually(func() float64 {
			return test_metrics.FindMetric(metrics, "sds_watchdogs").GetGauge().GetValue()
		}, "5s").Should(Equal(1.0))

	}))

	Context("should return new pair of + key", func() { // we cannot use DescribeTable because it does not support timeouts

		var firstExchangeResponse *envoy_discovery.DiscoveryResponse
		var stream envoy_service_secret.SecretDiscoveryService_StreamSecretsClient

		BeforeEach(test.Within(time.Minute, func() {
			// given
			ctx := metadata.AppendToOutgoingContext(context.Background(), "authorization", dpCredential)
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

			// and wait for the insight
			Eventually(func() error {
				return resManager.Get(context.Background(), core_mesh.NewDataplaneInsightResource(), core_store.GetByKey("backend-01", "default"))
			}, "30s", "1s").ShouldNot(HaveOccurred())
		}))

		AfterEach(func() {
			Expect(stream.CloseSend()).To(Succeed())
		})

		It("should return pair when CA is changed", test.Within(time.Minute, func() {
			// when
			meshRes := core_mesh.NewMeshResource()
			Expect(resManager.Get(context.Background(), meshRes, core_store.GetByKey(model.DefaultMesh, model.NoMesh))).To(Succeed())
			meshRes.Spec.Mtls.EnabledBackend = "" // we need to first disable mTLS
			Expect(resManager.Update(context.Background(), meshRes)).To(Succeed())
			meshRes.Spec.Mtls.EnabledBackend = "ca-2"
			Expect(resManager.Update(context.Background(), meshRes)).To(Succeed())

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

			// and insight is generated (insight is updated async, to does not have to be done before response is sent)
			Eventually(func() error {
				dpInsight := core_mesh.NewDataplaneInsightResource()
				err := resManager.Get(context.Background(), dpInsight, core_store.GetByKey("backend-01", "default"))
				if err != nil {
					return err
				}
				if dpInsight.Spec.MTLS.CertificateRegenerations != 2 {
					return errors.Errorf("Certs were generated %d times. Expected 2", dpInsight.Spec.MTLS.CertificateRegenerations)
				}
				return nil
			}, "30s", "1s").ShouldNot(HaveOccurred())
		}))

		It("should return a new pair when cert expired", test.Within(time.Minute, func() {
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

		}))

		It("should return a new pair when dataplane has changed", test.Within(time.Minute, func() {
			// when
			dpRes := core_mesh.NewDataplaneResource()
			Expect(resManager.Get(context.Background(), dpRes, core_store.GetByKey("backend-01", "default"))).To(Succeed())
			dpRes.Spec.Networking.Inbound[0].Tags["version"] = "xyz"

			// when new tag is added
			Expect(resManager.Update(context.Background(), dpRes)).To(Succeed())

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

		}))
	})

	It("should not return certs when DP is not authorized", test.Within(time.Minute, func() {
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
		Expect(err).To(MatchError("rpc error: code = Unknown desc = authentication failed: could not parse token: token contains an invalid number of segments"))
	}))
})
