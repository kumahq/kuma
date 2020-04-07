package get_test

import (
	"bytes"
	"context"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	gomega_types "github.com/onsi/gomega/types"

	"github.com/spf13/cobra"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/app/kumactl/cmd"
	kumactl_cmd "github.com/Kong/kuma/app/kumactl/pkg/cmd"
	config_proto "github.com/Kong/kuma/pkg/config/app/kumactl/v1alpha1"
	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/Kong/kuma/pkg/core/resources/model"
	core_store "github.com/Kong/kuma/pkg/core/resources/store"
	memory_resources "github.com/Kong/kuma/pkg/plugins/resources/memory"
	test_model "github.com/Kong/kuma/pkg/test/resources/model"
)

var _ = Describe("kumactl get dataplane NAME", func() {
	var rootCtx *kumactl_cmd.RootContext
	var rootCmd *cobra.Command
	var outbuf, errbuf *bytes.Buffer
	var store core_store.ResourceStore
	var dataplane *mesh_core.DataplaneResource
	BeforeEach(func() {
		// setup
		dataplane = &mesh_core.DataplaneResource{
			Meta: &test_model.ResourceMeta{
				Mesh: "default",
				Name: "dataplane-1",
			},
			Spec: mesh_proto.Dataplane{
				Networking: &mesh_proto.Dataplane_Networking{
					Address: "127.0.0.1",
					Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
						{
							Port:        8080,
							ServicePort: 80,
							Tags: map[string]string{
								"service": "mobile",
								"version": "v1",
							},
						},
						{
							Port:        8090,
							ServicePort: 90,
							Tags: map[string]string{
								"service": "metrics",
								"version": "v1",
							},
						},
					},
				},
			},
		}
		key := core_model.ResourceKey{
			Mesh: dataplane.Meta.GetMesh(),
			Name: dataplane.Meta.GetName(),
		}
		rootCtx = &kumactl_cmd.RootContext{
			Runtime: kumactl_cmd.RootRuntime{
				Now: time.Now,
				NewResourceStore: func(*config_proto.ControlPlaneCoordinates_ApiServer) (core_store.ResourceStore, error) {
					return store, nil
				},
			},
		}
		store = memory_resources.NewStore()
		err := store.Create(context.Background(), dataplane, core_store.CreateBy(key))
		Expect(err).ToNot(HaveOccurred())

		rootCmd = cmd.NewRootCmd(rootCtx)
		outbuf = &bytes.Buffer{}
		errbuf = &bytes.Buffer{}
		rootCmd.SetOut(outbuf)
		rootCmd.SetErr(errbuf)
	})
	It("should throw an error in case of no args", func() {
		// given
		rootCmd.SetArgs([]string{
			"get", "dataplane"})

		// when
		err := rootCmd.Execute()

		// then
		Expect(err).To(HaveOccurred())
		// and
		Expect(err.Error()).To(Equal("requires at least 1 arg(s), only received 0"))
		// and
		Expect(outbuf.String()).To(MatchRegexp(`Error: requires at least 1 arg\(s\), only received 0`))
		// and
		Expect(errbuf.Bytes()).To(BeEmpty())
	})
	It("should return error message if doesn't exist", func() {
		//given
		rootCmd.SetArgs([]string{
			"get", "dataplane", "dataplane-2"})

		// when
		err := rootCmd.Execute()

		// then
		Expect(err).To(HaveOccurred())
		// and
		Expect(outbuf.String()).To(Equal("Error: No resources found in default mesh\n"))
		// and
		Expect(errbuf.Bytes()).To(BeEmpty())

	})
	Describe("kumactl get dataplane NAME -o table|json|yaml", func() {

		type testCase struct {
			outputFormat string
			goldenFile   string
			matcher      func(interface{}) gomega_types.GomegaMatcher
		}

		DescribeTable("kumactl get dataplane -o table|json|yaml",
			func(given testCase) {
				// given
				rootCmd.SetArgs(append([]string{
					"get", "dataplane", "dataplane-1"}, given.outputFormat))

				// when
				err := rootCmd.Execute()
				// then
				Expect(err).ToNot(HaveOccurred())

				// when
				expected, err := ioutil.ReadFile(filepath.Join("testdata", given.goldenFile))
				// then
				Expect(err).ToNot(HaveOccurred())
				// and
				Expect(outbuf.String()).To(given.matcher(expected))
			},
			Entry("should support Table output by default", testCase{
				outputFormat: "",
				goldenFile:   "get-dataplane.golden.txt",
				matcher: func(expected interface{}) gomega_types.GomegaMatcher {
					return WithTransform(strings.TrimSpace, Equal(strings.TrimSpace(string(expected.([]byte)))))
				},
			}),
			Entry("should support Table output explicitly", testCase{
				outputFormat: "-otable",
				goldenFile:   "get-dataplane.golden.txt",
				matcher: func(expected interface{}) gomega_types.GomegaMatcher {
					return WithTransform(strings.TrimSpace, Equal(strings.TrimSpace(string(expected.([]byte)))))
				},
			}),
			Entry("should support JSON output", testCase{
				outputFormat: "-ojson",
				goldenFile:   "get-dataplane.golden.json",
				matcher:      MatchJSON,
			}),
			Entry("should support YAML output", testCase{
				outputFormat: "-oyaml",
				goldenFile:   "get-dataplane.golden.yaml",
				matcher:      MatchYAML,
			}),
		)
	})
})
