package main

func isValidConfFile(file string) bool {
	parsed, err := parseToHashMap(file)
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
	parsed, err := parseToHashMap(file)
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
