package api_server

import (
	"net"

	api_server "github.com/kumahq/kuma/pkg/api-server"
	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	"github.com/kumahq/kuma/pkg/core/runtime"
	"github.com/kumahq/kuma/pkg/dns/vips"
	"github.com/kumahq/kuma/pkg/xds/context"
	"github.com/kumahq/kuma/pkg/xds/server"
)

func NewApiServer(cfg kuma_cp.Config, runtime runtime.Runtime) (*api_server.ApiServer, error) {
	return api_server.NewApiServer(
		runtime.ResourceManager(),
		context.NewMeshContextBuilder(
			runtime.ResourceManager(),
			server.MeshResourceTypes(),
			net.LookupIP,
			cfg.Multizone.Zone.Name,
			vips.NewPersistence(
				runtime.ResourceManager(),
				runtime.ConfigManager(),
				cfg.Experimental.UseTagFirstVirtualOutboundModel,
			),
			cfg.DNSServer.Domain,
			cfg.DNSServer.ServiceVipPort,
			context.AnyToAnyReachableServicesGraphBuilder,
			cfg.Experimental.SkipPersistedVIPs,
		),
		runtime.APIInstaller(),
		registry.Global().ObjectDescriptors(model.HasWsEnabled()),
		&cfg,
		runtime.Metrics(),
		runtime.GetInstanceId,
		runtime.GetClusterId,
		runtime.APIServerAuthenticator(),
		runtime.Access(),
		runtime.EnvoyAdminClient(),
		runtime.TokenIssuers(),
		runtime.APIWebServiceCustomize(),
		runtime.GlobalInsightService(),
		runtime.XDS().Hooks.ResourceSetHooks(),
	)
}
