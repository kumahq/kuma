package api_server_test

import (
	"context"
	"io"
	"net/http"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	api_server "github.com/kumahq/kuma/pkg/api-server"
	config "github.com/kumahq/kuma/pkg/config/api-server"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/metrics"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
)

var _ = Describe("Service Insight Endpoints", func() {
	var apiServer *api_server.ApiServer
	var resourceStore store.ResourceStore
	var stop chan struct{}
	t1, _ := time.Parse(time.RFC3339, "2018-07-17T16:05:36.995+00:00")
	BeforeEach(func() {
		resourceStore = memory.NewStore()
		metrics, err := metrics.NewMetrics("Standalone")
		Expect(err).ToNot(HaveOccurred())
		apiServer = createTestApiServer(resourceStore, config.DefaultApiServerConfig(), true, metrics)
		client := resourceApiClient{
			address: apiServer.Address(),
			path:    "/meshes",
		}
		stop = make(chan struct{})
		go func() {
			defer GinkgoRecover()
			err := apiServer.Start(stop)
			Expect(err).ToNot(HaveOccurred())
		}()
		waitForServer(&client)
	})

	AfterEach(func() {
		close(stop)
	})

	BeforeEach(func() {
		err := resourceStore.Create(context.Background(), core_mesh.NewMeshResource(), store.CreateByKey("mesh-1", core_model.NoMesh), store.CreatedAt(t1))
		Expect(err).ToNot(HaveOccurred())

		err = resourceStore.Create(context.Background(), core_mesh.NewMeshResource(), store.CreateByKey("mesh-2", core_model.NoMesh), store.CreatedAt(t1))
		Expect(err).ToNot(HaveOccurred())
	})

	createServiceInsight := func(name, mesh string, serviceInsight *mesh_proto.ServiceInsight) {
		serviceInsightResource := core_mesh.ServiceInsightResource{
			Spec: serviceInsight,
		}
		err := resourceStore.Create(context.Background(), &serviceInsightResource, store.CreateByKey(name, mesh), store.CreatedAt(t1))
		Expect(err).ToNot(HaveOccurred())
	}

	BeforeEach(func() {
		createServiceInsight("all-services-mesh-1", "mesh-1", &mesh_proto.ServiceInsight{
			Services: map[string]*mesh_proto.ServiceInsight_Service{
				"backend": {
					Status: mesh_proto.ServiceInsight_Service_partially_degraded,
					Dataplanes: &mesh_proto.ServiceInsight_Service_DataplaneStat{
						Total:   100,
						Online:  70,
						Offline: 30,
					},
				},
				"frontend": {
					Status: mesh_proto.ServiceInsight_Service_partially_degraded,
					Dataplanes: &mesh_proto.ServiceInsight_Service_DataplaneStat{
						Total:   20,
						Online:  19,
						Offline: 1,
					},
				},
			},
		})

		createServiceInsight("all-services-mesh-2", "mesh-2", &mesh_proto.ServiceInsight{
			Services: map[string]*mesh_proto.ServiceInsight_Service{
				"db": {
					Status: mesh_proto.ServiceInsight_Service_partially_degraded,
					Dataplanes: &mesh_proto.ServiceInsight_Service_DataplaneStat{
						Total:   10,
						Online:  9,
						Offline: 1,
					},
				},
				"redis": {
					Status: mesh_proto.ServiceInsight_Service_partially_degraded,
					Dataplanes: &mesh_proto.ServiceInsight_Service_DataplaneStat{
						Total:   22,
						Online:  19,
						Offline: 3,
					},
				},
			},
		})
	})

	Describe("On GET", func() {
		It("should return an existing resource", func() {
			expected := `
{
  "total": 2,
  "items": [
	{
	  "type": "ServiceInsight",
	  "mesh": "mesh-1",
	  "name": "backend",
	  "creationTime": "2018-07-17T16:05:36.995Z",
	  "modificationTime": "2018-07-17T16:05:36.995Z",
	  "status": "partially_degraded",
	  "dataplanes": {
	  	"total": 100,
	  	"online": 70,
	  	"offline": 30
      }
	},
	{
	  "type": "ServiceInsight",
	  "mesh": "mesh-1",
	  "name": "frontend",
	  "creationTime": "2018-07-17T16:05:36.995Z",
	  "modificationTime": "2018-07-17T16:05:36.995Z",
      "status": "partially_degraded",
      "dataplanes": {
        "total": 20,
        "online": 19,
        "offline": 1
      }
	}
  ],
  "next": null
}`

			// when
			response, err := http.Get("http://" + apiServer.Address() + "/meshes/mesh-1/service-insights")
			Expect(err).ToNot(HaveOccurred())

			// then
			Expect(response.StatusCode).To(Equal(200))
			body, err := io.ReadAll(response.Body)
			Expect(err).ToNot(HaveOccurred())
			Expect(body).To(MatchJSON(expected))
		})

		It("should return stat for specific service", func() {
			expected := `
{
  "type": "ServiceInsight",
  "mesh": "mesh-1",
  "name": "backend",
  "creationTime": "2018-07-17T16:05:36.995Z",
  "modificationTime": "2018-07-17T16:05:36.995Z",
  "status": "partially_degraded",
  "dataplanes": {
    "total": 100,
    "online": 70,
    "offline": 30
  }
}`

			// when
			response, err := http.Get("http://" + apiServer.Address() + "/meshes/mesh-1/service-insights/backend")
			Expect(err).ToNot(HaveOccurred())

			// then
			Expect(response.StatusCode).To(Equal(200))
			body, err := io.ReadAll(response.Body)
			Expect(err).ToNot(HaveOccurred())
			Expect(body).To(MatchJSON(expected))
		})
	})

	type testCase struct {
		params   string
		expected string
	}
	DescribeTable("should paginate a list",
		func(given testCase) {
			// when
			response, err := http.Get("http://" + apiServer.Address() + "/service-insights" + given.params)
			Expect(err).ToNot(HaveOccurred())

			// then
			Expect(response.StatusCode).To(Equal(200))
			body, err := io.ReadAll(response.Body)
			Expect(err).ToNot(HaveOccurred())

			expected := strings.ReplaceAll(given.expected, "{{address}}", apiServer.Address())
			Expect(string(body)).To(MatchJSON(expected))
		},
		Entry("without pagination", testCase{
			params: "",
			expected: `
{
  "total": 4,
  "items": [
	{
	  "type": "ServiceInsight",
	  "mesh": "mesh-1",
	  "name": "backend",
	  "creationTime": "2018-07-17T16:05:36.995Z",
	  "modificationTime": "2018-07-17T16:05:36.995Z",
      "status": "partially_degraded",
      "dataplanes": {
	    "total": 100,
	    "online": 70,
	    "offline": 30
      }
	},
	{
	  "type": "ServiceInsight",
	  "mesh": "mesh-1",
	  "name": "frontend",
	  "creationTime": "2018-07-17T16:05:36.995Z",
	  "modificationTime": "2018-07-17T16:05:36.995Z",
      "status": "partially_degraded",
      "dataplanes": {
	    "total": 20,
	    "online": 19,
	    "offline": 1
      }
	},
	{
	  "type": "ServiceInsight",
	  "mesh": "mesh-2",
	  "name": "db",
	  "creationTime": "2018-07-17T16:05:36.995Z",
	  "modificationTime": "2018-07-17T16:05:36.995Z",
      "status": "partially_degraded",
      "dataplanes": {
	    "total": 10,
	    "online": 9,
	    "offline": 1
      }
	},
	{
	  "type": "ServiceInsight",
	  "mesh": "mesh-2",
	  "name": "redis",
	  "creationTime": "2018-07-17T16:05:36.995Z",
	  "modificationTime": "2018-07-17T16:05:36.995Z",
      "status": "partially_degraded",
      "dataplanes": {
	    "total": 22,
	    "online": 19,
	    "offline": 3
      }
	}
  ],
  "next": null
}
`,
		}),
		Entry("with initial page", testCase{
			params: "?size=2",
			expected: `
{
  "total": 4,
  "items": [
	{
	  "type": "ServiceInsight",
	  "mesh": "mesh-1",
	  "name": "backend",
	  "creationTime": "2018-07-17T16:05:36.995Z",
	  "modificationTime": "2018-07-17T16:05:36.995Z",
      "status": "partially_degraded",
      "dataplanes": {
	    "total": 100,
	    "online": 70,
	    "offline": 30
      }
	},
	{
	  "type": "ServiceInsight",
	  "mesh": "mesh-1",
	  "name": "frontend",
	  "creationTime": "2018-07-17T16:05:36.995Z",
	  "modificationTime": "2018-07-17T16:05:36.995Z",
      "status": "partially_degraded",
      "dataplanes": {
	    "total": 20,
	    "online": 19,
	    "offline": 1
      }
	}
  ],
  "next": "http://{{address}}/service-insights?offset=2&size=2"
}
`,
		}),
		Entry("with second page", testCase{
			params: "?offset=2&size=2",
			expected: `
{
  "total": 4,
  "items": [
	{
	  "type": "ServiceInsight",
	  "mesh": "mesh-2",
	  "name": "db",
	  "creationTime": "2018-07-17T16:05:36.995Z",
	  "modificationTime": "2018-07-17T16:05:36.995Z",
      "status": "partially_degraded",
      "dataplanes": {
	    "total": 10,
	    "online": 9,
	    "offline": 1
      }
	},
	{
	  "type": "ServiceInsight",
	  "mesh": "mesh-2",
	  "name": "redis",
	  "creationTime": "2018-07-17T16:05:36.995Z",
	  "modificationTime": "2018-07-17T16:05:36.995Z",
      "status": "partially_degraded",
      "dataplanes": {
	    "total": 22,
	    "online": 19,
	    "offline": 3
      }
	}
  ],
  "next": null
}
`,
		}),
	)
})
