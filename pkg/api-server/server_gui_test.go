package api_server_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	config "github.com/kumahq/kuma/pkg/config/api-server"
	gui_server "github.com/kumahq/kuma/pkg/config/gui-server"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
)

var _ = Describe("GUI Server", func() {

	var stop chan struct{}
	var baseUrl string

	beforeEach := func(enabelGUI bool) {
		// given
		cfg := config.DefaultApiServerConfig()

		// setup
		resourceStore := memory.NewStore()
		apiServer := createTestApiServer(resourceStore, cfg, enabelGUI)

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
		apiServer.GuiServerConfig = &gui_server.GuiServerConfig{
			GuiConfig: &gui_server.GuiConfig{
				ApiUrl:      "http://localhost:5681",
				Environment: "kubernetes",
			},
		}
	}

	Describe("enabled",
		func() {

			BeforeEach(func() { beforeEach(true) })

			AfterEach(func() {
				close(stop)
			})

			type testCase struct {
				urlPath      string
				expectedFile string
			}
			DescribeTable("should expose file",
				func(given testCase) {
					// when
					resp, err := http.Get(fmt.Sprintf("%s%s", baseUrl, given.urlPath))

					// then
					Expect(err).ToNot(HaveOccurred())

					// when
					received, err := ioutil.ReadAll(resp.Body)

					// then
					Expect(resp.Body.Close()).To(Succeed())
					Expect(err).ToNot(HaveOccurred())

					// when
					fileContent, err := ioutil.ReadFile(filepath.Join("..", "..", "app", "kuma-ui", "data", "resources", given.expectedFile))

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
			BeforeEach(func() { beforeEach(false) })

			AfterEach(func() {
				close(stop)
			})

			type testCase struct {
				urlPath  string
				expected string
			}
			DescribeTable("should not expose file",
				func(given testCase) {
					// when
					resp, err := http.Get(fmt.Sprintf("%s%s", baseUrl, given.urlPath))

					// then
					Expect(err).ToNot(HaveOccurred())

					// when
					received, err := ioutil.ReadAll(resp.Body)

					// then
					Expect(resp.Body.Close()).To(Succeed())
					Expect(err).ToNot(HaveOccurred())

					// then
					Expect(err).ToNot(HaveOccurred())
					Expect(string(received)).To(Equal(given.expected))
				},
				Entry("should not serve index.html without path", testCase{
					urlPath:  "/gui",
					expected: "GUI is disabled. If this is a Remote CP, please check the GUI on the Global CP.",
				}),
				Entry("should not serve index.html with / path", testCase{
					urlPath:  "/gui/",
					expected: "GUI is disabled. If this is a Remote CP, please check the GUI on the Global CP.",
				}),
			)
		})

})
