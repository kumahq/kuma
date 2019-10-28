package get_test

import (
	"bytes"
	"context"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"

	"github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/app/kumactl/cmd"
	kumactl_cmd "github.com/Kong/kuma/app/kumactl/pkg/cmd"
	config_proto "github.com/Kong/kuma/pkg/config/app/kumactl/v1alpha1"
	"github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/Kong/kuma/pkg/core/resources/model"
	core_store "github.com/Kong/kuma/pkg/core/resources/store"
	memory_resources "github.com/Kong/kuma/pkg/plugins/resources/memory"
	test_model "github.com/Kong/kuma/pkg/test/resources/model"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	gomega_types "github.com/onsi/gomega/types"
	"github.com/spf13/cobra"
)

var _ = Describe("kumactl get traffic-permissions", func() {

	trafficPermissionResources := []*mesh.TrafficPermissionResource{
		{
			Spec: v1alpha1.TrafficPermission{
				Rules: []*v1alpha1.TrafficPermission_Rule{
					{
						Sources: []*v1alpha1.Selector{
							{
								Match: map[string]string{
									"service": "web1",
									"version": "1.0",
								},
							},
						},
						Destinations: []*v1alpha1.Selector{
							{
								Match: map[string]string{
									"service": "backend1",
									"env":     "dev",
								},
							},
						},
					},
				},
			},
			Meta: &test_model.ResourceMeta{
				Mesh:      "default",
				Name:      "web1-to-backend1",
				Namespace: "",
			},
		},
		{
			Spec: v1alpha1.TrafficPermission{
				Rules: []*v1alpha1.TrafficPermission_Rule{
					{
						Sources: []*v1alpha1.Selector{
							{
								Match: map[string]string{
									"service": "web2",
									"version": "1.0",
								},
							},
						},
						Destinations: []*v1alpha1.Selector{
							{
								Match: map[string]string{
									"service": "backend2",
									"env":     "dev",
								},
							},
						},
					},
				},
			},
			Meta: &test_model.ResourceMeta{
				Mesh:      "default",
				Name:      "web2-to-backend2",
				Namespace: "",
			},
		},
	}

	Describe("GetTrafficPermissionsCmd", func() {

		var rootCtx *kumactl_cmd.RootContext
		var rootCmd *cobra.Command
		var buf *bytes.Buffer
		var store core_store.ResourceStore

		BeforeEach(func() {
			// setup
			rootCtx = &kumactl_cmd.RootContext{
				Runtime: kumactl_cmd.RootRuntime{
					Now: time.Now,
					NewResourceStore: func(*config_proto.ControlPlaneCoordinates_ApiServer) (core_store.ResourceStore, error) {
						return store, nil
					},
				},
			}

			store = memory_resources.NewStore()

			for _, ds := range trafficPermissionResources {
				err := store.Create(context.Background(), ds, core_store.CreateBy(core_model.MetaToResourceKey(ds.GetMeta())))
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

		DescribeTable("kumactl get traffic-permissions -o table|json|yaml",
			func(given testCase) {
				// given
				rootCmd.SetArgs(append([]string{
					"--config-file", filepath.Join("..", "testdata", "sample-kumactl.config.yaml"),
					"get", "traffic-permissions"}, given.outputFormat))

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
				goldenFile:   "get-traffic-permissions.golden.txt",
				matcher: func(expected interface{}) gomega_types.GomegaMatcher {
					return WithTransform(strings.TrimSpace, Equal(strings.TrimSpace(string(expected.([]byte)))))
				},
			}),
			Entry("should support Table output explicitly", testCase{
				outputFormat: "-otable",
				goldenFile:   "get-traffic-permissions.golden.txt",
				matcher: func(expected interface{}) gomega_types.GomegaMatcher {
					return WithTransform(strings.TrimSpace, Equal(strings.TrimSpace(string(expected.([]byte)))))
				},
			}),
			Entry("should support JSON output", testCase{
				outputFormat: "-ojson",
				goldenFile:   "get-traffic-permissions.golden.json",
				matcher:      MatchJSON,
			}),
			Entry("should support YAML output", testCase{
				outputFormat: "-oyaml",
				goldenFile:   "get-traffic-permissions.golden.yaml",
				matcher:      MatchYAML,
			}),
		)
	})

})
