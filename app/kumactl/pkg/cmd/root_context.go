package cmd

import (
	"time"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/app/kumactl/pkg/config"
	kumactl_resources "github.com/kumahq/kuma/app/kumactl/pkg/resources"
	"github.com/kumahq/kuma/app/kumactl/pkg/tokens"
	config_proto "github.com/kumahq/kuma/pkg/config/app/kumactl/v1alpha1"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	util_files "github.com/kumahq/kuma/pkg/util/files"
)

type RootArgs struct {
	ConfigFile string
	Mesh       string
}

type RootRuntime struct {
	Config                     config_proto.Configuration
	Now                        func() time.Time
	NewResourceStore           func(*config_proto.ControlPlaneCoordinates_ApiServer) (core_store.ResourceStore, error)
	NewDataplaneOverviewClient func(*config_proto.ControlPlaneCoordinates_ApiServer) (kumactl_resources.DataplaneOverviewClient, error)
	NewZoneOverviewClient      func(*config_proto.ControlPlaneCoordinates_ApiServer) (kumactl_resources.ZoneOverviewClient, error)
	NewServiceOverviewClient   func(*config_proto.ControlPlaneCoordinates_ApiServer) (kumactl_resources.ServiceOverviewClient, error)
	NewDataplaneTokenClient    func(*config_proto.ControlPlaneCoordinates_ApiServer) (tokens.DataplaneTokenClient, error)
	NewAPIServerClient         func(*config_proto.ControlPlaneCoordinates_ApiServer) (kumactl_resources.ApiServerClient, error)
}

type RootContext struct {
	Args    RootArgs
	Runtime RootRuntime
}

func DefaultRootContext() *RootContext {
	return &RootContext{
		Runtime: RootRuntime{
			Now:                        time.Now,
			NewResourceStore:           kumactl_resources.NewResourceStore,
			NewDataplaneOverviewClient: kumactl_resources.NewDataplaneOverviewClient,
			NewZoneOverviewClient:      kumactl_resources.NewZoneOverviewClient,
			NewServiceOverviewClient:   kumactl_resources.NewServiceOverviewClient,
			NewDataplaneTokenClient:    tokens.NewDataplaneTokenClient,
			NewAPIServerClient:         kumactl_resources.NewAPIServerClient,
		},
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
