package get_test

import (
	"bytes"
	"context"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"

	"github.com/golang/protobuf/ptypes/duration"
	"github.com/golang/protobuf/ptypes/wrappers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	gomega_types "github.com/onsi/gomega/types"
	"github.com/spf13/cobra"

	kuma_mesh "github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/app/kumactl/cmd"
	kumactl_cmd "github.com/Kong/kuma/app/kumactl/pkg/cmd"
	config_proto "github.com/Kong/kuma/pkg/config/app/kumactl/v1alpha1"
	"github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/Kong/kuma/pkg/core/resources/model"
	core_store "github.com/Kong/kuma/pkg/core/resources/store"
	memory_resources "github.com/Kong/kuma/pkg/plugins/resources/memory"
	test_model "github.com/Kong/kuma/pkg/test/resources/model"
)

var _ = Describe("kumactl get circuit-breakers", func() {

	circuitBreakerResources := []*mesh.CircuitBreakerResource{
		{
			Spec: kuma_mesh.CircuitBreaker{
				Sources: []*kuma_mesh.Selector{
					{
						Match: map[string]string{
							"service": "frontend",
							"version": "0.1",
						},
					},
				},
				Destinations: []*kuma_mesh.Selector{
					{
						Match: map[string]string{
							"service": "backend",
						},
					},
				},
				Conf: &kuma_mesh.CircuitBreaker_Conf{
					Interval:                    &duration.Duration{Seconds: 5},
					BaseEjectionTime:            &duration.Duration{Seconds: 5},
					MaxEjectionPercent:          &wrappers.UInt32Value{Value: 50},
					SplitExternalAndLocalErrors: false,
					Detectors: &kuma_mesh.CircuitBreaker_Conf_Detectors{
						TotalErrors:       &kuma_mesh.CircuitBreaker_Conf_Detectors_Errors{},
						GatewayErrors:     &kuma_mesh.CircuitBreaker_Conf_Detectors_Errors{},
						LocalErrors:       &kuma_mesh.CircuitBreaker_Conf_Detectors_Errors{},
						StandardDeviation: &kuma_mesh.CircuitBreaker_Conf_Detectors_StandardDeviation{},
						Failure:           &kuma_mesh.CircuitBreaker_Conf_Detectors_Failure{},
					},
				},
			},
			Meta: &test_model.ResourceMeta{
				Mesh: "default",
				Name: "cb1",
			},
		},
		{
			Spec: kuma_mesh.CircuitBreaker{
				Sources: []*kuma_mesh.Selector{
					{
						Match: map[string]string{
							"service": "web",
							"version": "0.1",
						},
					},
				},
				Destinations: []*kuma_mesh.Selector{
					{
						Match: map[string]string{
							"service": "redis",
						},
					},
				},
				Conf: &kuma_mesh.CircuitBreaker_Conf{
					Interval:                    &duration.Duration{Seconds: 5},
					BaseEjectionTime:            &duration.Duration{Seconds: 5},
					MaxEjectionPercent:          &wrappers.UInt32Value{Value: 50},
					SplitExternalAndLocalErrors: false,
					Detectors: &kuma_mesh.CircuitBreaker_Conf_Detectors{
						TotalErrors:   &kuma_mesh.CircuitBreaker_Conf_Detectors_Errors{Consecutive: &wrappers.UInt32Value{Value: 20}},
						GatewayErrors: &kuma_mesh.CircuitBreaker_Conf_Detectors_Errors{Consecutive: &wrappers.UInt32Value{Value: 10}},
						LocalErrors:   &kuma_mesh.CircuitBreaker_Conf_Detectors_Errors{Consecutive: &wrappers.UInt32Value{Value: 2}},
						StandardDeviation: &kuma_mesh.CircuitBreaker_Conf_Detectors_StandardDeviation{
							RequestVolume: &wrappers.UInt32Value{Value: 20},
							MinimumHosts:  &wrappers.UInt32Value{Value: 3},
							Factor:        &wrappers.DoubleValue{Value: 1.9},
						},
						Failure: &kuma_mesh.CircuitBreaker_Conf_Detectors_Failure{
							RequestVolume: &wrappers.UInt32Value{Value: 20},
							MinimumHosts:  &wrappers.UInt32Value{Value: 3},
							Threshold:     &wrappers.UInt32Value{Value: 85},
						},
					},
				},
			},
			Meta: &test_model.ResourceMeta{
				Mesh: "default",
				Name: "cb2",
			},
		},
	}

	Describe("GetCircuitBreakerCmd", func() {

		var rootCtx *kumactl_cmd.RootContext
		var rootCmd *cobra.Command
		var buf *bytes.Buffer
		var store core_store.ResourceStore
		rootTime, _ := time.Parse(time.RFC3339, "2008-04-27T16:05:36.995Z")
		BeforeEach(func() {
			// setup
			rootCtx = &kumactl_cmd.RootContext{
				Runtime: kumactl_cmd.RootRuntime{
					Now: func() time.Time { return rootTime },
					NewResourceStore: func(*config_proto.ControlPlaneCoordinates_ApiServer) (core_store.ResourceStore, error) {
						return store, nil
					},
				},
			}

			store = memory_resources.NewStore()

			for _, cb := range circuitBreakerResources {
				err := store.Create(context.Background(), cb, core_store.CreateBy(core_model.MetaToResourceKey(cb.GetMeta())))
				Expect(err).ToNot(HaveOccurred())
			}

			rootCmd = cmd.NewRootCmd(rootCtx)
			buf = &bytes.Buffer{}
			rootCmd.SetOut(buf)
		})

		type testCase struct {
			outputFormat string
			goldenFile   string
			pagination   string
			matcher      func(interface{}) gomega_types.GomegaMatcher
		}

		DescribeTable("kumactl get circuit-breakers -o table|json|yaml",
			func(given testCase) {
				// given
				rootCmd.SetArgs(append([]string{
					"--config-file", filepath.Join("..", "testdata", "sample-kumactl.config.yaml"),
					"get", "circuit-breakers"}, given.outputFormat, given.pagination))

				// when
				err := rootCmd.Execute()
				// then
				Expect(err).ToNot(HaveOccurred())

				// when
				expected, err := ioutil.ReadFile(filepath.Join("testdata", given.goldenFile))
				// then
				Expect(err).ToNot(HaveOccurred())
				// and
				Expect(buf.String()).To(given.matcher(expected))
			},
			Entry("should support Table output by default", testCase{
				outputFormat: "",
				goldenFile:   "get-circuit-breakers.golden.txt",
				matcher: func(expected interface{}) gomega_types.GomegaMatcher {
					return WithTransform(strings.TrimSpace, Equal(strings.TrimSpace(string(expected.([]byte)))))
				},
			}),
			Entry("should support Table output explicitly", testCase{
				outputFormat: "-otable",
				goldenFile:   "get-circuit-breakers.golden.txt",
				matcher: func(expected interface{}) gomega_types.GomegaMatcher {
					return WithTransform(strings.TrimSpace, Equal(strings.TrimSpace(string(expected.([]byte)))))
				},
			}),
			Entry("should support pagination", testCase{
				outputFormat: "-otable",
				goldenFile:   "get-circuit-breakers.pagination.golden.txt",
				pagination:   "--size=1",
				matcher: func(expected interface{}) gomega_types.GomegaMatcher {
					return WithTransform(strings.TrimSpace, Equal(strings.TrimSpace(string(expected.([]byte)))))
				},
			}),
			Entry("should support JSON output", testCase{
				outputFormat: "-ojson",
				goldenFile:   "get-circuit-breakers.golden.json",
				matcher:      MatchJSON,
			}),
			Entry("should support YAML output", testCase{
				outputFormat: "-oyaml",
				goldenFile:   "get-circuit-breakers.golden.yaml",
				matcher:      MatchYAML,
			}),
		)
	})
})
