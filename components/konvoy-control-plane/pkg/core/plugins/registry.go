package plugins

import (
	"github.com/pkg/errors"
)

type pluginType string

const (
	bootstrapPlugin     pluginType = "bootstrap"
	resourceStorePlugin pluginType = "resource-store"
	discoveryPlugin     pluginType = "discovery"
)

type PluginName string

const (
	Kubernetes PluginName = "k8s"
	Universal  PluginName = "universal"
	Memory     PluginName = "memory"
	Postgres   PluginName = "postgres"
)

type Registry interface {
	Bootstrap(PluginName) (BootstrapPlugin, error)
	ResourceStore(name PluginName) (ResourceStorePlugin, error)
	Discovery(name PluginName) (DiscoveryPlugin, error)
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
		discovery:     make(map[PluginName]DiscoveryPlugin),
	}
}

var _ MutableRegistry = &registry{}

type registry struct {
	bootstrap     map[PluginName]BootstrapPlugin
	resourceStore map[PluginName]ResourceStorePlugin
	discovery     map[PluginName]DiscoveryPlugin
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

func (r *registry) Discovery(name PluginName) (DiscoveryPlugin, error) {
	if p, ok := r.discovery[name]; ok {
		return p, nil
	} else {
		return nil, noSuchPluginError(discoveryPlugin, name)
	}
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
	if dp, ok := plugin.(DiscoveryPlugin); ok {
		if old, exists := r.discovery[name]; exists {
			return pluginAlreadyRegisteredError(discoveryPlugin, name, old, dp)
		}
		r.discovery[name] = dp
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
