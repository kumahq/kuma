package authn

import (
	"net"
	"net/http"
	"net/url"
	"strconv"
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
	if !isLoopbackRequestHost(r.Host) {
		return false
	}
	switch r.Header.Get("Sec-Fetch-Site") {
	case "", "same-origin", "none":
	default:
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

func isLoopbackRequestHost(hostport string) bool {
	host, _, ok := parseAuthority("http", hostport)
	if !ok {
		return false
	}
	return isLoopbackHost(host)
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
	originHost, originPort, ok := parseOrigin(u)
	if !ok {
		return false
	}
	host, port, ok := parseAuthority(u.Scheme, requestHost)
	if !ok {
		return false
	}
	return originHost == host && originPort == port
}

func parseOrigin(u *url.URL) (string, string, bool) {
	if u.User != nil || u.Host == "" || u.Path != "" || u.RawQuery != "" || u.Fragment != "" {
		return "", "", false
	}
	if strings.HasSuffix(u.Host, ":") {
		return "", "", false
	}
	return parseURLAuthority(u)
}

func parseAuthority(scheme, hostport string) (string, string, bool) {
	if strings.HasSuffix(hostport, ":") {
		return "", "", false
	}
	u, err := url.Parse(scheme + "://" + hostport)
	if err != nil || u.User != nil || u.Host == "" || u.Path != "" || u.RawQuery != "" || u.Fragment != "" {
		return "", "", false
	}
	return parseURLAuthority(u)
}

func parseURLAuthority(u *url.URL) (string, string, bool) {
	port := u.Port()
	if port == "" {
		port = defaultPort(u.Scheme)
	}
	if !validPort(port) {
		return "", "", false
	}
	host := canonicalHost(u.Hostname())
	if host == "" {
		return "", "", false
	}
	return host, port, true
}

func canonicalHost(host string) string {
	host = strings.TrimSuffix(strings.ToLower(host), ".")
	ip := net.ParseIP(host)
	if ip != nil {
		return ip.String()
	}
	return host
}

func defaultPort(scheme string) string {
	switch scheme {
	case "http":
		return "80"
	case "https":
		return "443"
	}
	return ""
}

func validPort(port string) bool {
	n, err := strconv.Atoi(port)
	return err == nil && n > 0 && n <= 65535
}
