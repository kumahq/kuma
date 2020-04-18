package bootstrap

import (
	"context"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	bootstrap_config "github.com/Kong/kuma/pkg/config/xds/bootstrap"
	"github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	"github.com/Kong/kuma/pkg/core/resources/manager"
	"github.com/Kong/kuma/pkg/core/resources/store"
	"github.com/Kong/kuma/pkg/plugins/resources/memory"
	"github.com/Kong/kuma/pkg/test"
)

var _ = Describe("Bootstrap Server", func() {

	var stop chan struct{}
	var resManager manager.ResourceManager
	var config *bootstrap_config.BootstrapParamsConfig
	var baseUrl string

	BeforeEach(func() {
		resManager = manager.NewResourceManager(memory.NewStore())
		config = bootstrap_config.DefaultBootstrapParamsConfig()
		config.XdsHost = "127.0.0.1"
		config.XdsPort = 5678

		port, err := test.GetFreePort()
		baseUrl = "http://localhost:" + strconv.Itoa(port)
		Expect(err).ToNot(HaveOccurred())
		server := BootstrapServer{
			Port:      uint32(port),
			Generator: NewDefaultBootstrapGenerator(resManager, config, ""),
		}
		stop = make(chan struct{})
		go func() {
			defer GinkgoRecover()
			err := server.Start(stop)
			Expect(err).ToNot(HaveOccurred())
		}()
		Eventually(func() bool {
			resp, err := http.Get(baseUrl)
			if err != nil {
				return false
			}
			Expect(resp.Body.Close()).To(Succeed())
			return true
		}).Should(BeTrue())
	}, 5)

	AfterEach(func() {
		close(stop)
	})

	BeforeEach(func() {
		err := resManager.Create(context.Background(), &mesh.MeshResource{}, store.CreateByKey("default", "default"))
		Expect(err).ToNot(HaveOccurred())
	})

	type testCase struct {
		dataplaneName      string
		body               string
		expectedConfigFile string
	}
	DescribeTable("should return configuration",
		func(given testCase) {
			// given
			res := mesh.DataplaneResource{
				Spec: mesh_proto.Dataplane{
					Networking: &mesh_proto.Dataplane_Networking{
						Address: "8.8.8.8",
						Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
							{
								Port:        443,
								ServicePort: 8443,
								Tags: map[string]string{
									"service": "backend",
								},
							},
						},
					},
				},
			}
			err := resManager.Create(context.Background(), &res, store.CreateByKey(given.dataplaneName, "default"))
			Expect(err).ToNot(HaveOccurred())

			// when
			resp, err := http.Post(baseUrl+"/bootstrap", "application/json", strings.NewReader(given.body))

			// then
			Expect(err).ToNot(HaveOccurred())
			received, err := ioutil.ReadAll(resp.Body)
			Expect(resp.Body.Close()).To(Succeed())
			Expect(err).ToNot(HaveOccurred())

			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			expected, err := ioutil.ReadFile(filepath.Join("testdata", given.expectedConfigFile))
			Expect(err).ToNot(HaveOccurred())

			Expect(received).To(MatchYAML(expected))
		},
		Entry("minimal data provided (universal)", testCase{
			dataplaneName:      "dp-1",
			body:               `{ "mesh": "default", "name": "dp-1" }`,
			expectedConfigFile: "bootstrap.universal.golden.yaml",
		}),
		Entry("minimal data provided (k8s)", testCase{
			dataplaneName:      "dp-1.default",
			body:               `{ "mesh": "default", "name": "dp-1.default" }`,
			expectedConfigFile: "bootstrap.k8s.golden.yaml",
		}),
		Entry("full data provided", testCase{
			dataplaneName:      "dp-1.default",
			body:               `{ "mesh": "default", "name": "dp-1.default", "adminPort": 1234, "dataplaneTokenPath": "/tmp/token" }`,
			expectedConfigFile: "bootstrap.overridden.golden.yaml",
		}),
	)

	It("should return 404 for unknown dataplane", func() {
		// when
		json := `
		{
			"mesh": "default",
			"name": "dp-1.default"
		}
		`

		resp, err := http.Post(baseUrl+"/bootstrap", "application/json", strings.NewReader(json))
		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(resp.Body.Close()).To(Succeed())
		Expect(resp.StatusCode).To(Equal(404))
	})
})
