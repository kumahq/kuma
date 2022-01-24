package plugins

var global = NewRegistry()

func Plugins() Registry {
	return global
}

func Register(name PluginName, plugin Plugin) {
	if err := global.Register(name, plugin); err != nil {
		panic(err)
	}
}
