package mesh_test

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
	"net"
	"net/http"
	"time"
)

var _ = Describe("Traffic Route WS", func() {
	var apiServer *api_server.ApiServer
	var resourceStore store.ResourceStore

	BeforeEach(func() {
		port, err := GetFreePort()
		Expect(err).NotTo(HaveOccurred())

		config := api_server.ApiServerConfig{BindAddress: fmt.Sprintf("localhost:%d", port)}
		resourceStore = memory.NewStore()
		apiServer = api_server.NewApiServer(resourceStore, config)

		apiServer.Start()

		time.Sleep(100 * time.Millisecond)
	})

	AfterEach(func() {
		//err := apiServer.Stop()
		//Expect(err).NotTo(HaveOccurred())
	})

	const namespace = "default"

	Describe("On GET", func() {
		It("should return an existing resource", func() {
			// given
			resource := sample_model.TrafficRouteResource{
				Spec: sample_proto.TrafficRoute{
					Path: "/sample-path",
				},
			}
			err := resourceStore.Create(context.Background(), &resource, store.CreateByName(namespace, "tr-1"))
			Expect(err).NotTo(HaveOccurred())

			// when
			response, err := http.Get("http://" + apiServer.Address() + "/meshes/default/traffic-routes/tr-1")
			Expect(err).NotTo(HaveOccurred())

			// then
			Expect(response.StatusCode).To(Equal(200))
			// expect body
		})

		It("should return 404 for non existing resource", func() {
			// when
			response, err := http.Get("http://" + apiServer.Address() + "/meshes/default/traffic-routes/tr-1")
			Expect(err).NotTo(HaveOccurred())

			// then
			Expect(response.StatusCode).To(Equal(404))
		})

		It("should list resources", func() {
			// given
			// resource 1
			resource := sample_model.TrafficRouteResource{
				Spec: sample_proto.TrafficRoute{
					Path: "/sample-path",
				},
			}
			err := resourceStore.Create(context.Background(), &resource, store.CreateByName(namespace, "tr-1"))
			Expect(err).NotTo(HaveOccurred())

			// resource 2
			resource = sample_model.TrafficRouteResource{
				Spec: sample_proto.TrafficRoute{
					Path: "/sample-path",
				},
			}
			err = resourceStore.Create(context.Background(), &resource, store.CreateByName(namespace, "tr-2"))
			Expect(err).NotTo(HaveOccurred())

			// when
			response, err := http.Get("http://" + apiServer.Address() + "/meshes/default/traffic-routes")
			Expect(err).NotTo(HaveOccurred())

			// then
			Expect(response.StatusCode).To(Equal(200))
			// expect body
		})
	})

	putResource := func(name string, route *sample_proto.TrafficRoute) *http.Response {
		jsonBytes, err := json.Marshal(&route)
		Expect(err).ToNot(HaveOccurred())
		request, err := http.NewRequest(
			"PUT",
			"http://"+apiServer.Address()+"/meshes/default/traffic-routes/" + name,
			bytes.NewBuffer(jsonBytes),
		)
		Expect(err).ToNot(HaveOccurred())
		request.Header.Add("content-type", "application/json")
		response, err := http.DefaultClient.Do(request)
		Expect(err).ToNot(HaveOccurred())
		return response
	}

	Describe("On PUT", func() {


		It("should create a resource when one does not exist", func() {
			// given
			route := sample_proto.TrafficRoute{
				Path: "/sample-path",
			}

			// when
			response := putResource("new-resource", &route)

			// then
			Expect(response.StatusCode).To(Equal(201))
		})

		It("should update a resource when one already exist", func() {
			// given
			name := "tr-1"
			route := sample_proto.TrafficRoute{
				Path: "/sample-path",
			}
			response := putResource(name, &route)
			Expect(response.StatusCode).To(Equal(201))

			// when
			route = sample_proto.TrafficRoute{
				Path: "/update-sample-path",
			}
			response = putResource(name, &route)
			Expect(response.StatusCode).To(Equal(200))

			// then
		})
	})

	Describe("On DELETE", func() {
		It("should delete existing resource", func() {
			// given
			name := "tr-1"
			route := sample_proto.TrafficRoute{
				Path: "/sample-path",
			}
			putResource(name, &route)

			// when
			request, err := http.NewRequest(
				"DELETE",
				"http://"+apiServer.Address()+"/meshes/default/traffic-routes/" + name,
				nil,
			)
			Expect(err).ToNot(HaveOccurred())
			response, err := http.DefaultClient.Do(request)
			Expect(err).ToNot(HaveOccurred())

			// then
			Expect(response.StatusCode).To(Equal(200))

			// when
			response, err = http.Get("http://" + apiServer.Address() + "/meshes/default/traffic-routes/" + name)
			Expect(err).NotTo(HaveOccurred())

			// then
			Expect(response.StatusCode).To(Equal(404))
		})

		It("should delete non-existing resource", func() {

		})
	})
})

func GetFreePort() (int, error) {
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
