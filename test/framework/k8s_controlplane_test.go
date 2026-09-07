package framework

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/v3/test/framework/portforward"
)

var _ = Describe("K8sControlPlane", func() {
	// An empty token leaves the cache cold, and the authn type below makes
	// retrieveAdminToken answer without reading anything.
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

	// The Helm universal path reads the token over HTTP rather than from a
	// Secret, so a server stands in for the control plane.
	universalControlPlane := func(srv *httptest.Server, env map[string]string) *K8sControlPlane {
		if env == nil {
			env = map[string]string{}
		}
		return &K8sControlPlane{
			t:       GinkgoT(),
			portFwd: portforward.Tunnel{Endpoint: srv.Listener.Addr().String()},
			cluster: &K8sCluster{opts: kumaDeploymentOptions{
				env:      env,
				helmOpts: map[string]string{"controlPlane.environment": "universal"},
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

		DescribeTable("should not look for a secret that is never bootstrapped",
			func(value string) {
				cp := &K8sControlPlane{cluster: &K8sCluster{opts: kumaDeploymentOptions{
					env: map[string]string{"KUMA_API_SERVER_AUTHN_TOKENS_BOOTSTRAP_ADMIN_TOKEN": value},
				}}}

				Expect(cp.retrieveAdminToken()).To(BeEmpty())
			},
			// Every spelling ParseBool reads as off.
			Entry("false", "false"),
			Entry("False", "False"),
			Entry("0", "0"),
			Entry("f", "f"),
		)

		It("should not read the universal secret over a loopback that is not admin", func() {
			var reads int
			srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { reads++ }))
			DeferCleanup(srv.Close)
			cp := universalControlPlane(srv, nil)

			Expect(cp.retrieveAdminToken()).To(BeEmpty())
			Expect(reads).To(BeZero())
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

		It("should read the secret once and answer from the cache after that", func() {
			var reads int
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				reads++
				_, _ = fmt.Fprintf(w, `{"data":%q}`, base64.StdEncoding.EncodeToString([]byte("admin-token")))
			}))
			DeferCleanup(srv.Close)
			cp := universalControlPlane(srv, map[string]string{
				"KUMA_API_SERVER_AUTHN_LOCALHOST_IS_ADMIN": "true",
			})

			Expect(cp.cachedAdminToken()).To(Equal("admin-token"))
			Expect(cp.cachedAdminToken()).To(Equal("admin-token"))
			Expect(reads).To(Equal(1))
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
