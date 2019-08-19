package api_server_test

import (
	"bytes"
	"context"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/api-server"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/api-server/definitions"
	config "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/config/api-server"
	core_resources "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/store"
	sample_proto "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/test/apis/sample/v1alpha1"
	sample_model "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/test/resources/apis/sample"

	. "github.com/onsi/gomega"
	"net"
	"net/http"

	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/model/rest"
)

type resourceApiClient struct {
	address string
	path    string
}

func (r *resourceApiClient) fullAddress() string {
	return "http://" + r.address + r.path
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

func (r *resourceApiClient) put(res rest.Resource) *http.Response {
	jsonBytes, err := res.MarshalJSON()
	Expect(err).ToNot(HaveOccurred())
	return r.putJson(res.Meta.Name, jsonBytes)
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

func waitForServer(client *resourceApiClient) {
	Eventually(func() bool {
		response, err := client.listOrError()
		return err == nil && response.StatusCode == 200
	}, "5s", "100ms").Should(BeTrue())
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

func createTestApiServer(store store.ResourceStore, config config.ApiServerConfig) *api_server.ApiServer {
	// we have to manually search for port and put it into config. There is no way to retrieve port of running
	// http.Server and we need it later for the client
	port, err := getFreePort()
	Expect(err).NotTo(HaveOccurred())
	config.Port = port
	defs := []definitions.ResourceWsDefinition{
		TrafficRouteWsDefinition,
		definitions.MeshWsDefinition,
	}
	resources := core_resources.Resources{Store: store}
	return api_server.NewApiServer(resources, defs, config)
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
