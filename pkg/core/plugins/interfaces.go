package plugins

import (
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/api-server/authn"
	core_ca "github.com/kumahq/kuma/pkg/core/ca"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	core_runtime "github.com/kumahq/kuma/pkg/core/runtime"
	secret_store "github.com/kumahq/kuma/pkg/core/secrets/store"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/events"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
)

type Plugin interface{}

type PluginConfig interface{}

type PluginContext = core_runtime.BuilderContext

type MutablePluginContext = core_runtime.Builder

// EnvironmentPreparingOrder describes an order at which base environment plugins (Universal/Kubernetes) configure the control plane.
var EnvironmentPreparingOrder = 0

// EnvironmentPreparedOrder describes an order at which you can put a plugin and expect that
// the base environment is already configured by Universal/Kubernetes plugins.
var EnvironmentPreparedOrder = EnvironmentPreparingOrder + 1

// BootstrapPlugin is responsible for environment-specific initialization at start up,
// e.g. Kubernetes-specific part of configuration.
// Unlike other plugins, can mutate plugin context directly.
type BootstrapPlugin interface {
	Plugin
	BeforeBootstrap(*MutablePluginContext, PluginConfig) error
	AfterBootstrap(*MutablePluginContext, PluginConfig) error
	Name() PluginName
	// Order defines an order in which plugins are applied on the control plane.
	// If you don't have specific need, consider using EnvironmentPreparedOrder
	Order() int
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

// PolicyPlugin a plugin to add a Policy to Kuma
type PolicyPlugin interface {
	Plugin
	// MatchedPolicies return all the policies of the plugins' type matching this dataplane. This is used in the inspect api and accessible in Apply through `proxy.Policies.Dynamic`
	MatchedPolicies(dataplane *core_mesh.DataplaneResource, resources xds_context.Resources) (core_xds.TypedMatchingPolicies, error)
	// Apply to `rs` using the `ctx` and `proxy` the mutation for all policies of the type this plugin implements.
	// You can access matching policies by using `proxy.Policies.Dynamic`.
	Apply(rs *core_xds.ResourceSet, ctx xds_context.Context, proxy *core_xds.Proxy) error
}
