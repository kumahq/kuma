package cmd

import (
	"context"
	"time"

	"github.com/pkg/errors"

	generate_context "github.com/kumahq/kuma/app/kumactl/cmd/generate/context"
	get_context "github.com/kumahq/kuma/app/kumactl/cmd/get/context"
	inspect_context "github.com/kumahq/kuma/app/kumactl/cmd/inspect/context"
	install_context "github.com/kumahq/kuma/app/kumactl/cmd/install/context"
	"github.com/kumahq/kuma/app/kumactl/pkg/client"
	"github.com/kumahq/kuma/app/kumactl/pkg/config"
	"github.com/kumahq/kuma/app/kumactl/pkg/plugins"
	kumactl_resources "github.com/kumahq/kuma/app/kumactl/pkg/resources"
	"github.com/kumahq/kuma/app/kumactl/pkg/tokens"
	"github.com/kumahq/kuma/pkg/api-server/types"
	config_proto "github.com/kumahq/kuma/pkg/config/app/kumactl/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/plugins/authn/api-server/tokens/cli"
	util_files "github.com/kumahq/kuma/pkg/util/files"
	util_http "github.com/kumahq/kuma/pkg/util/http"
	kuma_version "github.com/kumahq/kuma/pkg/version"
)

type RootArgs struct {
	ConfigFile string
	Mesh       string
	ApiTimeout time.Duration
}

type RootRuntime struct {
	Config                       config_proto.Configuration
	Now                          func() time.Time
	AuthnPlugins                 map[string]plugins.AuthnPlugin
	NewBaseAPIServerClient       func(*config_proto.ControlPlaneCoordinates_ApiServer, time.Duration) (util_http.Client, error)
	NewResourceStore             func(util_http.Client) core_store.ResourceStore
	NewDataplaneOverviewClient   func(util_http.Client) kumactl_resources.DataplaneOverviewClient
	NewDataplaneInspectClient    func(util_http.Client) kumactl_resources.DataplaneInspectClient
	NewInspectEnvoyProxyClient   func(core_model.ResourceTypeDescriptor, util_http.Client) kumactl_resources.InspectEnvoyProxyClient
	NewPolicyInspectClient       func(util_http.Client) kumactl_resources.PolicyInspectClient
	NewZoneIngressOverviewClient func(util_http.Client) kumactl_resources.ZoneIngressOverviewClient
	NewZoneEgressOverviewClient  func(util_http.Client) kumactl_resources.ZoneEgressOverviewClient
	NewZoneOverviewClient        func(util_http.Client) kumactl_resources.ZoneOverviewClient
	NewServiceOverviewClient     func(util_http.Client) kumactl_resources.ServiceOverviewClient
	NewDataplaneTokenClient      func(util_http.Client) tokens.DataplaneTokenClient
	NewZoneIngressTokenClient    func(util_http.Client) tokens.ZoneIngressTokenClient
	NewZoneTokenClient           func(util_http.Client) tokens.ZoneTokenClient
	NewAPIServerClient           func(util_http.Client) kumactl_resources.ApiServerClient
	Registry                     registry.TypeRegistry
}

// RootContext contains variables, functions and components that can be overridden when extending kumactl or running the test.
// Example:
//
// rootCtx := kumactl_cmd.DefaultRootContext()
// rootCtx.InstallCpContext.Args.ControlPlane_image_tag = "0.0.1"
// rootCmd := cmd.NewRootCmd(rootCtx)
// err := rootCmd.Execute()
type RootContext struct {
	Args                                RootArgs
	Runtime                             RootRuntime
	GetContext                          get_context.GetContext
	ListContext                         get_context.ListContext
	GenerateContext                     generate_context.GenerateContext
	InspectContext                      inspect_context.InspectContext
	InstallCpContext                    install_context.InstallCpContext
	InstallMetricsContext               install_context.InstallMetricsContext
	InstallCRDContext                   install_context.InstallCrdsContext
	InstallDemoContext                  install_context.InstallDemoContext
	InstallGatewayKongContext           install_context.InstallGatewayKongContext
	InstallGatewayKongEnterpriseContext install_context.InstallGatewayKongEnterpriseContext
	InstallTracingContext               install_context.InstallTracingContext
	InstallLoggingContext               install_context.InstallLoggingContext
}

func DefaultRootContext() *RootContext {
	return &RootContext{
		Runtime: RootRuntime{
			Now:                    time.Now,
			Registry:               registry.Global(),
			NewBaseAPIServerClient: client.ApiServerClient,
			AuthnPlugins: map[string]plugins.AuthnPlugin{
				cli.AuthType: &cli.TokenAuthnPlugin{},
			},
			NewResourceStore: func(client util_http.Client) core_store.ResourceStore {
				return kumactl_resources.NewResourceStore(client, registry.Global().ObjectDescriptors())
			},
			NewDataplaneOverviewClient:   kumactl_resources.NewDataplaneOverviewClient,
			NewDataplaneInspectClient:    kumactl_resources.NewDataplaneInspectClient,
			NewInspectEnvoyProxyClient:   kumactl_resources.NewInspectEnvoyProxyClient,
			NewPolicyInspectClient:       kumactl_resources.NewPolicyInspectClient,
			NewZoneIngressOverviewClient: kumactl_resources.NewZoneIngressOverviewClient,
			NewZoneEgressOverviewClient:  kumactl_resources.NewZoneEgressOverviewClient,
			NewZoneOverviewClient:        kumactl_resources.NewZoneOverviewClient,
			NewServiceOverviewClient:     kumactl_resources.NewServiceOverviewClient,
			NewDataplaneTokenClient:      tokens.NewDataplaneTokenClient,
			NewZoneIngressTokenClient:    tokens.NewZoneIngressTokenClient,
			NewZoneTokenClient:           tokens.NewZoneTokenClient,
			NewAPIServerClient:           kumactl_resources.NewAPIServerClient,
		},
		InstallCpContext:                    install_context.DefaultInstallCpContext(),
		InstallCRDContext:                   install_context.DefaultInstallCrdsContext(),
		InstallMetricsContext:               install_context.DefaultInstallMetricsContext(),
		InstallDemoContext:                  install_context.DefaultInstallDemoContext(),
		InstallGatewayKongContext:           install_context.DefaultInstallGatewayKongContext(),
		InstallGatewayKongEnterpriseContext: install_context.DefaultInstallGatewayKongEnterpriseContext(),
		InstallTracingContext:               install_context.DefaultInstallTracingContext(),
		InstallLoggingContext:               install_context.DefaultInstallLoggingContext(),
		GenerateContext:                     generate_context.DefaultGenerateContext(),
	}
}

func (rc *RootContext) LoadConfig() error {
	return config.Load(rc.Args.ConfigFile, &rc.Runtime.Config)
}

func (rc *RootContext) SaveConfig() error {
	return config.Save(rc.Args.ConfigFile, &rc.Runtime.Config)
}

func (rc *RootContext) Config() *config_proto.Configuration {
	return &rc.Runtime.Config
}

func (rc *RootContext) CurrentContext() (*config_proto.Context, error) {
	if rc.Config().CurrentContext == "" {
		return nil, errors.Errorf("active Control Plane is not set. Use `kumactl config control-planes add` to add a Control Plane and make it active")
	}
	_, currentContext := rc.Config().GetContext(rc.Config().CurrentContext)
	if currentContext == nil {
		return nil, errors.Errorf("apparently, configuration is broken. Use `kumactl config control-planes add` to add a Control Plane and make it active")
	}
	return currentContext, nil
}

func (rc *RootContext) CurrentControlPlane() (*config_proto.ControlPlane, error) {
	currentContext, err := rc.CurrentContext()
	if err != nil {
		return nil, err
	}
	_, controlPlane := rc.Config().GetControlPlane(currentContext.ControlPlane)
	if controlPlane == nil {
		return nil, errors.Errorf("apparently, configuration is broken. Use `kumactl config control-planes add` to add a Control Plane and make it active")
	}
	return controlPlane, nil
}

func (rc *RootContext) CurrentMesh() string {
	if rc.Args.Mesh != "" {
		return rc.Args.Mesh
	}
	return core_model.DefaultMesh
}

func (rc *RootContext) Now() time.Time {
	return rc.Runtime.Now()
}

func (rc *RootContext) BaseAPIServerClient() (util_http.Client, error) {
	controlPlane, err := rc.CurrentControlPlane()
	if err != nil {
		return nil, err
	}
	client, err := rc.Runtime.NewBaseAPIServerClient(controlPlane.Coordinates.ApiServer, rc.Args.ApiTimeout)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create a client for Control Plane %q", controlPlane.Name)
	}

	if controlPlane.Coordinates.ApiServer.AuthType != "" {
		plugin, ok := rc.Runtime.AuthnPlugins[controlPlane.Coordinates.ApiServer.AuthType]
		if !ok {
			return nil, errors.Errorf("authentication plugin of type %q not found", controlPlane.Coordinates.ApiServer.AuthType)
		}
		client, err = plugin.DecorateClient(client, controlPlane.Coordinates.ApiServer.AuthConf)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to decorate client with authentication type %q", controlPlane.Coordinates.ApiServer.AuthType)
		}
	}

	return client, nil
}

func (rc *RootContext) CurrentResourceStore() (core_store.ResourceStore, error) {
	client, err := rc.BaseAPIServerClient()
	if err != nil {
		return nil, err
	}
	return rc.Runtime.NewResourceStore(client), nil
}

func (rc *RootContext) CurrentDataplaneOverviewClient() (kumactl_resources.DataplaneOverviewClient, error) {
	client, err := rc.BaseAPIServerClient()
	if err != nil {
		return nil, err
	}
	return rc.Runtime.NewDataplaneOverviewClient(client), nil
}

func (rc *RootContext) CurrentDataplaneInspectClient() (kumactl_resources.DataplaneInspectClient, error) {
	client, err := rc.BaseAPIServerClient()
	if err != nil {
		return nil, err
	}
	return rc.Runtime.NewDataplaneInspectClient(client), nil
}

func (rc *RootContext) CurrentInspectEnvoyProxyClient(resDesc core_model.ResourceTypeDescriptor) (kumactl_resources.InspectEnvoyProxyClient, error) {
	client, err := rc.BaseAPIServerClient()
	if err != nil {
		return nil, err
	}
	return rc.Runtime.NewInspectEnvoyProxyClient(resDesc, client), nil
}

func (rc *RootContext) CurrentPolicyInspectClient() (kumactl_resources.PolicyInspectClient, error) {
	client, err := rc.BaseAPIServerClient()
	if err != nil {
		return nil, err
	}
	return rc.Runtime.NewPolicyInspectClient(client), nil
}

func (rc *RootContext) CurrentZoneOverviewClient() (kumactl_resources.ZoneOverviewClient, error) {
	client, err := rc.BaseAPIServerClient()
	if err != nil {
		return nil, err
	}
	return rc.Runtime.NewZoneOverviewClient(client), nil
}

func (rc *RootContext) CurrentZoneIngressOverviewClient() (kumactl_resources.ZoneIngressOverviewClient, error) {
	client, err := rc.BaseAPIServerClient()
	if err != nil {
		return nil, err
	}
	return rc.Runtime.NewZoneIngressOverviewClient(client), nil
}

func (rc *RootContext) CurrentZoneEgressOverviewClient() (kumactl_resources.ZoneEgressOverviewClient, error) {
	client, err := rc.BaseAPIServerClient()
	if err != nil {
		return nil, err
	}
	return rc.Runtime.NewZoneEgressOverviewClient(client), nil
}

func (rc *RootContext) CurrentServiceOverviewClient() (kumactl_resources.ServiceOverviewClient, error) {
	client, err := rc.BaseAPIServerClient()
	if err != nil {
		return nil, err
	}
	return rc.Runtime.NewServiceOverviewClient(client), nil
}

func (rc *RootContext) CurrentDataplaneTokenClient() (tokens.DataplaneTokenClient, error) {
	client, err := rc.BaseAPIServerClient()
	if err != nil {
		return nil, err
	}
	return rc.Runtime.NewDataplaneTokenClient(client), nil
}

func (rc *RootContext) CurrentZoneIngressTokenClient() (tokens.ZoneIngressTokenClient, error) {
	client, err := rc.BaseAPIServerClient()
	if err != nil {
		return nil, err
	}
	return rc.Runtime.NewZoneIngressTokenClient(client), nil
}

func (rc *RootContext) CurrentZoneTokenClient() (tokens.ZoneTokenClient, error) {
	client, err := rc.BaseAPIServerClient()
	if err != nil {
		return nil, err
	}
	return rc.Runtime.NewZoneTokenClient(client), nil
}

func (rc *RootContext) IsFirstTimeUsage() bool {
	if rc.Args.ConfigFile != "" {
		return !util_files.FileExists(rc.Args.ConfigFile)
	}
	return !util_files.FileExists(config.DefaultConfigFile)
}

func (rc *RootContext) CurrentApiClient() (kumactl_resources.ApiServerClient, error) {
	client, err := rc.BaseAPIServerClient()
	if err != nil {
		return nil, err
	}
	return rc.Runtime.NewAPIServerClient(client), nil
}

func (rc *RootContext) CheckServerVersionCompatibility() error {
	kumactlLog := core.Log.WithName("kumactl")

	var kumaBuildVersion *types.IndexResponse

	client, err := rc.CurrentApiClient()
	if err != nil {
		kumactlLog.Error(err, "Unable to get index client")
	} else {
		kumaBuildVersion, err = client.GetVersion(context.Background())
		if err != nil {
			kumactlLog.Error(err, "Unable to retrieve server version")
		}
	}

	if kumaBuildVersion == nil {
		return errors.New("WARNING: Unable to confirm the server supports this kumactl version")
	}

	if kumaBuildVersion.Version != kuma_version.Build.Version || kumaBuildVersion.Tagline != kuma_version.Product {
		return errors.New("WARNING: You are using kumactl version " + kuma_version.Build.Version + " for " + kuma_version.Product + ", but the server returned version: " + kumaBuildVersion.Tagline + " " + kumaBuildVersion.Version)
	}

	return nil
}
