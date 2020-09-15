package get_test

import (
	"bytes"
	"context"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"

	"github.com/kumahq/kuma/app/kumactl/cmd"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	gomega_types "github.com/onsi/gomega/types"

	"github.com/spf13/cobra"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	kumactl_cmd "github.com/kumahq/kuma/app/kumactl/pkg/cmd"
	config_proto "github.com/kumahq/kuma/pkg/config/app/kumactl/v1alpha1"
	mesh_core "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	memory_resources "github.com/kumahq/kuma/pkg/plugins/resources/memory"

	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
)

var _ = Describe("kumactl get external-services", func() {

	var externalServices []*mesh_core.ExternalServiceResource
	BeforeEach(func() {
		// setup
		externalServices = []*mesh_core.ExternalServiceResource{
			{
				Meta: &test_model.ResourceMeta{
					Mesh: "default",
					Name: "experiment",
				},
				Spec: mesh_proto.ExternalService{
					Networking: &mesh_proto.ExternalService_Networking{
						Address: "127.0.0.1",
						Inbound: []*mesh_proto.ExternalService_Networking_Inbound{
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
			},
			{
				Meta: &test_model.ResourceMeta{
					Mesh: "default",
					Name: "example",
				},
				Spec: mesh_proto.ExternalService{
					Networking: &mesh_proto.ExternalService_Networking{
						Address: "127.0.0.2",
						Inbound: []*mesh_proto.ExternalService_Networking_Inbound{
							{
								Port:        8080,
								ServicePort: 80,
								Tags: map[string]string{
									"service": "web",
									"version": "v2",
								},
							},
						},
					},
				},
			},
		}
	})

	Describe("GetExternalServicesCmd", func() {

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

			for _, pt := range externalServices {
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
			pagination   string
			goldenFile   string
			matcher      func(interface{}) gomega_types.GomegaMatcher
		}

		DescribeTable("kumactl get external-services -o table|json|yaml",
			func(given testCase) {
				// given
				rootCmd.SetArgs(append([]string{
					"--config-file", filepath.Join("..", "testdata", "sample-kumactl.config.yaml"),
					"get", "external-services"}, given.outputFormat, given.pagination))

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
				goldenFile:   "get-external-services.golden.txt",
				matcher: func(expected interface{}) gomega_types.GomegaMatcher {
					return WithTransform(strings.TrimSpace, Equal(strings.TrimSpace(string(expected.([]byte)))))
				},
			}),
			Entry("should support Table output explicitly", testCase{
				outputFormat: "-otable",
				goldenFile:   "get-external-services.golden.txt",
				matcher: func(expected interface{}) gomega_types.GomegaMatcher {
					return WithTransform(strings.TrimSpace, Equal(strings.TrimSpace(string(expected.([]byte)))))
				},
			}),
			Entry("should support pagination", testCase{
				outputFormat: "-otable",
				goldenFile:   "get-external-services.pagination.golden.txt",
				pagination:   "--size=1",
				matcher: func(expected interface{}) gomega_types.GomegaMatcher {
					return WithTransform(strings.TrimSpace, Equal(strings.TrimSpace(string(expected.([]byte)))))
				},
			}),
			Entry("should support JSON output", testCase{
				outputFormat: "-ojson",
				goldenFile:   "get-external-services.golden.json",
				matcher:      MatchJSON,
			}),
			Entry("should support YAML output", testCase{
				outputFormat: "-oyaml",
				goldenFile:   "get-external-services.golden.yaml",
				matcher:      MatchYAML,
			}),
		)
	})
})
