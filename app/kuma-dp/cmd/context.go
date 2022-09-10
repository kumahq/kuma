package cmd

import (
	"crypto/tls"
	"net/http"
	"runtime"
	"time"

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
	DataplaneTokenGenerator  func() ([]byte, error)
	Config                   *kumadp.Config
	LogLevel                 log.LogLevel
}

var features = []string{core_xds.FeatureTCPAccessLogViaNamedPipe}

func DefaultRootContext() *RootContext {
	config := kumadp.DefaultConfig()
	return &RootContext{
		ComponentManager: component.NewManager(leader_memory.NewNeverLeaderElector()),
		BootstrapGenerator: envoy.NewRemoteBootstrapGenerator(&http.Client{
			Timeout:   10 * time.Second,
			Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
		}, runtime.GOOS, features),
		Config:                   &config,
		BootstrapDynamicMetadata: map[string]string{},
		DataplaneTokenGenerator:  nil,
	}
}
