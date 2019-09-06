package api_server_test

import (
	api_server "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/config/api-server"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/plugins/resources/memory"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"net/http"
)

var _ = Describe("Index WS", func() {
	It("should return the version of Kuma Control Plane", func(done Done) {
		// setup
		resourceStore := memory.NewStore()
		apiServer := createTestApiServer(resourceStore, *api_server.DefaultApiServerConfig())

		stop := make(chan struct{})
		go func() {
			defer GinkgoRecover()
			err := apiServer.Start(stop)
			Expect(err).ToNot(HaveOccurred())
		}()

		// wait for the server
		Eventually(func() error {
			_, err := http.Get("http://localhost" + apiServer.Address())
			return err
		}, "3s").ShouldNot(HaveOccurred())

		// when
		resp, err := http.Get("http://localhost" + apiServer.Address())
		Expect(err).ToNot(HaveOccurred())

		// then
		body, err := ioutil.ReadAll(resp.Body)
		Expect(err).ToNot(HaveOccurred())

		expected := `
		{
			"tagline": "Kuma",
			"version": "0.1.0"
		}
`
		Expect(body).To(MatchJSON(expected))
		close(done)
	}, 5)
})
