package api_server_test

import (
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/api-server"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/model/rest"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/store"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/plugins/resources/memory"
	sample_proto "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/test/apis/sample/v1alpha1"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/test/resources/apis/sample"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Read only Resource WS", func() {
	var apiServer *api_server.ApiServer
	var resourceStore store.ResourceStore
	var client resourceApiClient

	const resourceName = "tr-1"
	const mesh = "default-mesh"

	BeforeEach(func() {
		resourceStore = memory.NewStore()
		apiServer = createTestApiServer(resourceStore, api_server.ApiServerConfig{ReadOnly: true})
		client = resourceApiClient{
			address: apiServer.Address(),
			path:    "/meshes/" + mesh + "/traffic-routes",
		}
		apiServer.Start()
		waitForServer(&client)

		putSampleResourceIntoStore(resourceStore, resourceName, mesh)
	})

	AfterEach(func() {
		err := apiServer.Stop()
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("On GET", func() {
		It("should return an existing resource", func() {
			// when
			response := client.get(resourceName)

			// then
			Expect(response.StatusCode).To(Equal(200))
		})

		It("should list resources", func() {
			// when
			response := client.list()

			// then
			Expect(response.StatusCode).To(Equal(200))
		})
	})

	Describe("On PUT", func() {
		It("should return 405", func() {
			// given
			res := rest.Resource{
				Meta: rest.ResourceMeta{
					Name: "new-resource",
					Mesh: mesh,
					Type: string(sample.TrafficRouteType),
				},
				Spec: &sample_proto.TrafficRoute{
					Path: "/sample-path",
				},
			}

			// when
			response := client.put(res)

			// then
			Expect(response.StatusCode).To(Equal(405))
		})
	})

	Describe("On DELETE", func() {
		It("should return 405", func() {
			// when
			response := client.delete("res-1")

			// then
			Expect(response.StatusCode).To(Equal(405))
		})
	})
})
