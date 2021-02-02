package cmd

import (
	"crypto/tls"
	"net/http"
	"time"

	"github.com/kumahq/kuma/app/kuma-dp/pkg/dataplane/envoy"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	leader_memory "github.com/kumahq/kuma/pkg/plugins/leader/memory"
)

// RootContext contains variables, functions and components that can be overridden when extending kuma-dp or running the test.
type RootContext struct {
	ComponentManager   component.Manager
	BootstrapGenerator envoy.BootstrapConfigFactoryFunc
}

func DefaultRootContext() *RootContext {
	return &RootContext{
		ComponentManager: component.NewManager(leader_memory.NewNeverLeaderElector()),
		BootstrapGenerator: envoy.NewRemoteBootstrapGenerator(&http.Client{
			Timeout:   10 * time.Second,
			Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
		}),
	}
}
