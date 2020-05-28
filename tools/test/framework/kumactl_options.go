package framework

import "os"

// KumactlOptions represents common options necessary to specify for all Kumactl calls
type KumactlOptions struct {
	ContextName string
	ConfigPath  string
	Verbose     bool
}

// NewKumactlOptions will return a pointer to new instance of KumactlOptions with the configured options
func NewKumactlOptions(contextName string, configPath string, verbose bool) *KumactlOptions {
	if configPath == "" {
		configPath = os.ExpandEnv("${HOME}/.kumactl/config")
	}
	return &KumactlOptions{
		ContextName: contextName,
		ConfigPath:  configPath,
		Verbose:     verbose,
	}
}

// GetConfigPath will return a sensible default if the config path is not set on the options.
func (kumactlOptions *KumactlOptions) GetConfigPath() (string, error) {
	return kumactlOptions.ConfigPath, nil
}
