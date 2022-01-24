package tokens

import (
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/api-server/authn"
	config_access "github.com/kumahq/kuma/pkg/config/access"
	"github.com/kumahq/kuma/pkg/core/plugins"
	core_tokens "github.com/kumahq/kuma/pkg/core/tokens"
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
		core_tokens.NewValidator(
			core_tokens.NewSigningKeyAccessor(context.ResourceManager(), issuer.UserTokenSigningKeyPrefix),
			core_tokens.NewRevocations(context.ResourceManager(), issuer.UserTokenRevocationsGlobalSecretKey),
		),
	)
	return UserTokenAuthenticator(validator), nil
}

func (c plugin) BeforeBootstrap(*plugins.MutablePluginContext, plugins.PluginConfig) error {
	return nil
}

func (c plugin) AfterBootstrap(context *plugins.MutablePluginContext, config plugins.PluginConfig) error {
	signingKeyManager := core_tokens.NewSigningKeyManager(context.ResourceManager(), issuer.UserTokenSigningKeyPrefix)
	component := core_tokens.NewDefaultSigningKeyComponent(signingKeyManager, log)
	if err := context.ComponentManager().Add(component); err != nil {
		return err
	}
	accessFn, ok := AccessStrategies[context.Config().Access.Type]
	if !ok {
		return errors.Errorf("no Access strategy for type %q", context.Config().Access.Type)
	}
	tokenIssuer := issuer.NewUserTokenIssuer(core_tokens.NewTokenIssuer(signingKeyManager))
	if context.Config().ApiServer.Authn.Tokens.BootstrapAdminToken {
		if err := context.ComponentManager().Add(NewAdminTokenBootstrap(tokenIssuer, context.ResourceManager(), context.Config())); err != nil {
			return err
		}
	}
	webService := server.NewWebService(tokenIssuer, accessFn(context))
	context.APIManager().Add(webService)
	return nil
}

func (c plugin) Name() plugins.PluginName {
	return PluginName
}

func (c plugin) Order() int {
	return plugins.EnvironmentPreparedOrder
}
