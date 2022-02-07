package api_server_test

import (
	"bytes"
	"context"
	"net"
	"net/http"
	"path/filepath"

	. "github.com/onsi/gomega"

	api_server "github.com/kumahq/kuma/pkg/api-server"
	"github.com/kumahq/kuma/pkg/api-server/customization"
	config_api_server "github.com/kumahq/kuma/pkg/config/api-server"
	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	config_manager "github.com/kumahq/kuma/pkg/core/config/manager"
	resources_access "github.com/kumahq/kuma/pkg/core/resources/access"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/model/rest"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/runtime"
	"github.com/kumahq/kuma/pkg/dns/vips"
	"github.com/kumahq/kuma/pkg/envoy/admin/access"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
	"github.com/kumahq/kuma/pkg/plugins/authn/api-server/certs"
	"github.com/kumahq/kuma/pkg/test"
	sample_proto "github.com/kumahq/kuma/pkg/test/apis/sample/v1alpha1"
	sample_model "github.com/kumahq/kuma/pkg/test/resources/apis/sample"
	test_runtime "github.com/kumahq/kuma/pkg/test/runtime"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	"github.com/kumahq/kuma/pkg/xds/server"
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
		ok := err == nil && response.StatusCode == 200
		if response != nil {
			Expect(response.Body.Close()).To(Succeed())
		}
		return ok
	}, "5s", "100ms").Should(BeTrue())
}

func putSampleResourceIntoStore(resourceStore store.ResourceStore, name string, mesh string) {
	resource := sample_model.TrafficRouteResource{
		Spec: &sample_proto.TrafficRoute{
			Path: "/sample-path",
		},
	}
	err := resourceStore.Create(context.Background(), &resource, store.CreateByKey(name, mesh))
	Expect(err).NotTo(HaveOccurred())
}

type configModifier func(config *kuma_cp.Config)

func createTestApiServer(
	store store.ResourceStore,
	config *config_api_server.ApiServerConfig,
	enableGUI bool,
	metrics core_metrics.Metrics,
	modifiers ...configModifier,
) *api_server.ApiServer {
	// we have to manually search for port and put it into config. There is no way to retrieve port of running
	// http.Server and we need it later for the client
	port, err := test.GetFreePort()
	Expect(err).NotTo(HaveOccurred())
	config.HTTP.Port = uint32(port)

	port, err = test.GetFreePort()
	Expect(err).NotTo(HaveOccurred())
	config.HTTPS.Port = uint32(port)
	if config.HTTPS.TlsKeyFile == "" {
		config.HTTPS.TlsKeyFile = filepath.Join("..", "..", "test", "certs", "server-key.pem")
		config.HTTPS.TlsCertFile = filepath.Join("..", "..", "test", "certs", "server-cert.pem")
		config.Auth.ClientCertsDir = filepath.Join("..", "..", "test", "certs", "client")
	}

	cfg := kuma_cp.DefaultConfig()
	cfg.ApiServer = config
	for _, modifier := range modifiers {
		modifier(&cfg)
	}

	apiServer, err := api_server.NewApiServer(
		manager.NewResourceManager(store),
		xds_context.NewMeshContextBuilder(
			manager.NewResourceManager(store),
			server.MeshResourceTypes(server.HashMeshExcludedResources),
			net.LookupIP,
			cfg.Multizone.Zone.Name,
			vips.NewPersistence(manager.NewResourceManager(store), config_manager.NewConfigManager(store)),
			cfg.DNSServer.Domain,
		),
		customization.NewAPIList(),
		append(registry.Global().ObjectDescriptors(model.HasWsEnabled()), sample_model.TrafficRouteResourceTypeDescriptor),
		&cfg,
		enableGUI,
		metrics,
		func() string { return "instance-id" },
		func() string { return "cluster-id" },
		certs.ClientCertAuthenticator,
		runtime.Access{
			ResourceAccess:       resources_access.NewAdminResourceAccess(cfg.Access.Static.AdminResources),
			DataplaneTokenAccess: nil,
			ConfigDumpAccess:     access.NewStaticConfigDumpAccess(cfg.Access.Static.ViewConfigDump),
		},
		&test_runtime.DummyEnvoyAdminClient{},
	)
	Expect(err).ToNot(HaveOccurred())
	return apiServer
}
