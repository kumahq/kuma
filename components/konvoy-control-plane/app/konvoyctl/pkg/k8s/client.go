package k8s

import (
	"context"

	kube_core "k8s.io/api/core/v1"
	kube_apierrors "k8s.io/apimachinery/pkg/api/errors"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
)

type Client interface {
	HasNamespace(name string) (bool, error)
}

func newClient(cfg *kubeConfig) (Client, error) {
	kubeConfig, err := cfg.clientConfig.ClientConfig()
	if err != nil {
		return nil, err
	}
	kubeClient, err := kube_client.New(kubeConfig, kube_client.Options{})
	if err != nil {
		return nil, err
	}
	return &client{client: kubeClient}, nil
}

type client struct {
	client kube_client.Client
}

var _ Client = &client{}

func (c *client) HasNamespace(name string) (bool, error) {
	ns := &kube_core.Namespace{}
	if err := c.client.Get(context.Background(), kube_client.ObjectKey{Name: name}, ns); err != nil {
		if kube_apierrors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
