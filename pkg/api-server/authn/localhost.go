package authn

import (
	"net"

	"github.com/emicklei/go-restful"

	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/user"
)

var log = core.Log.WithName("api-server").WithName("authn")

func LocalhostAuthenticator(request *restful.Request, response *restful.Response, chain *restful.FilterChain) {
	host, _, err := net.SplitHostPort(request.Request.RemoteAddr)
	if err != nil {
		log.Error(err, "could not parse Remote Address from the Request")
		if err := response.WriteErrorString(500, "Internal Server Error"); err != nil {
			log.Error(err, "could not write the response")
			return
		}
	}
	if host == "127.0.0.1" || host == "::1" {
		log.V(1).Info("authenticated as admin because requests originates from the same machine")
		request.Request = request.Request.WithContext(user.UserCtx(request.Request.Context(), user.Admin))
	}
	chain.ProcessFilter(request, response)
}
