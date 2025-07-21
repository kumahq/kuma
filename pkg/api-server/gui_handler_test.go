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

	type testCase struct {
		urlPath      string
		expectedFile string
		basePath     string
		guiRootUrl   string
		disabled     bool
	}
	DescribeTable("should expose file", func(given testCase) {
		// given
		var apiServer *api_server.ApiServer
		var stop func()
		apiServer, _, stop = StartApiServer(NewTestApiServerConfigurer().WithConfigMutator(func(config *server.ApiServerConfig) {
			config.GUI.Enabled = !given.disabled
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
			// index-bjW8zAoh
			r := regexp.MustCompile(`index[\-.][A-Za-z0-9_-]+\.`).ReplaceAll(in, []byte("index."))
			r = regexp.MustCompile(`"version":"[^"]*"`).ReplaceAll(r, []byte(`"version":"0.0.0"`))
			if r[len(r)-1] != '\n' {
				r = append(r, '\n')
			}
			return r
		}, matchers.MatchGoldenEqual(filepath.Join("testdata", "gui", given.expectedFile))))
	},
		Entry("should serve index.html without path", testCase{
			urlPath:      "/gui",
			expectedFile: "index.html",
		}),
		Entry("should serve robots.txt correctly", testCase{
			urlPath:      "/gui/robots.txt",
			expectedFile: "robots.txt",
		}),
		Entry("should serve on different path", testCase{
			urlPath:      "/gui/meshes",
			expectedFile: "gui_other_files.html",
		}),
		Entry("should serve index.html with / path", testCase{
			urlPath:      "/gui/",
			expectedFile: "index.html",
		}),
		Entry("should serve index.html on alternative path", testCase{
			urlPath:      "/ui/foo",
			expectedFile: "gui_with_base_path.html",
			basePath:     "/ui",
		}),
		Entry("should serve index.html on alternative path with end /", testCase{
			urlPath:      "/ui/",
			expectedFile: "gui_with_base_path_with_slash.html",
			basePath:     "/ui/",
		}),
		Entry("should serve index.html with path from rootUrl", testCase{
			urlPath:      "/gui/",
			expectedFile: "gui_with_root_url.html",
			guiRootUrl:   "https://foo.com/gui/foo",
		}),
		Entry("should serve index.html with path from rootUrl even with basePath set", testCase{
			urlPath:      "/foo/",
			expectedFile: "gui_with_root_url_and_base_path.html",
			basePath:     "/foo",
			guiRootUrl:   "https://foo.com/gui/foo",
		}),
		Entry("should serve index.html on alternative path", testCase{
			urlPath:      "/ui/foo",
			expectedFile: "gui_with_base_path.html",
			basePath:     "/ui",
		}),
		Entry("should not serve index.html without path", testCase{
			urlPath:      "/gui",
			disabled:     true,
			expectedFile: "gui_disabled_index.html",
		}),
		Entry("disabled should serve index.html with / path and disabled:true", testCase{
			urlPath:      "/gui/",
			disabled:     true,
			expectedFile: "gui_disabled_index_with_slash.html",
		}),
		Entry("favicon and disabled:true", testCase{
			urlPath:      "/gui/robots.txt",
			disabled:     true,
			expectedFile: "robots.txt",
		}),
		Entry("should not serve config.json", testCase{
			urlPath:      "/gui/config.json",
			disabled:     true,
			expectedFile: "gui_disabled_config.html",
		}),
	)
})
