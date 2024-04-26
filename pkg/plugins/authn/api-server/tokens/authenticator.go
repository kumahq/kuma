package tokens

import (
	"context"
	"strings"

	"github.com/emicklei/go-restful/v3"

	"github.com/kumahq/kuma/pkg/api-server/authn"
	"github.com/kumahq/kuma/pkg/core"
	rest_errors "github.com/kumahq/kuma/pkg/core/rest/errors"
	"github.com/kumahq/kuma/pkg/core/user"
	"github.com/kumahq/kuma/pkg/plugins/authn/api-server/tokens/issuer"
)

const bearerPrefix = "Bearer "

var log = core.Log.WithName("plugins").WithName("authn").WithName("api-server").WithName("tokens")

func UserTokenAuthenticator(validator issuer.UserTokenValidator, extensions context.Context) authn.Authenticator {
	return func(request *restful.Request, response *restful.Response, chain *restful.FilterChain) {
		if authn.SkipAuth(request) {
			chain.ProcessFilter(request, response)
			return
		}
		authnHeader := request.Request.Header.Get("authorization")
		if user.FromCtx(request.Request.Context()).Name == user.Anonymous.Name && //nolint:contextcheck // do not overwrite existing user
			authnHeader != "" &&
			strings.HasPrefix(authnHeader, bearerPrefix) {
			token := strings.TrimPrefix(authnHeader, bearerPrefix)
			u, err := validator.Validate(request.Request.Context(), token) //nolint:contextcheck
			if err != nil {
				rest_errors.HandleError(request.Request.Context(), extensions, response, &rest_errors.Unauthenticated{}, "Invalid authentication data")
				log.Info("authentication rejected", "reason", err.Error())
				return
			}
			request.Request = request.Request.WithContext(user.Ctx(request.Request.Context(), u.Authenticated())) //nolint:contextcheck
		}
		chain.ProcessFilter(request, response)
	}
}
