package plugins

import (
	core_discovery "github.com/Kong/kuma/pkg/core/discovery"
	core_store "github.com/Kong/kuma/pkg/core/resources/store"
	core_runtime "github.com/Kong/kuma/pkg/core/runtime"
	secret_store "github.com/Kong/kuma/pkg/core/secrets/store"
)

type Plugin interface{}

type PluginConfig interface{}

type PluginContext = core_runtime.BuilderContext

type MutablePluginContext = core_runtime.Builder

// BootstrapPlugin is responsible for environment-specific initialization at start up,
// e.g. Kubernetes-specific part of configuration.
// Unlike other plugins, can mutate plugin context directly.
type BootstrapPlugin interface {
	Plugin
	Bootstrap(*MutablePluginContext, PluginConfig) error
}

// ResourceStorePlugin is responsible for instantiating a particular ResourceStore.
type ResourceStorePlugin interface {
	Plugin
	NewResourceStore(PluginContext, PluginConfig) (core_store.ResourceStore, error)
}

// SecretStorePlugin is responsible for instantiating a particular SecretStore.
type SecretStorePlugin interface {
	Plugin
	NewSecretStore(PluginContext, PluginConfig) (secret_store.SecretStore, error)
}

// DiscoveryPlugin is responsible for instantiating a particular DiscoverySource.
type DiscoveryPlugin interface {
	Plugin
	NewDiscoverySource(PluginContext, PluginConfig) (core_discovery.DiscoverySource, error)
}
