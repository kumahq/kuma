package api_server_test

import (
	"bytes"
	"context"
	"fmt"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/api-server"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/model/rest"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/store"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/plugins/resources/memory"
	sample_proto "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/test/apis/sample/v1alpha1"
	sample_model "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/test/resources/apis/sample"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"net"
	"net/http"
)

var _ = Describe("Resource WS", func() {
	var apiServer *api_server.ApiServer
	var resourceStore store.ResourceStore
	var client resourceApiClient

	const namespace = "default"
	const mesh = "default"

	BeforeEach(func() {
		resourceStore = memory.NewStore()
		apiServer = createTestApiServer(resourceStore, api_server.ApiServerConfig{})
		client = resourceApiClient{
			address: apiServer.Address(),
			mesh:    mesh,
		}
		apiServer.Start()
		waitForServer(&client)
	})

	AfterEach(func() {
		err := apiServer.Stop()
		Expect(err).NotTo(HaveOccurred())
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
			route := sample_proto.TrafficRoute{
				Path: "/sample-path",
			}

			// when
			response := client.put("new-resource", &route)

			// then
			Expect(response.StatusCode).To(Equal(201))
		})

		It("should update a resource when one already exist", func() {
			// given
			name := "tr-1"
			putSampleResourceIntoStore(resourceStore, name, mesh)

			// when
			route := sample_proto.TrafficRoute{
				Path: "/update-sample-path",
			}
			response := client.put(name, &route)
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

func waitForServer(client *resourceApiClient) {
	for {
		response, err := client.listOrError()
		if err == nil && response.StatusCode == 200 {
			return
		}
	}
}

func putSampleResourceIntoStore(resourceStore store.ResourceStore, name string, mesh string) {
	resource := sample_model.TrafficRouteResource{
		Spec: sample_proto.TrafficRoute{
			Path: "/sample-path",
		},
	}
	err := resourceStore.Create(context.Background(), &resource, store.CreateByKey("default", name, mesh))
	Expect(err).NotTo(HaveOccurred())
}

func createTestApiServer(store store.ResourceStore, config api_server.ApiServerConfig) *api_server.ApiServer {
	// we have to manually search for port and put it into config. There is no way to retrieve port of running
	// http.Server and we need it later for the client
	port, err := getFreePort()
	Expect(err).NotTo(HaveOccurred())
	config.BindAddress = fmt.Sprintf("localhost:%d", port)
	definitions := []api_server.ResourceWsDefinition{
		TrafficRouteWsDefinition,
	}
	return api_server.NewApiServer(store, definitions, config)
}

func getFreePort() (int, error) {
	ln, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, err
	}
	if err := ln.Close(); err != nil {
		return 0, err
	}
	return ln.Addr().(*net.TCPAddr).Port, nil
}

type resourceApiClient struct {
	address string
	mesh    string
}

func (r *resourceApiClient) fullAddress() string {
	return "http://" + r.address + "/meshes/" + r.mesh + "/traffic-routes"
}

func (r *resourceApiClient) get(name string) *http.Response {
	response, err := http.Get(r.fullAddress() + "/" + name)
	Expect(err).NotTo(HaveOccurred())
	return response
}

func (r *resourceApiClient) list() *http.Response {
	response, err := http.Get(r.fullAddress())
	Expect(err).NotTo(HaveOccurred())
	return response
}

func (r *resourceApiClient) listOrError() (*http.Response, error) {
	return http.Get(r.fullAddress())
}

func (r *resourceApiClient) delete(name string) *http.Response {
	request, err := http.NewRequest(
		"DELETE",
		r.fullAddress()+"/"+name,
		nil,
	)
	Expect(err).ToNot(HaveOccurred())
	response, err := http.DefaultClient.Do(request)
	Expect(err).ToNot(HaveOccurred())
	return response
}

func (r *resourceApiClient) put(name string, route *sample_proto.TrafficRoute) *http.Response {
	resResponse := rest.Resource{
		Meta: rest.ResourceMeta{
			Name: name,
			Type: string(sample_model.TrafficRouteType),
			Mesh: "default",
		},
		Spec: route,
	}
	jsonBytes, err := resResponse.MarshalJSON()
	Expect(err).ToNot(HaveOccurred())
	return r.putJson(name, jsonBytes)
}

func (r *resourceApiClient) putJson(name string, json []byte) *http.Response {
	request, err := http.NewRequest(
		"PUT",
		r.fullAddress()+"/"+name,
		bytes.NewBuffer(json),
	)
	Expect(err).ToNot(HaveOccurred())
	request.Header.Add("content-type", "application/json")
	response, err := http.DefaultClient.Do(request)
	Expect(err).ToNot(HaveOccurred())
	return response
}
