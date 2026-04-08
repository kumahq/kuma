package diagnostics_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"

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

	doRequest := func(method, path string, body string) *httptest.ResponseRecorder {
		var bodyReader *bytes.Reader
		if body != "" {
			bodyReader = bytes.NewReader([]byte(body))
		} else {
			bodyReader = bytes.NewReader(nil)
		}
		req := httptest.NewRequest(method, path, bodyReader)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		return w
	}

	parseComponents := func(w *httptest.ResponseRecorder) map[string]any {
		var result map[string]any
		Expect(json.Unmarshal(w.Body.Bytes(), &result)).To(Succeed())
		Expect(result).To(HaveKey("components"))
		return result["components"].(map[string]any)
	}

	Describe("GET /logging", func() {
		It("returns empty components map when no overrides", func() {
			w := doRequest(http.MethodGet, "/logging", "")
			Expect(w.Code).To(Equal(http.StatusOK))
			Expect(parseComponents(w)).To(BeEmpty())
		})

		It("returns Content-Type application/json", func() {
			w := doRequest(http.MethodGet, "/logging", "")
			Expect(w.Header().Get("Content-Type")).To(ContainSubstring("application/json"))
		})

		It("returns all active overrides with correct levels", func() {
			Expect(registry.SetLevel("xds", kuma_log.DebugLevel)).To(Succeed())
			Expect(registry.SetLevel("kds", kuma_log.InfoLevel)).To(Succeed())
			Expect(registry.SetLevel("mads", kuma_log.OffLevel)).To(Succeed())

			w := doRequest(http.MethodGet, "/logging", "")
			components := parseComponents(w)
			Expect(components).To(HaveLen(3))
			Expect(components["xds"]).To(Equal("debug"))
			Expect(components["kds"]).To(Equal("info"))
			Expect(components["mads"]).To(Equal("off"))
		})
	})

	Describe("PUT /logging", func() {
		It("sets component log level", func() {
			w := doRequest(http.MethodPut, "/logging", `{"component":"xds","level":"debug"}`)
			Expect(w.Code).To(Equal(http.StatusOK))

			Expect(parseComponents(doRequest(http.MethodGet, "/logging", ""))).To(
				HaveKeyWithValue("xds", "debug"),
			)
		})

		DescribeTable("accepts all valid log levels",
			func(level string) {
				w := doRequest(http.MethodPut, "/logging", fmt.Sprintf(`{"component":"xds","level":"%s"}`, level))
				Expect(w.Code).To(Equal(http.StatusOK))
				Expect(parseComponents(doRequest(http.MethodGet, "/logging", ""))).To(
					HaveKeyWithValue("xds", level),
				)
			},
			Entry("debug", "debug"),
			Entry("info", "info"),
			Entry("off", "off"),
		)

		It("updates existing override", func() {
			Expect(registry.SetLevel("xds", kuma_log.DebugLevel)).To(Succeed())

			w := doRequest(http.MethodPut, "/logging", `{"component":"xds","level":"off"}`)
			Expect(w.Code).To(Equal(http.StatusOK))

			overrides := registry.ListOverrides()
			Expect(overrides).To(HaveLen(1))
			Expect(overrides["xds"]).To(Equal(kuma_log.OffLevel))
		})

		It("returns 429 when max overrides exceeded", func() {
			for i := range kuma_log.MaxOverrides {
				Expect(registry.SetLevel(fmt.Sprintf("comp-%d", i), kuma_log.DebugLevel)).To(Succeed())
			}
			w := doRequest(http.MethodPut, "/logging", `{"component":"one-too-many","level":"debug"}`)
			Expect(w.Code).To(Equal(http.StatusTooManyRequests))
		})

		It("allows updating at-capacity registry (does not count as new)", func() {
			for i := range kuma_log.MaxOverrides {
				Expect(registry.SetLevel(fmt.Sprintf("comp-%d", i), kuma_log.DebugLevel)).To(Succeed())
			}
			w := doRequest(http.MethodPut, "/logging", `{"component":"comp-0","level":"info"}`)
			Expect(w.Code).To(Equal(http.StatusOK))
		})

		It("rejects oversized body (>4096 bytes)", func() {
			oversized := fmt.Sprintf(`{"component":"xds","level":"debug","extra":"%s"}`, strings.Repeat("x", 5000))
			w := doRequest(http.MethodPut, "/logging", oversized)
			Expect(w.Code).To(Equal(http.StatusBadRequest))
		})

		DescribeTable("rejects invalid requests",
			func(body string, expectedCode int) {
				w := doRequest(http.MethodPut, "/logging", body)
				Expect(w.Code).To(Equal(expectedCode))
			},
			Entry("invalid log level", `{"component":"xds","level":"verbose"}`, http.StatusBadRequest),
			Entry("empty component field", `{"level":"debug"}`, http.StatusBadRequest),
			Entry("malformed JSON", `{invalid`, http.StatusBadRequest),
			Entry("invalid component name chars", `{"component":"../../etc","level":"debug"}`, http.StatusBadRequest),
			Entry("component starts with dot", `{"component":".xds","level":"debug"}`, http.StatusBadRequest),
			Entry("empty body", ``, http.StatusBadRequest),
		)

		DescribeTable("accepts valid component names",
			func(component string) {
				body := fmt.Sprintf(`{"component":%q,"level":"debug"}`, component)
				w := doRequest(http.MethodPut, "/logging", body)
				Expect(w.Code).To(Equal(http.StatusOK))
			},
			Entry("simple name", "xds"),
			Entry("dot-separated hierarchy", "xds.auth"),
			Entry("deep hierarchy", "plugins.authn.api-server.tokens"),
			Entry("with dashes", "kds-mux-client"),
			Entry("with underscores", "xds_server"),
			Entry("starts with digit", "1xds"),
			Entry("max length (256 chars)", strings.Repeat("a", 256)),
		)

		It("rejects component name exceeding 256 chars", func() {
			long := strings.Repeat("a", 257)
			w := doRequest(http.MethodPut, "/logging", fmt.Sprintf(`{"component":%q,"level":"debug"}`, long))
			Expect(w.Code).To(Equal(http.StatusBadRequest))
		})
	})

	Describe("DELETE /logging", func() {
		It("removes all overrides and returns them", func() {
			Expect(registry.SetLevel("xds", kuma_log.DebugLevel)).To(Succeed())
			Expect(registry.SetLevel("kds", kuma_log.InfoLevel)).To(Succeed())

			w := doRequest(http.MethodDelete, "/logging", "")
			Expect(w.Code).To(Equal(http.StatusOK))
			Expect(w.Header().Get("Content-Type")).To(ContainSubstring("application/json"))

			components := parseComponents(w)
			Expect(components).To(HaveLen(2))
			Expect(components["xds"]).To(Equal("debug"))
			Expect(components["kds"]).To(Equal("info"))

			Expect(registry.ListOverrides()).To(BeEmpty())
		})

		It("returns empty components when no overrides exist", func() {
			w := doRequest(http.MethodDelete, "/logging", "")
			Expect(w.Code).To(Equal(http.StatusOK))
			Expect(parseComponents(w)).To(BeEmpty())
		})

		It("is idempotent (second call returns empty)", func() {
			Expect(registry.SetLevel("xds", kuma_log.DebugLevel)).To(Succeed())
			doRequest(http.MethodDelete, "/logging", "")

			w := doRequest(http.MethodDelete, "/logging", "")
			Expect(w.Code).To(Equal(http.StatusOK))
			Expect(parseComponents(w)).To(BeEmpty())
		})
	})

	Describe("DELETE /logging/{component}", func() {
		It("removes specific override and returns 200", func() {
			Expect(registry.SetLevel("xds", kuma_log.DebugLevel)).To(Succeed())
			Expect(registry.SetLevel("kds", kuma_log.InfoLevel)).To(Succeed())

			w := doRequest(http.MethodDelete, "/logging/xds", "")
			Expect(w.Code).To(Equal(http.StatusOK))

			overrides := registry.ListOverrides()
			Expect(overrides).To(HaveLen(1))
			Expect(overrides).To(HaveKey("kds"))
			Expect(overrides).NotTo(HaveKey("xds"))
		})

		It("is idempotent when component does not exist", func() {
			w := doRequest(http.MethodDelete, "/logging/nonexistent", "")
			Expect(w.Code).To(Equal(http.StatusOK))
		})

		DescribeTable("rejects invalid requests",
			func(method string, path string, expectedCode int) {
				w := doRequest(method, path, "")
				Expect(w.Code).To(Equal(expectedCode))
			},
			Entry("GET not allowed", http.MethodGet, "/logging/xds", http.StatusMethodNotAllowed),
			Entry("PUT not allowed", http.MethodPut, "/logging/xds", http.StatusMethodNotAllowed),
			Entry("invalid component name chars", http.MethodDelete, "/logging/invalid!name", http.StatusBadRequest),
		)
	})

	Describe("method routing on /logging", func() {
		DescribeTable("rejects unsupported methods",
			func(method string) {
				w := doRequest(method, "/logging", "")
				Expect(w.Code).To(Equal(http.StatusMethodNotAllowed))
			},
			Entry("POST", http.MethodPost),
			Entry("PATCH", http.MethodPatch),
			Entry("HEAD", http.MethodHead),
		)
	})

	Describe("concurrent access", func() {
		It("handles concurrent PUT and DELETE without data races", func() {
			const goroutines = 20
			var wg sync.WaitGroup
			wg.Add(goroutines)
			for i := range goroutines {
				go func(n int) {
					defer GinkgoRecover()
					defer wg.Done()
					component := fmt.Sprintf("comp-%d", n)
					doRequest(http.MethodPut, "/logging", fmt.Sprintf(`{"component":%q,"level":"debug"}`, component))
					doRequest(http.MethodGet, "/logging", "")
					doRequest(http.MethodDelete, fmt.Sprintf("/logging/%s", component), "")
				}(i)
			}
			wg.Wait()
			// After all goroutines finish, registry should be consistent
			overrides := registry.ListOverrides()
			Expect(len(overrides)).To(BeNumerically("<=", goroutines))
		})
	})
})
