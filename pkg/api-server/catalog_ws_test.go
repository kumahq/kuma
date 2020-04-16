package api_server_test

import (
	"fmt"
	"io/ioutil"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	config "github.com/Kong/kuma/pkg/config/api-server"
	"github.com/Kong/kuma/pkg/plugins/resources/memory"
)

var _ = Describe("Catalog WS", func() {

	It("should return the api catalog", func() {
		// given
		cfg := config.DefaultApiServerConfig()
		cfg.Catalog.Admin.LocalUrl = "http://localhost:1111"
		cfg.Catalog.Admin.PublicUrl = "https://kuma.internal:2222"
		cfg.Catalog.DataplaneToken.LocalUrl = "http://localhost:1111"
		cfg.Catalog.DataplaneToken.PublicUrl = "https://kuma.internal:2222"
		cfg.Catalog.Bootstrap.Url = "http://kuma.internal:3333"
		cfg.Catalog.MonitoringAssignment.Url = "grpc://kuma.internal:4444"
		cfg.Catalog.Sds.Url = "https://sds.kuma.io:5555"

		// setup
		resourceStore := memory.NewStore()
		apiServer := createTestApiServer(resourceStore, cfg)

		stop := make(chan struct{})
		go func() {
			defer GinkgoRecover()
			err := apiServer.Start(stop)
			Expect(err).ToNot(HaveOccurred())
		}()

		// wait for the server
		Eventually(func() error {
			_, err := http.Get(fmt.Sprintf("http://localhost%s/catalog", apiServer.Address()))
			return err
		}, "3s").ShouldNot(HaveOccurred())

		// when
		resp, err := http.Get(fmt.Sprintf("http://localhost%s/catalog", apiServer.Address()))
		Expect(err).ToNot(HaveOccurred())

		// then
		body, err := ioutil.ReadAll(resp.Body)
		Expect(err).ToNot(HaveOccurred())

		expected := `
		{
			"apis": {
				"bootstrap": {
					"url": "http://kuma.internal:3333"
				},
				"dataplaneToken": {
					"localUrl": "http://localhost:1111",
					"publicUrl": "https://kuma.internal:2222"
				},
				"admin": {
					"localUrl": "http://localhost:1111",
					"publicUrl": "https://kuma.internal:2222"
				},
				"monitoringAssignment": {
					"url": "grpc://kuma.internal:4444"
				}
			}
		}
`
		Expect(body).To(MatchJSON(expected))
	})
})
