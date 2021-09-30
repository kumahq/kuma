package authz

import (
	"github.com/emicklei/go-restful"
	rest_errors "github.com/kumahq/kuma/pkg/core/rest/errors"

	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/user"
)

var log = core.Log.WithName("api-server").WithName("autz")

func AdminFilter(roleAssignments user.RoleAssignments) restful.FilterFunction {
	return func(request *restful.Request, response *restful.Response, chain *restful.FilterChain) {
		u := user.FromCtx(request.Request.Context())
		if u == nil {
			rest_errors.HandleError(response, &rest_errors.Unauthenticated{}, "User did not authenticate.")
			return
		}
		role := roleAssignments.Role(*u)
		if role != user.AdminRole {
			rest_errors.HandleError(response, &rest_errors.AccessDenied{}, "To access this endpoint you need to be admin.")
			return
		}
		chain.ProcessFilter(request, response)
	}
}
