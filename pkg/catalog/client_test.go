package catalog_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/Kong/kuma/pkg/catalog"
	catalog_client "github.com/Kong/kuma/pkg/catalog/client"
	config_catalog "github.com/Kong/kuma/pkg/config/api-server/catalog"
)

var _ = Describe("Catalog client", func() {
	It("should return server catalog", func() {
		// given
		catCfg := config_catalog.CatalogConfig{
			Bootstrap: config_catalog.BootstrapApiConfig{
				Url: "http://kuma.internal:3333",
			},
			Admin: config_catalog.AdminApiConfig{
				LocalUrl:  "http://localhost:1111",
				PublicUrl: "https://kuma.internal:2222",
			},
		}

		expected := catalog.Catalog{
			Apis: catalog.Apis{
				Bootstrap: catalog.BootstrapApi{
					Url: "http://kuma.internal:3333",
				},
				Admin: catalog.AdminApi{
					LocalUrl:  "http://localhost:1111",
					PublicUrl: "https://kuma.internal:2222",
				},
			},
		}

		// setup
		mux := http.NewServeMux()
		server := httptest.NewServer(mux)
		defer server.Close()
		mux.HandleFunc("/catalog", func(writer http.ResponseWriter, req *http.Request) {
			defer GinkgoRecover()
			cat := catalog.FromConfig(catCfg)
			bytes, err := json.Marshal(cat)
			Expect(err).ToNot(HaveOccurred())
			_, err = writer.Write(bytes)
			Expect(err).ToNot(HaveOccurred())
		})

		// when
		client, err := catalog_client.NewCatalogClient(server.URL)

		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		catalog, err := client.Catalog()

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(catalog).To(Equal(expected))
	})

	It("should throw an error on invalid status code", func() {
		mux := http.NewServeMux()
		server := httptest.NewServer(mux)
		defer server.Close()
		mux.HandleFunc("/catalog", func(writer http.ResponseWriter, req *http.Request) {
			writer.WriteHeader(500)
		})

		// when
		client, err := catalog_client.NewCatalogClient(server.URL)

		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		_, err = client.Catalog()

		// then
		Expect(err).To(MatchError("unexpected status code 500. Expected 200"))
	})
})
