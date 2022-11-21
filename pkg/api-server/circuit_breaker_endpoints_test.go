package api_server_test

import (
	"context"
	"io"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/yaml"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	api_server "github.com/kumahq/kuma/pkg/api-server"
	"github.com/kumahq/kuma/pkg/core"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/model/rest/unversioned"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
)

var _ = Describe("CircuitBreaker Endpoints", func() {
	var apiServer *api_server.ApiServer
	var resourceStore store.ResourceStore
	var client resourceApiClient

	var stop = func() {}
	BeforeEach(func() {
		core.Now = func() time.Time {
			now, _ := time.Parse(time.RFC3339, "2018-07-17T16:05:36.995+00:00")
			return now
		}
		resourceStore = memory.NewStore()
		apiServer, _, stop = StartApiServer(NewTestApiServerConfigurer().WithStore(resourceStore))
		client = resourceApiClient{
			apiServer.Address(),
			"/meshes/default/circuit-breakers",
		}
	})

	AfterEach(func() {
		stop()
		core.Now = time.Now
	})

	BeforeEach(func() {
		// when
		err := resourceStore.Create(context.Background(), core_mesh.NewMeshResource(), store.CreateByKey(model.DefaultMesh, model.NoMesh))
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
            kuma.io/service: web
            kuma.io/protocol: http
        destinations:
        - match:
            kuma.io/service: backend
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
			resource := &unversioned.Resource{
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
			body, err := io.ReadAll(response.Body)
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
