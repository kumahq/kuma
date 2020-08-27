package cmd

import (
	"context"
	"crypto/tls"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/ghodss/yaml"

	kumactl_resources "github.com/kumahq/kuma/app/kumactl/pkg/resources"
	config_proto "github.com/kumahq/kuma/pkg/config/app/kumactl/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/model/rest"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/util/proto"

	"github.com/pkg/errors"
	"github.com/sethvargo/go-retry"
	"github.com/spf13/cobra"

	kumadp_config "github.com/kumahq/kuma/app/kuma-dp/pkg/config"
	"github.com/kumahq/kuma/app/kuma-dp/pkg/dataplane/accesslogs"
	"github.com/kumahq/kuma/app/kuma-dp/pkg/dataplane/envoy"
	"github.com/kumahq/kuma/pkg/catalog"
	"github.com/kumahq/kuma/pkg/catalog/client"
	"github.com/kumahq/kuma/pkg/config"
	kuma_dp "github.com/kumahq/kuma/pkg/config/app/kuma-dp"
	config_types "github.com/kumahq/kuma/pkg/config/types"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	leader_memory "github.com/kumahq/kuma/pkg/plugins/leader/memory"
	util_net "github.com/kumahq/kuma/pkg/util/net"
	kuma_version "github.com/kumahq/kuma/pkg/version"
)

type CatalogClientFactory func(string) (client.CatalogClient, error)

var (
	runLog = dataplaneLog.WithName("run")
	// overridable by tests
	bootstrapGenerator = envoy.NewRemoteBootstrapGenerator(&http.Client{
		Timeout:   10 * time.Second,
		Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
	},
	)
	catalogClientFactory    = client.NewCatalogClient
	newResourceStoreFactory = kumactl_resources.NewResourceStore
)

func newRunCmd() *cobra.Command {
	cfg := kuma_dp.DefaultConfig()
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Launch Dataplane (Envoy)",
		Long:  `Launch Dataplane (Envoy).`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			// only support configuration via environment variables and args
			if err := config.Load("", &cfg); err != nil {
				runLog.Error(err, "unable to load configuration")
				return err
			}
			if conf, err := config.ToJson(&cfg); err == nil {
				runLog.Info("effective configuration", "config", string(conf))
			} else {
				runLog.Error(err, "unable to format effective configuration", "config", cfg)
				return err
			}

			catalog, err := fetchCatalog(cfg)
			if err != nil {
				return err
			}
			if catalog.Apis.DataplaneToken.Enabled() {
				if cfg.DataplaneRuntime.TokenPath == "" && cfg.DataplaneRuntime.Token == "" {
					return errors.New("Kuma CP is configured with Dataplane Token Server therefore the Dataplane Token is required. " +
						"Generate token using 'kumactl generate dataplane-token > /path/file' and provide it via --dataplane-token-file=/path/file argument to Kuma DP")
				}
				if cfg.DataplaneRuntime.TokenPath != "" {
					if err := kumadp_config.ValidateTokenPath(cfg.DataplaneRuntime.TokenPath); err != nil {
						return err
					}
				}
			}

			dp, err := readDataplaneResource(cmd, cfg)
			if err != nil {
				return err
			}
			if dp != nil {
				if err := registerDataplaneResource(context.Background(), cfg.ControlPlane.ApiServer.URL, dp); err != nil {
					return err
				}
				defer func() {
					if err := unregisterDataplaneResource(context.Background(), cfg.ControlPlane.ApiServer.URL, dp); err != nil {
						runLog.Error(err, "unable to unregister Dataplane resource")
					}
				}()
			}

			if !cfg.Dataplane.AdminPort.Empty() {
				// unless a user has explicitly opted out of Envoy Admin API, pick a free port from the range
				adminPort, err := util_net.PickTCPPort("127.0.0.1", cfg.Dataplane.AdminPort.Lowest(), cfg.Dataplane.AdminPort.Highest())
				if err != nil {
					return errors.Wrapf(err, "unable to find a free port in the range %q for Envoy Admin API to listen on", cfg.Dataplane.AdminPort)
				}
				cfg.Dataplane.AdminPort = config_types.MustExactPort(adminPort)
				runLog.Info("picked a free port for Envoy Admin API to listen on", "port", cfg.Dataplane.AdminPort)
			}

			if cfg.DataplaneRuntime.ConfigDir == "" {
				tmpDir, err := ioutil.TempDir("", "kuma-dp-")
				if err != nil {
					runLog.Error(err, "unable to create a temporary directory to store generated Envoy config at")
					return err
				}
				defer func() {
					if err := os.RemoveAll(tmpDir); err != nil {
						runLog.Error(err, "unable to remove a temporary directory with a generated Envoy config")
					}
				}()
				cfg.DataplaneRuntime.ConfigDir = tmpDir
				runLog.Info("generated Envoy configuration will be stored in a temporary directory", "dir", tmpDir)
			}

			if cfg.DataplaneRuntime.Token != "" {
				path := filepath.Join(cfg.DataplaneRuntime.ConfigDir, cfg.Dataplane.Name)
				if err := writeFile(path, []byte(cfg.DataplaneRuntime.Token), 0600); err != nil {
					runLog.Error(err, "unable to create file with dataplane token")
					return err
				}
				cfg.DataplaneRuntime.TokenPath = path
			}

			dataplane, err := envoy.New(envoy.Opts{
				Catalog:   *catalog,
				Config:    cfg,
				Generator: bootstrapGenerator,
				Stdout:    cmd.OutOrStdout(),
				Stderr:    cmd.OutOrStderr(),
			})
			if err != nil {
				return err
			}
			server := accesslogs.NewAccessLogServer()

			componentMgr := component.NewManager(leader_memory.NewNeverLeaderElector())
			if err := componentMgr.Add(server, dataplane); err != nil {
				return err
			}

			runLog.Info("starting Kuma DP", "version", kuma_version.Build.Version)
			if err := componentMgr.Start(core.SetupSignalHandler()); err != nil {
				runLog.Error(err, "error while running Kuma DP")
				return err
			}
			runLog.Info("stopping Kuma DP")
			return nil
		},
	}

	cmd.PersistentFlags().StringVar(&cfg.Dataplane.Name, "name", cfg.Dataplane.Name, "Name of the Dataplane")
	cmd.PersistentFlags().Var(&cfg.Dataplane.AdminPort, "admin-port", `Port (or range of ports to choose from) for Envoy Admin API to listen on. Empty value indicates that Envoy Admin API should not be exposed over TCP. Format: "9901 | 9901-9999 | 9901- | -9901"`)
	cmd.PersistentFlags().StringVar(&cfg.Dataplane.Mesh, "mesh", cfg.Dataplane.Mesh, "Mesh that Dataplane belongs to")
	cmd.PersistentFlags().StringVar(&cfg.ControlPlane.ApiServer.URL, "cp-address", cfg.ControlPlane.ApiServer.URL, "URL of the Control Plane API Server")
	cmd.PersistentFlags().StringVar(&cfg.DataplaneRuntime.BinaryPath, "binary-path", cfg.DataplaneRuntime.BinaryPath, "Binary path of Envoy executable")
	cmd.PersistentFlags().StringVar(&cfg.DataplaneRuntime.ConfigDir, "config-dir", cfg.DataplaneRuntime.ConfigDir, "Directory in which Envoy config will be generated")
	cmd.PersistentFlags().StringVar(&cfg.DataplaneRuntime.TokenPath, "dataplane-token-file", cfg.DataplaneRuntime.TokenPath, "Path to a file with dataplane token (use 'kumactl generate dataplane-token' to get one)")
	cmd.PersistentFlags().StringVar(&cfg.DataplaneRuntime.Token, "dataplane-token", cfg.DataplaneRuntime.Token, "Dataplane Token")
	cmd.PersistentFlags().StringVarP(&cfg.DataplaneRuntime.DataplaneTemplate, "dataplane-template", "t", "", "Path to Dataplane template to apply")
	return cmd
}

func writeFile(filename string, data []byte, perm os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(filename), perm); err != nil {
		return err
	}
	return ioutil.WriteFile(filename, data, perm)
}

// fetchCatalog tries to fetch Kuma CP catalog several times
// The main reason for introducing retries here is situation when DP is deployed in the same time as CP (ex. Ingress for Remote CP)
func fetchCatalog(cfg kuma_dp.Config) (*catalog.Catalog, error) {
	runLog.Info("connecting to the Control Plane API for Bootstrap API location")
	catalogClient, err := catalogClientFactory(cfg.ControlPlane.ApiServer.URL)
	if err != nil {
		return nil, errors.Wrap(err, "could not create catalog client")
	}

	backoff, err := retry.NewConstant(cfg.ControlPlane.ApiServer.Retry.Backoff)
	if err != nil {
		return nil, errors.Wrap(err, "could not create retry backoff")
	}
	backoff = retry.WithMaxDuration(cfg.ControlPlane.ApiServer.Retry.MaxDuration, backoff)
	var c catalog.Catalog
	err = retry.Do(context.Background(), backoff, func(ctx context.Context) error {
		c, err = catalogClient.Catalog()
		if err != nil {
			runLog.Info("could not connect to the Control Plane API. Retrying.", "backoff", cfg.ControlPlane.ApiServer.Retry.Backoff, "err", err.Error())
			return retry.RetryableError(err)
		}
		return nil
	})

	if err != nil {
		return nil, errors.Wrap(err, "could not retrieve catalog")
	}
	runLog.Info("connection successful", "catalog", c)
	return &c, nil
}

func registerDataplaneResource(ctx context.Context, address string, dataplane *core_mesh.DataplaneResource) error {
	rs, err := newResourceStoreFactory(&config_proto.ControlPlaneCoordinates_ApiServer{Url: address})
	if err != nil {
		return err
	}
	existing := &core_mesh.DataplaneResource{}
	err = rs.Get(ctx, existing, store.GetBy(model.MetaToResourceKey(dataplane.GetMeta())))
	if err == nil {
		return errors.Errorf("provided Dataplane %s already exists in %s mesh", dataplane.GetMeta().GetName(), dataplane.GetMeta().GetMesh())
	}
	if store.IsResourceNotFound(err) {
		return rs.Create(ctx, dataplane, store.CreateBy(model.MetaToResourceKey(dataplane.GetMeta())))
	}
	return err
}

func unregisterDataplaneResource(ctx context.Context, address string, dataplane *core_mesh.DataplaneResource) error {
	rs, err := newResourceStoreFactory(&config_proto.ControlPlaneCoordinates_ApiServer{Url: address})
	if err != nil {
		return err
	}
	return rs.Delete(ctx, dataplane, store.DeleteBy(model.MetaToResourceKey(dataplane.GetMeta())))
}

func readDataplaneResource(cmd *cobra.Command, cfg kuma_dp.Config) (*core_mesh.DataplaneResource, error) {
	var b []byte
	var err error
	switch cfg.DataplaneRuntime.DataplaneTemplate {
	case "":
		return nil, nil
	case "-":
		if b, err = ioutil.ReadAll(cmd.InOrStdin()); err != nil {
			return nil, err
		}
	default:
		if b, err = ioutil.ReadFile(cfg.DataplaneRuntime.DataplaneTemplate); err != nil {
			return nil, errors.Wrap(err, "error while reading provided file")
		}
	}
	return parseDataplane(b)
}

func parseDataplane(bytes []byte) (*core_mesh.DataplaneResource, error) {
	resMeta := rest.ResourceMeta{}
	if err := yaml.Unmarshal(bytes, &resMeta); err != nil {
		return nil, err
	}
	if resMeta.Name == "" {
		return nil, errors.New("Name field cannot be empty")
	}
	if resMeta.Mesh == "" {
		return nil, errors.New("Mesh field cannot be empty")
	}
	dp := &core_mesh.DataplaneResource{}
	if err := proto.FromYAML(bytes, dp.GetSpec()); err != nil {
		return nil, err
	}
	dp.SetMeta(meta{
		Name: resMeta.Name,
		Mesh: resMeta.Mesh,
	})
	return dp, nil
}

var _ model.ResourceMeta = &meta{}

type meta struct {
	Name string
	Mesh string
}

func (m meta) GetName() string {
	return m.Name
}

func (m meta) GetNameExtensions() model.ResourceNameExtensions {
	return model.ResourceNameExtensionsUnsupported
}

func (m meta) GetVersion() string {
	return ""
}

func (m meta) GetMesh() string {
	return m.Mesh
}

func (m meta) GetCreationTime() time.Time {
	return time.Unix(0, 0).UTC() // the date doesn't matter since it is set on server side anyways
}

func (m meta) GetModificationTime() time.Time {
	return time.Unix(0, 0).UTC() // the date doesn't matter since it is set on server side anyways
}
