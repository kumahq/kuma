package cmd

import (
	"context"
	"io"
	"path/filepath"
	"runtime"

	"github.com/pkg/errors"

	kumadp_config "github.com/kumahq/kuma/app/kuma-dp/pkg/config"
	"github.com/kumahq/kuma/app/kuma-dp/pkg/dataplane/envoy"
	kumadp "github.com/kumahq/kuma/pkg/config/app/kuma-dp"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	"github.com/kumahq/kuma/pkg/log"
	leader_memory "github.com/kumahq/kuma/pkg/plugins/leader/memory"
)

// RootContext contains variables, functions and components that can be overridden when extending kuma-dp or running the test.
type RootContext struct {
	ComponentManager         component.Manager
	BootstrapClient          envoy.BootstrapClient
	BootstrapDynamicMetadata map[string]string
	DataplaneTokenGenerator  func(cfg *kumadp.Config) (component.Component, error)
	Config                   *kumadp.Config
	LogLevel                 log.LogLevel
	// DynamicConfigHandlers are handlers for dynamic configuration files. This is useful to add extra handlers in plugins
	DynamicConfigHandlers map[string]func(ctx context.Context, reader io.Reader) error
	// Features is a list of features that are enabled in this kuma-dp instance. This is useful to enable/disable features in plugins.
	Features []string
}

// defaultDataplaneTokenGenerator uses only given tokens or paths from the
// config.
func defaultDataplaneTokenGenerator(cfg *kumadp.Config) (component.Component, error) {
	if cfg.DataplaneRuntime.Token != "" {
		path := filepath.Join(cfg.DataplaneRuntime.WorkDir, cfg.Dataplane.Name)
		if err := writeFile(path, []byte(cfg.DataplaneRuntime.Token), 0o600); err != nil {
			runLog.Error(err, "unable to create file with dataplane token")
			return nil, err
		}
		cfg.DataplaneRuntime.TokenPath = path
	}

	if cfg.DataplaneRuntime.TokenPath != "" {
		if err := kumadp_config.ValidateTokenPath(cfg.DataplaneRuntime.TokenPath); err != nil {
			return nil, errors.Wrapf(err, "dataplane token is invalid, in Kubernetes you must mount a serviceAccount token, in universal you must start your proxy with a generated token.")
		}
	}

	return component.ComponentFunc(func(<-chan struct{}) error {
		return nil
	}), nil
}

func DefaultRootContext() *RootContext {
	config := kumadp.DefaultConfig()
	return &RootContext{
		ComponentManager:         component.NewManager(leader_memory.NewNeverLeaderElector()),
		BootstrapClient:          envoy.NewRemoteBootstrapClient(runtime.GOOS),
		Config:                   &config,
		BootstrapDynamicMetadata: map[string]string{},
		DynamicConfigHandlers:    map[string]func(ctx context.Context, reader io.Reader) error{},
		DataplaneTokenGenerator:  defaultDataplaneTokenGenerator,
	}
}
