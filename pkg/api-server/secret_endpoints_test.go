package api_server_test

import (
	"context"
	"io/ioutil"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/ghodss/yaml"

	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	api_server "github.com/kumahq/kuma/pkg/api-server"
	config "github.com/kumahq/kuma/pkg/config/api-server"
	"github.com/kumahq/kuma/pkg/core"
	mesh_core "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/model/rest"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/metrics"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
)

var _ = Describe("Secret Endpoints", func() {
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
		metrics, err := metrics.NewMetrics("Standalone")
		Expect(err).ToNot(HaveOccurred())
		apiServer = createTestApiServer(resourceStore, config.DefaultApiServerConfig(), true, metrics)
		client = resourceApiClient{
			apiServer.Address(),
			"/meshes/default/secrets",
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
		err := resourceStore.Create(context.Background(), mesh_core.NewMeshResource(), store.CreateByKey(model.DefaultMesh, model.NoMesh))
		// then
		Expect(err).ToNot(HaveOccurred())
	})

	Describe("PUT => GET => DELETE => GET of meshed scoped secret", func() {

		given := `
        type: Secret
        name: sec-1
        mesh: default
        creationTime: "2018-07-17T16:05:36.995Z"
        modificationTime: "2018-07-17T16:05:36.995Z"
        data: dGVzdAo=
`
		It("GET should return data saved by PUT", func() {
			// given resource
			resource := rest.Resource{
				Spec: &system_proto.Secret{},
			}

			err := yaml.Unmarshal([]byte(given), &resource)
			Expect(err).ToNot(HaveOccurred())

			// when PUT a new resource
			response := client.put(resource)

			// then
			Expect(response.StatusCode).To(Equal(201))

			// when retrieve the resource
			response = client.get("sec-1")

			// then the resource is retrieved
			Expect(response.StatusCode).To(Equal(200))
			body, err := ioutil.ReadAll(response.Body)
			Expect(err).ToNot(HaveOccurred())

			// and the resource is equal to the one we saved
			actual, err := yaml.JSONToYAML(body)
			Expect(err).ToNot(HaveOccurred())
			Expect(actual).To(MatchYAML(given))

			// when delete resource
			response = client.delete("sec-1")

			// then resource is deleted
			Expect(response.StatusCode).To(Equal(200))

			// and not available via API
			response = client.get("sec-1")
			Expect(response.StatusCode).To(Equal(404))
		})
	})

	Describe("PUT => GET => DELETE => GET of global scoped secret", func() {

		given := `
        type: Secret
        name: sec-1
        creationTime: "2018-07-17T16:05:36.995Z"
        modificationTime: "2018-07-17T16:05:36.995Z"
        data: dGVzdAo=
`
		It("GET should return data saved by PUT", func() {
			// setup
			client = resourceApiClient{
				apiServer.Address(),
				"/secrets",
			}

			// given resource
			resource := rest.Resource{
				Spec: &system_proto.Secret{},
			}

			err := yaml.Unmarshal([]byte(given), &resource)
			Expect(err).ToNot(HaveOccurred())

			// when PUT a new resource
			response := client.put(resource)

			// then
			Expect(response.StatusCode).To(Equal(201))

			// when retrieve the resource
			response = client.get("sec-1")

			// then the resource is retrieved
			Expect(response.StatusCode).To(Equal(200))
			body, err := ioutil.ReadAll(response.Body)
			Expect(err).ToNot(HaveOccurred())

			// and the resource is equal to the one we saved
			actual, err := yaml.JSONToYAML(body)
			Expect(err).ToNot(HaveOccurred())
			Expect(actual).To(MatchYAML(given))

			// when delete resource
			response = client.delete("sec-1")

			// then resource is deleted
			Expect(response.StatusCode).To(Equal(200))

			// and not available via API
			response = client.get("sec-1")
			Expect(response.StatusCode).To(Equal(404))
		})
	})
})
