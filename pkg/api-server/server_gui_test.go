package api_server_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"

	types "github.com/kumahq/kuma/pkg/api-server/types"

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

	guiConfig := types.GuiConfig{
		ApiUrl:      "http://localhost:5681",
		Environment: "kubernetes",
	}

	BeforeEach(func() {
		// given
		cfg := config.DefaultApiServerConfig()

		// setup
		resourceStore := memory.NewStore()
		apiServer := createTestApiServer(resourceStore, cfg)

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
	})

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

	It("should serve the gui config", func() {
		// when
		resp, err := http.Get(fmt.Sprintf("%s/gui/config", baseUrl))

		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		received, err := ioutil.ReadAll(resp.Body)

		// then
		Expect(resp.Body.Close()).To(Succeed())
		Expect(err).ToNot(HaveOccurred())

		// and
		Expect(resp.Header.Get("content-type")).To(Equal("application/json"))

		// when
		cfg := types.GuiConfig{}
		Expect(json.Unmarshal(received, &cfg)).To(Succeed())

		// then
		Expect(cfg).To(Equal(guiConfig))
	})

	It("should proxy requests to api server", func() {
		// when
		resp, err := http.Get(fmt.Sprintf("%s/api/meshes", baseUrl))

		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		received, err := ioutil.ReadAll(resp.Body)

		// then
		Expect(resp.Body.Close()).To(Succeed())
		Expect(err).ToNot(HaveOccurred())

		// and
		Expect(resp.Header.Get("content-type")).To(Equal("application/json"))

		// and
		Expect(string(received)).To(Equal(`{
 "total": 0,
 "items": [],
 "next": null
}
`))
	})

})
