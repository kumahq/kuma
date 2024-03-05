package runtime_test

import (
	"context"
	"net/http"

	"github.com/emicklei/go-restful/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	test_api_server "github.com/kumahq/kuma/pkg/test/api_server"
	test_runtime "github.com/kumahq/kuma/pkg/test/runtime"
)

func return200ForPath(path string) func(ws *restful.WebService) error {
	return func(ws *restful.WebService) error {
		ws.Route(
			ws.GET(path).To(func(req *restful.Request, res *restful.Response) {
				res.WriteHeader(http.StatusOK)
			}),
		)

		return nil
	}
}

var _ = Describe("APIWebServiceCustomize", func() {
	var stop chan struct{}

	AfterEach(func() {
		if stop != nil {
			close(stop)
		}
	})

	It("should allow for multiple customization functions", func() {
		// given
		cfg := kuma_cp.DefaultConfig()
		cfg.ApiServer.HTTPS.Enabled = false
		cfg.ApiServer.HTTP.Interface = "127.0.0.1"
		builder, err := test_runtime.BuilderFor(context.Background(), cfg)
		Expect(err).ToNot(HaveOccurred())

		// when
		rt, err := builder.
			WithAPIWebServiceCustomize(return200ForPath("/foo")).
			WithAPIWebServiceCustomize(return200ForPath("/bar")).
			Build()
		Expect(err).ToNot(HaveOccurred())

		apiServer, err := test_api_server.NewApiServer(cfg, rt)
		Expect(err).To(Succeed())

		stop = make(chan struct{})
		go func() {
			defer GinkgoRecover()

			Expect(apiServer.Start(stop)).To(Succeed())
		}()

		// then
		Eventually(func(g Gomega) {
			fooRes, err := http.Get("http://" + apiServer.Address() + "/foo")
			g.Expect(err).To(Succeed())
			g.Expect(fooRes.StatusCode).To(Equal(http.StatusOK))

			barRes, err := http.Get("http://" + apiServer.Address() + "/bar")
			g.Expect(err).To(Succeed())
			g.Expect(barRes.StatusCode).To(Equal(http.StatusOK))
		}).Should(Succeed())
	})
})
