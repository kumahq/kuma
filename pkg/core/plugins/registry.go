package plugins

import (
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

type Registry interface {
	BootstrapPlugins() []BootstrapPlugin
	BootstrapPlugin(name PluginName) (BootstrapPlugin, error)
	ResourceStore(name PluginName) (ResourceStorePlugin, error)
	SecretStore(name PluginName) (SecretStorePlugin, error)
	ConfigStore(name PluginName) (ConfigStorePlugin, error)
	RuntimePlugins() map[PluginName]RuntimePlugin
	CaPlugins() map[PluginName]CaPlugin
	AuthnAPIServer() map[PluginName]AuthnAPIServerPlugin
	PolicyPlugins() map[PluginName]PolicyPlugin
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
		bootstrap:      make(map[PluginName]BootstrapPlugin),
		resourceStore:  make(map[PluginName]ResourceStorePlugin),
		secretStore:    make(map[PluginName]SecretStorePlugin),
		configStore:    make(map[PluginName]ConfigStorePlugin),
		runtime:        make(map[PluginName]RuntimePlugin),
		ca:             make(map[PluginName]CaPlugin),
		authnAPIServer: make(map[PluginName]AuthnAPIServerPlugin),
		policy:         make(map[PluginName]PolicyPlugin),
	}
}

var _ MutableRegistry = &registry{}

type registry struct {
	bootstrap      map[PluginName]BootstrapPlugin
	resourceStore  map[PluginName]ResourceStorePlugin
	secretStore    map[PluginName]SecretStorePlugin
	configStore    map[PluginName]ConfigStorePlugin
	runtime        map[PluginName]RuntimePlugin
	ca             map[PluginName]CaPlugin
	authnAPIServer map[PluginName]AuthnAPIServerPlugin
	policy         map[PluginName]PolicyPlugin
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

func (r *registry) PolicyPlugins() map[PluginName]PolicyPlugin {
	return r.policy
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
		if old, exists := r.resourceStore[name]; exists {
			return pluginAlreadyRegisteredError(resourceStorePlugin, name, old, rsp)
		}
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
		if old, exists := r.policy[name]; exists {
			return pluginAlreadyRegisteredError(policyPlugin, name, old, policy)
		}
		r.policy[name] = policy
	}
	return nil
}

func noSuchPluginError(typ pluginType, name PluginName) error {
	return errors.Errorf("there is no plugin registered with type=%q and name=%s", typ, name)
}

func pluginAlreadyRegisteredError(typ pluginType, name PluginName, old, new Plugin) error {
	return errors.Errorf("plugin with type=%q and name=%s has already been registered: old=%#v new=%#v",
		typ, name, old, new)
}
