package server_test

import (
	"bytes"
	"fmt"
	"net/http"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	v3 "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"github.com/golang/protobuf/jsonpb"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/prometheus/common/config"

	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	"github.com/kumahq/kuma/pkg/config/mads"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/runtime"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	"github.com/kumahq/kuma/pkg/mads/server"
	mads_v1_client "github.com/kumahq/kuma/pkg/mads/v1/client"
	"github.com/kumahq/kuma/pkg/metrics"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/pkg/test"
)

type testRuntime struct {
	runtime.Runtime
	rm         manager.ResourceManager
	config     kuma_cp.Config
	components []component.Component
	metrics    metrics.Metrics
}

func (t *testRuntime) ReadOnlyResourceManager() manager.ReadOnlyResourceManager {
	return t.rm
}

func (t *testRuntime) Add(component ...component.Component) error {
	t.components = append(t.components, component...)
	return nil
}

func (t *testRuntime) Config() kuma_cp.Config {
	return t.config
}

func (t *testRuntime) Metrics() metrics.Metrics {
	return t.metrics
}

var _ = Describe("MADS Server", func() {

	var rt *testRuntime
	var stopCh chan struct{}
	var errCh chan error
	var port uint32

	BeforeEach(func() {
		m, err := metrics.NewMetrics("zone-1")
		Expect(err).ToNot(HaveOccurred())

		rt = &testRuntime{
			rm:      manager.NewResourceManager(memory.NewStore()),
			config:  kuma_cp.Config{MonitoringAssignmentServer: mads.DefaultMonitoringAssignmentServerConfig()},
			metrics: m,
		}

		port, err = test.FindFreePort("127.0.0.1")
		Expect(err).ToNot(HaveOccurred())

		rt.config.MonitoringAssignmentServer.Port = port
		err = server.SetupServer(rt)
		Expect(err).ToNot(HaveOccurred())
		Expect(rt.components).To(HaveLen(1))

		stopCh = make(chan struct{})
		errCh = make(chan error)
		go func() {
			defer close(errCh)
			if err := rt.components[0].Start(stopCh); err != nil {
				errCh <- err
			}
		}()
	})

	AfterEach(func() {
		close(stopCh)
		err := <-errCh
		Expect(err).ToNot(HaveOccurred())
	})

	It("should serve GRPC requests", func() {
		client, err := mads_v1_client.New(fmt.Sprintf("grpc://127.0.0.1:%d", port))
		Expect(err).ToNot(HaveOccurred())

		var stream *mads_v1_client.Stream
		Eventually(func() bool {
			stream, err = client.StartStream()
			return err == nil
		}, "10s", "100ms").Should(BeTrue())

		err = stream.RequestAssignments("client-1")
		Expect(err).ToNot(HaveOccurred())

		assignments, err := stream.WaitForAssignments()
		Expect(err).ToNot(HaveOccurred())
		Expect(assignments).To(HaveLen(0))
	})

	It("should serve HTTP/1.1 requests", func() {
		rt, err := config.NewRoundTripperFromConfig(config.HTTPClientConfig{TLSConfig: config.TLSConfig{InsecureSkipVerify: true}}, "mads", config.WithHTTP2Disabled())
		Expect(err).ToNot(HaveOccurred())

		client := &http.Client{Transport: rt}

		req := &v3.DiscoveryRequest{
			VersionInfo:   "1",
			ResponseNonce: "",
			TypeUrl:       "type.googleapis.com/kuma.observability.v1.MonitoringAssignment",
			ResourceNames: []string{},
			Node: &envoy_core.Node{
				Id: "client-1",
			},
		}

		reqbuf := new(bytes.Buffer)
		marshaller := &jsonpb.Marshaler{}
		err = marshaller.Marshal(reqbuf, req)
		Expect(err).ToNot(HaveOccurred())

		request, err := http.NewRequest("POST", fmt.Sprintf("http://127.0.0.1:%d/v3/discovery:monitoringassignments", port), reqbuf)
		Expect(err).ToNot(HaveOccurred())

		request.Header.Add("Content-Type", "application/json")
		request.Header.Add("Accept", "application/json")

		Eventually(func() bool {
			response, err := client.Do(request)
			ok := err == nil && response.StatusCode == 200
			if response != nil {
				Expect(response.Body.Close()).To(Succeed())
			}
			return ok
		}, "10s", "100ms").Should(BeTrue())
	})
})
