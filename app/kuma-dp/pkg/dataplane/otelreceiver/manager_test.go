package otelreceiver

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"sync/atomic"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc"

	mt_dpapi "github.com/kumahq/kuma/v2/pkg/plugins/policies/meshtrace/dpapi"
)

func testSocketPath(prefix string) string {
	return filepath.Join(GinkgoT().TempDir(), fmt.Sprintf("%s.sock", prefix))
}

var _ = Describe("Manager reconcile", func() {
	It("should add and remove backends", func() {
		manager := &Manager{
			registerService: func(_ *grpc.Server, _ mt_dpapi.OtelBackendConfig) (func(), error) {
				return func() {}, nil
			},
			running: map[string]*runningServer{},
		}
		DeferCleanup(manager.stopAll)

		socketA := testSocketPath("a")
		socketB := testSocketPath("b")
		backendA := mt_dpapi.OtelBackendConfig{
			SocketPath: socketA,
			Endpoint:   "collector-a:4317",
		}
		backendB := mt_dpapi.OtelBackendConfig{
			SocketPath: socketB,
			Endpoint:   "collector-b:4317",
		}

		Expect(manager.reconcile([]mt_dpapi.OtelBackendConfig{backendA, backendB})).To(Succeed())
		Expect(manager.running).To(HaveLen(2))
		Expect(manager.running).To(HaveKey(socketA))
		Expect(manager.running).To(HaveKey(socketB))

		Expect(manager.reconcile([]mt_dpapi.OtelBackendConfig{backendB})).To(Succeed())
		Expect(manager.running).To(HaveLen(1))
		Expect(manager.running).NotTo(HaveKey(socketA))
		Expect(manager.running).To(HaveKey(socketB))
	})

	It("should restart backend on config change", func() {
		var starts atomic.Int32
		var closes atomic.Int32

		manager := &Manager{
			registerService: func(_ *grpc.Server, _ mt_dpapi.OtelBackendConfig) (func(), error) {
				starts.Add(1)
				return func() {
					closes.Add(1)
				}, nil
			},
			running: map[string]*runningServer{},
		}
		DeferCleanup(manager.stopAll)

		socketPath := testSocketPath("collector")
		initialBackend := mt_dpapi.OtelBackendConfig{
			SocketPath: socketPath,
			Endpoint:   "collector:4317",
		}
		Expect(manager.reconcile([]mt_dpapi.OtelBackendConfig{initialBackend})).To(Succeed())

		originalServer := manager.running[socketPath].server
		Expect(starts.Load()).To(Equal(int32(1)))
		Expect(closes.Load()).To(Equal(int32(0)))

		Expect(manager.reconcile([]mt_dpapi.OtelBackendConfig{initialBackend})).To(Succeed())
		Expect(starts.Load()).To(Equal(int32(1)))
		Expect(closes.Load()).To(Equal(int32(0)))

		updatedBackend := mt_dpapi.OtelBackendConfig{
			SocketPath: socketPath,
			Endpoint:   "collector:4318",
			UseHTTP:    true,
			Path:       "/otlp",
		}
		Expect(manager.reconcile([]mt_dpapi.OtelBackendConfig{updatedBackend})).To(Succeed())
		Expect(starts.Load()).To(Equal(int32(2)))
		Expect(closes.Load()).To(Equal(int32(1)))
		Expect(manager.running[socketPath].server).NotTo(BeIdenticalTo(originalServer))
		Expect(sameBackendConfig(manager.running[socketPath].backend, updatedBackend)).To(BeTrue())
	})

	It("should return error when backend start fails", func() {
		manager := &Manager{
			registerService: func(_ *grpc.Server, _ mt_dpapi.OtelBackendConfig) (func(), error) {
				return nil, errors.New("register failed")
			},
			running: map[string]*runningServer{},
		}
		DeferCleanup(manager.stopAll)

		backend := mt_dpapi.OtelBackendConfig{
			SocketPath: testSocketPath("broken"),
			Endpoint:   "collector:4317",
		}

		err := manager.reconcile([]mt_dpapi.OtelBackendConfig{backend})
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("failed to start OTel receiver"))
		Expect(manager.running).To(BeEmpty())
	})
})

var _ = Describe("sameBackendConfig", func() {
	It("should return true for identical configs", func() {
		a := mt_dpapi.OtelBackendConfig{Endpoint: "a:4317", UseHTTP: true, Path: "/x"}
		Expect(sameBackendConfig(a, a)).To(BeTrue())
	})

	It("should return false when endpoint differs", func() {
		a := mt_dpapi.OtelBackendConfig{Endpoint: "a:4317"}
		b := mt_dpapi.OtelBackendConfig{Endpoint: "b:4317"}
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
