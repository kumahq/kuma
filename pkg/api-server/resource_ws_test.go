package api_server_test

import (
	"context"
	"fmt"
	"github.com/Kong/kuma/pkg/api-server"
	config "github.com/Kong/kuma/pkg/config/api-server"
	mesh_res "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	"github.com/Kong/kuma/pkg/core/resources/model/rest"
	"github.com/Kong/kuma/pkg/core/resources/store"
	"github.com/Kong/kuma/pkg/plugins/resources/memory"
	sample_proto "github.com/Kong/kuma/pkg/test/apis/sample/v1alpha1"
	sample_model "github.com/Kong/kuma/pkg/test/resources/apis/sample"
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
	const mesh = "default"

	BeforeEach(func() {
		resourceStore = memory.NewStore()
		apiServer = createTestApiServer(resourceStore, config.DefaultApiServerConfig())
		client = resourceApiClient{
			address: apiServer.Address(),
			path:    "/meshes/" + mesh + "/traffic-routes",
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

	BeforeEach(func() {
		// create default mesh
		err := resourceStore.Create(context.Background(), &mesh_res.MeshResource{}, store.CreateByKey(namespace, mesh, mesh))
		Expect(err).ToNot(HaveOccurred())
	})

	Describe("On GET", func() {
		It("should return an existing resource", func() {
			// given
			putSampleResourceIntoStore(resourceStore, "tr-1", mesh)

			// when
			response := client.get("tr-1")

			// then
			Expect(response.StatusCode).To(Equal(200))
			body, err := ioutil.ReadAll(response.Body)
			Expect(err).ToNot(HaveOccurred())
			json := `
			{
				"type": "TrafficRoute",
				"name": "tr-1",
				"mesh": "default",
				"path": "/sample-path"
			}`
			Expect(body).To(MatchJSON(json))
		})

		It("should return 404 for non existing resource", func() {
			// when
			response := client.get("non-existing-resource")

			// then
			Expect(response.StatusCode).To(Equal(404))

			// and
			bytes, err := ioutil.ReadAll(response.Body)
			Expect(err).ToNot(HaveOccurred())
			Expect(bytes).To(MatchJSON(`
			{
				"title": "Could not retrieve a resource",
				"details": "Not found"
			}
			`))
		})

		It("should list resources", func() {
			// given
			putSampleResourceIntoStore(resourceStore, "tr-1", mesh)
			putSampleResourceIntoStore(resourceStore, "tr-2", mesh)

			// when
			response := client.list()

			// then
			Expect(response.StatusCode).To(Equal(200))
			json1 := `
			{
				"type": "TrafficRoute",
				"name": "tr-1",
				"mesh": "default",
				"path": "/sample-path"
			}`
			json2 := `
			{
				"type": "TrafficRoute",
				"name": "tr-2",
				"mesh": "default",
				"path": "/sample-path"
			}`
			body, err := ioutil.ReadAll(response.Body)
			Expect(err).ToNot(HaveOccurred())
			Expect(body).To(Or(
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
					Name: "new-resource",
					Mesh: mesh,
					Type: string(sample_model.TrafficRouteType),
				},
				Spec: &sample_proto.TrafficRoute{
					Path: "/sample-path",
				},
			}

			// when
			response := client.put(res)

			// then
			Expect(response.StatusCode).To(Equal(201))
		})

		It("should update a resource when one already exist", func() {
			// given
			name := "tr-1"
			putSampleResourceIntoStore(resourceStore, name, mesh)

			// when
			res := rest.Resource{
				Meta: rest.ResourceMeta{
					Name: name,
					Mesh: mesh,
					Type: string(sample_model.TrafficRouteType),
				},
				Spec: &sample_proto.TrafficRoute{
					Path: "/update-sample-path",
				},
			}
			response := client.put(res)
			Expect(response.StatusCode).To(Equal(200))

			// then
			resource := sample_model.TrafficRouteResource{}
			err := resourceStore.Get(context.Background(), &resource, store.GetByKey(namespace, name, mesh))
			Expect(err).ToNot(HaveOccurred())
			Expect(resource.Spec.Path).To(Equal("/update-sample-path"))
		})

		It("should return 400 on the type in url that is different from request", func() {
			// given
			json := `
			{
				"type": "InvalidType",
				"name": "tr-1",
				"mesh": "default",
				"path": "/sample-path"
			}
			`

			// when
			response := client.putJson("tr-1", []byte(json))

			// then
			Expect(response.StatusCode).To(Equal(400))

			// and
			bytes, err := ioutil.ReadAll(response.Body)
			Expect(err).ToNot(HaveOccurred())
			Expect(bytes).To(MatchJSON(`
			{
				"title": "Could not process a resource",
				"details": "Resource is not valid",
				"causes": [
					{
						"field": "type",
						"message": "type from the URL has to be the same as in body"
					}
				]
			}
			`))
		})

		It("should return 400 on the name that is different from request", func() {
			// given
			json := `
			{
				"type": "TrafficRoute",
				"name": "different-name",
				"mesh": "default",
				"path": "/sample-path"
			}
			`

			// when
			response := client.putJson("tr-1", []byte(json))

			// then
			Expect(response.StatusCode).To(Equal(400))

			// and
			bytes, err := ioutil.ReadAll(response.Body)
			Expect(err).ToNot(HaveOccurred())
			Expect(bytes).To(MatchJSON(`
			{
				"title": "Could not process a resource",
				"details": "Resource is not valid",
				"causes": [
					{
						"field": "name",
						"message": "name from the URL has to be the same as in body"
					}
				]
			}
			`))
		})

		It("should return 400 on the mesh that is different from request", func() {
			// given
			json := `
			{
				"type": "TrafficRoute",
				"name": "tr-1",
				"mesh": "different-mesh",
				"path": "/sample-path"
			}
			`

			// when
			response := client.putJson("tr-1", []byte(json))

			// then
			Expect(response.StatusCode).To(Equal(400))

			// and
			bytes, err := ioutil.ReadAll(response.Body)
			Expect(err).ToNot(HaveOccurred())
			Expect(bytes).To(MatchJSON(`
			{
				"title": "Could not process a resource",
				"details": "Resource is not valid",
				"causes": [
					{
						"field": "mesh",
						"message": "mesh from the URL has to be the same as in body"
					}
				]
			}`))
		})

		It("should return 400 on validation error", func() {
			// given
			json := `
			{
				"type": "TrafficRoute",
				"name": "tr-1",
				"mesh": "default",
				"path": ""
			}
			`

			// when
			response := client.putJson("tr-1", []byte(json))

			// then
			Expect(response.StatusCode).To(Equal(400))

			// when
			respBytes, err := ioutil.ReadAll(response.Body)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(respBytes).To(MatchJSON(`
			{
				"title": "Could not create a resource",
				"details": "Resource is not valid",
				"causes": [
					{
						"field": "path",
						"message": "cannot be empty"
					}
				]
			}
			`))
		})

		It("should return 400 on invalid name and mesh", func() {
			// given
			json := `
			{
				"type": "TrafficRoute",
				"name": "invalid@",
				"mesh": "invalid$",
				"path": "/path"
			}
			`

			// when
			client = resourceApiClient{
				address: apiServer.Address(),
				path:    "/meshes/invalid$/traffic-routes",
			}
			response := client.putJson("invalid@", []byte(json))

			// then
			Expect(response.StatusCode).To(Equal(400))

			// when
			respBytes, err := ioutil.ReadAll(response.Body)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(respBytes).To(MatchJSON(`
			{
				"title": "Could not process a resource",
				"details": "Resource is not valid",
				"causes": [
					{
						"field": "name",
						"message": "invalid characters. Valid characters are numbers, letters and '-', '_' symbols."
					},
					{
						"field": "mesh",
						"message": "invalid characters. Valid characters are numbers, letters and '-', '_' symbols."
					}
				]
			}
			`))
		})

		It("should return 400 when mesh does not exist", func() {
			// setup
			err := resourceStore.Delete(context.Background(), &mesh_res.MeshResource{}, store.DeleteByKey("default", "default", "default"))
			Expect(err).ToNot(HaveOccurred())

			// given
			res := rest.Resource{
				Meta: rest.ResourceMeta{
					Name: "new-resource",
					Mesh: "default",
					Type: string(sample_model.TrafficRouteType),
				},
				Spec: &sample_proto.TrafficRoute{
					Path: "/sample-path",
				},
			}

			// when
			response := client.put(res)

			// when
			respBytes, err := ioutil.ReadAll(response.Body)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(respBytes).To(MatchJSON(`
			{
				"title": "Could not create a resource",
				"details": "Mesh is not found",
				"causes": [
					{
						"field": "mesh",
						"message": "mesh of name default is not found"
					}
          		]
			}
			`))
		})
	})

	Describe("On DELETE", func() {
		It("should delete existing resource", func() {
			// given
			name := "tr-1"
			putSampleResourceIntoStore(resourceStore, name, mesh)

			// when
			response := client.delete(name)

			// then
			Expect(response.StatusCode).To(Equal(200))

			// and
			resource := sample_model.TrafficRouteResource{}
			err := resourceStore.Get(context.Background(), &resource, store.GetByKey(namespace, name, mesh))
			Expect(err).To(Equal(store.ErrorResourceNotFound(resource.GetType(), namespace, name, mesh)))
		})

		It("should delete non-existing resource", func() {
			// when
			response := client.delete("non-existing-resource")

			// then
			Expect(response.StatusCode).To(Equal(200))
		})
	})
})
