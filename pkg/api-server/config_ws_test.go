package api_server_test

import (
	"fmt"
	"io"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v2"

	api_server "github.com/kumahq/kuma/pkg/api-server"
	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
)

var _ = Describe("Config WS", func() {
	var stop func()
	var apiServer *api_server.ApiServer
	apiServer, stop = StartApiServer(NewTestApiServerConfigurer())
	AfterEach(func() {
		stop()
	})

	It("should return the config", func() {
		// when
		resp, err := http.Get(fmt.Sprintf("http://%s/config", apiServer.Address()))
		Expect(err).ToNot(HaveOccurred())

		// then
		body, err := io.ReadAll(resp.Body)
		Expect(err).ToNot(HaveOccurred())

		// when
		parsedCfg := kuma_cp.Config{}
		Expect(yaml.Unmarshal(body, &parsedCfg)).To(Succeed())
	})
})
