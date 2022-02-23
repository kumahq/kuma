package api_server_test

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	config "github.com/kumahq/kuma/pkg/config/api-server"
	"github.com/kumahq/kuma/pkg/metrics"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
)

var _ = Describe("GUI Server", func() {

	var stop chan struct{}
	var baseUrl string

	setupServer := func(enabelGUI bool) {
		// given
		cfg := config.DefaultApiServerConfig()

		// setup
		resourceStore := memory.NewStore()
		metrics, err := metrics.NewMetrics("Standalone")
		Expect(err).ToNot(HaveOccurred())
		apiServer := createTestApiServer(resourceStore, cfg, enabelGUI, metrics)

		stop = make(chan struct{})
		go func() {
			defer GinkgoRecover()
			err := apiServer.Start(stop)
			Expect(err).ToNot(HaveOccurred())
		}()
		port := strings.Split(apiServer.Address(), ":")[1]

		// wait for the server
		Eventually(func() error {
			_, err := http.Get(fmt.Sprintf("http://localhost:%s/config", port))
			return err
		}, "3s").ShouldNot(HaveOccurred())

		baseUrl = "http://localhost:" + port
	}

	Describe("enabled", func() {

		BeforeEach(func() { setupServer(true) })

		AfterEach(func() {
			close(stop)
		})

		type testCase struct {
			urlPath      string
			expectedFile string
		}
		DescribeTable("should expose file", func(given testCase) {
			// when
			resp, err := http.Get(fmt.Sprintf("%s%s", baseUrl, given.urlPath))

			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			received, err := io.ReadAll(resp.Body)

			// then
			Expect(resp.Body.Close()).To(Succeed())
			Expect(err).ToNot(HaveOccurred())

			// when
			fileContent, err := os.ReadFile(filepath.Join("..", "..", "app", "kuma-ui", "pkg", "resources", "data", given.expectedFile))

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(fileContent).To(Equal(received))
		},
			Entry("should serve index.html without path", testCase{
				urlPath:      "/gui",
				expectedFile: "index.html",
			}),
			Entry("should serve index.html with / path", testCase{
				urlPath:      "/gui/",
				expectedFile: "index.html",
			}),
		)
	})

	Describe("disabled",
		func() {
			BeforeEach(func() { setupServer(false) })

			AfterEach(func() {
				close(stop)
			})

			type testCase struct {
				urlPath  string
				expected string
			}
			DescribeTable("should not expose file", func(given testCase) {
				// when
				resp, err := http.Get(fmt.Sprintf("%s%s", baseUrl, given.urlPath))

				// then
				Expect(err).ToNot(HaveOccurred())

				// when
				received, err := io.ReadAll(resp.Body)

				// then
				Expect(resp.Body.Close()).To(Succeed())
				Expect(err).ToNot(HaveOccurred())

				// then
				Expect(err).ToNot(HaveOccurred())
				Expect(string(received)).To(ContainSubstring(given.expected))
			},
				Entry("should not serve index.html without path", testCase{
					urlPath:  "/gui",
					expected: "GUI is disabled. If this is a Zone CP, please check the GUI on the Global CP.",
				}),
				Entry("should not serve index.html with / path", testCase{
					urlPath:  "/gui/",
					expected: "GUI is disabled. If this is a Zone CP, please check the GUI on the Global CP.",
				}),
			)
		})

})
