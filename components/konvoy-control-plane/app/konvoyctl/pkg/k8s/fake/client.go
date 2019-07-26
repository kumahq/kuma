package fake

import (
	konvoyctl_k8s "github.com/Kong/konvoy/components/konvoy-control-plane/app/konvoyctl/pkg/k8s"
)

var _ konvoyctl_k8s.Client = &FakeClient{}

type FakeClient struct {
	NamespaceExists    bool
	NamespaceExistsErr error
}

func (c *FakeClient) HasNamespace(string) (bool, error) {
	return c.NamespaceExists, c.NamespaceExistsErr
}
