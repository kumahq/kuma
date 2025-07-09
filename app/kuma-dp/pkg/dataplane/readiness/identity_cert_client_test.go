package readiness_test

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/kumahq/kuma/app/kuma-dp/pkg/dataplane/readiness"
	"github.com/kumahq/kuma/pkg/test"
	kuma_tls "github.com/kumahq/kuma/pkg/tls"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"io"
	"net/http"
	"testing"
	"time"
)

func TestReadinessReporter(t *testing.T) {
	test.RunSpecs(t, "Readiness Reporter Test Suite")
}

var _ = Describe("IdentityCertClient", func() {

	It("should return ready when adminPort is 0", func() {
		checkClient := createCheckClient(0, false, "")

		ready, err := checkClient.CheckIfReady()

		Expect(err).ToNot(HaveOccurred())
		Expect(ready).To(BeTrue())
	})

	It("should return ready when no identity cert is found", func() {
		checkClient := createCheckClient(9901, true, "{}")

		ready, err := checkClient.CheckIfReady()

		Expect(err).ToNot(HaveOccurred())
		Expect(ready).To(BeTrue())
	})

	It("should return ready identity cert is initialized and not expired", func() {
		certNotExpired := time.Now().UTC().Add(time.Hour)
		okIdentityCertResp := createIdentityCertDump(certNotExpired)
		checkClient := createCheckClient(9901, true, okIdentityCertResp)

		ready, err := checkClient.CheckIfReady()

		Expect(err).ToNot(HaveOccurred())
		Expect(ready).To(BeTrue())
	})

	It("should return not ready when envoy is not listening when it should be listening", func() {
		checkClient := createCheckClient(9901, false, "")

		ready, err := checkClient.CheckIfReady()

		Expect(err).ToNot(BeNil())
		Expect(err.Error()).To(ContainSubstring("could not request envoy"))
		Expect(ready).To(BeFalse())
	})

	It("should return not ready identity cert is expired", func() {
		certExpired := time.Now().UTC().Add(-time.Hour)
		expiredIdentityCertResp := createIdentityCertDump(certExpired)
		checkClient := createCheckClient(9901, true, expiredIdentityCertResp)

		ready, err := checkClient.CheckIfReady()

		Expect(err).ToNot(HaveOccurred())
		Expect(ready).To(BeTrue())
	})

	It("should return not ready identity cert is uninitialized", func() {
		identityCertNotInitializedResp := `
{
 "configs": [
    {
      "name": "identity_cert:secret:mesh-dev",
      "version_info": "uninitialized",
      "last_updated": "2025-07-09T01:39:32.075Z",
      "secret": {
        "@type": "type.googleapis.com/envoy.extensions.transport_sockets.tls.v3.Secret",
        "name": "identity_cert:secret:mesh-dev"
      }
    }
 ]
} 
`
		checkClient := createCheckClient(9901, true, identityCertNotInitializedResp)

		ready, err := checkClient.CheckIfReady()

		Expect(err).ToNot(HaveOccurred())
		Expect(ready).To(BeTrue())
	})
})

func createCheckClient(adminPort uint32, useMockResponse bool, mockResponse string) *readiness.IdentityCertClient {
	httpClient := http.DefaultClient
	if useMockResponse {
		httpClient = mockHttpClient(mockResponse)
	}

	return &readiness.IdentityCertClient{EnvoyAdminAddress: "127.0.0.1", EnvoyAdminPort: adminPort,
		HttpClient: httpClient}
}

func createIdentityCertDump(certExpiry time.Time) string {
	originalValidity := kuma_tls.DefaultValidityPeriod
	defer func() { kuma_tls.DefaultValidityPeriod = originalValidity }()

	validity := certExpiry.Sub(time.Now().UTC())
	kuma_tls.DefaultValidityPeriod = validity
	certKeyPair, err := kuma_tls.NewSelfSignedCert("server", kuma_tls.DefaultKeyType, "dp-name")
	if err != nil {
		panic(fmt.Errorf("could not generate self-signed certificate: %s", err))
	}

	certBytesBase64 := base64.StdEncoding.EncodeToString(certKeyPair.CertPEM)
	return fmt.Sprintf(`
{
 "configs": [
  {
   "@type": "type.googleapis.com/envoy.admin.v3.SecretsConfigDump.DynamicSecret",
   "name": "identity_cert:secret:default",
   "version_info": "2db9701c-e978-406a-94aa-9514a9fe6841",
   "last_updated": "2025-07-09T01:39:32.075Z",
   "secret": {
    "@type": "type.googleapis.com/envoy.extensions.transport_sockets.tls.v3.Secret",
    "name": "identity_cert:secret:default",
    "tls_certificate": {
     "certificate_chain": {
      "inline_bytes": "%s"
     },
     "private_key": {
      "inline_bytes": "W3JlZGFjdGVkXQ=="
     }
    }
   }
  }
 ]
}
`, certBytesBase64)
}

func mockHttpClient(response string) *http.Client {
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(response)),
	}
	mockRT := &mockRoundTripper{Response: resp}
	return &http.Client{Transport: mockRT}
}

type mockRoundTripper struct {
	Response *http.Response
}

func (m *mockRoundTripper) RoundTrip(*http.Request) (*http.Response, error) {
	return m.Response, nil
}
