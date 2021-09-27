package certs

import (
	"github.com/kumahq/kuma/pkg/api-server/authn"
	"github.com/kumahq/kuma/pkg/core/plugins"
)

const PluginName = "clientCerts"

type plugin struct {
}

func init() {
	plugins.Register(PluginName, &plugin{})
}

var _ plugins.AuthnAPIServerPlugin = plugin{}

func (c plugin) NewAuthenticator(_ plugins.PluginContext) (authn.Authenticator, error) {
	return ClientCertAuthenticator, nil
}
