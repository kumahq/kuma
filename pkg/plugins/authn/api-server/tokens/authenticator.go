package tokens

import (
	"strings"

	"github.com/emicklei/go-restful"

	"github.com/kumahq/kuma/pkg/api-server/authn"
	rest_errors "github.com/kumahq/kuma/pkg/core/rest/errors"
	"github.com/kumahq/kuma/pkg/core/user"
	"github.com/kumahq/kuma/pkg/plugins/authn/api-server/tokens/issuer"
)

const bearerPrefix = "Bearer "

func UserTokenAuthenticator(issuer issuer.UserTokenIssuer) authn.Authenticator {
	return func(request *restful.Request, response *restful.Response, chain *restful.FilterChain) {
		authnHeader := request.Request.Header.Get("authorization")
		if user.FromCtx(request.Request.Context()) == nil && // do not overwrite existing user
			authnHeader != "" &&
			strings.HasPrefix(authnHeader, bearerPrefix) {
			token := strings.TrimPrefix(authnHeader, bearerPrefix)
			u, err := issuer.Validate(token)
			if err != nil {
				rest_errors.HandleError(response, &rest_errors.Unauthenticated{}, "invalid authentication data: "+err.Error())
				return
			}
			request.Request = request.Request.WithContext(user.Ctx(request.Request.Context(), u))
		}
		chain.ProcessFilter(request, response)
	}
}
