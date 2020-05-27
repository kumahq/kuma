package framework

// KumactlOptions represents common options necessary to specify for all Kumactl calls
type KumactlOptions struct {
	ContextName string
	ConfigPath  string
}

// NewKumactlOptions will return a pointer to new instance of KumactlOptions with the configured options
func NewKumactlOptions(contextName string, configPath string) *KumactlOptions {
	return &KumactlOptions{
		ContextName: contextName,
		ConfigPath:  configPath,
	}
}

// GetConfigPath will return a sensible default if the config path is not set on the options.
func (kumactlOptions *KumactlOptions) GetConfigPath() (string, error) {
	kumaConfigPath := kumactlOptions.ConfigPath
	if kumaConfigPath == "" {
		kumaConfigPath = "~/.kumactl/config"
	}
	return kumaConfigPath, nil
}
