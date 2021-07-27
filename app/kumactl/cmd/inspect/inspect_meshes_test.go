package inspect_test

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	gomega_types "github.com/onsi/gomega/types"
	"github.com/spf13/cobra"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/app/kumactl/cmd"
	kumactl_cmd "github.com/kumahq/kuma/app/kumactl/pkg/cmd"
	config_proto "github.com/kumahq/kuma/pkg/config/app/kumactl/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	memory_resources "github.com/kumahq/kuma/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/pkg/test/matchers"
	"github.com/kumahq/kuma/pkg/test/resources/model"
	"github.com/kumahq/kuma/pkg/util/test"
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

		var rootCtx *kumactl_cmd.RootContext
		var rootCmd *cobra.Command
		var buf *bytes.Buffer
		var store core_store.ResourceStore
		rootTime, _ := time.Parse(time.RFC3339, "2008-04-27T16:05:36.995Z")

		BeforeEach(func() {
			rootCtx = &kumactl_cmd.RootContext{
				Runtime: kumactl_cmd.RootRuntime{
					Now: func() time.Time { return rootTime },
					NewResourceStore: func(*config_proto.ControlPlaneCoordinates_ApiServer) (core_store.ResourceStore, error) {
						return store, nil
					},
					NewAPIServerClient: test.GetMockNewAPIServerClient(),
				},
			}

			store = memory_resources.NewStore()
			for _, cb := range meshInsightResources {
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
			matcher      func(interface{}) gomega_types.GomegaMatcher
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
				Expect(buf.String()).To(matchers.MatchGoldenEqual(filepath.Join("testdata", given.goldenFile)))
			},
			Entry("should support Table output by default", testCase{
				outputFormat: "",
				goldenFile:   "inspect-meshes.golden.txt",
				matcher: func(expected interface{}) gomega_types.GomegaMatcher {
					return WithTransform(strings.TrimSpace, Equal(strings.TrimSpace(string(expected.([]byte)))))
				},
			}),
			Entry("should support Table output explicitly", testCase{
				outputFormat: "-otable",
				goldenFile:   "inspect-meshes.golden.txt",
				matcher: func(expected interface{}) gomega_types.GomegaMatcher {
					return WithTransform(strings.TrimSpace, Equal(strings.TrimSpace(string(expected.([]byte)))))
				},
			}),
			Entry("should support JSON output", testCase{
				outputFormat: "-ojson",
				goldenFile:   "inspect-meshes.golden.json",
				matcher:      MatchJSON,
			}),
			Entry("should support YAML output", testCase{
				outputFormat: "-oyaml",
				goldenFile:   "inspect-meshes.golden.yaml",
				matcher:      MatchYAML,
			}),
		)
	})
})
