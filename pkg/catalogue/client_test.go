package catalogue_test

import (
	"encoding/json"
	"github.com/Kong/kuma/pkg/catalogue"
	catalogue_client "github.com/Kong/kuma/pkg/catalogue/client"
	config_catalogue "github.com/Kong/kuma/pkg/config/api-server/catalogue"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net/http"
	"net/http/httptest"
)

var _ = Describe("Catalogue client", func() {
	It("should return server catalogue", func() {
		// given
		catCfg := config_catalogue.CatalogueConfig{
			Bootstrap: config_catalogue.BootstrapApiConfig{
				Url: "http://kuma.internal:3333",
			},
			DataplaneToken: config_catalogue.DataplaneTokenApiConfig{
				LocalUrl:  "http://localhost:1111",
				PublicUrl: "https://kuma.internal:2222",
			},
		}

		expected := catalogue.Catalogue{
			Apis: catalogue.Apis{
				Bootstrap: catalogue.BootstrapApi{
					Url: "http://kuma.internal:3333",
				},
				DataplaneToken: catalogue.DataplaneTokenApi{
					LocalUrl:  "http://localhost:1111",
					PublicUrl: "https://kuma.internal:2222",
				},
			},
		}

		// setup
		mux := http.NewServeMux()
		server := httptest.NewServer(mux)
		defer server.Close()
		mux.HandleFunc("/catalogue", func(writer http.ResponseWriter, req *http.Request) {
			defer GinkgoRecover()
			cat := catalogue.FromConfig(catCfg)
			bytes, err := json.Marshal(cat)
			Expect(err).ToNot(HaveOccurred())
			_, err = writer.Write(bytes)
			Expect(err).ToNot(HaveOccurred())
		})

		// when
		client, err := catalogue_client.NewCatalogueClient(server.URL)

		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		catalogue, err := client.Catalogue()

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(catalogue).To(Equal(expected))
	})

	It("should throw an error on invalid status code", func() {
		mux := http.NewServeMux()
		server := httptest.NewServer(mux)
		defer server.Close()
		mux.HandleFunc("/catalogue", func(writer http.ResponseWriter, req *http.Request) {
			writer.WriteHeader(500)
		})

		// when
		client, err := catalogue_client.NewCatalogueClient(server.URL)

		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		_, err = client.Catalogue()

		// then
		Expect(err).To(MatchError("unexpected status code 500. Expected 200"))
	})
})
