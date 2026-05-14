package plugins

import "github.com/kumahq/kuma/v2/pkg/core"

var log = core.Log.WithName("plugins")

var global = NewRegistry()

func Plugins() Registry {
	return global
}

func Register(name PluginName, plugin Plugin) {
	if err := global.Register(name, plugin); err != nil {
		panic(err)
	}
	log.Info("plugin registered", "name", name, "kind", pluginKind(plugin))
}

// pluginKind returns the type string for a plugin, reusing the constants from registry.go.
func pluginKind(p Plugin) string {
	switch p.(type) {
	case BootstrapPlugin:
		return string(bootstrapPlugin)
	case ResourceStorePlugin:
		return string(resourceStorePlugin)
	case SecretStorePlugin:
		return string(secretStorePlugin)
	case ConfigStorePlugin:
		return string(configStorePlugin)
	case RuntimePlugin:
		return string(runtimePlugin)
	case CaPlugin:
		return string(caPlugin)
	case AuthnAPIServerPlugin:
		return string(authnAPIServer)
	case PolicyPlugin:
		return string(policyPlugin)
	case ProxyPlugin:
		return string(proxyPlugin)
	default:
		return "unknown"
	}
}
