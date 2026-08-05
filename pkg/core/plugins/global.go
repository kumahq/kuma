package plugins

import "github.com/kumahq/kuma/v3/pkg/core"

var log = core.Log.WithName("plugins")

var global = NewRegistry()

var registeredPlugins []registeredPlugin

var loggedRegisteredPlugins int

type registeredPlugin struct {
	name PluginName
	kind string
}

func Plugins() Registry {
	return global
}

func Register(name PluginName, plugin Plugin) {
	if err := global.Register(name, plugin); err != nil {
		panic(err)
	}
	registeredPlugins = append(registeredPlugins, registeredPlugin{name: name, kind: pluginKind(plugin)})
}

func LogRegistered() {
	for _, plugin := range registeredPlugins[loggedRegisteredPlugins:] {
		log.Info("plugin registered", "name", plugin.name, "kind", plugin.kind)
	}
	loggedRegisteredPlugins = len(registeredPlugins)
}

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
	case CoreResourcePlugin:
		return string(coreResourcePlugin)
	case IdentityProviderPlugin:
		return string(identityProviderPlugin)
	default:
		return "unknown"
	}
}
