package authz

import (
	"github.com/emicklei/go-restful"

	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/user"
)

var log = core.Log.WithName("api-server").WithName("autz")

func AdminFilter(roleAssignments user.RoleAssignments) restful.FilterFunction {
	return func(request *restful.Request, response *restful.Response, chain *restful.FilterChain) {
		u := user.UserFromCtx(request.Request.Context())
		if u == nil {
			if err := response.WriteErrorString(401, "Access Denied. User did not authenticate. To access this endpoint you need to be admin."); err != nil {
				log.Error(err, "could not write the response")
				return
			}
			return
		}
		role := roleAssignments.Role(*u)
		if role != user.AdminRole {
			if err := response.WriteErrorString(403, "Access Denied. User is not an admin. To access this endpoint you need to be admin."); err != nil {
				log.Error(err, "could not write the response")
				return
			}
		}
		chain.ProcessFilter(request, response)
	}
}
