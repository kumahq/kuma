package plugins

import (
	"cmp"
	"slices"
	"sort"

	"github.com/pkg/errors"
)

type pluginType string

const (
	bootstrapPlugin     pluginType = "bootstrap"
	resourceStorePlugin pluginType = "resource-store"
	secretStorePlugin   pluginType = "secret-store"
	configStorePlugin   pluginType = "config-store"
	runtimePlugin       pluginType = "runtime"
	caPlugin            pluginType = "ca"
	authnAPIServer      pluginType = "authn-api-server"
	policyPlugin        pluginType = "policy"
	proxyPlugin         pluginType = "proxy"
)

type PluginName string

const (
	Kubernetes PluginName = "k8s"
	Universal  PluginName = "universal"
	Memory     PluginName = "memory"
	Postgres   PluginName = "postgres"

	CaBuiltin  PluginName = "builtin"
	CaProvided PluginName = "provided"
)

type RegisteredPolicyPlugin struct {
	Plugin PolicyPlugin
	Name   PluginName
}

type Registry interface {
	BootstrapPlugins() []BootstrapPlugin
	ResourceStore(name PluginName) (ResourceStorePlugin, error)
	SecretStore(name PluginName) (SecretStorePlugin, error)
	ConfigStore(name PluginName) (ConfigStorePlugin, error)
	RuntimePlugins() map[PluginName]RuntimePlugin
	CaPlugins() map[PluginName]CaPlugin
	AuthnAPIServer() map[PluginName]AuthnAPIServerPlugin
	PolicyPlugins() []RegisteredPolicyPlugin
	ProxyPlugins() map[PluginName]ProxyPlugin
	CoreResourcePlugins() map[PluginName]CoreResourcePlugin
	IdentityProviders() map[PluginName]IdentityProviderPlugin
}

type RegistryMutator interface {
	Register(PluginName, Plugin) error
}

type MutableRegistry interface {
	Registry
	RegistryMutator
}

func NewRegistry() MutableRegistry {
	return &registry{
		bootstrap:                   make(map[PluginName]BootstrapPlugin),
		resourceStore:               make(map[PluginName]ResourceStorePlugin),
		secretStore:                 make(map[PluginName]SecretStorePlugin),
		configStore:                 make(map[PluginName]ConfigStorePlugin),
		runtime:                     make(map[PluginName]RuntimePlugin),
		ca:                          make(map[PluginName]CaPlugin),
		authnAPIServer:              make(map[PluginName]AuthnAPIServerPlugin),
		proxy:                       make(map[PluginName]ProxyPlugin),
		registeredResources:         make(map[PluginName]CoreResourcePlugin),
		registeredIdentityProviders: make(map[PluginName]IdentityProviderPlugin),
	}
}

var _ MutableRegistry = &registry{}

type registry struct {
	bootstrap                   map[PluginName]BootstrapPlugin
	resourceStore               map[PluginName]ResourceStorePlugin
	secretStore                 map[PluginName]SecretStorePlugin
	configStore                 map[PluginName]ConfigStorePlugin
	runtime                     map[PluginName]RuntimePlugin
	proxy                       map[PluginName]ProxyPlugin
	ca                          map[PluginName]CaPlugin
	authnAPIServer              map[PluginName]AuthnAPIServerPlugin
	registeredPolicies          []RegisteredPolicyPlugin
	registeredResources         map[PluginName]CoreResourcePlugin
	registeredIdentityProviders map[PluginName]IdentityProviderPlugin
}

func (r *registry) ResourceStore(name PluginName) (ResourceStorePlugin, error) {
	if p, ok := r.resourceStore[name]; ok {
		return p, nil
	} else {
		return nil, noSuchPluginError(resourceStorePlugin, name)
	}
}

func (r *registry) SecretStore(name PluginName) (SecretStorePlugin, error) {
	if p, ok := r.secretStore[name]; ok {
		return p, nil
	} else {
		return nil, noSuchPluginError(secretStorePlugin, name)
	}
}

func (r *registry) ConfigStore(name PluginName) (ConfigStorePlugin, error) {
	if p, ok := r.configStore[name]; ok {
		return p, nil
	} else {
		return nil, noSuchPluginError(configStorePlugin, name)
	}
}

func (r *registry) CaPlugins() map[PluginName]CaPlugin {
	return r.ca
}

func (r *registry) RuntimePlugins() map[PluginName]RuntimePlugin {
	return r.runtime
}

func (r *registry) ProxyPlugins() map[PluginName]ProxyPlugin {
	return r.proxy
}

// PolicyPlugins returns policy plugins sorted by their order.
func (r *registry) PolicyPlugins() []RegisteredPolicyPlugin {
	return r.registeredPolicies
}

func (r *registry) CoreResourcePlugins() map[PluginName]CoreResourcePlugin {
	return r.registeredResources
}

func (r *registry) IdentityProviders() map[PluginName]IdentityProviderPlugin {
	return r.registeredIdentityProviders
}

func (r *registry) BootstrapPlugins() []BootstrapPlugin {
	var plugins []BootstrapPlugin
	for _, plugin := range r.bootstrap {
		plugins = append(plugins, plugin)
	}
	sort.Slice(plugins, func(i, j int) bool {
		return plugins[i].Order() < plugins[j].Order()
	})
	return plugins
}

func (r *registry) BootstrapPlugin(name PluginName) (BootstrapPlugin, error) {
	if p, ok := r.bootstrap[name]; ok {
		return p, nil
	} else {
		return nil, noSuchPluginError(bootstrapPlugin, name)
	}
}

func (r *registry) AuthnAPIServer() map[PluginName]AuthnAPIServerPlugin {
	return r.authnAPIServer
}

func (r *registry) Register(name PluginName, plugin Plugin) error {
	if bp, ok := plugin.(BootstrapPlugin); ok {
		if old, exists := r.bootstrap[name]; exists {
			return pluginAlreadyRegisteredError(bootstrapPlugin, name, old, bp)
		}
		r.bootstrap[name] = bp
	}
	if rsp, ok := plugin.(ResourceStorePlugin); ok {
		r.resourceStore[name] = rsp
	}
	if ssp, ok := plugin.(SecretStorePlugin); ok {
		if old, exists := r.secretStore[name]; exists {
			return pluginAlreadyRegisteredError(secretStorePlugin, name, old, ssp)
		}
		r.secretStore[name] = ssp
	}
	if csp, ok := plugin.(ConfigStorePlugin); ok {
		if old, exists := r.configStore[name]; exists {
			return pluginAlreadyRegisteredError(configStorePlugin, name, old, csp)
		}
		r.configStore[name] = csp
	}
	if rp, ok := plugin.(RuntimePlugin); ok {
		if old, exists := r.runtime[name]; exists {
			return pluginAlreadyRegisteredError(runtimePlugin, name, old, rp)
		}
		r.runtime[name] = rp
	}
	if cp, ok := plugin.(CaPlugin); ok {
		if old, exists := r.ca[name]; exists {
			return pluginAlreadyRegisteredError(caPlugin, name, old, cp)
		}
		r.ca[name] = cp
	}
	if authn, ok := plugin.(AuthnAPIServerPlugin); ok {
		if old, exists := r.authnAPIServer[name]; exists {
			return pluginAlreadyRegisteredError(authnAPIServer, name, old, authn)
		}
		r.authnAPIServer[name] = authn
	}
	if policy, ok := plugin.(PolicyPlugin); ok {
		for _, existing := range r.registeredPolicies {
			if existing.Name == name {
				return pluginAlreadyRegisteredError(policyPlugin, name, existing.Plugin, policy)
			}
		}
		entry := RegisteredPolicyPlugin{Plugin: policy, Name: name}
		pos, found := slices.BinarySearchFunc(r.registeredPolicies, entry, func(a, b RegisteredPolicyPlugin) int {
			return cmp.Compare(a.Plugin.Order(), b.Plugin.Order())
		})
		if found {
			return errors.Errorf("policy plugin %q has the same order (%d) as plugin %q; each plugin must declare a unique order", name, policy.Order(), r.registeredPolicies[pos].Name)
		}
		r.registeredPolicies = slices.Insert(r.registeredPolicies, pos, entry)
	}
	if proxy, ok := plugin.(ProxyPlugin); ok {
		if old, exists := r.proxy[name]; exists {
			return pluginAlreadyRegisteredError(proxyPlugin, name, old, proxy)
		}
		r.proxy[name] = proxy
	}
	if coreRes, ok := plugin.(CoreResourcePlugin); ok {
		if old, exists := r.registeredResources[name]; exists {
			return pluginAlreadyRegisteredError(proxyPlugin, name, old, coreRes)
		}
		r.registeredResources[name] = coreRes
	}
	if identityProvider, ok := plugin.(IdentityProviderPlugin); ok {
		if old, exists := r.registeredIdentityProviders[name]; exists {
			return pluginAlreadyRegisteredError(proxyPlugin, name, old, identityProvider)
		}
		r.registeredIdentityProviders[name] = identityProvider
	}
	return nil
}

func noSuchPluginError(typ pluginType, name PluginName) error {
	return errors.Errorf("there is no plugin registered with type=%q and name=%s", typ, name)
}

func pluginAlreadyRegisteredError(typ pluginType, name PluginName, old, n Plugin) error {
	return errors.Errorf("plugin with type=%q and name=%s has already been registered: old=%#v new=%#v",
		typ, name, old, n)
}
