package bootstrap_test

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	dp_server_cfg "github.com/kumahq/kuma/pkg/config/dp-server"
	bootstrap_config "github.com/kumahq/kuma/pkg/config/xds/bootstrap"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/dp-server/server"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/pkg/test/matchers"
	test_metrics "github.com/kumahq/kuma/pkg/test/metrics"
	"github.com/kumahq/kuma/pkg/xds/bootstrap"
)

var _ = Describe("Bootstrap Server", func() {

	var stop chan struct{}
	var resManager manager.ResourceManager
	var baseUrl string
	var metrics core_metrics.Metrics

	core.TempDir = func() string {
		return "/tmp"
	}

	version := `
	"version": {
	  "kumaDp": {
	    "version": "0.0.1",
	    "gitTag": "v0.0.1",
	    "gitCommit": "91ce236824a9d875601679aa80c63783fb0e8725",
	    "buildDate": "2019-08-07T11:26:06Z"
	  },
	  "envoy": {
	    "version": "1.15.0",
	    "build": "hash/1.15.0/RELEASE"
	  }
	}`

	httpClient := &http.Client{
		Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
	}

	BeforeEach(func() {
		resManager = manager.NewResourceManager(memory.NewStore())
		config := bootstrap_config.DefaultBootstrapServerConfig()
		config.Params.XdsHost = "localhost"
		config.Params.XdsPort = 5678

		port, err := test.GetFreePort()
		baseUrl = "https://localhost:" + strconv.Itoa(port)
		Expect(err).ToNot(HaveOccurred())
		metrics, err = core_metrics.NewMetrics("Standalone")
		Expect(err).ToNot(HaveOccurred())

		dpServerCfg := dp_server_cfg.DpServerConfig{
			Port:        port,
			TlsCertFile: filepath.Join("..", "..", "..", "test", "certs", "server-cert.pem"),
			TlsKeyFile:  filepath.Join("..", "..", "..", "test", "certs", "server-key.pem"),
		}
		dpServer := server.NewDpServer(dpServerCfg, metrics)

		generator, err := bootstrap.NewDefaultBootstrapGenerator(resManager, config, filepath.Join("..", "..", "..", "test", "certs", "server-cert.pem"), true, true, 0)
		Expect(err).ToNot(HaveOccurred())
		bootstrapHandler := bootstrap.BootstrapHandler{
			Generator: generator,
		}
		dpServer.HTTPMux().HandleFunc("/bootstrap", bootstrapHandler.Handle)

		stop = make(chan struct{})
		go func() {
			defer GinkgoRecover()
			err := dpServer.Start(stop)
			Expect(err).ToNot(HaveOccurred())
		}()
		Eventually(func() bool {
			resp, err := httpClient.Get(baseUrl)
			if err != nil {
				return false
			}
			Expect(resp.Body.Close()).To(Succeed())
			return true
		}, 5).Should(BeTrue())
	})

	AfterEach(func() {
		close(stop)
	})

	BeforeEach(func() {
		err := resManager.Create(context.Background(), mesh.NewMeshResource(), store.CreateByKey(model.DefaultMesh, model.NoMesh))
		Expect(err).ToNot(HaveOccurred())
		core.Now = func() time.Time {
			now, _ := time.Parse(time.RFC3339, "2018-07-17T16:05:36.995+00:00")
			return now
		}
	})

	defaultDataplane := func() *mesh.DataplaneResource {
		return &mesh.DataplaneResource{
			Spec: &mesh_proto.Dataplane{
				Networking: &mesh_proto.Dataplane_Networking{
					Address: "8.8.8.8",
					Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
						{
							Port:        443,
							ServicePort: 8443,
							Tags: map[string]string{
								"kuma.io/service": "backend",
							},
						},
					},
					Admin: &mesh_proto.EnvoyAdmin{},
				},
			},
		}
	}

	type testCase struct {
		dataplaneName      string
		dataplane          func() *mesh.DataplaneResource
		body               string
		expectedConfigFile string
	}
	DescribeTable("should return configuration",
		func(given testCase) {
			// given
			err := resManager.Create(context.Background(), given.dataplane(), store.CreateByKey(given.dataplaneName, "default"))
			Expect(err).ToNot(HaveOccurred())

			// when
			resp, err := httpClient.Post(baseUrl+"/bootstrap", "application/json", strings.NewReader(given.body))

			// then
			Expect(err).ToNot(HaveOccurred())
			received, err := io.ReadAll(resp.Body)
			Expect(resp.Body.Close()).To(Succeed())
			Expect(err).ToNot(HaveOccurred())

			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(received).To(matchers.MatchGoldenYAML(filepath.Join("testdata", given.expectedConfigFile)))
		},
		Entry("minimal data provided (universal)", testCase{
			dataplaneName:      "dp-1",
			dataplane:          defaultDataplane,
			body:               fmt.Sprintf(`{ "mesh": "default", "name": "dp-1", "dataplaneToken": "token", %s }`, version),
			expectedConfigFile: "bootstrap.universal.golden.yaml",
		}),
		Entry("minimal data provided (k8s)", testCase{
			dataplaneName:      "dp-1.default",
			dataplane:          defaultDataplane,
			body:               fmt.Sprintf(`{ "mesh": "default", "name": "dp-1.default", "dataplaneToken": "token", %s }`, version),
			expectedConfigFile: "bootstrap.k8s.golden.yaml",
		}),
		Entry("full data provided", testCase{
			dataplaneName: "dp-1.default",
			dataplane: func() *mesh.DataplaneResource {
				dp := defaultDataplane()
				dp.Spec.Networking.Admin.Port = 1234
				return dp
			},
			body:               fmt.Sprintf(`{ "mesh": "default", "name": "dp-1.default", "dataplaneToken": "token", %s }`, version),
			expectedConfigFile: "bootstrap.overridden.golden.yaml",
		}),
	)

	It("should return 404 for unknown dataplane", func() {
		// when
		json := `
		{
			"mesh": "default",
			"name": "dp-1.default",
			"dataplaneToken": "token"
		}
		`

		resp, err := httpClient.Post(baseUrl+"/bootstrap", "application/json", strings.NewReader(json))
		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(resp.Body.Close()).To(Succeed())
		Expect(resp.StatusCode).To(Equal(404))
	})

	It("should return 422 for the lack of the dataplane token", func() {
		// given
		res := mesh.DataplaneResource{
			Spec: &mesh_proto.Dataplane{
				Networking: &mesh_proto.Dataplane_Networking{
					Address: "8.8.8.8",
					Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
						{
							Port:        443,
							ServicePort: 8443,
							Tags: map[string]string{
								"kuma.io/service": "backend",
							},
						},
					},
				},
			},
		}
		err := resManager.Create(context.Background(), &res, store.CreateByKey("dp-1", "default"))
		Expect(err).ToNot(HaveOccurred())

		// when
		json := `
		{
			"mesh": "default",
			"name": "dp-1"
		}
		`

		resp, err := httpClient.Post(baseUrl+"/bootstrap", "application/json", strings.NewReader(json))
		// then
		Expect(err).ToNot(HaveOccurred())
		bytes, err := io.ReadAll(resp.Body)
		Expect(err).ToNot(HaveOccurred())
		Expect(resp.Body.Close()).To(Succeed())
		Expect(resp.StatusCode).To(Equal(422))
		Expect(string(bytes)).To(Equal("Dataplane Token is required. Generate token using 'kumactl generate dataplane-token > /path/file' and provide it via --dataplane-token-file=/path/file argument to Kuma DP"))

	})

	It("should publish metrics", func() {
		// given
		res := mesh.DataplaneResource{
			Spec: &mesh_proto.Dataplane{
				Networking: &mesh_proto.Dataplane_Networking{
					Address: "8.8.8.8",
					Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
						{
							Port:        443,
							ServicePort: 8443,
							Tags: map[string]string{
								"kuma.io/service": "backend",
							},
						},
					},
				},
			},
		}
		err := resManager.Create(context.Background(), &res, store.CreateByKey("dp-1", "default"))
		Expect(err).ToNot(HaveOccurred())

		// when
		_, err = httpClient.Post(baseUrl+"/bootstrap", "application/json", strings.NewReader(`{ "mesh": "default", "name": "dp-1" }`))

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(test_metrics.FindMetric(metrics, "dp_server_http_request_duration_seconds", "handler", "/bootstrap")).ToNot(BeNil())
		Expect(test_metrics.FindMetric(metrics, "dp_server_http_requests_inflight", "handler", "/bootstrap")).ToNot(BeNil())
		Expect(test_metrics.FindMetric(metrics, "dp_server_http_response_size_bytes", "handler", "/bootstrap")).ToNot(BeNil())
	})
})
