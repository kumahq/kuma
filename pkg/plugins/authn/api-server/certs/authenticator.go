package certs

import (
	"github.com/emicklei/go-restful"

	"github.com/kumahq/kuma/pkg/core/user"
)

// backwards compatibility with Kuma 1.3.x
func ClientCertAuthenticator(request *restful.Request, response *restful.Response, chain *restful.FilterChain) {
	if user.FromCtx(request.Request.Context()) == nil && // do not overwrite existing user
		request.Request.TLS != nil &&
		request.Request.TLS.HandshakeComplete &&
		len(request.Request.TLS.PeerCertificates) > 0 {
		request.Request = request.Request.WithContext(user.Ctx(request.Request.Context(), user.Admin))
	}
	chain.ProcessFilter(request, response)
}
