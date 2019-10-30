package cmd

import (
	"time"

	"github.com/pkg/errors"

	"github.com/Kong/kuma/app/kumactl/pkg/config"
	kumactl_resources "github.com/Kong/kuma/app/kumactl/pkg/resources"
	"github.com/Kong/kuma/app/kumactl/pkg/tokens"
	"github.com/Kong/kuma/pkg/catalogue"
	catalogue_client "github.com/Kong/kuma/pkg/catalogue/client"
	config_proto "github.com/Kong/kuma/pkg/config/app/kumactl/v1alpha1"
	kumactl_config "github.com/Kong/kuma/pkg/config/app/kumactl/v1alpha1"
	core_model "github.com/Kong/kuma/pkg/core/resources/model"
	core_store "github.com/Kong/kuma/pkg/core/resources/store"
	util_files "github.com/Kong/kuma/pkg/util/files"
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
	NewDataplaneTokenClient    func(string, *kumactl_config.Context_DataplaneTokenApiCredentials) (tokens.DataplaneTokenClient, error)
	NewCatalogueClient         func(string) (catalogue_client.CatalogueClient, error)
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
			NewDataplaneTokenClient:    tokens.NewDataplaneTokenClient,
			NewCatalogueClient:         catalogue_client.NewCatalogueClient,
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

func (rc *RootContext) catalogue() (catalogue.Catalogue, error) {
	controlPlane, err := rc.CurrentControlPlane()
	if err != nil {
		return catalogue.Catalogue{}, err
	}
	client, err := rc.Runtime.NewCatalogueClient(controlPlane.Coordinates.ApiServer.Url)
	if err != nil {
		return catalogue.Catalogue{}, errors.Wrap(err, "could not create components client")
	}
	return client.Catalogue()
}

func (rc *RootContext) CurrentDataplaneTokenClient() (tokens.DataplaneTokenClient, error) {
	components, err := rc.catalogue()
	if err != nil {
		return nil, err
	}
	ctx, err := rc.CurrentContext()
	if err != nil {
		return nil, err
	}

	var url string
	if ctx.DataplaneTokenApiCredentials.TlsEnabled() {
		if components.Apis.DataplaneToken.PublicUrl == "" {
			return nil, errors.New("dataplane token server is not configured with TLS. Either configure kuma-cp to start server with tls or configure kumactl without certificates")
		}
		url = components.Apis.DataplaneToken.PublicUrl
	} else {
		url = components.Apis.DataplaneToken.LocalUrl
	}
	return rc.Runtime.NewDataplaneTokenClient(url, ctx.DataplaneTokenApiCredentials)
}

func (rc *RootContext) IsFirstTimeUsage() bool {
	return rc.Args.ConfigFile == "" && !util_files.FileExists(config.DefaultConfigFile)
}
