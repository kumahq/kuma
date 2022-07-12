package main

func isValidConfFile(file string) bool {
	parsed, err := parseFileToHashMap(file)
	if err != nil {
		log.Error(err, "could not unmarshal config file")
		return false
	}

	configType, ok := parsed["type"]
	if ok {
		log.V(1).Info("config valid", "file", file, "type", configType)
		return true
	}
	log.V(1).Info("config not valid", "file", file)
	return false
}

func isValidConflistFile(file string) bool {
	parsed, err := parseFileToHashMap(file)
	if err != nil {
		log.Error(err, "could not unmarshal config file")
		return false
	}

	configName, hasName := parsed["name"]
	plugins, hasPlugins := parsed["plugins"]
	if hasName && hasPlugins {
		log.V(1).Info("config valid", "file", file, "name", configName, "plugins", plugins)
		return true
	}

	return false
}

func checkInstall(cniConfPath string, isPluginChained bool) bool {
	if fileExists(cniConfPath) {
		parsed, err := parseFileToHashMap(cniConfPath)
		if err != nil {
			return false
		}
		if isPluginChained {
			if isValidConflistFile(cniConfPath) {
				plugins, err := getPluginsArray(parsed)
				if err != nil {
					return false
				}
				index, err := findKumaCniConfigIndex(plugins)
				if err != nil {
					return false
				}

				if index >= 0 {
					return true
				}
			}
		} else {
			if isValidConfFile(cniConfPath) {
				pluginType, ok := parsed["type"]
				if !ok {
					return false
				}
				if pluginType == "kuma-cni" {
					return true
				}
			}
		}
	}

	return false
}
