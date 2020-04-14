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

	"github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/app/kumactl/cmd"
	kumactl_cmd "github.com/Kong/kuma/app/kumactl/pkg/cmd"
	config_proto "github.com/Kong/kuma/pkg/config/app/kumactl/v1alpha1"
	"github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/Kong/kuma/pkg/core/resources/model"
	core_store "github.com/Kong/kuma/pkg/core/resources/store"
	memory_resources "github.com/Kong/kuma/pkg/plugins/resources/memory"
	test_model "github.com/Kong/kuma/pkg/test/resources/model"
)

var _ = Describe("kumactl get meshes", func() {

	sampleMeshes := []*mesh.MeshResource{
		{
			Spec: v1alpha1.Mesh{
				Mtls: &v1alpha1.Mesh_Mtls{
					Enabled: true,
					Ca: &v1alpha1.CertificateAuthority{
						Type: &v1alpha1.CertificateAuthority_Builtin_{
							Builtin: &v1alpha1.CertificateAuthority_Builtin{},
						},
					},
				},
			},
			Meta: &test_model.ResourceMeta{
				Mesh: "mesh1",
				Name: "mesh1",
			},
		},
		{
			Spec: v1alpha1.Mesh{
				Mtls: &v1alpha1.Mesh_Mtls{
					Enabled: true,
					Ca: &v1alpha1.CertificateAuthority{
						Type: &v1alpha1.CertificateAuthority_Provided_{
							Provided: &v1alpha1.CertificateAuthority_Provided{},
						},
					},
				},
			},
			Meta: &test_model.ResourceMeta{
				Mesh: "mesh2",
				Name: "mesh2",
			},
		},
		{
			Spec: v1alpha1.Mesh{
				Mtls: &v1alpha1.Mesh_Mtls{
					Enabled: false,
					Ca: &v1alpha1.CertificateAuthority{
						Type: &v1alpha1.CertificateAuthority_Provided_{
							Provided: &v1alpha1.CertificateAuthority_Provided{},
						},
					},
				},
				Metrics: &v1alpha1.Metrics{
					Prometheus: &v1alpha1.Metrics_Prometheus{
						Port: 1234,
						Path: "/non-standard-path",
					},
				},
				Logging: &v1alpha1.Logging{
					Backends: []*v1alpha1.LoggingBackend{
						{
							Name: "logstash",
							Type: &v1alpha1.LoggingBackend_Tcp_{
								Tcp: &v1alpha1.LoggingBackend_Tcp{
									Address: "127.0.0.1:5000",
								},
							},
						},
						{
							Name: "file",
							Type: &v1alpha1.LoggingBackend_File_{
								File: &v1alpha1.LoggingBackend_File{
									Path: "/tmp/service.log",
								},
							},
						},
					},
				},
				Tracing: &v1alpha1.Tracing{
					Backends: []*v1alpha1.TracingBackend{
						{
							Name: "zipkin-us",
							Type: &v1alpha1.TracingBackend_Zipkin_{
								Zipkin: &v1alpha1.TracingBackend_Zipkin{
									Url: "http://zipkin.us:8080/v1/spans",
								},
							},
						},
						{
							Name: "zipkin-eu",
							Type: &v1alpha1.TracingBackend_Zipkin_{
								Zipkin: &v1alpha1.TracingBackend_Zipkin{
									Url: "http://zipkin.eu:8080/v1/spans",
								},
							},
						},
					},
				},
			},
			Meta: &test_model.ResourceMeta{
				Mesh: "mesh3",
				Name: "mesh3",
			},
		},
		{
			Spec: v1alpha1.Mesh{
				Mtls: &v1alpha1.Mesh_Mtls{
					Enabled: false,
					Ca: &v1alpha1.CertificateAuthority{
						Type: &v1alpha1.CertificateAuthority_Provided_{
							Provided: &v1alpha1.CertificateAuthority_Provided{},
						},
					},
				},
				Metrics: &v1alpha1.Metrics{
					Prometheus: &v1alpha1.Metrics_Prometheus{
						Port: 1234,
						Path: "/non-standard-path",
					},
				},
				Logging: &v1alpha1.Logging{
					Backends: []*v1alpha1.LoggingBackend{},
				},
				Tracing: &v1alpha1.Tracing{
					Backends: []*v1alpha1.TracingBackend{},
				},
			},
			Meta: &test_model.ResourceMeta{
				Mesh: "mesh4",
				Name: "mesh4",
			},
		},
	}

	Describe("GetMeshesCmd", func() {

		var rootCtx *kumactl_cmd.RootContext
		var rootCmd *cobra.Command
		var buf *bytes.Buffer
		var store core_store.ResourceStore
		var t1 time.Time
		BeforeEach(func() {
			// setup
			t1, _ = time.Parse(time.RFC3339, "2018-07-17T16:05:36.995+00:00")
			rootCtx = &kumactl_cmd.RootContext{
				Runtime: kumactl_cmd.RootRuntime{
					Now: time.Now,
					NewResourceStore: func(*config_proto.ControlPlaneCoordinates_ApiServer) (core_store.ResourceStore, error) {
						return store, nil
					},
				},
			}

			store = memory_resources.NewStore()

			for _, ds := range sampleMeshes {
				key := core_model.ResourceKey{
					Mesh: ds.Meta.GetMesh(),
					Name: ds.Meta.GetName(),
				}
				err := store.Create(context.Background(), ds, core_store.CreateBy(key), core_store.CreatedAt(t1))
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

		DescribeTable("kumactl get meshes -o table|json|yaml",
			func(given testCase) {
				// given
				rootCmd.SetArgs(append([]string{
					"--config-file", filepath.Join("..", "testdata", "sample-kumactl.config.yaml"),
					"get", "meshes"}, given.outputFormat))

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
				goldenFile:   "get-meshes.golden.txt",
				matcher: func(expected interface{}) gomega_types.GomegaMatcher {
					return WithTransform(strings.TrimSpace, Equal(strings.TrimSpace(string(expected.([]byte)))))
				},
			}),
			Entry("should support Table output explicitly", testCase{
				outputFormat: "-otable",
				goldenFile:   "get-meshes.golden.txt",
				matcher: func(expected interface{}) gomega_types.GomegaMatcher {
					return WithTransform(strings.TrimSpace, Equal(strings.TrimSpace(string(expected.([]byte)))))
				},
			}),
			Entry("should support JSON output", testCase{
				outputFormat: "-ojson",
				goldenFile:   "get-meshes.golden.json",
				matcher:      MatchJSON,
			}),
			Entry("should support YAML output", testCase{
				outputFormat: "-oyaml",
				goldenFile:   "get-meshes.golden.yaml",
				matcher:      MatchYAML,
			}),
		)
	})

})
