package coordinates_test

import (
	"encoding/json"
	kuma_cp "github.com/Kong/kuma/pkg/config/app/kuma-cp"
	"github.com/Kong/kuma/pkg/coordinates"
	coordinates_client "github.com/Kong/kuma/pkg/coordinates/client"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net/http"
	"net/http/httptest"
)

var _ = Describe("Coordinates client", func() {
	It("should return server coordinates", func() {
		// given
		cfg := kuma_cp.DefaultConfig()
		cfg.Hostname = "kuma.internal"
		cfg.DataplaneTokenServer.Local.Port = 1111
		cfg.DataplaneTokenServer.Public.Interface = "192.168.0.1"
		cfg.DataplaneTokenServer.Public.Port = 2222
		cfg.BootstrapServer.Port = 3333

		expected := coordinates.Coordinates{
			Apis: coordinates.Apis{
				Bootstrap: coordinates.BootstrapApi{
					Url: "http://kuma.internal:3333",
				},
				DataplaneToken: coordinates.DataplaneTokenApi{
					LocalUrl:  "http://localhost:1111",
					PublicUrl: "https://kuma.internal:2222",
				},
			},
		}

		// setup
		mux := http.NewServeMux()
		server := httptest.NewServer(mux)
		mux.HandleFunc("/coordinates", func(writer http.ResponseWriter, req *http.Request) {
			defer GinkgoRecover()
			coords := coordinates.FromConfig(cfg)
			bytes, err := json.Marshal(coords)
			Expect(err).ToNot(HaveOccurred())
			_, err = writer.Write(bytes)
			Expect(err).ToNot(HaveOccurred())
		})

		// when
		client, err := coordinates_client.NewCoordinatesClient(server.URL)

		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		coordinates, err := client.Coordinates()

		// then
		Expect(coordinates).To(Equal(expected))
	})

	It("should throw an error on invalid status code", func() {
		mux := http.NewServeMux()
		server := httptest.NewServer(mux)
		mux.HandleFunc("/coordinates", func(writer http.ResponseWriter, req *http.Request) {
			writer.WriteHeader(500)
		})

		// when
		client, err := coordinates_client.NewCoordinatesClient(server.URL)

		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		_, err = client.Coordinates()

		// then
		Expect(err).To(MatchError("unexpected status code 500. Expected 200"))
	})
})
