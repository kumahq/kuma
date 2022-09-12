package api_server_test

import (
	"io"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	api_server "github.com/kumahq/kuma/pkg/api-server"
	config "github.com/kumahq/kuma/pkg/config/api-server"
	"github.com/kumahq/kuma/pkg/core/resources/model/rest/unversioned"
	rest_v1alpha1 "github.com/kumahq/kuma/pkg/core/resources/model/rest/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	sample_proto "github.com/kumahq/kuma/pkg/test/apis/sample/v1alpha1"
	"github.com/kumahq/kuma/pkg/test/resources/apis/sample"
)

var _ = Describe("Read only Resource Endpoints", func() {
	var apiServer *api_server.ApiServer
	var resourceStore store.ResourceStore
	var client resourceApiClient
	var stop = func() {}

	const resourceName = "tr-1"
	const mesh = "default-mesh"

	BeforeEach(func() {
		resourceStore = memory.NewStore()
		apiServer, _, stop = StartApiServer(NewTestApiServerConfigurer().WithStore(resourceStore).WithConfigMutator(func(serverConfig *config.ApiServerConfig) {
			serverConfig.ReadOnly = true
		}))
		client = resourceApiClient{
			address: apiServer.Address(),
			path:    "/meshes/" + mesh + "/sample-traffic-routes",
		}
		putSampleResourceIntoStore(resourceStore, resourceName, mesh)
	})

	AfterEach(func() {
		stop()
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
			res := &unversioned.Resource{
				Meta: rest_v1alpha1.ResourceMeta{
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
			body, err := io.ReadAll(response.Body)

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
			body, err := io.ReadAll(response.Body)

			Expect(err).NotTo(HaveOccurred())
			Expect(string(body)).To(Equal(
				"On Kubernetes you cannot change the state of Kuma resources with 'kumactl apply' or via the HTTP API." +
					" As a best practice, you should always be using 'kubectl apply' instead." +
					" You can still use 'kumactl' or the HTTP API to make read-only operations. On Universal this limitation does not apply.\n"))
		})
	})
})
