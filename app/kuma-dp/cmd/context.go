package cmd

import (
	"net/http"
	"path/filepath"
	"runtime"
	"time"

	"github.com/pkg/errors"

	kumadp_config "github.com/kumahq/kuma/app/kuma-dp/pkg/config"
	"github.com/kumahq/kuma/app/kuma-dp/pkg/dataplane/envoy"
	kumadp "github.com/kumahq/kuma/pkg/config/app/kuma-dp"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/log"
	leader_memory "github.com/kumahq/kuma/pkg/plugins/leader/memory"
)

// RootContext contains variables, functions and components that can be overridden when extending kuma-dp or running the test.
type RootContext struct {
	ComponentManager         component.Manager
	BootstrapGenerator       envoy.BootstrapConfigFactoryFunc
	BootstrapDynamicMetadata map[string]string
	DataplaneTokenGenerator  func(*kumadp.Config) error
	Config                   *kumadp.Config
	LogLevel                 log.LogLevel
}

var features = []string{core_xds.FeatureTCPAccessLogViaNamedPipe}

// defaultDataplaneTokenGenerator uses only given tokens or paths from the
// config.
func defaultDataplaneTokenGenerator(cfg *kumadp.Config) error {
	if cfg.DataplaneRuntime.Token != "" {
		path := filepath.Join(cfg.DataplaneRuntime.ConfigDir, cfg.Dataplane.Name)
		if err := writeFile(path, []byte(cfg.DataplaneRuntime.Token), 0600); err != nil {
			runLog.Error(err, "unable to create file with dataplane token")
			return err
		}
		cfg.DataplaneRuntime.TokenPath = path
	}

	if cfg.DataplaneRuntime.TokenPath != "" {
		if err := kumadp_config.ValidateTokenPath(cfg.DataplaneRuntime.TokenPath); err != nil {
			return errors.Wrapf(err, "dataplane token is invalid, in Kubernetes you must mount a serviceAccount token, in universal you must start your proxy with a generated token.")
		}
	}

	return nil
}

func DefaultRootContext() *RootContext {
	config := kumadp.DefaultConfig()
	return &RootContext{
		ComponentManager: component.NewManager(leader_memory.NewNeverLeaderElector()),
		BootstrapGenerator: envoy.NewRemoteBootstrapGenerator(&http.Client{
			Timeout: 10 * time.Second,
		}, runtime.GOOS, features),
		Config:                   &config,
		BootstrapDynamicMetadata: map[string]string{},
		DataplaneTokenGenerator:  defaultDataplaneTokenGenerator,
	}
}
