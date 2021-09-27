package plugins

import (
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/api-server/authn"
	core_ca "github.com/kumahq/kuma/pkg/core/ca"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	core_runtime "github.com/kumahq/kuma/pkg/core/runtime"
	secret_store "github.com/kumahq/kuma/pkg/core/secrets/store"
	"github.com/kumahq/kuma/pkg/events"
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
	BeforeBootstrap(*MutablePluginContext, PluginConfig) error
	AfterBootstrap(*MutablePluginContext, PluginConfig) error
}

// ResourceStorePlugin is responsible for instantiating a particular ResourceStore.
type DbVersion = uint
type ResourceStorePlugin interface {
	Plugin
	NewResourceStore(PluginContext, PluginConfig) (core_store.ResourceStore, error)
	Migrate(PluginContext, PluginConfig) (DbVersion, error)
	EventListener(PluginContext, events.Emitter) error
}

var AlreadyMigrated = errors.New("database already migrated")

// ConfigStorePlugin is responsible for instantiating a particular ConfigStore.
type ConfigStorePlugin interface {
	Plugin
	NewConfigStore(PluginContext, PluginConfig) (core_store.ResourceStore, error)
}

// SecretStorePlugin is responsible for instantiating a particular SecretStore.
type SecretStorePlugin interface {
	Plugin
	NewSecretStore(PluginContext, PluginConfig) (secret_store.SecretStore, error)
}

// RuntimePlugin is responsible for registering environment-specific components,
// e.g. Kubernetes admission web hooks.
type RuntimePlugin interface {
	Plugin
	Customize(core_runtime.Runtime) error
}

// CaPlugin is responsible for providing Certificate Authority Manager
type CaPlugin interface {
	Plugin
	NewCaManager(PluginContext, PluginConfig) (core_ca.Manager, error)
}

// AuthnAPIServerPlugin is responsible for providing authenticator for API Server.
type AuthnAPIServerPlugin interface {
	Plugin
	NewAuthenticator(PluginContext) (authn.Authenticator, error)
}
