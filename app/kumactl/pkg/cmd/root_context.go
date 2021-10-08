package cmd

import (
	"context"
	"time"

	"github.com/pkg/errors"

	get_context "github.com/kumahq/kuma/app/kumactl/cmd/get/context"
	inspect_context "github.com/kumahq/kuma/app/kumactl/cmd/inspect/context"
	install_context "github.com/kumahq/kuma/app/kumactl/cmd/install/context"
	"github.com/kumahq/kuma/app/kumactl/pkg/config"
	kumactl_resources "github.com/kumahq/kuma/app/kumactl/pkg/resources"
	"github.com/kumahq/kuma/app/kumactl/pkg/tokens"
	"github.com/kumahq/kuma/pkg/api-server/types"
	config_proto "github.com/kumahq/kuma/pkg/config/app/kumactl/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	util_files "github.com/kumahq/kuma/pkg/util/files"
	kuma_version "github.com/kumahq/kuma/pkg/version"
)

type RootArgs struct {
	ConfigFile string
	Mesh       string
}

type RootRuntime struct {
	Config                       config_proto.Configuration
	Now                          func() time.Time
	NewResourceStore             func(*config_proto.ControlPlaneCoordinates_ApiServer) (core_store.ResourceStore, error)
	NewDataplaneOverviewClient   func(*config_proto.ControlPlaneCoordinates_ApiServer) (kumactl_resources.DataplaneOverviewClient, error)
	NewZoneIngressOverviewClient func(*config_proto.ControlPlaneCoordinates_ApiServer) (kumactl_resources.ZoneIngressOverviewClient, error)
	NewZoneOverviewClient        func(*config_proto.ControlPlaneCoordinates_ApiServer) (kumactl_resources.ZoneOverviewClient, error)
	NewServiceOverviewClient     func(*config_proto.ControlPlaneCoordinates_ApiServer) (kumactl_resources.ServiceOverviewClient, error)
	NewDataplaneTokenClient      func(*config_proto.ControlPlaneCoordinates_ApiServer) (tokens.DataplaneTokenClient, error)
	NewZoneIngressTokenClient    func(*config_proto.ControlPlaneCoordinates_ApiServer) (tokens.ZoneIngressTokenClient, error)
	NewAPIServerClient           func(*config_proto.ControlPlaneCoordinates_ApiServer) (kumactl_resources.ApiServerClient, error)
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
			Now:      time.Now,
			Registry: registry.Global(),
			NewResourceStore: func(server *config_proto.ControlPlaneCoordinates_ApiServer) (core_store.ResourceStore, error) {
				return kumactl_resources.NewResourceStore(server, registry.Global().ObjectDescriptors())
			},
			NewDataplaneOverviewClient:   kumactl_resources.NewDataplaneOverviewClient,
			NewZoneIngressOverviewClient: kumactl_resources.NewZoneIngressOverviewClient,
			NewZoneOverviewClient:        kumactl_resources.NewZoneOverviewClient,
			NewServiceOverviewClient:     kumactl_resources.NewServiceOverviewClient,
			NewDataplaneTokenClient:      tokens.NewDataplaneTokenClient,
			NewZoneIngressTokenClient:    tokens.NewZoneIngressTokenClient,
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

func (rc *RootContext) CurrentResourceStore() (core_store.ResourceStore, error) {
	controlPlane, err := rc.CurrentControlPlane()
	if err != nil {
		return nil, err
	}
	rs, err := rc.Runtime.NewResourceStore(controlPlane.Coordinates.ApiServer)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create a client for Control Plane %q", controlPlane.Name)
	}
	return rs, nil
}

func (rc *RootContext) CurrentDataplaneOverviewClient() (kumactl_resources.DataplaneOverviewClient, error) {
	controlPlane, err := rc.CurrentControlPlane()
	if err != nil {
		return nil, err
	}
	return rc.Runtime.NewDataplaneOverviewClient(controlPlane.Coordinates.ApiServer)
}

func (rc *RootContext) CurrentZoneOverviewClient() (kumactl_resources.ZoneOverviewClient, error) {
	controlPlane, err := rc.CurrentControlPlane()
	if err != nil {
		return nil, err
	}
	return rc.Runtime.NewZoneOverviewClient(controlPlane.Coordinates.ApiServer)
}

func (rc *RootContext) CurrentZoneIngressOverviewClient() (kumactl_resources.ZoneIngressOverviewClient, error) {
	controlPlane, err := rc.CurrentControlPlane()
	if err != nil {
		return nil, err
	}
	return rc.Runtime.NewZoneIngressOverviewClient(controlPlane.Coordinates.ApiServer)
}

func (rc *RootContext) CurrentServiceOverviewClient() (kumactl_resources.ServiceOverviewClient, error) {
	controlPlane, err := rc.CurrentControlPlane()
	if err != nil {
		return nil, err
	}
	return rc.Runtime.NewServiceOverviewClient(controlPlane.Coordinates.ApiServer)
}

func (rc *RootContext) CurrentDataplaneTokenClient() (tokens.DataplaneTokenClient, error) {
	controlPlane, err := rc.CurrentControlPlane()
	if err != nil {
		return nil, err
	}
	return rc.Runtime.NewDataplaneTokenClient(controlPlane.Coordinates.ApiServer)
}

func (rc *RootContext) CurrentZoneIngressTokenClient() (tokens.ZoneIngressTokenClient, error) {
	controlPlane, err := rc.CurrentControlPlane()
	if err != nil {
		return nil, err
	}
	return rc.Runtime.NewZoneIngressTokenClient(controlPlane.Coordinates.ApiServer)
}

func (rc *RootContext) IsFirstTimeUsage() bool {
	if rc.Args.ConfigFile != "" {
		return !util_files.FileExists(rc.Args.ConfigFile)
	}
	return !util_files.FileExists(config.DefaultConfigFile)
}

func (rc *RootContext) CurrentApiClient() (kumactl_resources.ApiServerClient, error) {
	controlPlane, err := rc.CurrentControlPlane()
	if err != nil {
		return nil, err
	}
	return rc.Runtime.NewAPIServerClient(controlPlane.Coordinates.ApiServer)
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
