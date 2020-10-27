package auth

import (
	"net"

	"github.com/emicklei/go-restful"

	"github.com/kumahq/kuma/pkg/core"
)

var log = core.Log.WithName("api-server").WithName("auth")

// AdminAuth validates that the client can access admin endpoints (like Secrets or Dataplane Token)
// You can access the endpoint in two cases
// 1) Request originates from localhost. We assume that if someone has an access to VM/Pod with server, they can do whatever they want. This is also for better UX
// 2) Request originates from outside of localhost but client certs are configured for HTTPS.
//    Client certs are essentially self signed CAs. For now we do not support SAN validation with the same CA that was used to sign server cert
func AdminAuth(request *restful.Request, response *restful.Response, chain *restful.FilterChain) {
	host, _, err := net.SplitHostPort(request.Request.RemoteAddr)
	if err != nil {
		response.WriteHeader(500)
		log.Error(err, "could not parse Remote Address from the Request")
		_, err := response.Write([]byte("Internal Server Error"))
		if err != nil {
			log.Error(err, "could not write the response")
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
	log.Info("attempt to access admin server from the outside of the same machine without allowed certificates")
	response.WriteHeader(403)
	_, err = response.Write([]byte("Access Denied. To access this endpoint you need to do it either from the same machine or by configuring HTTPS on API Server and providing valid certificates"))
	if err != nil {
		log.Error(err, "could not write the response")
	}
}
