package authn

import (
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/emicklei/go-restful/v3"

	"github.com/kumahq/kuma/v2/pkg/core"
	"github.com/kumahq/kuma/v2/pkg/core/user"
)

var log = core.Log.WithName("api-server").WithName("authn")

func LocalhostAuthenticator(request *restful.Request, response *restful.Response, chain *restful.FilterChain) {
	if isDirectLoopbackRequest(request.Request) {
		log.V(1).Info("authenticated as admin because request is a direct localhost call")
		request.Request = request.Request.WithContext(user.Ctx(request.Request.Context(), user.Admin.Authenticated()))
	}
	chain.ProcessFilter(request, response)
}

// isDirectLoopbackRequest returns true only when all conditions hold:
//   - RemoteAddr is a loopback address,
//   - no proxy-hop headers are present (Forwarded, X-Forwarded-For, X-Real-IP),
//   - the Host header is localhost or a loopback address,
//   - the browser did not mark the request as cross-site, and
//   - if an Origin header is present it is same-origin with the Host.
//
// This prevents browsers on the same machine, which connect over loopback,
// from triggering admin access on behalf of a non-localhost origin.
func isDirectLoopbackRequest(r *http.Request) bool {
	remoteHost, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil || !isLoopbackHost(remoteHost) {
		return false
	}
	if hasProxyHeaders(r.Header) {
		return false
	}
	if !isLoopbackHost(hostWithoutPort(r.Host)) {
		return false
	}
	if r.Header.Get("Sec-Fetch-Site") == "cross-site" {
		return false
	}
	return isSameOriginLoopback(r.Header.Get("Origin"), r.Host)
}

func isLoopbackHost(host string) bool {
	host = strings.TrimSuffix(strings.ToLower(host), ".")
	if host == "localhost" {
		return true
	}
	if strings.HasPrefix(host, "[") && strings.HasSuffix(host, "]") {
		host = strings.TrimPrefix(strings.TrimSuffix(host, "]"), "[")
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func hostWithoutPort(hostport string) string {
	host, _, err := net.SplitHostPort(hostport)
	if err != nil {
		return hostport
	}
	return host
}

func hasProxyHeaders(h http.Header) bool {
	return h.Get("Forwarded") != "" || h.Get("X-Forwarded-For") != "" || h.Get("X-Real-IP") != ""
}

func isSameOriginLoopback(origin, requestHost string) bool {
	if origin == "" {
		return true
	}
	u, err := url.Parse(origin)
	if err != nil || u.Host == "" {
		return false
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return false
	}
	if !isLoopbackHost(u.Hostname()) {
		return false
	}
	return strings.EqualFold(u.Host, requestHost)
}
