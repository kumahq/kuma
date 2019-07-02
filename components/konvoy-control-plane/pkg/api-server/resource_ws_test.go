package api_server_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/api-server"
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

var _ = Describe("Traffic Route WS", func() {
	var apiServer *api_server.ApiServer
	var resourceStore store.ResourceStore
	var client resourceApiClient

	BeforeEach(func() {
		resourceStore = memory.NewStore()
		apiServer = createTestApiServer(resourceStore, api_server.ApiServerConfig{})
		client = resourceApiClient{address: apiServer.Address()}
		apiServer.Start()
	})

	AfterEach(func() {
		err := apiServer.Stop()
		Expect(err).NotTo(HaveOccurred())
	})

	const namespace = "default"

	Describe("On GET", func() {
		It("should return an existing resource", func() {
			// given
			putSampleResourceIntoStore(resourceStore, "tr-1")

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
			putSampleResourceIntoStore(resourceStore, "tr-1")
			putSampleResourceIntoStore(resourceStore, "tr-2")

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
			putSampleResourceIntoStore(resourceStore, name)

			// when
			route := sample_proto.TrafficRoute{
				Path: "/update-sample-path",
			}
			response := client.put(name, &route)
			Expect(response.StatusCode).To(Equal(200))

			// then
			resource := sample_model.TrafficRouteResource{}
			err := resourceStore.Get(context.Background(), &resource, store.GetByName(namespace, name))
			Expect(err).ToNot(HaveOccurred())
			Expect(resource.Spec.Path).To(Equal("/update-sample-path"))
		})
	})

	Describe("On DELETE", func() {
		It("should delete existing resource", func() {
			// given
			name := "tr-1"
			putSampleResourceIntoStore(resourceStore, name)

			// when
			response := client.delete(name)

			// then
			Expect(response.StatusCode).To(Equal(200))

			// and
			resource := sample_model.TrafficRouteResource{}
			err := resourceStore.Get(context.Background(), &resource, store.GetByName(namespace, name))
			Expect(err).To(Equal(store.ErrorResourceNotFound(resource.GetType(), namespace, name)))
		})

		It("should delete non-existing resource", func() {
			// when
			response := client.delete("non-existing-resource")

			// then
			Expect(response.StatusCode).To(Equal(200))
		})
	})
})

func putSampleResourceIntoStore(resourceStore store.ResourceStore, name string) {
	resource := sample_model.TrafficRouteResource{
		Spec: sample_proto.TrafficRoute{
			Path: "/sample-path",
		},
	}
	err := resourceStore.Create(context.Background(), &resource, store.CreateByName("default", name))
	Expect(err).NotTo(HaveOccurred())
}

func createTestApiServer(store store.ResourceStore, config api_server.ApiServerConfig) *api_server.ApiServer {
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
	err = ln.Close()
	if err != nil {
		return 0, err
	}
	return ln.Addr().(*net.TCPAddr).Port, nil
}

type resourceApiClient struct {
	address string
}

func (r *resourceApiClient) fullAddress() string {
	return "http://" + r.address + "/meshes/default/traffic-routes"
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

func (r *resourceApiClient) delete(name string) *http.Response {
	request, err := http.NewRequest(
		"DELETE",
		r.fullAddress() + "/" + name,
		nil,
	)
	Expect(err).ToNot(HaveOccurred())
	response, err := http.DefaultClient.Do(request)
	Expect(err).ToNot(HaveOccurred())
	return response
}

func (r *resourceApiClient) put(name string, route *sample_proto.TrafficRoute) *http.Response {
	jsonBytes, err := json.Marshal(&route)
	Expect(err).ToNot(HaveOccurred())
	request, err := http.NewRequest(
		"PUT",
		r.fullAddress() + "/" + name,
		bytes.NewBuffer(jsonBytes),
	)
	Expect(err).ToNot(HaveOccurred())
	request.Header.Add("content-type", "application/json")
	response, err := http.DefaultClient.Do(request)
	Expect(err).ToNot(HaveOccurred())
	return response
}
