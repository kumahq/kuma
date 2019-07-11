package k8s

import (
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

type KubeConfig interface {
	Filename() string
	CurrentContext() string
	NewClient() (Client, error)
}

type kubeConfig struct {
	clientConfig clientcmd.ClientConfig
	config       *clientcmdapi.Config
}

var _ KubeConfig = &kubeConfig{}

func (c *kubeConfig) Filename() string {
	return c.clientConfig.ConfigAccess().GetDefaultFilename()
}
func (c *kubeConfig) CurrentContext() string {
	return c.config.CurrentContext
}
func (c *kubeConfig) NewClient() (Client, error) {
	return newClient(c)
}

func DetectKubeConfig() (KubeConfig, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	config, err := clientConfig.ConfigAccess().GetStartingConfig()
	if err != nil {
		return nil, err
	}
	return &kubeConfig{
		clientConfig: clientConfig,
		config:       config,
	}, nil
}
