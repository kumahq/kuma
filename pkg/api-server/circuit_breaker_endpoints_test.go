package api_server_test

import (
	"context"
	"io/ioutil"
	"time"

	"github.com/ghodss/yaml"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	api_server "github.com/Kong/kuma/pkg/api-server"
	config "github.com/Kong/kuma/pkg/config/api-server"
	"github.com/Kong/kuma/pkg/core"
	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	"github.com/Kong/kuma/pkg/core/resources/model/rest"
	"github.com/Kong/kuma/pkg/core/resources/store"
	"github.com/Kong/kuma/pkg/plugins/resources/memory"
)

var _ = Describe("CircuitBreaker Endpoints", func() {
	var apiServer *api_server.ApiServer
	var resourceStore store.ResourceStore
	var client resourceApiClient
	var stop chan struct{}

	BeforeEach(func() {
		core.Now = func() time.Time {
			now, _ := time.Parse(time.RFC3339, "2018-07-17T16:05:36.995+00:00")
			return now
		}
		resourceStore = memory.NewStore()
		apiServer = createTestApiServer(resourceStore, config.DefaultApiServerConfig())
		client = resourceApiClient{
			apiServer.Address(),
			"/meshes/default/circuit-breakers",
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
		core.Now = time.Now
	})

	BeforeEach(func() {
		// when
		err := resourceStore.Create(context.Background(), &mesh_core.MeshResource{}, store.CreateByKey("default", "default"))
		// then
		Expect(err).ToNot(HaveOccurred())
	})

	Describe("PUT => GET", func() {

		given := `
        type: CircuitBreaker
        name: web-to-backend
        mesh: default
        creationTime: "2018-07-17T16:05:36.995Z"
        modificationTime: "2018-07-17T16:05:36.995Z"
        sources:
        - match:
            service: web
            protocol: http
        destinations:
        - match:
            service: backend
        conf:
          baseEjectionTime: 5s
          detectors:
            gatewayErrors: 
              consecutive: 10
            localErrors: 
              consecutive: 5
            totalErrors: 
              consecutive: 20
            failure:
              minimumHosts: 3
              requestVolume: 20
              threshold: 85
            standardDeviation:
              factor: 1.9
              minimumHosts: 3
              requestVolume: 20
          interval: 5s
          maxEjectionPercent: 50
`
		It("GET should return data saved by PUT", func() {
			// given
			resource := rest.Resource{
				Spec: &mesh_proto.CircuitBreaker{},
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
