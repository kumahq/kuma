package plugins

type PluginInitializer struct {
	InitFn      func()
	Initialized bool
}

func InitAll(plugins map[string]*PluginInitializer) {
	for _, initializer := range plugins {
		if !initializer.Initialized {
			initializer.InitFn()
			initializer.Initialized = true
		}
	}
}

func Init(enabledPlugins []string, plugins map[string]*PluginInitializer) {
	for _, policy := range enabledPlugins {
		initializer, ok := plugins[policy]
		if ok && !initializer.Initialized {
			initializer.InitFn()
			initializer.Initialized = true
		} else {
			panic("plugin " + policy + " not found")
		}
	}
}

func InitAllIf(enabledPlugins []string, pluginName string, plugins map[string]*PluginInitializer) {
	for _, plugin := range enabledPlugins {
		if plugin == pluginName {
			for _, initializer := range plugins {
				if !initializer.Initialized {
					initializer.InitFn()
					initializer.Initialized = true
				}
			}
			break;
		}
	}
}
