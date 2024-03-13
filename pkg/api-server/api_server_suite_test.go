package api_server_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/emicklei/go-restful/v3"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	api_server "github.com/kumahq/kuma/pkg/api-server"
	"github.com/kumahq/kuma/pkg/api-server/customization"
	config_api_server "github.com/kumahq/kuma/pkg/config/api-server"
	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	config_core "github.com/kumahq/kuma/pkg/config/core"
	config_manager "github.com/kumahq/kuma/pkg/core/config/manager"
	resources_access "github.com/kumahq/kuma/pkg/core/resources/access"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/model/rest"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/runtime"
	"github.com/kumahq/kuma/pkg/dns/vips"
	"github.com/kumahq/kuma/pkg/envoy/admin/access"
	"github.com/kumahq/kuma/pkg/insights/globalinsight"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
	"github.com/kumahq/kuma/pkg/plugins/authn/api-server/certs"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/pkg/test/matchers"
	test_runtime "github.com/kumahq/kuma/pkg/test/runtime"
	test_store "github.com/kumahq/kuma/pkg/test/store"
	"github.com/kumahq/kuma/pkg/tokens/builtin"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	"github.com/kumahq/kuma/pkg/xds/server"
)

func TestWs(t *testing.T) {
	test.RunSpecs(t, "API Server")
}

type resourceApiClient struct {
	address string
	path    string
}

type TestMeta struct {
	Type string `json:"type"`
	Name string `json:"name"`
	Mesh string `json:"mesh"`
}

type TestListResponse struct {
	Total int        `json:"total"`
	Next  string     `json:"next"`
	Items []TestMeta `json:"items"`
}

func MatchListResponse(r TestListResponse) types.GomegaMatcher {
	return And(
		HaveHTTPStatus(http.StatusOK),
		WithTransform(func(response *http.Response) (TestListResponse, error) {
			res := TestListResponse{}
			body, err := io.ReadAll(response.Body)
			if err != nil {
				return res, nil
			}
			if err := json.Unmarshal(body, &res); err != nil {
				return res, err
			}
			return res, nil
		}, Equal(r)),
	)
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
	jsonBytes, err := json.Marshal(res)
	Expect(err).ToNot(HaveOccurred())
	return r.putJson(res.GetMeta().Name, jsonBytes)
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

func putSampleResourceIntoStore(resourceStore store.ResourceStore, name string, mesh string, keyAndValue ...string) {
	resource := core_mesh.TrafficRouteResource{
		Spec: &mesh_proto.TrafficRoute{
			Conf: &mesh_proto.TrafficRoute_Conf{
				Destination: map[string]string{
					"path": "/sample-path",
				},
			},
		},
	}
	labels := map[string]string{}
	for i := 0; i < len(keyAndValue); i += 2 {
		labels[keyAndValue[i]] = keyAndValue[i+1]
	}
	err := resourceStore.Create(context.Background(), &resource, store.CreateByKey(name, mesh), store.CreateWithLabels(labels))
	Expect(err).NotTo(HaveOccurred())
}

type testApiServerConfigurer struct {
	store                        store.ResourceStore
	config                       *config_api_server.ApiServerConfig
	metrics                      func() core_metrics.Metrics
	zone                         string
	global                       bool
	disableOriginLabelValidation bool
}

func NewTestApiServerConfigurer() *testApiServerConfigurer {
	t := &testApiServerConfigurer{
		metrics: func() core_metrics.Metrics {
			m, _ := core_metrics.NewMetrics("Zone")
			return m
		},
		config: config_api_server.DefaultApiServerConfig(),
		store:  memory.NewStore(),
	}
	t.config.GUI.Enabled = false
	return t
}

func (t *testApiServerConfigurer) WithZone(z string) *testApiServerConfigurer {
	t.global = false
	t.zone = z
	return t
}

func (t *testApiServerConfigurer) WithGlobal() *testApiServerConfigurer {
	t.zone = ""
	t.global = true
	return t
}

func (t *testApiServerConfigurer) WithStore(resourceStore store.ResourceStore) *testApiServerConfigurer {
	t.store = resourceStore
	return t
}

func (t *testApiServerConfigurer) WithDisableOriginLabelValidation(disable bool) *testApiServerConfigurer {
	t.disableOriginLabelValidation = disable
	return t
}

// WithMetrics a function that creates metrics (needs to be a function as these can't be reused in case of failed startups)
func (t *testApiServerConfigurer) WithMetrics(metricsFn func() core_metrics.Metrics) *testApiServerConfigurer {
	t.metrics = metricsFn
	return t
}

func (t *testApiServerConfigurer) WithConfigMutator(fn func(*config_api_server.ApiServerConfig)) *testApiServerConfigurer {
	fn(t.config)
	return t
}

func StartApiServer(t *testApiServerConfigurer) (*api_server.ApiServer, kuma_cp.Config, func()) {
	var apiServer *api_server.ApiServer
	var cfg kuma_cp.Config
	var stop func()

	Eventually(func() error {
		var err error
		apiServer, cfg, stop, err = tryStartApiServer(t)
		return err
	}).
		WithTimeout(time.Second * 30).
		WithPolling(time.Millisecond * 500).
		WithOffset(1).
		Should(Succeed())

	return apiServer, cfg, stop
}

func tryStartApiServer(t *testApiServerConfigurer) (*api_server.ApiServer, kuma_cp.Config, func(), error) {
	ctx, stop := context.WithCancel(context.Background())
	// we have to manually search for port and put it into config. There is no way to retrieve port of running
	// http.Server and we need it later for the client
	port, err := test.GetFreePort()
	if err != nil {
		return nil, kuma_cp.Config{}, stop, err
	}
	t.config.HTTP.Port = uint32(port)

	port, err = test.GetFreePort()
	if err != nil {
		return nil, kuma_cp.Config{}, stop, err
	}
	t.config.HTTPS.Port = uint32(port)
	if t.config.HTTPS.TlsKeyFile == "" {
		t.config.HTTPS.TlsKeyFile = filepath.Join("..", "..", "test", "certs", "server-key.pem")
		t.config.HTTPS.TlsCertFile = filepath.Join("..", "..", "test", "certs", "server-cert.pem")
		t.config.Auth.ClientCertsDir = filepath.Join("..", "..", "test", "certs", "client")
	}

	cfg := kuma_cp.DefaultConfig()
	cfg.ApiServer = t.config
	if t.zone != "" {
		cfg.Mode = config_core.Zone
		cfg.Multizone.Zone.Name = t.zone
		cfg.Multizone.Zone.GlobalAddress = "grpcs://global:5685"
	} else if t.global {
		cfg.Mode = config_core.Global
	}

	cfg.Multizone.Zone.DisableOriginLabelValidation = t.disableOriginLabelValidation

	resManager := manager.NewResourceManager(t.store)
	apiServer, err := api_server.NewApiServer(
		resManager,
		xds_context.NewMeshContextBuilder(
			resManager,
			server.MeshResourceTypes(),
			net.LookupIP,
			cfg.Multizone.Zone.Name,
			vips.NewPersistence(resManager, config_manager.NewConfigManager(t.store), false),
			cfg.DNSServer.Domain,
			80,
			xds_context.AnyToAnyReachableServicesGraphBuilder,
		),
		customization.NewAPIList(),
		registry.Global().ObjectDescriptors(model.HasWsEnabled()),
		&cfg,
		t.metrics(),
		func() string { return "instance-id" },
		func() string { return "cluster-id" },
		certs.ClientCertAuthenticator,
		runtime.Access{
			ResourceAccess:       resources_access.NewAdminResourceAccess(cfg.Access.Static.AdminResources),
			DataplaneTokenAccess: nil,
			EnvoyAdminAccess: access.NewStaticEnvoyAdminAccess(
				cfg.Access.Static.ViewConfigDump,
				cfg.Access.Static.ViewStats,
				cfg.Access.Static.ViewClusters,
			),
		},
		&test_runtime.DummyEnvoyAdminClient{},
		builtin.TokenIssuers{
			DataplaneToken: builtin.NewDataplaneTokenIssuer(resManager),
			ZoneToken:      builtin.NewZoneTokenIssuer(resManager),
		},
		func(*restful.WebService) error { return nil },
		globalinsight.NewDefaultGlobalInsightService(t.store),
	)
	if err != nil {
		return nil, cfg, stop, err
	}
	errChan := make(chan error)
	go func() {
		err := apiServer.Start(ctx.Done()) //nolint:contextcheck
		errChan <- err
	}()

	tick := time.NewTicker(time.Millisecond * 500)
	leftTicks := 10
	for {
		if leftTicks == 0 {
			stop()
			return nil, cfg, stop, errors.New("no more ticks left")
		}
		select {
		case err = <-errChan:
			return nil, cfg, stop, err
		case <-tick.C:
			leftTicks--
			r, err := http.Get("http://" + apiServer.Address() + "/config")
			if err != nil {
				return nil, cfg, stop, err
			}
			r.Body.Close()
			if r.StatusCode == http.StatusOK {
				return apiServer, cfg, stop, nil
			}
		}
	}
}

// apiTest takes an `<testName>.input.yaml` which contains as first line a comment #<urlToRun> <statusCode> and then a set of yaml to preload in the resourceStore.
// It will then run against the apiServer check that the status code matches and that the output matches `<testName>.golden.json`
// Combined with `test.EntriesForFolder` and `ginkgo.DescribeTable` is a way to create a lot of api tests by only creating the input files.
func apiTest(inputResourceFile string, apiServer *api_server.ApiServer, resourceStore store.ResourceStore) {
	inputs, err := os.ReadFile(inputResourceFile)
	Expect(err).NotTo(HaveOccurred())
	parts := strings.SplitN(string(inputs), "\n", 2)
	Expect(parts[0]).To(HavePrefix("#"), "the first line of the input is not a comment with the url path")
	actions := strings.Split(strings.Trim(parts[0], "# "), " ")
	Expect(actions).To(HaveLen(2), "the first line of the input should be: # <path> <statusCode>")
	url := fmt.Sprintf("http://%s%s", apiServer.Address(), actions[0])
	Expect(url).ToNot(BeEmpty())
	status, err := strconv.Atoi(actions[1])
	Expect(err).NotTo(HaveOccurred(), "status is not an int")

	Expect(test_store.LoadResources(context.Background(), resourceStore, string(inputs))).To(Succeed())

	req, err := http.NewRequest("GET", url, nil)
	Expect(err).NotTo(HaveOccurred())

	response, err := http.DefaultClient.Do(req)
	Expect(err).NotTo(HaveOccurred())
	defer response.Body.Close()
	Expect(response).To(HaveHTTPStatus(status))

	// then
	b, err := io.ReadAll(response.Body)
	result := strings.ReplaceAll(string(b), apiServer.Address(), "{{address}}")
	Expect(err).ToNot(HaveOccurred())
	goldenFile := strings.ReplaceAll(inputResourceFile, ".input.yaml", ".golden.json")

	Expect(result).To(matchers.MatchGoldenJSON(goldenFile))
}
