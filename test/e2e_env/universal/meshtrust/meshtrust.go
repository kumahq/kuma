package meshtrust

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/v2/test/framework"
	"github.com/kumahq/kuma/v2/test/framework/envs/universal"
)

func MeshTrust() {
	It("should create MeshTrust resource", func() {
		meshName := "meshtrust-test"
		cluster := universal.Cluster

		err := cluster.Install(framework.MeshUniversal(meshName))
		Expect(err).ToNot(HaveOccurred())

		yaml := `
								type: MeshTrust
								mesh: meshtrust-test
								name: trust-1
								spec:
								  trustDomain: example.com
								  caBundles:
								  - type: Pem
								    pem:
								      value: |
								        -----BEGIN CERTIFICATE-----
								        MIIB9jCCAZ2gAwIBAgIUJ1gLZ/fvZhGq51qWrJzL6z2XWoQwCgYIKoZIzj0EAwIw
								        EjEQMA4GA1UEChMHVGVzdCBDQTAeFw0yNTA3MzAxMjAwMDBaFw0zNTA3MjcxMjAw
								        MDBaMBIxEDAOBgNVBAoTB1Rlc3QgQ0EwWTATBgcqhkjOPQIBBggqhkjOPQMBBwNC
								        AAQaXDFzOPslZ4e8n2KjsNkG+Wxi37L2zRWdMczDi7VqLrO03lczkB98/vzrdMKF
								        JgYxycULx10EYgMGQkgLrf1po2QwYjAdBgNVHQ4EFgQUUQBd5VjEO3N4XcgrxgMK
								        NU9xIQswHwYDVR0jBBgwFoAUUQBd5VjEO3N4XcgrxgMKNU9xIQswDwYDVR0TAQH/
								        BAUwAwEB/zAKBggqhkjOPQQDAgNHADBEAiA3dhhIQCzNkeGSjj6jK+jGE8fEKVmp
								        c9Vh+kJkmPUJZQIgQBr2GkV8uSfq/5ZKHD6jz6MJvKsg06dMBdvZBIA2ujg=
								        -----END CERTIFICATE-----
								  origin:
								    kri: kri-test-value
								`

		err = cluster.GetKumactlOptions().KumactlApplyFromString(yaml)
		Expect(err).ToNot(HaveOccurred())

		// Verify it was created
		out, err := cluster.GetKumactlOptions().RunKumactlAndGetOutput("get", "meshtrust", "-m", meshName, "trust-1", "-o", "yaml")
		Expect(err).ToNot(HaveOccurred())
		Expect(out).To(ContainSubstring("trustDomain: example.com"))
		Expect(out).To(ContainSubstring("origin:\n    kri: kri-test-value"))
		Expect(out).To(ContainSubstring("status: {}")) // status field should be present but empty

		// Try to create a MeshTrust with status.origin set by the user (should fail)
		invalidYaml := `
						type: MeshTrust
						mesh: meshtrust-test
						name: trust-2
						spec:
						  trustDomain: invalid.com
						  caBundles:
						  - type: Pem
						    pem:
						      value: |
						        -----BEGIN CERTIFICATE-----
						        MIIB9jCCAZ2gAwIBAgIUJ1gLZ/fvZhGq51qWrJzL6z2XWoQwCgYIKoZIzj0EAwIw
						        EjEQMA4GA1UEChMHVGVzdCBDQTAeFw0yNTA3MzAxMjAwMDBaFw0zNTA3MjcxMjAw
						        MDBaMBIxEDAOBgNVBAoTB1Rlc3QgQ0EwWTATBgcqhkjOPQIBBggqhkjOPQMBBwNC
						        AAQaXDFzOPslZ4e8n2KjsNkG+Wxi37L2zRWdMczDi7VqLrO03lczkB98/vzrdMKF
						        JgYxycULx10EYgMGQkgLrf1po2QwYjAdBgNVHQ4EFgQUUQBd5VjEO3N4XcgrxgMK
						        NU9xIQswHwYDVR0jBBgwFoAUUQBd5VjEO3N4XcgrxgMKNU9xIQswDwYDVR0TAQH/
						        BAUwAwEB/zAKBggqhkjOPQQDAgNHADBEAiA3dhhIQCzNkeGSjj6jK+jGE8fEKVmp
						        c9Vh+kJkmPUJZQIgQBr2GkV8uSfq/5ZKHD6jz6MJvKsg06dMBdvZBIA2ujg=
						        -----END CERTIFICATE-----
						status:
						  origin:
						    kri: user-set-kri
						`
		apiServerUrl := fmt.Sprintf("https://%s", net.JoinHostPort(cluster.GetKuma().(*framework.UniversalControlPlane).Networking().IP, "5678"))
		req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/meshes/%s/meshtrusts", apiServerUrl, meshName), bytes.NewBufferString(invalidYaml))
		Expect(err).ToNot(HaveOccurred())
		req.Header.Set("Content-Type", "application/x-yaml")

		cp := cluster.GetKuma()
		jsonOutput, _, err := cp.Exec("curl", "-s", "--fail", "--show-error", "http://localhost:5681/global-secrets/admin-user-token")
		Expect(err).ToNot(HaveOccurred())
		adminToken, err := framework.ExtractSecretDataFromResponse(jsonOutput)
		Expect(err).ToNot(HaveOccurred())
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", adminToken))

		client := &http.Client{Timeout: 10 * time.Second, Transport: &http.Transport{
			// #nosec G402
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}}
		resp, err := client.Do(req)
		Expect(err).ToNot(HaveOccurred())
		defer resp.Body.Close()

		Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
		respBody, err := io.ReadAll(resp.Body)
		Expect(err).ToNot(HaveOccurred())
		Expect(string(respBody)).To(ContainSubstring("status.origin: field is read-only"))

		// Clean up
		err = cluster.GetKumactlOptions().KumactlDelete("mesh", meshName, "")
		Expect(err).ToNot(HaveOccurred())
	})
}
