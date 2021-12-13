package api_server_test

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"path"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/types/known/durationpb"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	config "github.com/kumahq/kuma/pkg/config/api-server"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/metrics"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/pkg/test/kds/samples"
	"github.com/kumahq/kuma/pkg/test/matchers"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

type staticMatchedPolicyGetter struct {
	mp *core_xds.MatchedPolicies
}

func (s *staticMatchedPolicyGetter) Get(ctx context.Context, dataplaneKey model.ResourceKey) (*core_xds.MatchedPolicies, error) {
	return s.mp, nil
}

func inbound(ip string, dpPort, workloadPort uint32) mesh_proto.InboundInterface {
	return mesh_proto.InboundInterface{
		DataplaneIP:   ip,
		DataplanePort: dpPort,
		WorkloadPort:  workloadPort,
	}
}

func outbound(ip string, port uint32) mesh_proto.OutboundInterface {
	return mesh_proto.OutboundInterface{
		DataplaneIP:   ip,
		DataplanePort: port,
	}
}

var _ = Describe("Inspect WS", func() {

	type testCase struct {
		path            string
		goldenFile      string
		matchedPolicies *core_xds.MatchedPolicies
	}

	DescribeTable("should return policies matched for specific dataplane",
		func(given testCase) {
			// setup
			resourceStore := memory.NewStore()
			metrics, err := metrics.NewMetrics("Standalone")
			Expect(err).ToNot(HaveOccurred())
			apiServer := createTestApiServer(resourceStore, config.DefaultApiServerConfig(),
				true, metrics, &staticMatchedPolicyGetter{mp: given.matchedPolicies})

			stop := make(chan struct{})
			go func() {
				defer GinkgoRecover()
				err := apiServer.Start(stop)
				Expect(err).ToNot(HaveOccurred())
			}()

			// when
			var resp *http.Response
			Eventually(func() error {
				r, err := http.Get((&url.URL{
					Scheme: "http",
					Host:   apiServer.Address(),
					Path:   given.path,
				}).String())
				resp = r
				return err
			}, "3s").ShouldNot(HaveOccurred())

			// then
			bytes, err := io.ReadAll(resp.Body)
			Expect(err).ToNot(HaveOccurred())
			Expect(bytes).To(matchers.MatchGoldenJSON(path.Join("testdata", given.goldenFile)))
		},
		Entry("full example", testCase{
			path:       "/inspect/meshes/default/dataplane/backend-1",
			goldenFile: "inspect.json",
			matchedPolicies: &core_xds.MatchedPolicies{
				TrafficPermissions: core_xds.TrafficPermissionMap{
					inbound("192.168.0.1", 80, 81): {Spec: samples.TrafficPermission},
				},
				HealthChecks: core_xds.HealthCheckMap{
					"backend":  {Spec: samples.HealthCheck},
					"gateway":  {Spec: samples.HealthCheck},
					"postgres": {Spec: samples.HealthCheck},
				},
				FaultInjections: core_xds.FaultInjectionMap{
					inbound("192.168.0.1", 80, 81): {
						{Spec: &mesh_proto.FaultInjection{Conf: &mesh_proto.FaultInjection_Conf{
							Delay: &mesh_proto.FaultInjection_Conf_Delay{
								Value:      durationpb.New(5 * time.Second),
								Percentage: util_proto.Double(90),
							},
						}}},
						{Spec: &mesh_proto.FaultInjection{Conf: &mesh_proto.FaultInjection_Conf{
							Abort: &mesh_proto.FaultInjection_Conf_Abort{
								HttpStatus: util_proto.UInt32(500),
								Percentage: util_proto.Double(80),
							},
						}}},
					},
				},
				Timeouts: core_xds.TimeoutMap{
					outbound("192.168.0.2", 8080): {Spec: samples.Timeout},
				},
			},
		}),
	)
})
