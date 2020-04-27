package cmd

import (
	"fmt"
	"net"
	"net/url"
	"time"

	"github.com/pkg/errors"

	"github.com/Kong/kuma/app/kumactl/pkg/config"
	kumactl_resources "github.com/Kong/kuma/app/kumactl/pkg/resources"
	"github.com/Kong/kuma/app/kumactl/pkg/tokens"
	"github.com/Kong/kuma/pkg/catalog"
	catalog_client "github.com/Kong/kuma/pkg/catalog/client"
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
	NewAdminResourceStore      func(string, *kumactl_config.Context_AdminApiCredentials) (core_store.ResourceStore, error)
	NewDataplaneOverviewClient func(*config_proto.ControlPlaneCoordinates_ApiServer) (kumactl_resources.DataplaneOverviewClient, error)
	NewDataplaneTokenClient    func(string, *kumactl_config.Context_AdminApiCredentials) (tokens.DataplaneTokenClient, error)
	NewCatalogClient           func(string) (catalog_client.CatalogClient, error)
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
			NewAdminResourceStore:      kumactl_resources.NewAdminResourceStore,
			NewDataplaneOverviewClient: kumactl_resources.NewDataplaneOverviewClient,
			NewDataplaneTokenClient:    tokens.NewDataplaneTokenClient,
			NewCatalogClient:           catalog_client.NewCatalogClient,
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

func (rc *RootContext) CurrentAdminResourceStore() (core_store.ResourceStore, error) {
	ctx, err := rc.CurrentContext()
	if err != nil {
		return nil, err
	}

	adminServerUrl, err := rc.adminServerUrl()
	if err != nil {
		return nil, err
	}
	rs, err := rc.Runtime.NewAdminResourceStore(adminServerUrl, ctx.GetCredentials().GetAdminApi())
	if err != nil {
		return nil, err
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

func (rc *RootContext) catalog() (catalog.Catalog, error) {
	controlPlane, err := rc.CurrentControlPlane()
	if err != nil {
		return catalog.Catalog{}, err
	}
	client, err := rc.Runtime.NewCatalogClient(controlPlane.Coordinates.ApiServer.Url)
	if err != nil {
		return catalog.Catalog{}, errors.Wrap(err, "could not create components client")
	}
	return client.Catalog()
}

func (rc *RootContext) CurrentDataplaneTokenClient() (tokens.DataplaneTokenClient, error) {
	// todo(jakubdyszkiewicz) check enable/disable by checking cp config
	components, err := rc.catalog()
	if err != nil {
		return nil, err
	}
	if !components.Apis.DataplaneToken.Enabled() {
		return nil, errors.New("Enable the server to be able to generate tokens.")
	}

	ctx, err := rc.CurrentContext()
	if err != nil {
		return nil, err
	}

	adminServerUrl, err := rc.adminServerUrl()
	if err != nil {
		return nil, err
	}
	return rc.Runtime.NewDataplaneTokenClient(adminServerUrl, ctx.GetCredentials().GetAdminApi())
}

func (rc *RootContext) adminServerUrl() (string, error) {
	components, err := rc.catalog()
	if err != nil {
		return "", err
	}

	ctx, err := rc.CurrentContext()
	if err != nil {
		return "", err
	}

	sameMachine, err := rc.cpOnTheSameMachine()
	if err != nil {
		return "", errors.Wrap(err, "could not determine if cp is on the same machine")
	}
	if sameMachine {
		return components.Apis.Admin.LocalUrl, nil
	} else {
		if err := validateRemoteAdminServerSettings(ctx, components); err != nil {
			return "", err
		}
		return components.Apis.DataplaneToken.PublicUrl, nil
	}
}

func validateRemoteAdminServerSettings(ctx *kumactl_config.Context, components catalog.Catalog) error {
	reason := ""
	clientConfigured := ctx.GetCredentials().GetAdminApi().HasClientCert()
	serverConfigured := components.Apis.Admin.PublicUrl != ""
	if !clientConfigured && serverConfigured {
		reason = "admin server in kuma-cp is configured with TLS and kumactl is not."
	}
	if clientConfigured && !serverConfigured {
		reason = "kumactl is configured with TLS and admin server in kuma-cp is not."
	}
	if !clientConfigured && !serverConfigured {
		reason = "both kumactl and admin server in kuma-cp are not configured with TLS."
	}
	if reason != "" { // todo(jakubdyszkiewicz) once docs are in place, put a link to it in 1)
		msg := fmt.Sprintf(`kumactl is trying to access admin server in remote machine but: %s. This can be solved in several ways:
1) Configure kuma-cp admin server with certificate and then use this certificate to configure kumactl.
2) Use kumactl on the same machine as kuma-cp.
3) Use SSH port forwarding so that kuma-cp could be accessed on a remote machine with kumactl on a loopback address.`, reason)
		return errors.New(msg)
	}
	return nil
}

func (rc *RootContext) cpOnTheSameMachine() (bool, error) {
	controlPlane, err := rc.CurrentControlPlane()
	if err != nil {
		return false, err
	}
	cpUrl, err := url.Parse(controlPlane.Coordinates.ApiServer.Url)
	if err != nil {
		return false, err
	}
	host, _, err := net.SplitHostPort(cpUrl.Host)
	if err != nil {
		return false, err
	}
	ip, err := net.ResolveIPAddr("", host)
	if err != nil {
		return false, err
	}
	return ip.IP.IsLoopback(), nil
}

func (rc *RootContext) IsFirstTimeUsage() bool {
	return rc.Args.ConfigFile == "" && !util_files.FileExists(config.DefaultConfigFile)
}
