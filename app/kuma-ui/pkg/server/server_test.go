package server_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Kong/kuma/app/kuma-ui/pkg/server"
	"github.com/Kong/kuma/app/kuma-ui/pkg/server/types"
	gui_server "github.com/Kong/kuma/pkg/config/gui-server"
	"github.com/Kong/kuma/pkg/test"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("GUI Server", func() {

	var stop chan struct{}
	var baseUrl string

	guiConfig := types.GuiConfig{
		ApiUrl:      "http://kuma.internal:5681",
		Environment: "kubernetes",
	}

	BeforeEach(func() {
		// setup api server
		mux := http.NewServeMux()
		mux.HandleFunc("/some-endpoint", func(writer http.ResponseWriter, request *http.Request) {
			defer GinkgoRecover()
			writer.WriteHeader(200)
			_, err := writer.Write([]byte(fmt.Sprintf("response from api server: %s", request.URL.Path)))
			Expect(err).ToNot(HaveOccurred())
		})
		apiSrv := httptest.NewServer(mux)
		apiSrvPort, err := strconv.Atoi(strings.Split(apiSrv.Listener.Addr().String(), ":")[1])
		Expect(err).ToNot(HaveOccurred())

		// setup gui server
		port, err := test.GetFreePort()
		Expect(err).ToNot(HaveOccurred())
		baseUrl = "http://localhost:" + strconv.Itoa(port)

		srv := server.Server{
			Config: &gui_server.GuiServerConfig{
				Port: uint32(port),
				GuiConfig: &gui_server.GuiConfig{
					ApiUrl:      "http://kuma.internal:5681",
					Environment: "kubernetes",
				},
			},
			ApiServerPort: apiSrvPort,
		}
		stop = make(chan struct{})
		go func() {
			defer GinkgoRecover()
			err := srv.Start(stop)
			Expect(err).ToNot(HaveOccurred())
		}()
		Eventually(func() bool {
			resp, err := http.Get(baseUrl)
			if err != nil {
				return false
			}
			Expect(resp.Body.Close()).To(Succeed())
			return true
		}).Should(BeTrue())
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
			fileContent, err := ioutil.ReadFile(filepath.Join("..", "..", "data", "resources", given.expectedFile))

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(fileContent).To(Equal(received))
		},
		Entry("should serve index.html without path", testCase{
			urlPath:      "",
			expectedFile: "index.html",
		}),
		Entry("should serve index.html with / path", testCase{
			urlPath:      "/",
			expectedFile: "index.html",
		}),
	)

	It("should serve the gui config", func() {
		// when
		resp, err := http.Get(fmt.Sprintf("%s/config", baseUrl))

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
		resp, err := http.Get(fmt.Sprintf("%s/api/some-endpoint", baseUrl))

		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		received, err := ioutil.ReadAll(resp.Body)

		// then
		Expect(resp.Body.Close()).To(Succeed())
		Expect(err).ToNot(HaveOccurred())

		// and
		Expect(string(received)).To(Equal("response from api server: /some-endpoint"))
	})

})
