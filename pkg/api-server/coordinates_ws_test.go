package api_server_test

import (
	"fmt"
	kuma_cp "github.com/Kong/kuma/pkg/config/app/kuma-cp"
	"github.com/Kong/kuma/pkg/plugins/resources/memory"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"net/http"
)

var _ = Describe("Components WS", func() {

	It("should return the components location", func() {
		// given
		cfg := kuma_cp.DefaultConfig()
		cfg.Hostname = "kuma.internal"
		cfg.DataplaneTokenServer.Local.Port = 1111
		cfg.DataplaneTokenServer.Public.Interface = "192.168.0.1"
		cfg.DataplaneTokenServer.Public.Port = 2222
		cfg.BootstrapServer.Port = 3333

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
			_, err := http.Get(fmt.Sprintf("http://localhost%s/coordinates", apiServer.Address()))
			return err
		}, "3s").ShouldNot(HaveOccurred())

		// when
		resp, err := http.Get(fmt.Sprintf("http://localhost%s/coordinates", apiServer.Address()))
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
				}
			}
		}
`
		Expect(body).To(MatchJSON(expected))
	})
})
