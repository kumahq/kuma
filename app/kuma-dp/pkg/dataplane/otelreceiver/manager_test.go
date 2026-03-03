package otelreceiver

import (
	"fmt"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	core_xds "github.com/kumahq/kuma/v2/pkg/core/xds"
)

func testSocketPath(prefix string) string {
	return filepath.Join(GinkgoT().TempDir(), fmt.Sprintf("%s.sock", prefix))
}

var _ = Describe("Manager reconcile", func() {
	It("should add and remove backends", func() {
		manager := &Manager{
			running: map[string]*runningServer{},
		}
		DeferCleanup(manager.stopAll)

		socketA := testSocketPath("a")
		socketB := testSocketPath("b")
		backendA := core_xds.OtelPipeBackend{
			SocketPath: socketA,
			Endpoint:   "collector-a:4317",
		}
		backendB := core_xds.OtelPipeBackend{
			SocketPath: socketB,
			Endpoint:   "collector-b:4317",
		}

		Expect(manager.reconcile([]core_xds.OtelPipeBackend{backendA, backendB})).To(Succeed())
		Expect(manager.running).To(HaveLen(2))
		Expect(manager.running).To(HaveKey(socketA))
		Expect(manager.running).To(HaveKey(socketB))

		Expect(manager.reconcile([]core_xds.OtelPipeBackend{backendB})).To(Succeed())
		Expect(manager.running).To(HaveLen(1))
		Expect(manager.running).NotTo(HaveKey(socketA))
		Expect(manager.running).To(HaveKey(socketB))
	})

	It("should restart backend on config change", func() {
		manager := &Manager{
			running: map[string]*runningServer{},
		}
		DeferCleanup(manager.stopAll)

		socketPath := testSocketPath("collector")
		initialBackend := core_xds.OtelPipeBackend{
			SocketPath: socketPath,
			Endpoint:   "collector:4317",
		}
		Expect(manager.reconcile([]core_xds.OtelPipeBackend{initialBackend})).To(Succeed())

		originalServer := manager.running[socketPath].server

		// Same config - no restart.
		Expect(manager.reconcile([]core_xds.OtelPipeBackend{initialBackend})).To(Succeed())
		Expect(manager.running[socketPath].server).To(BeIdenticalTo(originalServer))

		updatedBackend := core_xds.OtelPipeBackend{
			SocketPath: socketPath,
			Endpoint:   "collector:4318",
			UseHTTP:    true,
			Path:       "/otlp",
		}
		Expect(manager.reconcile([]core_xds.OtelPipeBackend{updatedBackend})).To(Succeed())
		Expect(manager.running[socketPath].server).NotTo(BeIdenticalTo(originalServer))
		Expect(sameBackendConfig(manager.running[socketPath].backend, updatedBackend)).To(BeTrue())
	})

	It("should return error for invalid socket path", func() {
		manager := &Manager{
			running: map[string]*runningServer{},
		}
		DeferCleanup(manager.stopAll)

		backend := core_xds.OtelPipeBackend{
			SocketPath: "/nonexistent/dir/broken.sock",
			Endpoint:   "collector:4317",
		}

		err := manager.reconcile([]core_xds.OtelPipeBackend{backend})
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("failed to start OTel receiver"))
		Expect(manager.running).To(BeEmpty())
	})
})

var _ = Describe("sameBackendConfig", func() {
	It("should return true for identical configs", func() {
		a := core_xds.OtelPipeBackend{Endpoint: "a:4317", UseHTTP: true, Path: "/x"}
		Expect(sameBackendConfig(a, a)).To(BeTrue())
	})

	It("should return false when endpoint differs", func() {
		a := core_xds.OtelPipeBackend{Endpoint: "a:4317"}
		b := core_xds.OtelPipeBackend{Endpoint: "b:4317"}
		Expect(sameBackendConfig(a, b)).To(BeFalse())
	})
})

var _ = Describe("testSocketPath helper", func() {
	It("should create unique paths in temp dir", func() {
		p := testSocketPath("foo")
		Expect(p).To(ContainSubstring("foo.sock"))
		Expect(strings.HasPrefix(p, "/")).To(BeTrue())
	})
})
