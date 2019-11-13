package server_test

import (
	"fmt"
	"github.com/Kong/kuma/app/kuma-ui/pkg/server"
	gui_server "github.com/Kong/kuma/pkg/config/gui-server"
	"github.com/Kong/kuma/pkg/test"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strconv"
)

var _ = Describe("GUI Server", func() {

	var stop chan struct{}
	var baseUrl string

	BeforeEach(func() {
		port, err := test.GetFreePort()
		Expect(err).ToNot(HaveOccurred())
		baseUrl = "http://localhost:" + strconv.Itoa(port)

		srv := server.Server{
			Config: &gui_server.GuiServerConfig{
				Port: uint32(port),
			},
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
		Entry("should serve data.js", testCase{
			urlPath:      "/data.js",
			expectedFile: "data.js",
		}),
	)
})
