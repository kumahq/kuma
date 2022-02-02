package customization_test

import (
	"net"
	"path/filepath"
	"testing"

	. "github.com/onsi/gomega"

	api_server "github.com/kumahq/kuma/pkg/api-server"
	"github.com/kumahq/kuma/pkg/api-server/customization"
	config_api_server "github.com/kumahq/kuma/pkg/config/api-server"
	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	config_manager "github.com/kumahq/kuma/pkg/core/config/manager"
	resources_access "github.com/kumahq/kuma/pkg/core/resources/access"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/runtime"
	"github.com/kumahq/kuma/pkg/dns/vips"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
	"github.com/kumahq/kuma/pkg/plugins/authn/api-server/certs"
	"github.com/kumahq/kuma/pkg/test"
	test_runtime "github.com/kumahq/kuma/pkg/test/runtime"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	"github.com/kumahq/kuma/pkg/xds/server"
)

func TestWs(t *testing.T) {
	test.RunSpecs(t, "API Server Customization")
}

func createTestApiServer(store store.ResourceStore, config *config_api_server.ApiServerConfig, enableGUI bool, metrics core_metrics.Metrics, wsManager customization.APIManager) *api_server.ApiServer {
	// we have to manually search for port and put it into config. There is no way to retrieve port of running
	// http.Server and we need it later for the client
	port, err := test.GetFreePort()
	Expect(err).NotTo(HaveOccurred())
	config.HTTP.Port = uint32(port)

	port, err = test.GetFreePort()
	Expect(err).NotTo(HaveOccurred())
	config.HTTPS.Port = uint32(port)
	if config.HTTPS.TlsKeyFile == "" {
		config.HTTPS.TlsKeyFile = filepath.Join("..", "..", "..", "test", "certs", "server-key.pem")
		config.HTTPS.TlsCertFile = filepath.Join("..", "..", "..", "test", "certs", "server-cert.pem")
		config.Auth.ClientCertsDir = filepath.Join("..", "..", "..", "test", "certs", "client")
	}

	if wsManager == nil {
		wsManager = customization.NewAPIList()
	}
	cfg := kuma_cp.DefaultConfig()
	cfg.ApiServer = config
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
		wsManager,
		registry.Global().ObjectDescriptors(core_model.HasWsEnabled()),
		&cfg,
		enableGUI,
		metrics,
		func() string { return "instance-id" },
		func() string { return "cluster-id" },
		certs.ClientCertAuthenticator,
		runtime.Access{
			ResourceAccess:       resources_access.NewAdminResourceAccess(cfg.Access.Static.AdminResources),
			DataplaneTokenAccess: nil,
		},
		&test_runtime.DummyEnvoyAdminClient{},
	)
	Expect(err).ToNot(HaveOccurred())
	return apiServer
}
