package k8s

import (
	"context"

	kube_core "k8s.io/api/core/v1"
	kube_apierrors "k8s.io/apimachinery/pkg/api/errors"
	kube_scheme "k8s.io/client-go/kubernetes/scheme"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"

	mesh_v1alpha1 "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/plugins/resources/k8s/native/api/v1alpha1"
)

type Client interface {
	HasNamespace(name string) (bool, error)
}

func newClient(cfg *kubeConfig) (Client, error) {
	kubeConfig, err := cfg.clientConfig.ClientConfig()
	if err != nil {
		return nil, err
	}
	scheme := kube_scheme.Scheme
	mesh_v1alpha1.AddToScheme(scheme)
	kubeClient, err := kube_client.New(kubeConfig, kube_client.Options{Scheme: scheme})
	if err != nil {
		return nil, err
	}
	return &client{Client: kubeClient}, nil
}

type client struct {
	kube_client.Client
}

var _ Client = &client{}

func (c *client) HasNamespace(name string) (bool, error) {
	ns := &kube_core.Namespace{}
	if err := c.Get(context.Background(), kube_client.ObjectKey{Name: name}, ns); err != nil {
		if kube_apierrors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
