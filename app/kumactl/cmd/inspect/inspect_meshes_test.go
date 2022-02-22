package inspect_test

import (
	"bytes"
	"context"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	gomega_types "github.com/onsi/gomega/types"
	"github.com/spf13/cobra"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/app/kumactl/cmd"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	memory_resources "github.com/kumahq/kuma/pkg/plugins/resources/memory"
	test_kumactl "github.com/kumahq/kuma/pkg/test/kumactl"
	"github.com/kumahq/kuma/pkg/test/matchers"
	"github.com/kumahq/kuma/pkg/test/resources/model"
)

var _ = Describe("kumactl inspect meshes", func() {

	meshInsightResources := []*mesh.MeshInsightResource{
		{
			Meta: &model.ResourceMeta{Name: "default"},
			Spec: &mesh_proto.MeshInsight{
				Dataplanes: &mesh_proto.MeshInsight_DataplaneStat{
					Total:   100,
					Online:  90,
					Offline: 10,
				},
				Policies: map[string]*mesh_proto.MeshInsight_PolicyStat{
					string(mesh.TrafficTraceType):      {Total: 1},
					string(mesh.TrafficRouteType):      {Total: 2},
					string(mesh.TrafficLogType):        {Total: 3},
					string(mesh.HealthCheckType):       {Total: 4},
					string(mesh.CircuitBreakerType):    {Total: 5},
					string(mesh.FaultInjectionType):    {Total: 6},
					string(mesh.TrafficPermissionType): {Total: 7},
					string(mesh.ProxyTemplateType):     {Total: 8},
					string(mesh.ExternalServiceType):   {Total: 9},
					string(mesh.RateLimitType):         {Total: 10},
				},
			},
		},
		{
			Meta: &model.ResourceMeta{Name: "mesh-1"},
			Spec: &mesh_proto.MeshInsight{
				Dataplanes: &mesh_proto.MeshInsight_DataplaneStat{
					Total:   100,
					Online:  90,
					Offline: 10,
				},
				Policies: map[string]*mesh_proto.MeshInsight_PolicyStat{
					string(mesh.TrafficTraceType):      {Total: 10},
					string(mesh.TrafficRouteType):      {Total: 20},
					string(mesh.TrafficLogType):        {Total: 30},
					string(mesh.HealthCheckType):       {Total: 40},
					string(mesh.CircuitBreakerType):    {Total: 50},
					string(mesh.FaultInjectionType):    {Total: 60},
					string(mesh.TrafficPermissionType): {Total: 70},
					string(mesh.ProxyTemplateType):     {Total: 80},
					string(mesh.ExternalServiceType):   {Total: 90},
					string(mesh.RateLimitType):         {Total: 100},
				},
			},
		},
	}

	Describe("InspectMeshesCmd", func() {

		var rootCmd *cobra.Command
		var buf *bytes.Buffer
		var store core_store.ResourceStore
		rootTime, _ := time.Parse(time.RFC3339, "2008-04-27T16:05:36.995Z")

		BeforeEach(func() {
			store = memory_resources.NewStore()
			for _, cb := range meshInsightResources {
				err := store.Create(context.Background(), cb, core_store.CreateBy(core_model.MetaToResourceKey(cb.GetMeta())))
				Expect(err).ToNot(HaveOccurred())
			}

			rootCtx, err := test_kumactl.MakeRootContext(rootTime, store)
			Expect(err).ToNot(HaveOccurred())

			rootCmd = cmd.NewRootCmd(rootCtx)
			buf = &bytes.Buffer{}
			rootCmd.SetOut(buf)
		})

		type testCase struct {
			outputFormat string
			goldenFile   string
			matcher      func(path ...string) gomega_types.GomegaMatcher
		}

		DescribeTable("kumactl inspect meshes -o table|json|yaml",
			func(given testCase) {
				// given
				rootCmd.SetArgs(append([]string{
					"--config-file", filepath.Join("..", "testdata", "sample-kumactl.config.yaml"),
					"inspect", "meshes"}, given.outputFormat))

				// when
				err := rootCmd.Execute()
				// then
				Expect(err).ToNot(HaveOccurred())
				Expect(buf.String()).To(given.matcher("testdata", given.goldenFile))
			},
			Entry("should support Table output by default", testCase{
				outputFormat: "",
				goldenFile:   "inspect-meshes.golden.txt",
				matcher:      matchers.MatchGoldenEqual,
			}),
			Entry("should support Table output explicitly", testCase{
				outputFormat: "-otable",
				goldenFile:   "inspect-meshes.golden.txt",
				matcher:      matchers.MatchGoldenEqual,
			}),
			Entry("should support JSON output", testCase{
				outputFormat: "-ojson",
				goldenFile:   "inspect-meshes.golden.json",
				matcher:      matchers.MatchGoldenJSON,
			}),
			Entry("should support YAML output", testCase{
				outputFormat: "-oyaml",
				goldenFile:   "inspect-meshes.golden.yaml",
				matcher:      matchers.MatchGoldenYAML,
			}),
		)
	})
})
