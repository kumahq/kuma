package get_test

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"
	"time"

	"github.com/kumahq/kuma/pkg/test/matchers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	gomega_types "github.com/onsi/gomega/types"
	"github.com/spf13/cobra"

	kumactl_resources "github.com/kumahq/kuma/app/kumactl/pkg/resources"

	"github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/app/kumactl/cmd"
	kumactl_cmd "github.com/kumahq/kuma/app/kumactl/pkg/cmd"
	config_proto "github.com/kumahq/kuma/pkg/config/app/kumactl/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	memory_resources "github.com/kumahq/kuma/pkg/plugins/resources/memory"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
)

var _ = Describe("kumactl get virtual-outbounds", func() {

	virtualOutboundResources := []*mesh.VirtualOutboundResource{
		{
			Spec: &v1alpha1.VirtualOutbound{
				Selectors: []*v1alpha1.Selector{
					{
						Match: map[string]string{
							"service": "web1",
							"version": "1.0",
						},
					},
				},
				Conf: &v1alpha1.VirtualOutbound_Conf{
					Host: "foo",
					Port: "80",
				},
			},
			Meta: &test_model.ResourceMeta{
				Mesh: "default",
				Name: "simple",
			},
		},
		{
			Spec: &v1alpha1.VirtualOutbound{
				Selectors: []*v1alpha1.Selector{
					{
						Match: map[string]string{
							"service": "web2",
							"version": "1.0",
						},
					},
				},
				Conf: &v1alpha1.VirtualOutbound_Conf{
					Host: "{{service}}.{{version}}.mesh",
					Port: "{{port}}",
					Parameters: map[string]string{
						"service": "service",
						"port":    "port",
						"version": "version",
					},
				},
			},
			Meta: &test_model.ResourceMeta{
				Mesh: "default",
				Name: "with-parameters",
			},
		},
	}

	Describe("GetVirtualOutboundCmd", func() {

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
					NewAPIServerClient: kumactl_resources.NewAPIServerClient,
				},
			}

			store = core_store.NewPaginationStore(memory_resources.NewStore())

			for _, ds := range virtualOutboundResources {
				err := store.Create(context.Background(), ds, core_store.CreateBy(core_model.MetaToResourceKey(ds.GetMeta())))
				Expect(err).ToNot(HaveOccurred())
			}

			rootCmd = cmd.NewRootCmd(rootCtx)
			buf = &bytes.Buffer{}
			rootCmd.SetOut(buf)
		})

		type testCase struct {
			outputFormat string
			pagination   string
			goldenFile   string
			matcher      func(interface{}) gomega_types.GomegaMatcher
		}

		DescribeTable("kumactl get virtual-outbounds -o table|json|yaml",
			func(given testCase) {
				// given
				rootCmd.SetArgs(append([]string{
					"--config-file", filepath.Join("..", "testdata", "sample-kumactl.config.yaml"),
					"get", "virtual-outbounds"}, given.outputFormat, given.pagination))

				// when
				err := rootCmd.Execute()
				// then
				Expect(err).ToNot(HaveOccurred())
				Expect(buf.String()).To(matchers.MatchGolden(given.matcher, filepath.Join("testdata", given.goldenFile)))
			},
			Entry("should support Table output by default", testCase{
				outputFormat: "",
				goldenFile:   "get-virtual-outbounds.golden.txt",
				matcher: func(expected interface{}) gomega_types.GomegaMatcher {
					return WithTransform(strings.TrimSpace, Equal(strings.TrimSpace(string(expected.([]byte)))))
				},
			}),
			Entry("should support Table output explicitly", testCase{
				outputFormat: "-otable",
				goldenFile:   "get-virtual-outbounds.golden.txt",
				matcher: func(expected interface{}) gomega_types.GomegaMatcher {
					return WithTransform(strings.TrimSpace, Equal(strings.TrimSpace(string(expected.([]byte)))))
				},
			}),
			Entry("should support pagination", testCase{
				outputFormat: "-otable",
				pagination:   "--size=1",
				goldenFile:   "get-virtual-outbounds.pagination.golden.txt",
				matcher: func(expected interface{}) gomega_types.GomegaMatcher {
					return WithTransform(strings.TrimSpace, Equal(strings.TrimSpace(string(expected.([]byte)))))
				},
			}),
			Entry("should support JSON output", testCase{
				outputFormat: "-ojson",
				goldenFile:   "get-virtual-outbounds.golden.json",
				matcher:      MatchJSON,
			}),
			Entry("should support YAML output", testCase{
				outputFormat: "-oyaml",
				goldenFile:   "get-virtual-outbounds.golden.yaml",
				matcher:      MatchYAML,
			}),
		)
	})

})
