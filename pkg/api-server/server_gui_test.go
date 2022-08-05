package api_server_test

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	api_server "github.com/kumahq/kuma/pkg/api-server"
)

var _ = Describe("GUI Server", func() {

	var baseUrl string

	Describe("enabled", func() {

		type testCase struct {
			urlPath      string
			expectedFile string
		}
		DescribeTable("should expose file", func(given testCase) {
			// given
			var apiServer *api_server.ApiServer
			var stop func()
			apiServer, _, stop = StartApiServer(NewTestApiServerConfigurer().WithGui())
			baseUrl = "http://" + apiServer.Address()
			defer stop()

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
			type testCase struct {
				urlPath  string
				expected string
			}
			DescribeTable("should not expose file", func(given testCase) {
				// given
				var apiServer *api_server.ApiServer
				var stop func()
				apiServer, _, stop = StartApiServer(NewTestApiServerConfigurer())
				baseUrl = "http://" + apiServer.Address()
				defer stop()

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
