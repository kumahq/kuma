package diagnostics_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/v2/pkg/diagnostics"
	kuma_log "github.com/kumahq/kuma/v2/pkg/log"
)

var _ = Describe("Logging handlers", func() {
	var (
		registry *kuma_log.ComponentLevelRegistry
		mux      *http.ServeMux
	)

	BeforeEach(func() {
		registry = kuma_log.NewComponentLevelRegistry()
		mux = http.NewServeMux()
		diagnostics.AddLoggingHandlers(mux, registry)
	})

	It("should return current log levels", func() {
		req := httptest.NewRequest(http.MethodGet, "/logging", http.NoBody)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		Expect(w.Code).To(Equal(http.StatusOK))
		var result map[string]any
		Expect(json.Unmarshal(w.Body.Bytes(), &result)).To(Succeed())
		Expect(result).To(HaveKey("components"))
	})

	It("should set component log level", func() {
		body, err := json.Marshal(map[string]string{"component": "xds", "level": "debug"})
		Expect(err).ToNot(HaveOccurred())
		req := httptest.NewRequest(http.MethodPut, "/logging", bytes.NewReader(body))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		Expect(w.Code).To(Equal(http.StatusOK))

		req = httptest.NewRequest(http.MethodGet, "/logging", http.NoBody)
		w = httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		var result map[string]any
		Expect(json.Unmarshal(w.Body.Bytes(), &result)).To(Succeed())
		components := result["components"].(map[string]any)
		Expect(components["xds"]).To(Equal("debug"))
	})

	DescribeTable("should reject invalid PUT requests",
		func(body string, expectedCode int) {
			req := httptest.NewRequest(http.MethodPut, "/logging", bytes.NewReader([]byte(body)))
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, req)
			Expect(w.Code).To(Equal(expectedCode))
		},
		Entry("invalid log level", `{"component":"xds","level":"invalid"}`, http.StatusBadRequest),
		Entry("empty component", `{"level":"debug"}`, http.StatusBadRequest),
		Entry("malformed JSON", `{invalid`, http.StatusBadRequest),
		Entry("invalid component name", `{"component":"../../etc","level":"debug"}`, http.StatusBadRequest),
	)

	It("should reset single component level", func() {
		Expect(registry.SetLevel("xds", kuma_log.DebugLevel)).To(Succeed())

		req := httptest.NewRequest(http.MethodDelete, "/logging/xds", http.NoBody)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		Expect(w.Code).To(Equal(http.StatusOK))

		Expect(registry.ListOverrides()).To(BeEmpty())
	})

	It("should reset all component levels and return removed overrides", func() {
		Expect(registry.SetLevel("xds", kuma_log.DebugLevel)).To(Succeed())
		Expect(registry.SetLevel("kds", kuma_log.InfoLevel)).To(Succeed())

		req := httptest.NewRequest(http.MethodDelete, "/logging", http.NoBody)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		Expect(w.Code).To(Equal(http.StatusOK))

		var result map[string]any
		Expect(json.Unmarshal(w.Body.Bytes(), &result)).To(Succeed())
		components := result["components"].(map[string]any)
		Expect(components).To(HaveLen(2))

		Expect(registry.ListOverrides()).To(BeEmpty())
	})

	DescribeTable("should reject invalid requests on /logging/{component}",
		func(method string, path string, expectedCode int) {
			req := httptest.NewRequest(method, path, http.NoBody)
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, req)
			Expect(w.Code).To(Equal(expectedCode))
		},
		Entry("GET not allowed", http.MethodGet, "/logging/xds", http.StatusMethodNotAllowed),
		Entry("invalid component name on DELETE", http.MethodDelete, "/logging/invalid!name", http.StatusBadRequest),
	)
})
