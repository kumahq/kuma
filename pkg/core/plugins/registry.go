package plugins

import (
	"github.com/pkg/errors"

	"github.com/Kong/kuma/pkg/core"
)

type pluginType string

const (
	bootstrapPlugin     pluginType = "bootstrap"
	resourceStorePlugin pluginType = "resource-store"
	secretStorePlugin   pluginType = "secret-store"
	discoveryPlugin     pluginType = "discovery"
	runtimePlugin       pluginType = "runtime"
	caPlugin            pluginType = "ca"
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
	Bootstrap(PluginName) (BootstrapPlugin, error)
	ResourceStore(name PluginName) (ResourceStorePlugin, error)
	SecretStore(name PluginName) (SecretStorePlugin, error)
	Discovery(name PluginName) (DiscoveryPlugin, error)
	Runtime(name PluginName) (RuntimePlugin, error)
	Ca(name PluginName) (CaPlugin, error)
	CaPlugins() map[PluginName]CaPlugin
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
		bootstrap:     make(map[PluginName]BootstrapPlugin),
		resourceStore: make(map[PluginName]ResourceStorePlugin),
		secretStore:   make(map[PluginName]SecretStorePlugin),
		discovery:     make(map[PluginName]DiscoveryPlugin),
		runtime:       make(map[PluginName]RuntimePlugin),
		ca:            make(map[PluginName]CaPlugin),
	}
}

var _ MutableRegistry = &registry{}

type registry struct {
	bootstrap     map[PluginName]BootstrapPlugin
	resourceStore map[PluginName]ResourceStorePlugin
	secretStore   map[PluginName]SecretStorePlugin
	discovery     map[PluginName]DiscoveryPlugin
	runtime       map[PluginName]RuntimePlugin
	ca            map[PluginName]CaPlugin
}

func (r *registry) Bootstrap(name PluginName) (BootstrapPlugin, error) {
	if p, ok := r.bootstrap[name]; ok {
		return p, nil
	} else {
		return nil, noSuchPluginError(bootstrapPlugin, name)
	}
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

func (r *registry) Discovery(name PluginName) (DiscoveryPlugin, error) {
	if p, ok := r.discovery[name]; ok {
		return p, nil
	} else {
		return nil, noSuchPluginError(discoveryPlugin, name)
	}
}

func (r *registry) Runtime(name PluginName) (RuntimePlugin, error) {
	if p, ok := r.runtime[name]; ok {
		return p, nil
	} else {
		return nil, noSuchPluginError(runtimePlugin, name)
	}
}

func (r *registry) Ca(name PluginName) (CaPlugin, error) {
	if p, ok := r.ca[name]; ok {
		return p, nil
	} else {
		return nil, noSuchPluginError(runtimePlugin, name)
	}
}

func (r *registry) CaPlugins() map[PluginName]CaPlugin {
	return r.ca
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
	if dp, ok := plugin.(DiscoveryPlugin); ok {
		if old, exists := r.discovery[name]; exists {
			return pluginAlreadyRegisteredError(discoveryPlugin, name, old, dp)
		}
		r.discovery[name] = dp
	}
	if rp, ok := plugin.(RuntimePlugin); ok {
		if old, exists := r.runtime[name]; exists {
			return pluginAlreadyRegisteredError(runtimePlugin, name, old, rp)
		}
		r.runtime[name] = rp
	}
	if cp, ok := plugin.(CaPlugin); ok {
		core.Log.Info("Registering plugin", "name", name, "type", caPlugin)
		if old, exists := r.ca[name]; exists {
			return pluginAlreadyRegisteredError(runtimePlugin, name, old, cp)
		}
		r.ca[name] = cp
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
