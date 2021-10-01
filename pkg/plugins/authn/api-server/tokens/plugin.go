package tokens

import (
	"github.com/kumahq/kuma/pkg/api-server/authn"
	"github.com/kumahq/kuma/pkg/api-server/authz"
	"github.com/kumahq/kuma/pkg/core/plugins"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/plugins/authn/api-server/tokens/issuer"
	"github.com/kumahq/kuma/pkg/plugins/authn/api-server/tokens/ws/server"
)

const PluginName = "tokens"

type plugin struct {
}

var _ plugins.AuthnAPIServerPlugin = plugin{}
var _ plugins.BootstrapPlugin = plugin{}

func init() {
	plugins.Register(PluginName, &plugin{})
}

func (c plugin) NewAuthenticator(context plugins.PluginContext) (authn.Authenticator, error) {
	return UserTokenAuthenticator(tokenIssuer(context.ResourceManager())), nil
}

func (c plugin) BeforeBootstrap(*plugins.MutablePluginContext, plugins.PluginConfig) error {
	return nil
}

func (c plugin) AfterBootstrap(context *plugins.MutablePluginContext, config plugins.PluginConfig) error {
	if err := context.ComponentManager().Add(issuer.NewDefaultSigningKeyComponent(issuer.NewSigningKeyManager(context.ResourceManager()))); err != nil {
		return err
	}
	webService := server.NewWebService(tokenIssuer(context.ResourceManager()))
	webService.Filter(authz.AdminFilter(context.RoleAssignments()))
	context.APIManager().Add(webService)
	return nil
}

func tokenIssuer(resManager manager.ResourceManager) issuer.UserTokenIssuer {
	return issuer.NewUserTokenIssuer(issuer.NewSigningKeyManager(resManager), issuer.NewTokenRevocations(resManager))
}
