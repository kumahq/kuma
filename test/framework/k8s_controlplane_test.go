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
	// controlPlaneWithToken returns a control plane pointed at srv. A non-empty
	// token is seeded into the cache; an empty one leaves the cache cold and
	// lets retrieveAdminToken answer from the cluster options, so nothing here
	// reads a secret either way.
	controlPlaneWithToken := func(srv *httptest.Server, token string, apiHeaders ...string) *K8sControlPlane {
		return &K8sControlPlane{
			portFwd:    portforward.Tunnel{Endpoint: srv.Listener.Addr().String()},
			apiHeaders: apiHeaders,
			adminToken: token,
			cluster: &K8sCluster{opts: kumaDeploymentOptions{
				env: map[string]string{"KUMA_API_SERVER_AUTHN_TYPE": "external"},
			}},
		}
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

	Describe("parseAPIHeaders", func() {
		It("should keep a value that contains the separator", func() {
			Expect(parseAPIHeaders([]string{"Authorization=Bearer abc=="})).
				To(Equal(map[string]string{"Authorization": "Bearer abc=="}))
		})

		It("should skip an entry that is not a pair", func() {
			Expect(parseAPIHeaders([]string{"Authorization"})).To(BeEmpty())
		})
	})

	Describe("cachedAdminToken", func() {
		It("should keep answering when there is no token to cache", func() {
			cp := &K8sControlPlane{cluster: &K8sCluster{opts: kumaDeploymentOptions{
				env: map[string]string{"KUMA_API_SERVER_AUTHN_TYPE": "external"},
			}}}

			Expect(cp.cachedAdminToken()).To(BeEmpty())
			Expect(cp.cachedAdminToken()).To(BeEmpty())
		})

		It("should answer from the cache once it holds a token", func() {
			// No cluster, so a second read would panic rather than pass.
			cp := &K8sControlPlane{adminToken: "admin-token"}

			Expect(cp.cachedAdminToken()).To(Equal("admin-token"))
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
