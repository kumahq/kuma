package k8s

import (
	"net/http"

	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

type KubeConfig interface {
	GetFilename() string
	GetCurrentContext() string
	NewClient() (Client, error)
	NewServiceProxyTransport(namespace, service string) (http.RoundTripper, error)
}

type kubeConfig struct {
	clientConfig clientcmd.ClientConfig
	config       *clientcmdapi.Config
}

var _ KubeConfig = &kubeConfig{}

func (c *kubeConfig) GetFilename() string {
	return c.clientConfig.ConfigAccess().GetDefaultFilename()
}
func (c *kubeConfig) GetCurrentContext() string {
	return c.config.CurrentContext
}
func (c *kubeConfig) NewClient() (Client, error) {
	cfg, err := c.clientConfig.ClientConfig()
	if err != nil {
		return nil, err
	}
	return NewClient(cfg)
}
func (c *kubeConfig) NewServiceProxyTransport(namespace, name string) (http.RoundTripper, error) {
	cfg, err := c.clientConfig.ClientConfig()
	if err != nil {
		return nil, err
	}
	kubeApiProxy, err := NewKubeApiProxyTransport(cfg)
	if err != nil {
		return nil, err
	}
	return NewServiceProxyTransport(kubeApiProxy, namespace, name), nil
}

func DetectKubeConfig() (KubeConfig, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	return newKubeConfig(clientConfig)
}

func GetKubeConfig(kubeconfig, context, namespace string) (KubeConfig, error) {
	loadingRules := &clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfig}
	configOverrides := &clientcmd.ConfigOverrides{
		Context: clientcmdapi.Context{
			Namespace: namespace,
		},
		CurrentContext: context,
	}
	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	return newKubeConfig(clientConfig)
}

func newKubeConfig(clientConfig clientcmd.ClientConfig) (KubeConfig, error) {
	config, err := clientConfig.ConfigAccess().GetStartingConfig()
	if err != nil {
		return nil, err
	}
	return &kubeConfig{
		clientConfig: clientConfig,
		config:       config,
	}, nil
}
