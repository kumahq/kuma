package fake

import (
	"net/http"

	konvoyctl_k8s "github.com/Kong/konvoy/components/konvoy-control-plane/app/konvoyctl/pkg/k8s"
)

var _ konvoyctl_k8s.KubeConfig = &FakeKubeConfig{}

type FakeKubeConfig struct {
	Filename                 string
	CurrentContext           string
	Client                   konvoyctl_k8s.Client
	ClientErr                error
	ServiceProxyTransport    http.RoundTripper
	ServiceProxyTransportErr error
}

func (c *FakeKubeConfig) GetFilename() string {
	return c.Filename
}
func (c *FakeKubeConfig) GetCurrentContext() string {
	return c.CurrentContext
}
func (c *FakeKubeConfig) NewClient() (konvoyctl_k8s.Client, error) {
	return c.Client, c.ClientErr
}
func (c *FakeKubeConfig) NewServiceProxyTransport(namespace, name string) (http.RoundTripper, error) {
	return c.ServiceProxyTransport, c.ServiceProxyTransportErr
}
