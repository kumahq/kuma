package api_server_test

import (
	"fmt"
	"github.com/Kong/kuma/pkg/config"
	api_server_config "github.com/Kong/kuma/pkg/config/api-server"
	kuma_cp "github.com/Kong/kuma/pkg/config/app/kuma-cp"
	"github.com/Kong/kuma/pkg/plugins/resources/memory"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"net/http"
)

var _ = Describe("Config WS", func() {

	It("should return the config", func() {
		// given
		cfg := api_server_config.DefaultApiServerConfig()

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
			_, err := http.Get(fmt.Sprintf("http://localhost%s/config", apiServer.Address()))
			return err
		}, "3s").ShouldNot(HaveOccurred())

		// when
		resp, err := http.Get(fmt.Sprintf("http://localhost%s/config", apiServer.Address()))
		Expect(err).ToNot(HaveOccurred())

		// then
		body, err := ioutil.ReadAll(resp.Body)
		Expect(err).ToNot(HaveOccurred())

		// when
		expectedConfig := kuma_cp.DefaultConfig()
		expectedConfig.ApiServer = cfg
		cfgJson, err := config.ConfigForDisplayJson(&expectedConfig)
		Expect(err).ToNot(HaveOccurred())

		// then
		Expect(cfgJson).To(MatchJSON(body))
	})
})
