package api_server_test

import (
	"io/ioutil"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	api_server "github.com/Kong/kuma/pkg/api-server"
	config "github.com/Kong/kuma/pkg/config/api-server"
	"github.com/Kong/kuma/pkg/core/resources/model/rest"
	"github.com/Kong/kuma/pkg/core/resources/store"
	"github.com/Kong/kuma/pkg/plugins/resources/memory"
	sample_proto "github.com/Kong/kuma/pkg/test/apis/sample/v1alpha1"
	"github.com/Kong/kuma/pkg/test/resources/apis/sample"
)

var _ = Describe("Read only Resource Endpoints", func() {
	var apiServer *api_server.ApiServer
	var resourceStore store.ResourceStore
	var client resourceApiClient
	var stop chan struct{}

	const resourceName = "tr-1"
	const mesh = "default-mesh"

	BeforeEach(func() {
		resourceStore = memory.NewStore()
		cfg := config.DefaultApiServerConfig()
		cfg.ReadOnly = true
		apiServer = createTestApiServer(resourceStore, cfg)
		client = resourceApiClient{
			address: apiServer.Address(),
			path:    "/meshes/" + mesh + "/sample-traffic-routes",
		}
		stop = make(chan struct{})
		go func() {
			defer GinkgoRecover()
			err := apiServer.Start(stop)
			Expect(err).ToNot(HaveOccurred())
		}()
		waitForServer(&client)

		putSampleResourceIntoStore(resourceStore, resourceName, mesh)
	}, 5)

	AfterEach(func() {
		close(stop)
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
			body, err := ioutil.ReadAll(response.Body)

			Expect(err).NotTo(HaveOccurred())
			Expect(string(body)).To(Equal(
				"On Kubernetes you cannot change the state of Kuma resources with 'kumactl apply' or via the HTTP API." +
					" As a best practice, you should always be using 'kubectl apply' instead." +
					" You can still use 'kumactl' or the HTTP API to make read-only operations. On Universal this limitation does not apply.\n"))
		})
	})

	Describe("On DELETE", func() {
		It("should return 405", func() {
			// when
			response := client.delete("res-1")

			// then
			Expect(response.StatusCode).To(Equal(405))
			body, err := ioutil.ReadAll(response.Body)

			Expect(err).NotTo(HaveOccurred())
			Expect(string(body)).To(Equal(
				"On Kubernetes you cannot change the state of Kuma resources with 'kumactl apply' or via the HTTP API." +
					" As a best practice, you should always be using 'kubectl apply' instead." +
					" You can still use 'kumactl' or the HTTP API to make read-only operations. On Universal this limitation does not apply.\n"))
		})
	})
})
