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
	var registry *kuma_log.ComponentLevelRegistry
	var mux *http.ServeMux

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

	It("should reject invalid log level", func() {
		body, err := json.Marshal(map[string]string{"component": "xds", "level": "invalid"})
		Expect(err).ToNot(HaveOccurred())
		req := httptest.NewRequest(http.MethodPut, "/logging", bytes.NewReader(body))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		Expect(w.Code).To(Equal(http.StatusBadRequest))
	})

	It("should reject empty component", func() {
		body, err := json.Marshal(map[string]string{"level": "debug"})
		Expect(err).ToNot(HaveOccurred())
		req := httptest.NewRequest(http.MethodPut, "/logging", bytes.NewReader(body))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		Expect(w.Code).To(Equal(http.StatusBadRequest))
	})

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

	It("should reject malformed JSON body", func() {
		req := httptest.NewRequest(http.MethodPut, "/logging", bytes.NewReader([]byte("{invalid")))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		Expect(w.Code).To(Equal(http.StatusBadRequest))
	})

	It("should reject invalid component name", func() {
		body, err := json.Marshal(map[string]string{"component": "../../etc", "level": "debug"})
		Expect(err).ToNot(HaveOccurred())
		req := httptest.NewRequest(http.MethodPut, "/logging", bytes.NewReader(body))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		Expect(w.Code).To(Equal(http.StatusBadRequest))
	})

	It("should return method not allowed for GET on /logging/{component}", func() {
		req := httptest.NewRequest(http.MethodGet, "/logging/xds", http.NoBody)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		Expect(w.Code).To(Equal(http.StatusMethodNotAllowed))
	})

	It("should reject invalid component name on DELETE", func() {
		req := httptest.NewRequest(http.MethodDelete, "/logging/invalid!name", http.NoBody)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		Expect(w.Code).To(Equal(http.StatusBadRequest))
	})
})
