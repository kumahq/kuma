package readiness_test

import (
	"context"
	"github.com/kumahq/kuma/app/kuma-dp/pkg/dataplane/readiness"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"net"
	"net/http"
	"net/http/httptest"
)

var (
	liveServerStats = "cluster_manager.cds.update_success: 1\nlistener_manager.lds.update_success: 1\nserver.state: 0\nlistener_manager.workers_started: 1"
	onlyServerStats = "server.state: 0"
	initServerStats = "cluster_manager.cds.update_success: 1\nlistener_manager.lds.update_success: 1\nserver.state: 2"
	noServerStats   = ""
)

var _ = Describe("Envoy readiness", func() {
	It("should handle successful stats", func() {
		ctx, cancel := context.WithCancel(context.Background())
		server := StartEnvoyStatsServer(liveServerStats)
		defer server.Close()
		defer cancel()
		probe := readiness.EnvoyReadinessProbe{Context: ctx, AdminPort: uint16(server.Listener.Addr().(*net.TCPAddr).Port)}

		err := probe.Check()

		Expect(err).ToNot(HaveOccurred())
	})

	It("should handle draining stats", func() {
		ctx, cancel := context.WithCancel(context.Background())
		server := StartEnvoyStatsServer(liveServerStats)
		defer server.Close()

		probe := readiness.EnvoyReadinessProbe{Context: ctx, AdminPort: uint16(server.Listener.Addr().(*net.TCPAddr).Port)}
		cancel()

		err := probe.Check()

		Expect(err).To(HaveOccurred())
	})

	It("should handle envoy initializing", func() {
		ctx, cancel := context.WithCancel(context.Background())
		server := StartEnvoyStatsServer(initServerStats)
		defer server.Close()
		defer cancel()

		probe := readiness.EnvoyReadinessProbe{Context: ctx, AdminPort: uint16(server.Listener.Addr().(*net.TCPAddr).Port)}

		err := probe.Check()

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("server.state: 2"))
	})

	It("should handle no envoy stats", func() {
		ctx, cancel := context.WithCancel(context.Background())
		server := StartEnvoyStatsServer(noServerStats)
		defer server.Close()
		defer cancel()

		probe := readiness.EnvoyReadinessProbe{Context: ctx, AdminPort: uint16(server.Listener.Addr().(*net.TCPAddr).Port)}
		err := probe.Check()

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("config not received from XDS server (is Kuma control plane running?): cds updates: 0 successful, 0 rejected; lds updates: 0 successful, 0 rejected"))
	})

	It("should cache envoy check results", func() {
		ctx, cancel := context.WithCancel(context.Background())
		server := StartEnvoyStatsServer(noServerStats)
		defer server.Close()
		defer cancel()

		// 1. a failing server, it should always fail
		probe := readiness.EnvoyReadinessProbe{Context: ctx, AdminPort: uint16(server.Listener.Addr().(*net.TCPAddr).Port)}

		err := probe.Check()
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("config not received from XDS server (is Kuma control plane running?): cds updates: 0 successful, 0 rejected; lds updates: 0 successful, 0 rejected"))

		err = probe.Check()
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("config not received from XDS server (is Kuma control plane running?): cds updates: 0 successful, 0 rejected; lds updates: 0 successful, 0 rejected"))
		server.Close()

		// 2. a good server, it should succeed and cache
		server = StartEnvoyStatsServer(liveServerStats)
		probe.AdminPort = uint16(server.Listener.Addr().(*net.TCPAddr).Port)
		err = probe.Check()
		Expect(err).ToNot(HaveOccurred())
		server.Close()

		// 3. switch to a fail server again, it should still succeed
		server = StartEnvoyStatsServer(noServerStats)
		probe.AdminPort = uint16(server.Listener.Addr().(*net.TCPAddr).Port)
		err = probe.Check()
		Expect(err).ToNot(HaveOccurred())

		// 4. even when the server is closed, it should always succeed
		server.Close()
		err = probe.Check()
		Expect(err).ToNot(HaveOccurred())
	})

	type testCaseEntry struct {
		stats           string
		expectedMessage string
	}

	DescribeTable("Different Envoy stats",
		func(testCase testCaseEntry) {
			server := StartEnvoyStatsServer(testCase.stats)
			defer server.Close()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			probe := readiness.EnvoyReadinessProbe{Context: ctx, AdminPort: uint16(server.Listener.Addr().(*net.TCPAddr).Port)}
			err := probe.Check()

			// Expect no error
			if testCase.expectedMessage == "" {
				Expect(err).ToNot(HaveOccurred())
			} else {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal(testCase.expectedMessage))
			}
		},
		Entry("only lds", testCaseEntry{
			stats:           "listener_manager.lds.update_success: 1",
			expectedMessage: "config not fully received from XDS server: cds updates: 0 successful, 0 rejected; lds updates: 1 successful, 0 rejected",
		}),
		Entry("only cds", testCaseEntry{
			stats:           "cluster_manager.cds.update_success: 1",
			expectedMessage: "config not fully received from XDS server: cds updates: 1 successful, 0 rejected; lds updates: 0 successful, 0 rejected",
		}),
		Entry("reject cds", testCaseEntry{
			stats: `cluster_manager.cds.update_rejected: 1
listener_manager.lds.update_success: 1`,
			expectedMessage: "config received from XDS server, but was rejected: cds updates: 0 successful, 1 rejected; lds updates: 1 successful, 0 rejected",
		}),
		Entry("empty config", testCaseEntry{
			stats:           "",
			expectedMessage: "config not received from XDS server (is Kuma control plane running?): cds updates: 0 successful, 0 rejected; lds updates: 0 successful, 0 rejected",
		}),
		Entry("workers not started", testCaseEntry{
			stats: `
cluster_manager.cds.update_success: 1
listener_manager.lds.update_success: 1
listener_manager.workers_started: 0
server.state: 0`,
			expectedMessage: "workers have not yet started",
		}),
		Entry("full", testCaseEntry{
			stats: `
cluster_manager.cds.update_success: 1
listener_manager.lds.update_success: 1
listener_manager.workers_started: 1
server.state: 0`,
			expectedMessage: "",
		}),
	)

})

// StartEnvoyStatsServer starts a server and returns the response passed.
func StartEnvoyStatsServer(response string) *httptest.Server {
	return createHTTPServer(createDefaultFuncMap("/stats", response))
}

func createHTTPServer(handlers map[string]func(rw http.ResponseWriter, _ *http.Request)) *httptest.Server {
	mux := http.NewServeMux()
	for k, v := range handlers {
		mux.HandleFunc(k, v)
	}

	// Start a local HTTP server
	server := httptest.NewUnstartedServer(mux)

	l, err := net.Listen("tcp", ":0")
	if err != nil {
		panic("Could not create listener for test: " + err.Error())
	}
	server.Listener = l
	server.Start()
	return server
}

func createDefaultFuncMap(pathPrefix, responseText string) map[string]func(rw http.ResponseWriter, _ *http.Request) {
	return map[string]func(rw http.ResponseWriter, _ *http.Request){
		pathPrefix: func(rw http.ResponseWriter, _ *http.Request) {
			// Send response to be tested
			_, err := rw.Write([]byte(responseText))
			if err != nil {
				panic("Could not write response: " + err.Error())
			}
		},
	}
}
