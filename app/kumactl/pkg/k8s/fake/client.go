package fake

import (
	kumactl_k8s "github.com/kumahq/kuma/app/kumactl/pkg/k8s"
)

var _ kumactl_k8s.Client = &FakeClient{}

type FakeClient struct {
	NamespaceExists    bool
	NamespaceExistsErr error
}

func (c *FakeClient) HasNamespace(string) (bool, error) {
	return c.NamespaceExists, c.NamespaceExistsErr
}
