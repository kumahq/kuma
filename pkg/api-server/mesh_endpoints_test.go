package api_server_test

import (
	"context"
	"fmt"
	"io"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/api/mesh/v1alpha1"
	api_server "github.com/kumahq/kuma/pkg/api-server"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/model/rest/unversioned"
	rest_v1alpha1 "github.com/kumahq/kuma/pkg/core/resources/model/rest/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

var _ = Describe("Resource Endpoints", func() {
	var apiServer *api_server.ApiServer
	var resourceStore store.ResourceStore
	var client resourceApiClient
	var stop = func() {}
	t1, _ := time.Parse(time.RFC3339, "2018-07-17T16:05:36.995+00:00")
	BeforeEach(func() {
		resourceStore = memory.NewStore()
		apiServer, _, stop = StartApiServer(NewTestApiServerConfigurer().WithStore(resourceStore))
		client = resourceApiClient{
			apiServer.Address(),
			"/meshes",
		}
	})

	AfterEach(func() {
		stop()
	})

	Describe("On GET", func() {
		It("should return an existing resource", func() {
			// given
			putMeshIntoStore(resourceStore, "mesh-1", t1)

			// when
			response := client.get("mesh-1")

			// then
			Expect(response.StatusCode).To(Equal(200))
			body, err := io.ReadAll(response.Body)
			Expect(err).ToNot(HaveOccurred())
			json := `
			{
				"type": "Mesh",
				"name": "mesh-1",
				"creationTime": "2018-07-17T16:05:36.995Z",
				"modificationTime": "2018-07-17T16:05:36.995Z"
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
			putMeshIntoStore(resourceStore, "mesh-1", t1)
			putMeshIntoStore(resourceStore, "mesh-2", t1)

			// when
			response := client.list()

			// then
			Expect(response.StatusCode).To(Equal(200))
			json1 := `
			{
				"type": "Mesh",
				"name": "mesh-1",
				"creationTime": "2018-07-17T16:05:36.995Z",
				"modificationTime": "2018-07-17T16:05:36.995Z"
			}`
			json2 := `
			{
				"type": "Mesh",
				"name": "mesh-2",
				"creationTime": "2018-07-17T16:05:36.995Z",
				"modificationTime": "2018-07-17T16:05:36.995Z"
			}`
			body, err := io.ReadAll(response.Body)
			Expect(err).ToNot(HaveOccurred())

			Expect(string(body)).To(Or(
				MatchJSON(fmt.Sprintf(`{"total": %d, "items": [%s,%s], "next": null}`, 2, json1, json2)),
				MatchJSON(fmt.Sprintf(`{"total": %d, "items": [%s,%s], "next": null}`, 2, json2, json1)),
			))
		})
	})

	Describe("On PUT", func() {
		It("should create a resource when one does not exist", func() {
			// given
			res := &unversioned.Resource{
				Meta: rest_v1alpha1.ResourceMeta{
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
			putMeshIntoStore(resourceStore, name, t1)

			// when
			res := &unversioned.Resource{
				Meta: rest_v1alpha1.ResourceMeta{
					Name: "mesh-1",
					Type: string(mesh.MeshType),
				},
				Spec: &v1alpha1.Mesh{
					Tracing: &v1alpha1.Tracing{
						Backends: []*v1alpha1.TracingBackend{
							{
								Name: "zipkin-us",
								Type: v1alpha1.TracingZipkinType,
								Conf: util_proto.MustToStruct(&v1alpha1.ZipkinTracingBackendConfig{
									Url: "http://zipkin-us:9090/v2/spans",
								}),
							},
						},
					},
				},
			}
			response := client.put(res)
			Expect(response.StatusCode).To(Equal(200))

			// then
			resource := mesh.NewMeshResource()
			err := resourceStore.Get(context.Background(), resource, store.GetByKey(name, model.NoMesh))
			Expect(err).ToNot(HaveOccurred())
			Expect(resource.Spec.Tracing.Backends[0].Name).To(Equal("zipkin-us"))
		})

		It("should return 400 on the type in url that is different from request", func() {
			// given
			json := `
			{
				"type": "Mesh-1",
				"name": "tr-1"
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
				"name": "different-name"
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
				"name": "different-mesh"
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
			putMeshIntoStore(resourceStore, name, t1)

			// when
			response := client.delete(name)

			// then
			Expect(response.StatusCode).To(Equal(200))

			// and
			resource := mesh.NewMeshResource()
			err := resourceStore.Get(context.Background(), resource, store.GetByKey(name, name))
			Expect(err).To(Equal(store.ErrorResourceNotFound(resource.Descriptor().Name, name, name)))
		})

		It("should delete non-existing resource", func() {
			// when
			response := client.delete("non-existing-resource")

			// then
			Expect(response.StatusCode).To(Equal(404))
		})
	})
})

func putMeshIntoStore(resourceStore store.ResourceStore, name string, createdAt time.Time) {
	err := resourceStore.Create(context.Background(), mesh.NewMeshResource(), store.CreateByKey(name, model.NoMesh), store.CreatedAt(createdAt))
	Expect(err).NotTo(HaveOccurred())
}
