package api_server_test

import (
	"io"
	"path"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	api_server "github.com/kumahq/kuma/pkg/api-server"
	config "github.com/kumahq/kuma/pkg/config/api-server"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model/rest/unversioned"
	rest_v1alpha1 "github.com/kumahq/kuma/pkg/core/resources/model/rest/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/pkg/test/matchers"
)

var _ = Describe("Read only Resource Endpoints", func() {
	var apiServer *api_server.ApiServer
	var resourceStore store.ResourceStore
	var client resourceApiClient
	stop := func() {}

	const resourceName = "tr-1"
	const mesh = "default-mesh"

	BeforeEach(func() {
		resourceStore = memory.NewStore()
		apiServer, _, stop = StartApiServer(NewTestApiServerConfigurer().WithStore(resourceStore).WithConfigMutator(func(serverConfig *config.ApiServerConfig) {
			serverConfig.ReadOnly = true
		}))
		client = resourceApiClient{
			address: apiServer.Address(),
			path:    "/meshes/" + mesh + "/traffic-routes",
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
					Type: string(core_mesh.TrafficRouteType),
				},
				Spec: &mesh_proto.TrafficRoute{
					Conf: &mesh_proto.TrafficRoute_Conf{
						Destination: map[string]string{
							"path": "/sample-path",
						},
					},
				},
			}

			// when
			response := client.put(res)

			// then
			Expect(response.StatusCode).To(Equal(405))
			bytes, err := io.ReadAll(response.Body)

			Expect(err).NotTo(HaveOccurred())
			Expect(bytes).To(matchers.MatchGoldenJSON(path.Join("testdata", "resource-read-only_put.golden.json")))
		})
	})

	Describe("On DELETE", func() {
		It("should return 405", func() {
			// when
			response := client.delete("res-1")

			// then
			Expect(response.StatusCode).To(Equal(405))
			bytes, err := io.ReadAll(response.Body)

			Expect(err).NotTo(HaveOccurred())
			Expect(bytes).To(matchers.MatchGoldenJSON(path.Join("testdata", "resource-read-only_delete.golden.json")))
		})
	})
})
