package bootstrap

import (
	"context"
	mesh_proto "github.com/Kong/konvoy/components/konvoy-control-plane/api/mesh/v1alpha1"
	xds_config "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/config/xds"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/apis/mesh"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/manager"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/store"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/plugins/resources/memory"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/test"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
)

var _ = Describe("Bootstrap Server", func() {

	var stop chan struct{}
	var resManager manager.ResourceManager
	var config *xds_config.XdsBootstrapParamsConfig
	var baseUrl string

	BeforeEach(func() {
		resManager = manager.NewResourceManager(memory.NewStore())
		config = xds_config.DefaultXdsBootstrapParamsConfig()

		port, err := test.GetFreePort()
		baseUrl = "http://localhost:" + strconv.Itoa(port)
		Expect(err).ToNot(HaveOccurred())
		server := BootstrapServer{
			Port:      port,
			Generator: NewDefaultBootstrapGenerator(resManager, config),
		}
		stop = make(chan struct{})
		go func() {
			defer GinkgoRecover()
			err := server.Start(stop)
			Expect(err).ToNot(HaveOccurred())
		}()
		Eventually(func() bool {
			_, err := http.Get(baseUrl)
			return err == nil
		}).Should(BeTrue())
	}, 5)

	AfterEach(func() {
		close(stop)
	})

	BeforeEach(func() {
		err := resManager.Create(context.Background(), &mesh.MeshResource{}, store.CreateByKey("default", "default", "default"))
		Expect(err).ToNot(HaveOccurred())
	})

	It("should return configuration", func() {
		// given
		res := mesh.DataplaneResource{
			Spec: mesh_proto.Dataplane{
				Networking: &mesh_proto.Dataplane_Networking{
					Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
						{
							Interface: "8.8.8.8:443:8443",
							Tags: map[string]string{
								"service": "backend",
							},
						},
					},
				},
			},
		}
		err := resManager.Create(context.Background(), &res, store.CreateByKey("default", "dp-1", "default"))
		Expect(err).ToNot(HaveOccurred())

		// when
		json := `
		{
			"nodeId": "dp-1.default.default"
		}
		`

		resp, err := http.Post(baseUrl+"/bootstrap", "application/json", strings.NewReader(json))

		// then
		Expect(err).ToNot(HaveOccurred())
		received, err := ioutil.ReadAll(resp.Body)
		Expect(err).ToNot(HaveOccurred())

		expected, err := ioutil.ReadFile(filepath.Join("testdata", "bootstrap.golden.yaml"))
		Expect(err).ToNot(HaveOccurred())

		Expect(received).To(MatchYAML(expected))
	})

	It("should return 404 for unknown dataplane", func() {
		// when
		json := `
		{
			"nodeId": "dp-1.default.default"
		}
		`

		resp, err := http.Post(baseUrl+"/bootstrap", "application/json", strings.NewReader(json))
		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(resp.StatusCode).To(Equal(404))
	})
})
