package api_server_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	api_server "github.com/kumahq/kuma/v2/pkg/api-server"
	kuma_log "github.com/kumahq/kuma/v2/pkg/log"
)

var _ = Describe("Logging Endpoint", func() {
	var apiServer *api_server.ApiServer
	var stop func()
	var baseURL string

	BeforeEach(func() {
		kuma_log.GlobalComponentLevelRegistry().ResetAll()
		apiServer, _, stop = StartApiServer(NewTestApiServerConfigurer())
		baseURL = fmt.Sprintf("http://%s", apiServer.Address())
	})
	AfterEach(func() {
		stop()
		kuma_log.GlobalComponentLevelRegistry().ResetAll()
	})

	It("should return current log levels", func() {
		resp, err := http.Get(baseURL + "/logging")
		Expect(err).ToNot(HaveOccurred())
		Expect(resp.StatusCode).To(Equal(http.StatusOK))

		body, err := io.ReadAll(resp.Body)
		Expect(err).ToNot(HaveOccurred())

		var result map[string]any
		Expect(json.Unmarshal(body, &result)).To(Succeed())
		Expect(result).To(HaveKey("global"))
		Expect(result).To(HaveKey("components"))
	})

	It("should set component log level", func() {
		// set xds to debug
		reqBody, _ := json.Marshal(map[string]string{
			"component": "xds",
			"level":     "debug",
		})
		req, _ := http.NewRequest(http.MethodPut, baseURL+"/logging", bytes.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		Expect(err).ToNot(HaveOccurred())
		Expect(resp.StatusCode).To(Equal(http.StatusOK))

		// verify it shows up in GET
		resp, err = http.Get(baseURL + "/logging")
		Expect(err).ToNot(HaveOccurred())

		body, err := io.ReadAll(resp.Body)
		Expect(err).ToNot(HaveOccurred())

		var result map[string]any
		Expect(json.Unmarshal(body, &result)).To(Succeed())
		components := result["components"].(map[string]any)
		Expect(components["xds"]).To(Equal("debug"))
	})

	It("should reject invalid log level", func() {
		reqBody, _ := json.Marshal(map[string]string{
			"component": "xds",
			"level":     "invalid",
		})
		req, _ := http.NewRequest(http.MethodPut, baseURL+"/logging", bytes.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		Expect(err).ToNot(HaveOccurred())
		Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
	})

	It("should reset single component level", func() {
		kuma_log.GlobalComponentLevelRegistry().SetLevel("xds", kuma_log.DebugLevel)

		req, _ := http.NewRequest(http.MethodDelete, baseURL+"/logging/xds", nil)
		resp, err := http.DefaultClient.Do(req)
		Expect(err).ToNot(HaveOccurred())
		Expect(resp.StatusCode).To(Equal(http.StatusOK))

		overrides := kuma_log.GlobalComponentLevelRegistry().ListOverrides()
		Expect(overrides).To(BeEmpty())
	})

	It("should reset all component levels", func() {
		kuma_log.GlobalComponentLevelRegistry().SetLevel("xds", kuma_log.DebugLevel)
		kuma_log.GlobalComponentLevelRegistry().SetLevel("kds", kuma_log.InfoLevel)

		req, _ := http.NewRequest(http.MethodDelete, baseURL+"/logging", nil)
		resp, err := http.DefaultClient.Do(req)
		Expect(err).ToNot(HaveOccurred())
		Expect(resp.StatusCode).To(Equal(http.StatusOK))

		overrides := kuma_log.GlobalComponentLevelRegistry().ListOverrides()
		Expect(overrides).To(BeEmpty())
	})
})
