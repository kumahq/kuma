package tokens

import (
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/api-server/authn"
	config_access "github.com/kumahq/kuma/pkg/config/access"
	"github.com/kumahq/kuma/pkg/core/plugins"
	"github.com/kumahq/kuma/pkg/plugins/authn/api-server/tokens/access"
	"github.com/kumahq/kuma/pkg/plugins/authn/api-server/tokens/issuer"
	"github.com/kumahq/kuma/pkg/plugins/authn/api-server/tokens/ws/server"
)

const PluginName = "tokens"

type plugin struct {
}

var _ plugins.AuthnAPIServerPlugin = plugin{}
var _ plugins.BootstrapPlugin = plugin{}

// We declare AccessStrategies and not into Runtime because it's a plugin.
var AccessStrategies = map[string]func(*plugins.MutablePluginContext) access.GenerateUserTokenAccess{
	config_access.StaticType: func(context *plugins.MutablePluginContext) access.GenerateUserTokenAccess {
		return access.NewStaticGenerateUserTokenAccess(context.Config().Access.Static.GenerateUserToken)
	},
}

func init() {
	plugins.Register(PluginName, &plugin{})
}

func (c plugin) NewAuthenticator(context plugins.PluginContext) (authn.Authenticator, error) {
	validator := issuer.NewUserTokenValidator(
		issuer.NewSigningKeyAccessor(context.ResourceManager()),
		issuer.NewTokenRevocations(context.ResourceManager()),
	)
	return UserTokenAuthenticator(validator), nil
}

func (c plugin) BeforeBootstrap(*plugins.MutablePluginContext, plugins.PluginConfig) error {
	return nil
}

func (c plugin) AfterBootstrap(context *plugins.MutablePluginContext, config plugins.PluginConfig) error {
	if err := context.ComponentManager().Add(issuer.NewDefaultSigningKeyComponent(issuer.NewSigningKeyManager(context.ResourceManager()))); err != nil {
		return err
	}
	accessFn, ok := AccessStrategies[context.Config().Access.Type]
	if !ok {
		return errors.Errorf("no Access strategy for type %q", context.Config().Access.Type)
	}
	tokenIssuer := issuer.NewUserTokenIssuer(issuer.NewSigningKeyManager(context.ResourceManager()))
	if context.Config().ApiServer.Authn.Tokens.BootstrapAdminToken {
		if err := context.ComponentManager().Add(NewAdminTokenBootstrap(tokenIssuer, context.ResourceManager(), context.Config())); err != nil {
			return err
		}
	}
	webService := server.NewWebService(tokenIssuer, accessFn(context))
	context.APIManager().Add(webService)
	return nil
}
