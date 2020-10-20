package auth

import (
	"net"

	"github.com/emicklei/go-restful"

	"github.com/kumahq/kuma/pkg/core"
)

var log = core.Log.WithName("api-server").WithName("auth")

func AdminAuth(request *restful.Request, response *restful.Response, chain *restful.FilterChain) {
	host, _, err := net.SplitHostPort(request.Request.RemoteAddr)
	if err != nil {
		response.WriteHeader(500)
		_, err := response.Write([]byte("asdf")) // todo
		if err != nil {
			// log
		}
	}
	if host == "127.0.0.1" || host == "::1" {
		log.V(1).Info("passing the request because it originates from the same machine")
		chain.ProcessFilter(request, response)
		return
	}
	if request.Request.TLS != nil && request.Request.TLS.HandshakeComplete && len(request.Request.TLS.PeerCertificates) > 0 {
		log.V(1).Info("passing the request because it was authenticated via certificate")
		chain.ProcessFilter(request, response)
		return
	}
	log.Info("attempt to access admin server from outside of the same machine without certificates")
	response.WriteHeader(403)
	// todo body
}
