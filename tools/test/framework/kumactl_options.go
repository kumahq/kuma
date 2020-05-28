package framework

// KumactlOptions represents common options necessary to specify for all Kumactl calls
type KumactlOptions struct {
	ContextName string
	ConfigPath  string
	Silent bool
}

// NewKumactlOptions will return a pointer to new instance of KumactlOptions with the configured options
func NewKumactlOptions(contextName string, configPath string, silent bool) *KumactlOptions {
	if configPath == "" {
		configPath = os.ExpandEnv("${HOME}/.kumactl/config")
	}
	return &KumactlOptions{
		ContextName: contextName,
		ConfigPath:  configPath,
		Silent: silent,
	}
}

// GetConfigPath will return a sensible default if the config path is not set on the options.
func (kumactlOptions *KumactlOptions) GetConfigPath() (string, error) {

	return kumactlOptions.ConfigPath, nil
}
