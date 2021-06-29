package k8s

import (
	"errors"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func DefaultClientConfig(kubeconfig, context string) (*rest.Config, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	loadingRules.DefaultClientConfig = &clientcmd.DefaultClientConfig
	loadingRules.ExplicitPath = kubeconfig
	configOverrides := &clientcmd.ConfigOverrides{
		ClusterDefaults: clientcmd.ClusterDefaults,
		CurrentContext:  context,
	}
	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	config, err := clientConfig.ClientConfig()
	if err != nil {
		// translate default error since it's confusing. What user really wants to set is KUBECONFIG, not KUBERNETES_MASTER
		if err.Error() == ("invalid configuration: " + clientcmd.ErrEmptyConfig.Error()) {
			return nil, errors.New("make sure you set KUBECONFIG environment variable to a valid Kubernetes config file")
		}
		return nil, err
	}
	return config, nil
}
