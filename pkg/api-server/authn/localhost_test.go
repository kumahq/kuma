package authn_test

import (
	"context"
	"net/http"
	"net/http/httptest"

	"github.com/emicklei/go-restful/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/api-server/authn"
	"github.com/kumahq/kuma/pkg/core/user"
)

// runLocalhost runs LocalhostAuthenticator with the given request parameters
// and returns the name of the resulting user from context.
func runLocalhost(remoteAddr, host, origin string, extraHeaders map[string]string) string {
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "http://localhost/test", http.NoBody)
	req.RemoteAddr = remoteAddr
	req.Host = host
	if origin != "" {
		req.Header.Set("Origin", origin)
	}
	for k, v := range extraHeaders {
		req.Header.Set(k, v)
	}

	rr := httptest.NewRecorder()
	restfulReq := &restful.Request{Request: req}
	restfulResp := restful.NewResponse(rr)

	chain := &restful.FilterChain{
		Target: func(_ *restful.Request, _ *restful.Response) {},
	}
	authn.LocalhostAuthenticator(restfulReq, restfulResp, chain)

	return user.FromCtx(restfulReq.Request.Context()).Name
}

var _ = Describe("LocalhostAuthenticator", func() {
	DescribeTable("direct loopback requests",
		func(remoteAddr, host, origin string, extraHeaders map[string]string, expectAdmin bool) {
			name := runLocalhost(remoteAddr, host, origin, extraHeaders)
			if expectAdmin {
				Expect(name).To(Equal(user.Admin.Name))
			} else {
				Expect(name).NotTo(Equal(user.Admin.Name))
			}
		},
		// Admin-granted cases
		Entry("direct loopback IPv4, no Origin",
			"127.0.0.1:54321", "localhost:5681", "", nil, true),
		Entry("direct loopback IPv6, no Origin",
			"[::1]:54321", "localhost:5681", "", nil, true),
		Entry("direct loopback IPv6 Host without explicit port",
			"[::1]:54321", "[::1]", "", nil, true),
		Entry("direct loopback, same-origin http://localhost",
			"127.0.0.1:54321", "localhost:5681", "http://localhost:5681", nil, true),
		Entry("direct loopback, same-origin http://127.0.0.1",
			"127.0.0.1:54321", "127.0.0.1:5681", "http://127.0.0.1:5681", nil, true),
		Entry("direct loopback, same-origin Host case",
			"127.0.0.1:54321", "LOCALHOST:5681", "http://localhost:5681", nil, true),
		Entry("direct loopback, same-origin implicit HTTP default port",
			"127.0.0.1:54321", "localhost:80", "http://localhost", nil, true),
		Entry("direct loopback, same-origin implicit HTTPS default port",
			"127.0.0.1:54321", "localhost:443", "https://localhost", nil, true),
		Entry("direct loopback, same-origin explicit origin HTTP default port",
			"127.0.0.1:54321", "localhost", "http://localhost:80", nil, true),
		Entry("direct loopback, same-origin bracketed IPv6 default port",
			"[::1]:54321", "[::1]:80", "http://[::1]", nil, true),
		Entry("direct loopback, same-origin bracketed IPv6 explicit origin default port",
			"[::1]:54321", "[::1]", "http://[::1]:80", nil, true),
		Entry("direct loopback with same-origin fetch metadata",
			"127.0.0.1:54321", "localhost:5681", "", map[string]string{"Sec-Fetch-Site": "same-origin"}, true),
		Entry("direct loopback with user-initiated fetch metadata",
			"127.0.0.1:54321", "localhost:5681", "", map[string]string{"Sec-Fetch-Site": "none"}, true),
		// Blocked cases
		Entry("cross-origin evil.com",
			"127.0.0.1:54321", "localhost:5681", "https://evil.com", nil, false),
		Entry("cross-site browser request without Origin",
			"127.0.0.1:54321", "localhost:5681", "", map[string]string{"Sec-Fetch-Site": "cross-site"}, false),
		Entry("same-site browser request without Origin",
			"127.0.0.1:54321", "localhost:5681", "", map[string]string{"Sec-Fetch-Site": "same-site"}, false),
		Entry("cross-origin local app on different port",
			"127.0.0.1:54321", "localhost:5681", "http://127.0.0.1:3000", nil, false),
		Entry("cross-origin implicit HTTP default port to non-default request port",
			"127.0.0.1:54321", "localhost:5681", "http://localhost", nil, false),
		Entry("cross-origin implicit HTTPS default port to non-default request port",
			"127.0.0.1:54321", "localhost:5682", "https://localhost", nil, false),
		Entry("cross-origin explicit non-default origin port to implicit request port",
			"127.0.0.1:54321", "localhost", "http://localhost:5681", nil, false),
		Entry("X-Forwarded-For header present",
			"127.0.0.1:54321", "localhost:5681", "", map[string]string{"X-Forwarded-For": "1.2.3.4"}, false),
		Entry("Forwarded header present",
			"127.0.0.1:54321", "localhost:5681", "", map[string]string{"Forwarded": "for=1.2.3.4"}, false),
		Entry("X-Real-IP header present",
			"127.0.0.1:54321", "localhost:5681", "", map[string]string{"X-Real-IP": "1.2.3.4"}, false),
		Entry("non-loopback Host (reverse proxy public domain)",
			"127.0.0.1:54321", "api.example.com", "", nil, false),
		Entry("malformed bracketed IPv6 Host",
			"[::1]:54321", "[::1", "", nil, false),
		Entry("malformed Host port",
			"127.0.0.1:54321", "localhost:bad", "", nil, false),
		Entry("empty Host port",
			"127.0.0.1:54321", "localhost:", "", nil, false),
		Entry("Origin: null",
			"127.0.0.1:54321", "localhost:5681", "null", nil, false),
		Entry("malformed Origin",
			"127.0.0.1:54321", "localhost:5681", "not-a-url", nil, false),
		Entry("malformed Origin port",
			"127.0.0.1:54321", "localhost:5681", "http://localhost:bad", nil, false),
		Entry("Origin with path",
			"127.0.0.1:54321", "localhost:5681", "http://localhost:5681/path", nil, false),
		Entry("Origin with query",
			"127.0.0.1:54321", "localhost:5681", "http://localhost:5681?x=1", nil, false),
		Entry("Origin with userinfo",
			"127.0.0.1:54321", "localhost:5681", "http://user@localhost:5681", nil, false),
		Entry("remote RemoteAddr is not granted admin",
			"203.0.113.1:54321", "localhost:5681", "", nil, false),
	)
})
