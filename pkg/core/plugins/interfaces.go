package plugins

import (
	"context"

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
type (
	DbVersion           = uint
	ResourceStorePlugin interface {
		Plugin
		NewResourceStore(PluginContext, PluginConfig) (core_store.ResourceStore, core_store.Transactions, error)
		Migrate(PluginContext, PluginConfig) (DbVersion, error)
		EventListener(PluginContext, events.Emitter) error
	}
)

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

type MatchedPoliciesConfig struct {
	IncludeShadow bool
}

func NewMatchedPoliciesConfig(opts ...MatchedPoliciesOption) *MatchedPoliciesConfig {
	cfg := &MatchedPoliciesConfig{}
	for _, opt := range opts {
		opt(cfg)
	}
	return cfg
}

type MatchedPoliciesOption func(*MatchedPoliciesConfig)

func IncludeShadow() MatchedPoliciesOption {
	return func(cfg *MatchedPoliciesConfig) {
		cfg.IncludeShadow = true
	}
}

// PolicyPlugin a plugin to add a Policy to Kuma
type PolicyPlugin interface {
	Plugin
	// MatchedPolicies return all the policies of the plugins' type matching this dataplane. This is used in the inspect api and accessible in Apply through `proxy.Policies.Dynamic`
	MatchedPolicies(dataplane *core_mesh.DataplaneResource, resources xds_context.Resources, opts ...MatchedPoliciesOption) (core_xds.TypedMatchingPolicies, error)
	// Apply to `rs` using the `ctx` and `proxy` the mutation for all policies of the type this plugin implements.
	// You can access matching policies by using `proxy.Policies.Dynamic`.
	Apply(rs *core_xds.ResourceSet, ctx xds_context.Context, proxy *core_xds.Proxy) error
}

type EgressPolicyPlugin interface {
	PolicyPlugin
	// EgressMatchedPolicies returns all the policies of the plugins' type matching the external service that
	// should be applied on the zone egress.
	EgressMatchedPolicies(tags map[string]string, resources xds_context.Resources, opts ...MatchedPoliciesOption) (core_xds.TypedMatchingPolicies, error)
}

// ProxyPlugin a plugin to modify the proxy. This happens before any `PolicyPlugin` or any envoy generation. and it is applied both for Dataplanes and ZoneProxies
type ProxyPlugin interface {
	Plugin
	// Apply mutate the proxy as needed.
	Apply(ctx context.Context, meshCtx xds_context.MeshContext, proxy *core_xds.Proxy) error
}

// CoreResourcePlugin a plugin to generate xDS resources based on core resources.
type CoreResourcePlugin interface {
	Plugin
	// Apply to `rs` using `proxy` the mutation or new resources.
	Generate(rs *core_xds.ResourceSet, xdsCtx xds_context.Context, proxy *core_xds.Proxy) error
}
