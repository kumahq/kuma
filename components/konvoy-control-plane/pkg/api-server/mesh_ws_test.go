package api_server_test

import (
	"context"
	"fmt"
	"github.com/Kong/konvoy/components/konvoy-control-plane/api/mesh/v1alpha1"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/api-server"
	config "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/config/api-server"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/apis/mesh"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/model/rest"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/store"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/plugins/resources/memory"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
)

var _ = Describe("Resource WS", func() {
	var apiServer *api_server.ApiServer
	var resourceStore store.ResourceStore
	var client resourceApiClient
	var stop chan struct{}

	const namespace = "default"

	BeforeEach(func() {
		resourceStore = memory.NewStore()
		apiServer = createTestApiServer(resourceStore, *config.DefaultApiServerConfig())
		client = resourceApiClient{
			apiServer.Address(),
			"/meshes",
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

	Describe("On GET", func() {
		It("should return an existing resource", func() {
			// given
			putMeshIntoStore(resourceStore, "mesh-1")

			// when
			response := client.get("mesh-1")

			// then
			Expect(response.StatusCode).To(Equal(200))
			body, err := ioutil.ReadAll(response.Body)
			Expect(err).ToNot(HaveOccurred())
			json := `
			{
				"type": "Mesh",
				"name": "mesh-1"
			}`
			Expect(body).To(MatchJSON(json))
		})

		It("should return 404 for non existing resource", func() {
			// when
			response := client.get("non-existing-resource")

			// then
			Expect(response.StatusCode).To(Equal(404))
		})

		It("should list resources", func() {
			// given
			putMeshIntoStore(resourceStore, "mesh-1")
			putMeshIntoStore(resourceStore, "mesh-2")

			// when
			response := client.list()

			// then
			Expect(response.StatusCode).To(Equal(200))
			json1 := `
			{
				"type": "Mesh",
				"name": "mesh-1"
			}`
			json2 := `
			{
				"type": "Mesh",
				"name": "mesh-2"
			}`
			body, err := ioutil.ReadAll(response.Body)
			Expect(err).ToNot(HaveOccurred())

			Expect(string(body)).To(Or(
				MatchJSON(fmt.Sprintf(`{"items": [%s,%s]}`, json1, json2)),
				MatchJSON(fmt.Sprintf(`{"items": [%s,%s]}`, json2, json1)),
			))
		})
	})

	Describe("On PUT", func() {
		It("should create a resource when one does not exist", func() {
			// given
			res := rest.Resource{
				Meta: rest.ResourceMeta{
					Name: "new-mesh",
					Type: string(mesh.MeshType),
				},
				Spec: &v1alpha1.Mesh{},
			}

			// when
			response := client.put(res)

			// then
			Expect(response.StatusCode).To(Equal(201))
		})

		It("should update a resource when one already exist", func() {
			// given
			name := "mesh-1"
			putMeshIntoStore(resourceStore, name)

			// when
			res := rest.Resource{
				Meta: rest.ResourceMeta{
					Name: "mesh-1",
					Type: string(mesh.MeshType),
				},
				Spec: &v1alpha1.Mesh{
					Tracing: &v1alpha1.Tracing{
						Type: &v1alpha1.Tracing_Zipkin_{},
					},
				},
			}
			response := client.put(res)
			Expect(response.StatusCode).To(Equal(200))

			// then
			resource := mesh.MeshResource{}
			err := resourceStore.Get(context.Background(), &resource, store.GetByKey(namespace, name, name))
			Expect(err).ToNot(HaveOccurred())
			Expect(resource.Spec.Tracing.Type).To(Equal(&v1alpha1.Tracing_Zipkin_{}))
		})

		It("should return 400 on the type in url that is different from request", func() {
			// given
			json := `
			{
				"type": "Mesh-1",
				"name": "tr-1",
			}
			`

			// when
			response := client.putJson("tr-1", []byte(json))

			// then
			Expect(response.StatusCode).To(Equal(400))
		})

		It("should return 400 on the name that is different from request", func() {
			// given
			json := `
			{
				"type": "Mesh",
				"name": "different-name",
			}
			`

			// when
			response := client.putJson("mesh-1", []byte(json))

			// then
			Expect(response.StatusCode).To(Equal(400))
		})

		It("should return 400 on the mesh that is different from request", func() {
			// given
			json := `
			{
				"type": "Mesh",
				"name": "different-mesh",
			}
			`

			// when
			response := client.putJson("mesh-1", []byte(json))

			// then
			Expect(response.StatusCode).To(Equal(400))
		})
	})

	Describe("On DELETE", func() {
		It("should delete existing resource", func() {
			// given
			name := "mesh-1"
			putMeshIntoStore(resourceStore, name)

			// when
			response := client.delete(name)

			// then
			Expect(response.StatusCode).To(Equal(200))

			// and
			resource := mesh.MeshResource{}
			err := resourceStore.Get(context.Background(), &resource, store.GetByKey(namespace, name, name))
			Expect(err).To(Equal(store.ErrorResourceNotFound(resource.GetType(), namespace, name, name)))
		})

		It("should delete non-existing resource", func() {
			// when
			response := client.delete("non-existing-resource")

			// then
			Expect(response.StatusCode).To(Equal(200))
		})
	})
})

func putMeshIntoStore(resourceStore store.ResourceStore, name string) {
	resource := mesh.MeshResource{}
	err := resourceStore.Create(context.Background(), &resource, store.CreateByKey("default", name, name))
	Expect(err).NotTo(HaveOccurred())
}
