package api_server_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"path"

	"github.com/emicklei/go-restful/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	api_server "github.com/kumahq/kuma/pkg/api-server"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/model/rest/unversioned"
	rest_v1alpha1 "github.com/kumahq/kuma/pkg/core/resources/model/rest/v1alpha1"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
	"github.com/kumahq/kuma/pkg/plugins/policies/meshtrafficpermission/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/pkg/test/matchers"
	test_metrics "github.com/kumahq/kuma/pkg/test/metrics"
	"github.com/kumahq/kuma/pkg/test/resources/builders"
)

var _ = Describe("Resource Endpoints Zone", func() {
	var apiServer *api_server.ApiServer
	var resourceStore core_store.ResourceStore
	var client resourceApiClient
	stop := func() {}

	const mesh = "default"

	BeforeEach(func() {
		resourceStore = core_store.NewPaginationStore(memory.NewStore())
		apiServer, _, stop = StartApiServer(
			NewTestApiServerConfigurer().
				WithStore(resourceStore).
				WithZone("custom-long-zone-name-to-check-if-name-length-check-works"),
		)
		client = resourceApiClient{
			address: apiServer.Address(),
			path:    "/meshes/" + mesh + "/dataplanes",
		}
	})

	AfterEach(func() {
		stop()
	})

	BeforeEach(func() {
		// create default mesh
		err := resourceStore.Create(context.Background(), core_mesh.NewMeshResource(), core_store.CreateByKey(mesh, model.NoMesh))
		Expect(err).ToNot(HaveOccurred())
	})

	Describe("On PUT", func() {
		It("should create a resource when one does not exist", func() {
			// given
			res := &unversioned.Resource{
				Meta: rest_v1alpha1.ResourceMeta{
					Name: "new-resource",
					Mesh: mesh,
					Type: string(core_mesh.DataplaneType),
				},
				Spec: &mesh_proto.Dataplane{
					Networking: &mesh_proto.Dataplane_Networking{
						Address:           "192.168.1.1",
						AdvertisedAddress: "192.168.1.1",
						Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
							{
								Port:        8000,
								ServicePort: 8080,
								Tags: map[string]string{
									mesh_proto.ServiceTag: "service-1",
								},
							},
						},
					},
				},
			}

			// when
			response := client.put(res)

			// then
			Expect(response.StatusCode).To(Equal(201))
		})

		It("should return 400 when resource name is too long", func() {
			// given
			res := &unversioned.Resource{
				Meta: rest_v1alpha1.ResourceMeta{
					Name: "super-long-resource-name-that-exceed-the-length-of-the-resource" +
						"-that-can-be-created-on-the-global-control-plane" +
						"we-want-to-check-if-resource-prefixed-with-zone-and-suffixed-with-namespace" +
						"can-fit-into-global-control-plane",
					Mesh: mesh,
					Type: string(core_mesh.DataplaneType),
				},
				Spec: &mesh_proto.Dataplane{
					Networking: &mesh_proto.Dataplane_Networking{
						Address:           "192.168.1.1",
						AdvertisedAddress: "192.168.1.1",
						Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
							{
								Port:        8000,
								ServicePort: 8080,
								Tags: map[string]string{
									mesh_proto.ServiceTag: "service-1",
								},
							},
						},
					},
				},
			}

			// when
			response := client.put(res)

			// then
			Expect(response.StatusCode).To(Equal(400))

			// and
			bytes, err := io.ReadAll(response.Body)
			Expect(err).ToNot(HaveOccurred())
			Expect(bytes).To(matchers.MatchGoldenJSON(path.Join("testdata", "resource_name-too-long.golden.json")))
		})
	})
})

var _ = Describe("Resource Endpoints", func() {
	var apiServer *api_server.ApiServer
	var resourceStore core_store.ResourceStore
	var client resourceApiClient
	stop := func() {}
	var metrics core_metrics.Metrics

	const mesh = "default"

	BeforeEach(func() {
		resourceStore = core_store.NewPaginationStore(memory.NewStore())
		apiServer, _, stop = StartApiServer(NewTestApiServerConfigurer().WithStore(resourceStore).WithMetrics(func() core_metrics.Metrics {
			m, _ := core_metrics.NewMetrics("Zone")
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
		err := resourceStore.Create(context.Background(), core_mesh.NewMeshResource(), core_store.CreateByKey(mesh, model.NoMesh))
		Expect(err).ToNot(HaveOccurred())
	})

	Describe("On GET", func() {
		It("should return an existing resource", func() {
			// given
			putSampleResourceIntoStore(resourceStore, "tr-1", mesh, "foo", "bar")

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
				"labels": {
					"foo": "bar"
				},
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
			Expect(bytes).To(matchers.MatchGoldenJSON(path.Join("testdata", "resource_not-found.golden.json")))
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

		It("should list external services with filters", func() {
			esWithTags := func(svc string, kv ...string) *core_mesh.ExternalServiceResource {
				tags := map[string]string{
					"kuma.io/service": svc,
				}
				for i := 0; i < len(kv); i += 2 {
					tags[kv[i]] = kv[i+1]
				}
				return &core_mesh.ExternalServiceResource{
					Spec: &mesh_proto.ExternalService{
						Tags: tags,
					},
				}
			}
			// given three resources
			for i := 0; i < 3; i++ {
				err := resourceStore.Create(context.Background(), esWithTags("my-svc"), core_store.CreateByKey(fmt.Sprintf("dp-%02d", i), mesh))
				Expect(err).NotTo(HaveOccurred())
			}
			err := resourceStore.Create(context.Background(), esWithTags("other-svc"), core_store.CreateByKey("dp-not-good", mesh))
			Expect(err).NotTo(HaveOccurred())

			// when ask for dataplanes with "my-svc" filter
			client = resourceApiClient{
				address: apiServer.Address(),
				path:    "/external-services?tag=kuma.io/service:my-svc",
			}
			response := client.list()
			Expect(response).To(MatchListResponse(TestListResponse{
				Total: 3,
				Next:  "",
				Items: []TestMeta{
					{
						Mesh: "default",
						Name: "dp-00",
						Type: "ExternalService",
					},
					{
						Mesh: "default",
						Name: "dp-01",
						Type: "ExternalService",
					},
					{
						Mesh: "default",
						Name: "dp-02",
						Type: "ExternalService",
					},
				},
			}))
		})

		It("should list dp with tag filters", func() {
			dpWithService := func(n string) *core_mesh.DataplaneResource {
				return &core_mesh.DataplaneResource{
					Spec: &mesh_proto.Dataplane{
						Networking: &mesh_proto.Dataplane_Networking{
							Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
								{
									Tags: map[string]string{"kuma.io/service": n},
								},
							},
						},
					},
				}
			}
			// given three resources
			for i := 0; i < 3; i++ {
				err := resourceStore.Create(context.Background(), dpWithService("my-svc"), core_store.CreateByKey(fmt.Sprintf("dp-%02d", i), mesh))
				Expect(err).NotTo(HaveOccurred())
			}
			err := resourceStore.Create(context.Background(), dpWithService("other-svc"), core_store.CreateByKey("dp-not-good", mesh))
			Expect(err).NotTo(HaveOccurred())

			// when ask for dataplanes with "my-svc" filter
			client = resourceApiClient{
				address: apiServer.Address(),
				path:    "/dataplanes?tag=kuma.io/service:my-svc",
			}
			response := client.list()
			Expect(response).To(MatchListResponse(TestListResponse{
				Total: 3,
				Next:  "",
				Items: []TestMeta{
					{
						Mesh: "default",
						Name: "dp-00",
						Type: "Dataplane",
					},
					{
						Mesh: "default",
						Name: "dp-01",
						Type: "Dataplane",
					},
					{
						Mesh: "default",
						Name: "dp-02",
						Type: "Dataplane",
					},
				},
			}))
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
			Expect(bytes).To(matchers.MatchGoldenJSON(path.Join("testdata", "resource_page-offset-invalid.golden.json")))
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
			Expect(bytes).To(matchers.MatchGoldenJSON(path.Join("testdata", "resource_page-size-invalid.golden.json")))
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
			Expect(bytes).To(matchers.MatchGoldenJSON(path.Join("testdata", "resource_page-size-too-large.golden.json")))
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
					Labels: map[string]string{
						"foo": "bar",
					},
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

			// and then
			resource := core_mesh.NewTrafficRouteResource()
			err := resourceStore.Get(context.Background(), resource, core_store.GetByKey("new-resource", mesh))
			Expect(err).ToNot(HaveOccurred())
			Expect(resource.Spec.Conf.Destination["path"]).To(Equal("/sample-path"))
			Expect(resource.Meta.GetLabels()).To(Equal(map[string]string{
				"foo":            "bar",
				"kuma.io/origin": "zone",
			}))
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
					Labels: map[string]string{
						"foo":      "barbar",
						"newlabel": "newvalue",
					},
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
			err := resourceStore.Get(context.Background(), resource, core_store.GetByKey(name, mesh))
			Expect(err).ToNot(HaveOccurred())
			Expect(resource.Spec.Conf.Destination["path"]).To(Equal("/update-sample-path"))
			Expect(resource.Meta.GetLabels()).To(Equal(map[string]string{"foo": "barbar", "newlabel": "newvalue"}))
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
								"action": "Allow"
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
			Expect(bytes).To(matchers.MatchGoldenJSON(path.Join("testdata", "resource_incoherent-types.golden.json")))
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
			Expect(bytes).To(matchers.MatchGoldenJSON(path.Join("testdata", "resource_incoherent-names.golden.json")))
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
			Expect(bytes).To(matchers.MatchGoldenJSON(path.Join("testdata", "resource_incoherent-mesh.golden.json")))
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
			bytes, err := io.ReadAll(response.Body)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(bytes).To(matchers.MatchGoldenJSON(path.Join("testdata", "resource_invalid-conf.golden.json")))
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
				path:    "/meshes/default/traffic-routes",
			}
			response := client.putJson("invalid@", []byte(json))

			// then
			Expect(response.StatusCode).To(Equal(400))

			// when
			bytes, err := io.ReadAll(response.Body)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(bytes).To(matchers.MatchGoldenJSON(path.Join("testdata", "resource_name-mesh.golden.json")))
		})

		It("should return 400 when json is invalid", func() {
			// given
			json := `{"foo": }`

			// when
			response := client.putJson("sample", []byte(json))

			// when
			bytes, err := io.ReadAll(response.Body)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(response.StatusCode).To(Equal(http.StatusBadRequest))
			Expect(bytes).To(matchers.MatchGoldenJSON(path.Join("testdata", "resource_400onInvalidJson.golden.json")))
		})

		It("should return 400 when resourceType is empty", func() {
			// given
			json := `{"type": "", "name": "foo"}`

			// when
			response := client.putJson("sample", []byte(json))

			// when
			bytes, err := io.ReadAll(response.Body)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(response.StatusCode).To(Equal(http.StatusBadRequest))
			Expect(bytes).To(matchers.MatchGoldenJSON(path.Join("testdata", "resource_400onEmptyResourceType.golden.json")))
		})

		It("should return 400 when meta is invalid", func() {
			// given
			json := `{"type": "TrafficRoute", "name": 4}`

			// when
			response := client.putJson("sample", []byte(json))

			// when
			bytes, err := io.ReadAll(response.Body)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(response.StatusCode).To(Equal(http.StatusBadRequest))
			Expect(bytes).To(matchers.MatchGoldenJSON(path.Join("testdata", "resource_400onInvalidMetaSchema.golden.json")))
		})

		It("should return 400 when spec is invalid", func() {
			// given
			json := `{"type": "TrafficRoute", "sample": "foo", "sources": [
					{
						"match": {
							"kuma.io/service": 4
						}
					}
				]}`

			// when
			response := client.putJson("sample", []byte(json))

			// when
			bytes, err := io.ReadAll(response.Body)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(response.StatusCode).To(Equal(http.StatusBadRequest))
			Expect(bytes).To(matchers.MatchGoldenJSON(path.Join("testdata", "resource_400onInvalidSpecSchema.golden.json")))
		})

		It("should return 400 when spec is invalid on new policies", func() {
			// given
			json := `{
				"type": "MeshTrafficPermission",
				"name": "sample",
				"spec": {
					"targetRef": {
						"kind": "MeshService",
						"name": 2
					},
					"from": [{"targetRef":{"kind":"Mesh"}}]
				}
			}`

			// when
			cl := resourceApiClient{
				address: apiServer.Address(),
				path:    "/meshes/" + mesh + "/meshtrafficpermissions",
			}
			response := cl.putJson("sample", []byte(json))

			// when
			bytes, err := io.ReadAll(response.Body)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(response.StatusCode).To(Equal(http.StatusBadRequest))
			Expect(bytes).To(matchers.MatchGoldenJSON(path.Join("testdata", "resource_400onInvalidSpecNewSchema.golden.json")))
		})

		It("should return 400 when meta type doesn't match on new policies", func() {
			// given
			json := `{
				"type": "TrafficPermission",
				"name": "sample",
				"spec": {
					"targetRef": {
						"kind": "Mesh"
					},
					"from": [{"targetRef":{"kind":"Mesh"}}]
				}
			}`

			// when
			cl := resourceApiClient{
				address: apiServer.Address(),
				path:    "/meshes/" + mesh + "/meshtrafficpermissions",
			}
			response := cl.putJson("sample", []byte(json))

			// when
			bytes, err := io.ReadAll(response.Body)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(response.StatusCode).To(Equal(http.StatusBadRequest))
			Expect(bytes).To(matchers.MatchGoldenJSON(path.Join("testdata", "resource_400onInvalidMetaTypeWithNewSchema.golden.json")))
		})

		It("should return 400 when labels are invalid", func() {
			// given
			json := `{
				"type": "MeshTrafficPermission",
				"name": "sample",
				"mesh": "default",
				"labels": {
					"foo/bar/baz": "bar",
					"": "bar",
					"foo": "^*bar"
				},
				"spec": {
					"targetRef": {
						"kind": "Mesh"
					},
					"from": [{"targetRef":{"kind":"Mesh"}, "default":{"action":"Allow"}}]
				}
			}`

			// when
			cl := resourceApiClient{
				address: apiServer.Address(),
				path:    "/meshes/" + mesh + "/meshtrafficpermissions",
			}
			response := cl.putJson("sample", []byte(json))

			// when
			bytes, err := io.ReadAll(response.Body)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(response.StatusCode).To(Equal(http.StatusBadRequest))
			Expect(bytes).To(matchers.MatchGoldenJSON(path.Join("testdata", "resource_400onInvalidLabels.golden.json")))
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
			err := resourceStore.Get(context.Background(), resource, core_store.GetByKey(name, mesh))
			Expect(err).To(Equal(core_store.ErrorResourceNotFound(resource.Descriptor().Name, name, mesh)))
		})

		It("should delete non-existing resource", func() {
			// when
			response := client.delete("non-existing-resource")

			// then
			Expect(response.StatusCode).To(Equal(404))

			// and
			bytes, err := io.ReadAll(response.Body)
			Expect(err).ToNot(HaveOccurred())
			Expect(bytes).To(matchers.MatchGoldenJSON(path.Join("testdata", "resource-delete_not-found.golden.json")))
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

	DescribeTable("inspect for policies /meshes/{mesh}/{policyType}/{policyName}/_resources/dataplanes", func(inputFile string) {
		apiTest(inputFile, apiServer, resourceStore)
	}, test.EntriesForFolder("resources/inspect/policies/_resources/dataplanes"))

	DescribeTable("inspect dataplane rules /meshes/{mesh}/dataplanes/{dpName}/_rules", func(inputFile string) {
		format.MaxLength = 0
		apiTest(inputFile, apiServer, resourceStore)
	}, test.EntriesForFolder("resources/inspect/dataplanes/_rules"))

	DescribeTable("inspect meshgateway rules /meshes/{mesh}/meshgateways/{gwName}/_rules", func(inputFile string) {
		format.MaxLength = 0
		apiTest(inputFile, apiServer, resourceStore)
	}, test.EntriesForFolder("resources/inspect/meshgateways/_rules"))

	DescribeTable("resources CRUD", func(inputFile string) {
		format.MaxLength = 0
		apiTest(inputFile, apiServer, resourceStore)
	}, test.EntriesForFolder("resources/crud"))
})

var _ = Describe("Resource Endpoints on Zone, label origin", func() {
	createServer := func(federatedZone, validateOriginLabel bool) (*api_server.ApiServer, core_store.ResourceStore, func()) {
		store := core_store.NewPaginationStore(memory.NewStore())
		zone := ""
		if federatedZone {
			zone = "zone-1"
		}
		apiServer, _, stop := StartApiServer(
			NewTestApiServerConfigurer().
				WithStore(store).
				WithDisableOriginLabelValidation(!validateOriginLabel).
				WithZone(zone),
		)
		return apiServer, store, stop
	}

	createMesh := func(s core_store.ResourceStore, mesh string) {
		// create default mesh
		err := s.Create(context.Background(), core_mesh.NewMeshResource(), core_store.CreateByKey(mesh, model.NoMesh))
		Expect(err).ToNot(HaveOccurred())
	}

	It("should return 400 when origin validation is enabled and origin label is not set", func() {
		// given
		apiServer, store, stop := createServer(true, true)
		defer stop()
		createMesh(store, "mesh-1")
		client := resourceApiClient{address: apiServer.Address(), path: "/meshes/mesh-1/meshtrafficpermissions"}

		// when
		res := &rest_v1alpha1.Resource{
			ResourceMeta: rest_v1alpha1.ResourceMeta{
				Name: "mtp-1",
				Mesh: "mesh-1",
				Type: string(v1alpha1.MeshTrafficPermissionType),
			},
			Spec: builders.MeshTrafficPermission().
				WithTargetRef(builders.TargetRefMesh()).
				AddFrom(builders.TargetRefMesh(), v1alpha1.Allow).
				Build().Spec,
		}
		resp := client.put(res)

		// then
		Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
		// and then
		bytes, err := io.ReadAll(resp.Body)
		Expect(err).ToNot(HaveOccurred())
		Expect(bytes).To(matchers.MatchGoldenJSON(path.Join("testdata", "resource_400onNoOriginLabel.golden.json")))
	})

	DescribeTable("should set origin label automatically when origin validation is disabled",
		func(federatedZone bool) {
			// given
			apiServer, store, stop := createServer(federatedZone, false)
			defer stop()
			createMesh(store, "mesh-1")
			client := resourceApiClient{address: apiServer.Address(), path: "/meshes/mesh-1/meshtrafficpermissions"}

			// when
			res := &rest_v1alpha1.Resource{
				ResourceMeta: rest_v1alpha1.ResourceMeta{
					Name: "mtp-1",
					Mesh: "mesh-1",
					Type: string(v1alpha1.MeshTrafficPermissionType),
				},
				Spec: builders.MeshTrafficPermission().
					WithTargetRef(builders.TargetRefMesh()).
					AddFrom(builders.TargetRefMesh(), v1alpha1.Allow).
					Build().Spec,
			}
			resp := client.put(res)

			// then
			Expect(resp.StatusCode).To(Equal(http.StatusCreated))
			// and then
			actualMtp := v1alpha1.NewMeshTrafficPermissionResource()
			Expect(store.Get(context.Background(), actualMtp, core_store.GetByKey("mtp-1", "mesh-1"))).To(Succeed())
			Expect(actualMtp.Meta.GetLabels()).To(Equal(map[string]string{
				mesh_proto.ResourceOriginLabel: string(mesh_proto.ZoneResourceOrigin),
			}))
		},
		Entry("non-federated zone", false),
		Entry("federated zone", true),
	)

	It("should set origin label automatically for DPPs", func() {
		// given
		apiServer, store, stop := createServer(false, false)
		defer stop()
		createMesh(store, "mesh-1")
		client := resourceApiClient{address: apiServer.Address(), path: "/meshes/mesh-1/dataplanes"}

		// when
		res := &unversioned.Resource{
			Meta: rest_v1alpha1.ResourceMeta{
				Name: "dpp-1",
				Mesh: "mesh-1",
				Type: string(core_mesh.DataplaneType),
			},
			Spec: builders.Dataplane().
				WithName("backend-1").
				WithHttpServices("backend").
				AddOutboundsToServices("redis", "elastic", "postgres").
				Build().Spec,
		}
		resp := client.put(res)

		// then
		Expect(resp.StatusCode).To(Equal(http.StatusCreated))
		// and then
		actualDpp := core_mesh.NewDataplaneResource()
		Expect(store.Get(context.Background(), actualDpp, core_store.GetByKey("dpp-1", "mesh-1"))).To(Succeed())
		Expect(actualDpp.Meta.GetLabels()).To(Equal(map[string]string{
			mesh_proto.ResourceOriginLabel: string(mesh_proto.ZoneResourceOrigin),
		}))
	})
})
