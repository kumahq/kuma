package get_test

import (
	"bytes"
	"context"
	"github.com/Kong/kuma/app/kumactl/cmd"
	"io/ioutil"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	gomega_types "github.com/onsi/gomega/types"

	"github.com/spf13/cobra"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	kumactl_cmd "github.com/Kong/kuma/app/kumactl/pkg/cmd"
	config_proto "github.com/Kong/kuma/pkg/config/app/kumactl/v1alpha1"
	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/Kong/kuma/pkg/core/resources/model"
	core_store "github.com/Kong/kuma/pkg/core/resources/store"
	memory_resources "github.com/Kong/kuma/pkg/plugins/resources/memory"

	test_model "github.com/Kong/kuma/pkg/test/resources/model"
)

var _ = Describe("kumactl get proxytemplates", func() {

	var sampleProxyTemplates []*mesh_core.ProxyTemplateResource

	BeforeEach(func() {
		sampleProxyTemplates = []*mesh_core.ProxyTemplateResource{
			{
				Meta: &test_model.ResourceMeta{
					Mesh: "default",
					Name: "custom-template",
				},
				Spec: mesh_proto.ProxyTemplate{},
			},
			{
				Meta: &test_model.ResourceMeta{
					Mesh: "default",
					Name: "another-template",
				},
				Spec: mesh_proto.ProxyTemplate{},
			},
			{
				Meta: &test_model.ResourceMeta{
					Mesh: "pilot",
					Name: "simple-template",
				},
				Spec: mesh_proto.ProxyTemplate{},
			},
		}
	})

	Describe("GetProxyTemplatesCmd", func() {

		var rootCtx *kumactl_cmd.RootContext
		var rootCmd *cobra.Command
		var buf *bytes.Buffer
		var store core_store.ResourceStore

		BeforeEach(func() {
			// setup

			rootCtx = &kumactl_cmd.RootContext{
				Runtime: kumactl_cmd.RootRuntime{
					NewResourceStore: func(*config_proto.ControlPlaneCoordinates_ApiServer) (core_store.ResourceStore, error) {
						return store, nil
					},
				},
			}

			store = memory_resources.NewStore()

			for _, pt := range sampleProxyTemplates {
				key := core_model.ResourceKey{
					Mesh: pt.Meta.GetMesh(),
					Name: pt.Meta.GetName(),
				}
				err := store.Create(context.Background(), pt, core_store.CreateBy(key))
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

		DescribeTable("kumactl get proxytemplates -o table|json|yaml",
			func(given testCase) {
				// given
				rootCmd.SetArgs(append([]string{
					"--config-file", filepath.Join("..", "testdata", "sample-kumactl.config.yaml"),
					"get", "proxytemplates"}, given.outputFormat))

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
				goldenFile:   "get-proxytemplates.golden.txt",
				matcher: func(expected interface{}) gomega_types.GomegaMatcher {
					return WithTransform(strings.TrimSpace, Equal(strings.TrimSpace(string(expected.([]byte)))))
				},
			}),
			Entry("should support Table output explicitly", testCase{
				outputFormat: "-otable",
				goldenFile:   "get-proxytemplates.golden.txt",
				matcher: func(expected interface{}) gomega_types.GomegaMatcher {
					return WithTransform(strings.TrimSpace, Equal(strings.TrimSpace(string(expected.([]byte)))))
				},
			}),
			Entry("should support JSON output", testCase{
				outputFormat: "-ojson",
				goldenFile:   "get-proxytemplates.golden.json",
				matcher:      MatchJSON,
			}),
			Entry("should support YAML output", testCase{
				outputFormat: "-oyaml",
				goldenFile:   "get-proxytemplates.golden.yaml",
				matcher:      MatchYAML,
			}),
		)
	})
})
