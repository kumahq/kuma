package api_server_test

import (
	"fmt"
	"io"
	"net/http"
	"path"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	config "github.com/kumahq/kuma/pkg/config/api-server"
	"github.com/kumahq/kuma/pkg/metrics"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/pkg/test/matchers"
)

var _ = Describe("Versions WS", func() {
	It("should return the supported versions", func() {
		// setup
		resourceStore := memory.NewStore()
		metrics, err := metrics.NewMetrics("Standalone")
		Expect(err).ToNot(HaveOccurred())
		apiServer := createTestApiServer(resourceStore, config.DefaultApiServerConfig(), true, metrics)

		stop := make(chan struct{})
		go func() {
			defer GinkgoRecover()
			err := apiServer.Start(stop)
			Expect(err).ToNot(HaveOccurred())
		}()

		// when
		var resp *http.Response
		Eventually(func() error {
			r, err := http.Get(fmt.Sprintf("http://%s/versions", apiServer.Address()))
			resp = r
			return err
		}, "3s").ShouldNot(HaveOccurred())

		// then
		bytes, err := io.ReadAll(resp.Body)
		Expect(err).ToNot(HaveOccurred())
		Expect(bytes).To(matchers.MatchGoldenJSON(path.Join("testdata", "versions.json")))
	})
})
