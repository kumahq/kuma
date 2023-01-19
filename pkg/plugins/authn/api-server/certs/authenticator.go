package certs

import (
	"github.com/emicklei/go-restful/v3"

	"github.com/kumahq/kuma/pkg/core/user"
)

// backwards compatibility with Kuma 1.3.x
// https://github.com/kumahq/kuma/issues/4004
func ClientCertAuthenticator(request *restful.Request, response *restful.Response, chain *restful.FilterChain) {
	if user.FromCtx(request.Request.Context()).Name == user.Anonymous.Name && // do not overwrite existing user
		request.Request.TLS != nil &&
		request.Request.TLS.HandshakeComplete &&
		len(request.Request.TLS.PeerCertificates) > 0 {
		request.Request = request.Request.WithContext(user.Ctx(request.Request.Context(), user.Admin.Authenticated()))
	}
	chain.ProcessFilter(request, response)
}
