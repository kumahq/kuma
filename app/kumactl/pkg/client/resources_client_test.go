package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	util_http "github.com/kumahq/kuma/pkg/util/http"
)

var _ = Describe("Resources Client", func() {
	It("should return resources list", func() {
		// given
		mux := http.NewServeMux()
		server := httptest.NewServer(mux)
		defer server.Close()
		mux.HandleFunc("/_resources", func(writer http.ResponseWriter, req *http.Request) {
			defer GinkgoRecover()
			Expect(req.Header.Get("accept")).To(Equal("application/json"))

			resp := `{
 "resources": [
  {
   "includeInDump": true,
   "name": "CircuitBreaker",
   "path": "circuit-breakers",
   "pluralDisplayName": "Circuit Breakers",
   "policy": {
    "hasFromTargetRef": false,
    "hasToTargetRef": false,
    "isTargetRef": false
   },
   "readOnly": false,
   "scope": "Mesh",
   "singularDisplayName": "Circuit Breaker"
  }
 ]
}`

			_, err := writer.Write([]byte(resp))
			Expect(err).ToNot(HaveOccurred())
		})
		serverURL, err := url.Parse(server.URL)
		Expect(err).ToNot(HaveOccurred())

		resListClient := NewHTTPResourcesListClient(
			util_http.ClientWithBaseURL(http.DefaultClient, serverURL, nil),
		)

		// when
		obj, err := resListClient.List(context.Background())

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(obj.Resources).To(HaveLen(1))
		Expect(obj.Resources[0].Name).To(Equal("CircuitBreaker"))
	})
})
