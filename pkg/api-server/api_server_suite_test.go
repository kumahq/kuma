package api_server_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	api_server "github.com/kumahq/kuma/pkg/api-server"
	"github.com/kumahq/kuma/pkg/api-server/customization"
	config_access "github.com/kumahq/kuma/pkg/config/access"
	config_api_server "github.com/kumahq/kuma/pkg/config/api-server"
	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	config_core "github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/pkg/core/access"
	config_manager "github.com/kumahq/kuma/pkg/core/config/manager"
	resources_access "github.com/kumahq/kuma/pkg/core/resources/access"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/runtime"
	"github.com/kumahq/kuma/pkg/dns/vips"
	envoyadmin_access "github.com/kumahq/kuma/pkg/envoy/admin/access"
	"github.com/kumahq/kuma/pkg/insights/globalinsight"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
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

type testApiServerConfigurer struct {
	store                        store.ResourceStore
	config                       *config_api_server.ApiServerConfig
	metrics                      func() core_metrics.Metrics
	zone                         string
	global                       bool
	disableOriginLabelValidation bool
	accessConfigMutator          func(config *config_access.AccessConfig)
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

func (t *testApiServerConfigurer) WithAccessConfigMutator(fn func(config *config_access.AccessConfig)) *testApiServerConfigurer {
	t.accessConfigMutator = fn
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
	if t.accessConfigMutator != nil {
		t.accessConfigMutator(&cfg.Access)
	}

	cfg.Multizone.Zone.DisableOriginLabelValidation = t.disableOriginLabelValidation

	resManager := manager.NewResourceManager(t.store)
	apiServer, err := api_server.NewApiServer(
		test_runtime.NewTestRuntime(
			resManager,
			cfg,
			t.metrics(),
			customization.NewAPIList(),
			runtime.Access{
				ResourceAccess:       resources_access.NewAdminResourceAccess(cfg.Access.Static.AdminResources),
				DataplaneTokenAccess: nil,
				EnvoyAdminAccess: envoyadmin_access.NewStaticEnvoyAdminAccess(
					cfg.Access.Static.ViewConfigDump,
					cfg.Access.Static.ViewStats,
					cfg.Access.Static.ViewClusters,
				),
				ControlPlaneMetadataAccess: access.NewStaticControlPlaneMetadataAccess(cfg.Access.Static.ControlPlaneMetadata),
			},
			builtin.TokenIssuers{
				DataplaneToken: builtin.NewDataplaneTokenIssuer(resManager),
				ZoneToken:      builtin.NewZoneTokenIssuer(resManager),
			},
			globalinsight.NewDefaultGlobalInsightService(t.store),
		),
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
		registry.Global().ObjectDescriptors(model.HasWsEnabled()),
		&cfg,
		nil,
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

type action struct {
	path   string
	status int
	method string
}

func (a action) String() string {
	return fmt.Sprintf("#%s %d method=%s", a.path, a.status, a.method)
}

func parseAction(in string) (action, error) {
	out := action{
		method: http.MethodGet,
	}
	actions := strings.Split(strings.Trim(in, "# "), " ")
	for i := range actions {
		switch i {
		case 0:
			out.path = actions[i]
		case 1:
			s, err := strconv.Atoi(actions[i])
			if err != nil {
				return out, errors.New("status code is not a number")
			}
			out.status = s
		default:
			opts := strings.Split(actions[i], "=")
			switch opts[0] {
			case "method":
				out.method = opts[1]
			default:
				return out, errors.New("unknown option: " + actions[i])
			}
		}
	}

	return out, nil
}

// apiTest takes an `<testName>.input.yaml` which contains as first line a comment #<urlToRun> <statusCode> and then a set of yaml to preload in the resourceStore.
// It will then run against the apiServer check that the status code matches and that the output matches `<testName>.golden.json`
// Combined with `test.EntriesForFolder` and `ginkgo.DescribeTable` is a way to create a lot of api tests by only creating the input files.
func apiTest(inputResourceFile string, apiServer *api_server.ApiServer, resourceStore store.ResourceStore) {
	memory.ClearStore(resourceStore)
	inputs, err := os.ReadFile(inputResourceFile)
	Expect(err).NotTo(HaveOccurred())
	parts := strings.Split(string(inputs), "\n")
	actions := []action{}
	for i, p := range parts {
		if strings.HasPrefix(p, "#") {
			if !strings.HasPrefix(p, "#/") {
				continue
			}
		} else {
			break
		}
		actionEntry, err := parseAction(p)
		Expect(err).ToNot(HaveOccurred(), fmt.Sprintf("failed parsing action, line:%d cmd:%s", i, p))
		actions = append(actions, actionEntry)
	}
	Expect(test_store.LoadResources(context.Background(), resourceStore, string(inputs))).To(Succeed())

	Expect(actions).ToNot(BeEmpty(), "You need at least one action")
	for i, act := range actions {
		ginkgo.By(fmt.Sprintf("when calling: %d %s", i, act))
		// Given
		url := fmt.Sprintf("http://%s%s", apiServer.Address(), act.path)
		var body io.Reader = nil
		if act.method == http.MethodPut || act.method == http.MethodPost {
			requestFile := strings.ReplaceAll(inputResourceFile, ".input.yaml", ".request.json")
			if len(actions) > 1 {
				requestFile = strings.ReplaceAll(requestFile, ".request.json", fmt.Sprintf("_%02d.request.json", i))
			}
			b, err := os.ReadFile(requestFile)
			Expect(err).NotTo(HaveOccurred())
			body = bytes.NewBuffer(b)
		}
		req, err := http.NewRequest(act.method, url, body)
		Expect(err).NotTo(HaveOccurred())
		if body != nil {
			req.Header.Add("content-type", "application/json")
		}

		// When
		response, err := http.DefaultClient.Do(req)
		Expect(err).NotTo(HaveOccurred())
		defer response.Body.Close()
		Expect(response).To(HaveHTTPStatus(act.status))

		// then
		b, err := io.ReadAll(response.Body)
		result := strings.ReplaceAll(string(b), apiServer.Address(), "{{address}}")
		// Cleanup times
		result = regexp.MustCompile("[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9][^\"]+").ReplaceAllString(result, "0001-01-01T00:00:00Z")
		Expect(err).ToNot(HaveOccurred())
		goldenFile := strings.ReplaceAll(inputResourceFile, ".input.yaml", ".golden.json")
		if len(actions) > 1 {
			goldenFile = strings.ReplaceAll(goldenFile, ".golden.json", fmt.Sprintf("_%02d.golden.json", i))
		}
		if result == "" {
			Expect(goldenFile).ToNot(BeAnExistingFile(), "golden file exists when result is empty")
		} else {
			Expect(result).To(matchers.MatchGoldenJSON(goldenFile))
		}
	}
}
