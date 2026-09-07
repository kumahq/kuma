package framework

import (
	"net/http"
	"net/http/httptest"
	"net/url"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/v3/test/framework/portforward"
)

var _ = Describe("K8sControlPlane", func() {
	// controlPlaneWithToken returns a control plane pointed at srv, with the
	// admin token already resolved to token so nothing reads a secret.
	controlPlaneWithToken := func(srv *httptest.Server, token string, apiHeaders ...string) *K8sControlPlane {
		cp := &K8sControlPlane{
			portFwd:    portforward.Tunnel{Endpoint: srv.Listener.Addr().String()},
			apiHeaders: apiHeaders,
		}
		cp.adminTokenOnce.Do(func() { cp.adminToken = token })
		return cp
	}

	Describe("retrieveAdminToken", func() {
		It("should not look for a secret when the control plane does not authenticate with tokens", func() {
			cp := &K8sControlPlane{cluster: &K8sCluster{opts: kumaDeploymentOptions{
				env: map[string]string{"KUMA_API_SERVER_AUTHN_TYPE": "external"},
			}}}

			Expect(cp.retrieveAdminToken()).To(BeEmpty())
		})

		It("should not look for a secret that is never bootstrapped", func() {
			cp := &K8sControlPlane{cluster: &K8sCluster{opts: kumaDeploymentOptions{
				env: map[string]string{"KUMA_API_SERVER_AUTHN_TOKENS_BOOTSTRAP_ADMIN_TOKEN": "false"},
			}}}

			Expect(cp.retrieveAdminToken()).To(BeEmpty())
		})
	})

	Describe("InspectEnvoyProxy", func() {
		var srv *httptest.Server
		var seen http.Header

		BeforeEach(func() {
			seen = nil
			srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				seen = r.Header.Clone()
				_, _ = w.Write([]byte("{}"))
			}))
			DeferCleanup(srv.Close)
		})

		It("should authenticate with the bootstrapped admin token", func() {
			cp := controlPlaneWithToken(srv, "admin-token")

			_, err := cp.InspectEnvoyProxy("/meshes/default/dataplanes/dp-1/stats", url.Values{})

			Expect(err).ToNot(HaveOccurred())
			Expect(seen.Get("Authorization")).To(Equal("Bearer admin-token"))
		})

		It("should send no credential when there is no token to attach", func() {
			cp := controlPlaneWithToken(srv, "")

			_, err := cp.InspectEnvoyProxy("/meshes/default/dataplanes/dp-1/stats", url.Values{})

			Expect(err).ToNot(HaveOccurred())
			Expect(seen).ToNot(HaveKey("Authorization"))
		})

		It("should let an explicit Authorization header win", func() {
			cp := controlPlaneWithToken(srv, "admin-token", "Authorization=Bearer caller-token")

			_, err := cp.InspectEnvoyProxy("/meshes/default/dataplanes/dp-1/stats", url.Values{})

			Expect(err).ToNot(HaveOccurred())
			Expect(seen.Get("Authorization")).To(Equal("Bearer caller-token"))
		})
	})
})
