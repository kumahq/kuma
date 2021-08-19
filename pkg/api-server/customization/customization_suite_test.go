package customization_test

import (
	"path/filepath"
	"testing"

	. "github.com/onsi/gomega"

	api_server "github.com/kumahq/kuma/pkg/api-server"
	"github.com/kumahq/kuma/pkg/api-server/customization"
	config_api_server "github.com/kumahq/kuma/pkg/config/api-server"
	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
	"github.com/kumahq/kuma/pkg/test"
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

	resources := manager.NewResourceManager(store)

	getInstanceId := func() string { return "instance-id" }
	getClusterId := func() string { return "cluster-id" }

	if wsManager == nil {
		wsManager = customization.NewAPIList()
	}
	cfg := kuma_cp.DefaultConfig()
	cfg.ApiServer = config
	apiServer, err := api_server.NewApiServer(resources, wsManager, registry.Global().ObjectDescriptors(core_model.HasWsEnabled()), &cfg, enableGUI, metrics, getInstanceId, getClusterId)
	Expect(err).ToNot(HaveOccurred())
	return apiServer
}
