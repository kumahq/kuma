package customization_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/emicklei/go-restful"

	"github.com/kumahq/kuma/pkg/api-server/customization"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	api_server_config "github.com/kumahq/kuma/pkg/config/api-server"
	"github.com/kumahq/kuma/pkg/metrics"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
)

var _ = Describe("API Manager", func() {

	It("should return the config", func() {
		// given
		cfg := api_server_config.DefaultApiServerConfig()

		// setup
		resourceStore := memory.NewStore()
		metrics, err := metrics.NewMetrics("Standalone")
		Expect(err).ToNot(HaveOccurred())
		wsManager := customization.NewAPIList()

		ws := new(restful.WebService).Path("foo")
		ws.Route(ws.GET("baz").To(func(request *restful.Request, response *restful.Response) {
			_ = response.WriteAsJson("bar")
		}))
		wsManager.Add(ws)

		apiServer := createTestApiServer(resourceStore, cfg, true, metrics, wsManager)

		stop := make(chan struct{})
		go func() {
			defer GinkgoRecover()
			err := apiServer.Start(stop)
			Expect(err).ToNot(HaveOccurred())
		}()
		port := strings.Split(apiServer.Address(), ":")[1]

		// wait for the server
		Eventually(func() error {
			_, err := http.Get(fmt.Sprintf("http://localhost:%s/foo", port))
			return err
		}, "3s").ShouldNot(HaveOccurred())

		// when
		resp, err := http.Get(fmt.Sprintf("http://localhost:%s/foo/baz", port))
		Expect(err).ToNot(HaveOccurred())

		// then
		body, err := ioutil.ReadAll(resp.Body)
		Expect(err).ToNot(HaveOccurred())

		// when
		Expect(string(body)).To(Equal("\"bar\"\n"))
		close(stop)
	})
})
