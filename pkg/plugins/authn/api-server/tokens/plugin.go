package tokens

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/api-server/authn"
	config_access "github.com/kumahq/kuma/pkg/config/access"
	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/plugins"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_tokens "github.com/kumahq/kuma/pkg/core/tokens"
	"github.com/kumahq/kuma/pkg/defaults"
	"github.com/kumahq/kuma/pkg/plugins/authn/api-server/tokens/access"
	"github.com/kumahq/kuma/pkg/plugins/authn/api-server/tokens/issuer"
	"github.com/kumahq/kuma/pkg/plugins/authn/api-server/tokens/ws/server"
)

const PluginName = "tokens"

type plugin struct {
	// TODO: properly run AfterBootstrap - https://github.com/kumahq/kuma/issues/6607
	isInitialised bool
}

var (
	_ plugins.AuthnAPIServerPlugin = &plugin{}
	_ plugins.BootstrapPlugin      = &plugin{}
)

// We declare AccessStrategies and not into Runtime because it's a plugin.
var AccessStrategies = map[string]func(*plugins.MutablePluginContext) access.GenerateUserTokenAccess{
	config_access.StaticType: func(context *plugins.MutablePluginContext) access.GenerateUserTokenAccess {
		return access.NewStaticGenerateUserTokenAccess(context.Config().Access.Static.GenerateUserToken)
	},
}

var NewUserTokenIssuer = func(signingKeyManager core_tokens.SigningKeyManager) issuer.UserTokenIssuer {
	return issuer.NewUserTokenIssuer(core_tokens.NewTokenIssuer(signingKeyManager))
}

func init() {
	plugins.Register(PluginName, &plugin{})
}

func (c *plugin) NewAuthenticator(context plugins.PluginContext) (authn.Authenticator, error) {
	publicKeys, err := core_tokens.PublicKeyFromConfig(context.Config().ApiServer.Authn.Tokens.Validator.PublicKeys)
	if err != nil {
		return nil, err
	}
	staticSigningKeyAccessor, err := core_tokens.NewStaticSigningKeyAccessor(publicKeys)
	if err != nil {
		return nil, err
	}
	accessors := []core_tokens.SigningKeyAccessor{staticSigningKeyAccessor}
	if context.Config().ApiServer.Authn.Tokens.Validator.UseSecrets {
		accessors = append(accessors, core_tokens.NewSigningKeyAccessor(context.ResourceManager(), system.UserTokenSigningKeyPrefix))
	}
	validator := issuer.NewUserTokenValidator(
		core_tokens.NewValidator(
			core.Log.WithName("tokens"),
			accessors,
			core_tokens.NewRevocations(context.ResourceManager(), core_model.ResourceKey{Name: system.UserTokenRevocations}),
			context.Config().Store.Type,
		),
	)
	c.isInitialised = true
	return UserTokenAuthenticator(validator), nil
}

func (c *plugin) BeforeBootstrap(*plugins.MutablePluginContext, plugins.PluginConfig) error {
	return nil
}

func (c *plugin) AfterBootstrap(context *plugins.MutablePluginContext, config plugins.PluginConfig) error {
	if !c.isInitialised {
		return nil
	}

	defaults.EnsureDefaultFuncs = append(defaults.EnsureDefaultFuncs, EnsureUserTokenSigningKeyExists)

	accessFn, ok := AccessStrategies[context.Config().Access.Type]
	if !ok {
		return errors.Errorf("no Access strategy for type %q", context.Config().Access.Type)
	}
	signingKeyManager := core_tokens.NewSigningKeyManager(context.ResourceManager(), system.UserTokenSigningKeyPrefix)
	tokenIssuer := NewUserTokenIssuer(signingKeyManager)
	if !context.Config().ApiServer.Authn.Tokens.EnableIssuer {
		tokenIssuer = issuer.DisabledIssuer{}
	}
	if context.Config().ApiServer.Authn.Tokens.BootstrapAdminToken {
		if err := context.ComponentManager().Add(NewAdminTokenBootstrap(tokenIssuer, context.ResourceManager(), context.Config())); err != nil {
			return err
		}
	}
	webService := server.NewWebService(tokenIssuer, accessFn(context))
	context.APIManager().Add(webService)
	return nil
}

func EnsureUserTokenSigningKeyExists(ctx context.Context, resManager core_manager.ResourceManager, logger logr.Logger, cfg kuma_cp.Config) error {
	return core_tokens.EnsureDefaultSigningKeyExist(system.UserTokenSigningKeyPrefix, ctx, resManager, logger)
}

func (c *plugin) Name() plugins.PluginName {
	return PluginName
}

func (c *plugin) Order() int {
	return plugins.EnvironmentPreparedOrder + 1
}
