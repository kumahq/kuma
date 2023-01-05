package api_server_test

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/emicklei/go-restful/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	api_server "github.com/kumahq/kuma/pkg/api-server"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/model/rest/unversioned"
	rest_v1alpha1 "github.com/kumahq/kuma/pkg/core/resources/model/rest/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	test_metrics "github.com/kumahq/kuma/pkg/test/metrics"
)

var _ = Describe("Resource Endpoints", func() {
	var apiServer *api_server.ApiServer
	var resourceStore store.ResourceStore
	var client resourceApiClient
	var stop = func() {}
	var metrics core_metrics.Metrics

	const mesh = "default"

	BeforeEach(func() {
		resourceStore = store.NewPaginationStore(memory.NewStore())
		apiServer, _, stop = StartApiServer(NewTestApiServerConfigurer().WithStore(resourceStore).WithMetrics(func() core_metrics.Metrics {
			m, _ := core_metrics.NewMetrics("Standalone")
			metrics = m
			return m
		}))
		client = resourceApiClient{
			address: apiServer.Address(),
			path:    "/meshes/" + mesh + "/traffic-routes",
		}
	})

	AfterEach(func() {
		stop()
	})

	BeforeEach(func() {
		// create default mesh
		err := resourceStore.Create(context.Background(), core_mesh.NewMeshResource(), store.CreateByKey(mesh, model.NoMesh))
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
			body, err := io.ReadAll(response.Body)
			Expect(err).ToNot(HaveOccurred())
			json := `
			{
				"type": "TrafficRoute",
				"name": "tr-1",
				"mesh": "default",
				"creationTime": "0001-01-01T00:00:00Z",
				"modificationTime": "0001-01-01T00:00:00Z",
				"conf": {
				  "destination": {
					"path": "/sample-path"
				  }
				}
			}`
			Expect(body).To(MatchJSON(json))
		})

		It("should return 404 for non existing resource", func() {
			// when
			response := client.get("non-existing-resource")

			// then
			Expect(response.StatusCode).To(Equal(404))

			// and
			bytes, err := io.ReadAll(response.Body)
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
				"creationTime": "0001-01-01T00:00:00Z",
				"modificationTime": "0001-01-01T00:00:00Z",
				"conf": {
				  "destination": {
					"path": "/sample-path"
				  }
				}
			}`
			json2 := `
			{
				"type": "TrafficRoute",
				"name": "tr-2",
				"mesh": "default",
				"creationTime": "0001-01-01T00:00:00Z",
				"modificationTime": "0001-01-01T00:00:00Z",
				"conf": {
				  "destination": {
					"path": "/sample-path"
				  }
				}
			}`
			body, err := io.ReadAll(response.Body)
			Expect(err).ToNot(HaveOccurred())
			Expect(body).To(Or(
				MatchJSON(fmt.Sprintf(`{"total": %d, "items": [%s,%s], "next": null}`, 2, json1, json2)),
				MatchJSON(fmt.Sprintf(`{"total": %d, "items": [%s,%s], "next": null}`, 2, json2, json1)),
			))
		})

		It("should list resources from all meshes", func() {
			// given
			putSampleResourceIntoStore(resourceStore, "tr-1", "mesh-1")
			putSampleResourceIntoStore(resourceStore, "tr-2", "mesh-2")

			// when
			client = resourceApiClient{
				address: apiServer.Address(),
				path:    "/traffic-routes",
			}
			response := client.list()

			// then
			Expect(response.StatusCode).To(Equal(200))
			json1 := `
			{
				"type": "TrafficRoute",
				"name": "tr-1",
				"mesh": "mesh-1",
				"creationTime": "0001-01-01T00:00:00Z",
				"modificationTime": "0001-01-01T00:00:00Z",
				"conf": {
				  "destination": {
					"path": "/sample-path"
				  }
				}
			}`
			json2 := `
			{
				"type": "TrafficRoute",
				"name": "tr-2",
				"mesh": "mesh-2",
				"creationTime": "0001-01-01T00:00:00Z",
				"modificationTime": "0001-01-01T00:00:00Z",
				"conf": {
				  "destination": {
					"path": "/sample-path"
				  }
				}
			}`
			body, err := io.ReadAll(response.Body)
			Expect(err).ToNot(HaveOccurred())
			Expect(body).To(Or(
				MatchJSON(fmt.Sprintf(`{"total": %d, "items": [%s,%s], "next": null}`, 2, json1, json2)),
				MatchJSON(fmt.Sprintf(`{"total": %d, "items": [%s,%s], "next": null}`, 2, json2, json1)),
			))
		})

		It("should list resources using pagination", func() {
			// given three resources
			putSampleResourceIntoStore(resourceStore, "tr-1", "mesh-1")
			putSampleResourceIntoStore(resourceStore, "tr-2", "mesh-1")
			putSampleResourceIntoStore(resourceStore, "tr-3", "mesh-1")

			// when ask for page with size 2
			client = resourceApiClient{
				address: apiServer.Address(),
				path:    "/traffic-routes?size=2",
			}
			response := client.list()

			// then one page is returned with next url
			Expect(response.StatusCode).To(Equal(200))
			json := fmt.Sprintf(`
			{
				"total": 3,
				"items": [
					{
						"type": "TrafficRoute",
						"name": "tr-1",
						"mesh": "mesh-1",
						"creationTime": "0001-01-01T00:00:00Z",
						"modificationTime": "0001-01-01T00:00:00Z",
						"conf": {
						  "destination": {
							"path": "/sample-path"
						  }
						}
					},
					{
						"type": "TrafficRoute",
						"name": "tr-2",
						"mesh": "mesh-1",
						"creationTime": "0001-01-01T00:00:00Z",
						"modificationTime": "0001-01-01T00:00:00Z",
						"conf": {
						  "destination": {
							"path": "/sample-path"
						  }
						}
					}
				],
				"next": "http://%s/traffic-routes?offset=2&size=2"
			}`, client.address)
			body, err := io.ReadAll(response.Body)
			Expect(err).ToNot(HaveOccurred())
			Expect(body).To(MatchJSON(json))

			// when query for next page
			client = resourceApiClient{
				address: apiServer.Address(),
				path:    "/traffic-routes?size=2&offset=2",
			}
			response = client.list()

			// then another page with one element left is returned
			Expect(response.StatusCode).To(Equal(200))
			json = `
			{
				"total": 3,
				"items": [
					{
						"type": "TrafficRoute",
						"name": "tr-3",
						"mesh": "mesh-1",
						"creationTime": "0001-01-01T00:00:00Z",
				        "modificationTime": "0001-01-01T00:00:00Z",
						"conf": {
						  "destination": {
							"path": "/sample-path"
						  }
						}
					}
				],
				"next": null
			}`
			body, err = io.ReadAll(response.Body)
			Expect(err).ToNot(HaveOccurred())
			Expect(body).To(MatchJSON(json))
		})

		It("should return 400 with error on invalid offset", func() {
			// when
			client = resourceApiClient{
				address: apiServer.Address(),
				path:    "/traffic-routes?size=2&offset=invalidoffset",
			}
			response := client.list()

			// then
			Expect(response.StatusCode).To(Equal(400))
			// and
			bytes, err := io.ReadAll(response.Body)
			Expect(err).ToNot(HaveOccurred())
			Expect(bytes).To(MatchJSON(`
			{
				"title": "Could not retrieve resources",
				"details": "Invalid offset",
				"causes": [
					{
						"field": "offset",
						"message": "Invalid format"
					}
				]
			}
			`))
		})

		It("should return 400 with error on invalid size type", func() {
			// when
			client = resourceApiClient{
				address: apiServer.Address(),
				path:    "/traffic-routes?size=invalid",
			}
			response := client.list()

			// then
			Expect(response.StatusCode).To(Equal(400))
			// and
			bytes, err := io.ReadAll(response.Body)
			Expect(err).ToNot(HaveOccurred())
			Expect(bytes).To(MatchJSON(`
			{
				"title": "Could not retrieve resources",
				"details": "Invalid page size",
				"causes": [
					{
						"field": "size",
						"message": "Invalid format"
					}
				]
			}
			`))
		})

		It("should return 400 with error when page size exceeded the limit", func() {
			// when
			client = resourceApiClient{
				address: apiServer.Address(),
				path:    "/traffic-routes?size=2000",
			}
			response := client.list()

			// then
			Expect(response.StatusCode).To(Equal(400))
			// and
			bytes, err := io.ReadAll(response.Body)
			Expect(err).ToNot(HaveOccurred())
			Expect(bytes).To(MatchJSON(`
			{
				"title": "Could not retrieve resources",
				"details": "Invalid page size",
				"causes": [
					{
						"field": "size",
						"message": "Invalid page size of 2000. Maximum page size is 1000"
					}
				]
			}
			`))
		})
	})

	Describe("On PUT", func() {
		It("should create a resource when one does not exist", func() {
			// given
			res := &unversioned.Resource{
				Meta: rest_v1alpha1.ResourceMeta{
					Name: "new-resource",
					Mesh: mesh,
					Type: string(core_mesh.TrafficRouteType),
				},
				Spec: &mesh_proto.TrafficRoute{
					Sources: []*mesh_proto.Selector{{Match: map[string]string{
						mesh_proto.ServiceTag: "*",
					}}},
					Destinations: []*mesh_proto.Selector{{Match: map[string]string{
						mesh_proto.ServiceTag: "*",
					}}},
					Conf: &mesh_proto.TrafficRoute_Conf{
						Destination: map[string]string{
							mesh_proto.ServiceTag: "*",
							"path":                "/sample-path",
						},
					},
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
			res := &unversioned.Resource{
				Meta: rest_v1alpha1.ResourceMeta{
					Name: name,
					Mesh: mesh,
					Type: string(core_mesh.TrafficRouteType),
				},
				Spec: &mesh_proto.TrafficRoute{
					Sources: []*mesh_proto.Selector{{Match: map[string]string{
						mesh_proto.ServiceTag: "*",
					}}},
					Destinations: []*mesh_proto.Selector{{Match: map[string]string{
						mesh_proto.ServiceTag: "*",
					}}},
					Conf: &mesh_proto.TrafficRoute_Conf{
						Destination: map[string]string{
							mesh_proto.ServiceTag: "*",
							"path":                "/update-sample-path",
						},
					},
				},
			}
			response := client.put(res)
			Expect(response.StatusCode).To(Equal(200))

			// then
			resource := core_mesh.NewTrafficRouteResource()
			err := resourceStore.Get(context.Background(), resource, store.GetByKey(name, mesh))
			Expect(err).ToNot(HaveOccurred())
			Expect(resource.Spec.Conf.Destination["path"]).To(Equal("/update-sample-path"))
		})

		It("should return 400 on the type in url that is different from request", func() {
			// given
			json := `
			{
				"type": "MeshTrafficPermission",
				"name": "tr-1",
				"mesh": "default",
				"spec": {
					"targetRef": {
						"kind": "Mesh"
					},
					"from": [
						{
							"targetRef": {
								"kind": "Mesh"
							},
							"default": {
								"action": "ALLOW"
							}
						}
					]
				}
			}
			`

			// when
			response := client.putJson("tr-1", []byte(json))

			// then
			Expect(response.StatusCode).To(Equal(400))

			// and
			bytes, err := io.ReadAll(response.Body)
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
				"sources": [
					{
						"match": {
							"kuma.io/service": "frontend"
						}
					}
				],
				"destinations": [
					{
						"match": {
							"kuma.io/service": "backend"
						}
					}
				],
				"conf": {
					"destination": {
						"kuma.io/service": "backend-v2"
					}
				}
			}
			`

			// when
			response := client.putJson("tr-1", []byte(json))

			// then
			Expect(response.StatusCode).To(Equal(400))

			// and
			bytes, err := io.ReadAll(response.Body)
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
				"sources": [
					{
						"match": {
							"kuma.io/service": "frontend"
						}
					}
				],
				"destinations": [
					{
						"match": {
							"kuma.io/service": "backend"
						}
					}
				],
				"conf": {
					"destination": {
						"kuma.io/service": "backend-v2"
					}
				}
			}
			`

			// when
			response := client.putJson("tr-1", []byte(json))

			// then
			Expect(response.StatusCode).To(Equal(400))

			// and
			bytes, err := io.ReadAll(response.Body)
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
				"sources": [
				  {
					"match": {
					  "kuma.io/service": "*"
					}
				  }
				],
				"destinations": [
				  {
					"match": {
					  "kuma.io/service": "*"
					}
				  }
				],
				"conf": {}
			}
			`

			// when
			response := client.putJson("tr-1", []byte(json))

			// then
			Expect(response.StatusCode).To(Equal(400))

			// when
			respBytes, err := io.ReadAll(response.Body)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(respBytes).To(MatchJSON(`
			{
				"title": "Could not process a resource",
				"details": "Resource is not valid",
				"causes": [
					{
						"field": "conf",
						"message": "requires either \"destination\" or \"split\""
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
				"sources": [
					{
						"match": {
							"kuma.io/service": "frontend"
						}
					}
				],
				"destinations": [
					{
						"match": {
							"kuma.io/service": "backend"
						}
					}
				],
				"conf": {
					"destination": {
						"kuma.io/service": "backend-v2"
					}
				}
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
			respBytes, err := io.ReadAll(response.Body)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(respBytes).To(MatchJSON(`
			{
				"title": "Could not process a resource",
				"details": "Resource is not valid",
				"causes": [
					{
						"field": "name",
						"message": "invalid characters. Valid characters are numbers, lowercase latin letters and '-', '_' symbols."
					},
					{
						"field": "mesh",
						"message": "invalid characters. Valid characters are numbers, lowercase latin letters and '-', '_' symbols."
					}
				]
			}
			`))
		})

		It("should return 400 when mesh does not exist", func() {
			// setup
			err := resourceStore.Delete(context.Background(), core_mesh.NewMeshResource(), store.DeleteByKey(model.DefaultMesh, model.NoMesh))
			Expect(err).ToNot(HaveOccurred())

			// given
			res := &unversioned.Resource{
				Meta: rest_v1alpha1.ResourceMeta{
					Name: "new-resource",
					Mesh: "default",
					Type: string(core_mesh.TrafficRouteType),
				},
				Spec: &mesh_proto.TrafficRoute{
					Sources: []*mesh_proto.Selector{{Match: map[string]string{
						mesh_proto.ServiceTag: "*",
					}}},
					Destinations: []*mesh_proto.Selector{{Match: map[string]string{
						mesh_proto.ServiceTag: "*",
					}}},
					Conf: &mesh_proto.TrafficRoute_Conf{
						Destination: map[string]string{
							mesh_proto.ServiceTag: "*",
							"path":                "/sample-path",
						},
					},
				},
			}

			// when
			response := client.put(res)

			// when
			respBytes, err := io.ReadAll(response.Body)

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
			resource := core_mesh.NewTrafficRouteResource()
			err := resourceStore.Get(context.Background(), resource, store.GetByKey(name, mesh))
			Expect(err).To(Equal(store.ErrorResourceNotFound(resource.Descriptor().Name, name, mesh)))
		})

		It("should delete non-existing resource", func() {
			// when
			response := client.delete("non-existing-resource")

			// then
			Expect(response.StatusCode).To(Equal(404))

			// and
			bytes, err := io.ReadAll(response.Body)
			Expect(err).ToNot(HaveOccurred())
			Expect(bytes).To(MatchJSON(`
			{
				"title": "Could not delete a resource",
				"details": "Not found"
			}
			`))
		})
	})

	It("should support CORS", func() {
		// given
		req, err := http.NewRequest("GET", fmt.Sprintf("http://%s/meshes/%s/traffic-routes", apiServer.Address(), mesh), nil)
		Expect(err).NotTo(HaveOccurred())
		req.Header.Add(restful.HEADER_Origin, "test")

		// when
		response, err := http.DefaultClient.Do(req)

		// then
		Expect(err).NotTo(HaveOccurred())

		// when
		value := response.Header.Get(restful.HEADER_AccessControlAllowOrigin)

		// then server returns that the domain is allowed
		Expect(value).To(Equal("test"))
	})

	It("should expose metrics", func() {
		// given
		req, err := http.NewRequest("GET", fmt.Sprintf("http://%s/meshes/%s/traffic-routes", apiServer.Address(), mesh), nil)
		Expect(err).NotTo(HaveOccurred())

		// when
		_, err = http.DefaultClient.Do(req)

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(test_metrics.FindMetric(metrics, "api_server_http_request_duration_seconds")).ToNot(BeNil())
		Expect(test_metrics.FindMetric(metrics, "api_server_http_requests_inflight")).ToNot(BeNil())
		Expect(test_metrics.FindMetric(metrics, "api_server_http_response_size_bytes")).ToNot(BeNil())
	})
})
