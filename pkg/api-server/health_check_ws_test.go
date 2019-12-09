package api_server_test

import (
	"context"
	"io/ioutil"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/ghodss/yaml"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	api_server "github.com/Kong/kuma/pkg/api-server"
	config "github.com/Kong/kuma/pkg/config/api-server"
	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	"github.com/Kong/kuma/pkg/core/resources/model/rest"
	"github.com/Kong/kuma/pkg/core/resources/store"
	"github.com/Kong/kuma/pkg/plugins/resources/memory"
)

var _ = Describe("HealthCheck WS", func() {
	var apiServer *api_server.ApiServer
	var resourceStore store.ResourceStore
	var client resourceApiClient
	var stop chan struct{}

	BeforeEach(func() {
		resourceStore = memory.NewStore()
		apiServer = createTestApiServer(resourceStore, config.DefaultApiServerConfig())
		client = resourceApiClient{
			apiServer.Address(),
			"/meshes/default/health-checks",
		}
		stop = make(chan struct{})
		go func() {
			defer GinkgoRecover()
			err := apiServer.Start(stop)
			Expect(err).ToNot(HaveOccurred())
		}()
		waitForServer(&client)
	}, 5)

	AfterEach(func() {
		close(stop)
	})

	BeforeEach(func() {
		// when
		err := resourceStore.Create(context.Background(), &mesh_core.MeshResource{}, store.CreateByKey("default", "default"))
		// then
		Expect(err).ToNot(HaveOccurred())
	})

	Describe("PUT => GET", func() {

		given := `
        type: HealthCheck
        name: web-to-backend
        mesh: default
        sources:
        - match:
            service: web
        destinations:
        - match:
            service: backend
        conf:
          activeChecks:
            interval: 10s
            timeout: 2s
            unhealthyThreshold: 3
            healthyThreshold: 1
          passiveChecks:
            unhealthyThreshold: 3
            penaltyInterval: 5s
`
		It("GET should return data saved by PUT", func() {
			// given
			resource := rest.Resource{
				Spec: &mesh_proto.HealthCheck{},
			}

			// when
			err := yaml.Unmarshal([]byte(given), &resource)
			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			response := client.put(resource)
			// then
			Expect(response.StatusCode).To(Equal(201))

			// when
			response = client.get("web-to-backend")
			// then
			Expect(response.StatusCode).To(Equal(200))
			// when
			body, err := ioutil.ReadAll(response.Body)
			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			actual, err := yaml.JSONToYAML(body)
			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(actual).To(MatchYAML(given))
		})
	})
})
