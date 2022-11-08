package api_server_test

import (
	"fmt"
	"io"
	"net/http"

	"github.com/google/go-cmp/cmp"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/yaml"

	api_server "github.com/kumahq/kuma/pkg/api-server"
	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
)

var _ = Describe("Config WS", func() {
	var apiServer *api_server.ApiServer
	var cfg kuma_cp.Config
	var stop func()

	BeforeEach(func() {
		apiServer, cfg, stop = StartApiServer(NewTestApiServerConfigurer())
	})
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
		Expect(parsedCfg.Store.Postgres.Password).To(Equal("*****"))
		Expect(parsedCfg).To(BeComparableTo(cfg, cmp.FilterPath(
			func(p cmp.Path) bool { return p.String() == "Store.Postgres.Password" }, cmp.Ignore())),
		)
	})
})
