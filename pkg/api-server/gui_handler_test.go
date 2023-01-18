package api_server_test

import (
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"regexp"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	api_server "github.com/kumahq/kuma/pkg/api-server"
	server "github.com/kumahq/kuma/pkg/config/api-server"
	"github.com/kumahq/kuma/pkg/test/matchers"
)

var _ = Describe("GUI Server", func() {

	var baseUrl string

	Describe("enabled", func() {

		type testCase struct {
			urlPath      string
			expectedFile string
			basePath     string
			guiRootUrl   string
		}
		DescribeTable("should expose file", func(given testCase) {
			// given
			var apiServer *api_server.ApiServer
			var stop func()
			apiServer, _, stop = StartApiServer(NewTestApiServerConfigurer().WithConfigMutator(func(config *server.ApiServerConfig) {
				config.GUI.Enabled = true
				config.RootUrl = "https://foo.bar.com:8080/foo"
				if given.basePath != "" {
					config.GUI.BasePath = given.basePath
				}
				if given.guiRootUrl != "" {
					config.GUI.RootUrl = given.guiRootUrl
				}
			}))
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

			Expect(received).To(WithTransform(func(in []byte) []byte {
				// Remove the part of the file name that changes always
				r := regexp.MustCompile(`index[\-\.][a-z0-9]+\.`).ReplaceAll(in, []byte("index."))
				r = regexp.MustCompile(`"[0-9]+\.[0-9]+\.[0-9]+[^"]*"`).ReplaceAll(r, []byte(`"0.0.0"`))
				r = regexp.MustCompile(`"unknown"`).ReplaceAll(r, []byte(`"0.0.0"`))
				if r[len(r)-1] != '\n' {
					r = append(r, '\n')
				}
				return r
			}, matchers.MatchGoldenEqual(given.expectedFile)))
		},
			Entry("should serve index.html without path", testCase{
				urlPath:      "/gui",
				expectedFile: filepath.Join("testdata", "index.html"),
			}),
			Entry("should serve robots.txt correctly", testCase{
				urlPath:      "/gui/robots.txt",
				expectedFile: filepath.Join("testdata", "robots.txt"),
			}),
			Entry("should serve on different path", testCase{
				urlPath:      "/gui/meshes",
				expectedFile: filepath.Join("testdata", "gui_other_files.html"),
			}),
			Entry("should serve index.html with / path", testCase{
				urlPath:      "/gui/",
				expectedFile: filepath.Join("testdata", "index.html"),
			}),
			Entry("should serve index.html on alternative path", testCase{
				urlPath:      "/ui/foo",
				expectedFile: filepath.Join("testdata", "gui_with_base_path.html"),
				basePath:     "/ui",
			}),
			Entry("should serve index.html on alternative path with end /", testCase{
				urlPath:      "/ui/",
				expectedFile: filepath.Join("testdata", "gui_with_base_path_with_slash.html"),
				basePath:     "/ui/",
			}),
			Entry("should serve index.html with path from rootUrl", testCase{
				urlPath:      "/gui/",
				expectedFile: filepath.Join("testdata", "gui_with_root_url.html"),
				guiRootUrl:   "https://foo.com/gui/foo",
			}),
			Entry("should serve index.html with path from rootUrl even with basePath set", testCase{
				urlPath:      "/foo/",
				expectedFile: filepath.Join("testdata", "gui_with_root_url_and_base_path.html"),
				basePath:     "/foo",
				guiRootUrl:   "https://foo.com/gui/foo",
			}),
			Entry("should serve index.html on alternative path", testCase{
				urlPath:      "/ui/foo",
				expectedFile: filepath.Join("testdata", "gui_with_base_path.html"),
				basePath:     "/ui",
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
				Entry("should not serve config.json", testCase{
					urlPath:  "/gui/config.json",
					expected: "GUI is disabled. If this is a Zone CP, please check the GUI on the Global CP.",
				}),
			)
		})

})
