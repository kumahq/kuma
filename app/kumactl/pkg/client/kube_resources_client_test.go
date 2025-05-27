package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	util_http "github.com/kumahq/kuma/pkg/util/http"
)

var _ = Describe("Kube Resources Client", func() {
	It("should return kube resource", func() {
		// given
		mux := http.NewServeMux()
		server := httptest.NewServer(mux)
		defer server.Close()
		mux.HandleFunc("/globalsecrets/zone-token-signing-key-1", func(writer http.ResponseWriter, req *http.Request) {
			defer GinkgoRecover()
			Expect(req.Header.Get("accept")).To(Equal("application/json"))
			Expect(req.URL.Query().Get("format")).To(Equal("k8s"))

			resp := `{
 "kind": "Secret",
 "apiVersion": "v1",
 "metadata": {
  "name": "zone-token-signing-key-1",
  "namespace": "kuma-system",
  "creationTimestamp": "2024-01-03T15:33:33Z"
 },
 "data": {
  "value": "XYZ"
 },
 "type": "system.kuma.io/global-secret"
}`

			_, err := writer.Write([]byte(resp))
			Expect(err).ToNot(HaveOccurred())
		})
		serverURL, err := url.Parse(server.URL)
		Expect(err).ToNot(HaveOccurred())

		kubeResClient := NewHTTPKubernetesResourcesClient(
			util_http.ClientWithBaseURL(http.DefaultClient, serverURL, nil),
			registry.Global().ObjectDescriptors(),
		)

		// when
		obj, err := kubeResClient.Get(context.Background(), system.NewGlobalSecretResource().Descriptor(), "zone-token-signing-key-1", model.NoMesh)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(obj["kind"]).To(Equal("Secret"))
	})
})
