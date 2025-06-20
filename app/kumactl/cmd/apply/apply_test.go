package apply_test

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"

	"github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/app/kumactl/cmd"
	kumactl_cmd "github.com/kumahq/kuma/app/kumactl/pkg/cmd"
	test_kumactl "github.com/kumahq/kuma/app/kumactl/pkg/test"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/rest/errors/types"
	memory_resources "github.com/kumahq/kuma/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/pkg/test/matchers"
	"github.com/kumahq/kuma/pkg/test/resources/model"
	test_store "github.com/kumahq/kuma/pkg/test/store"
	util_http "github.com/kumahq/kuma/pkg/util/http"
)

const defaultNetworkingSection = `networking:
  address: 2.2.2.2
  inbound:
    - port: 80
      tags:
        "kuma.io/service": "web"`

var _ = Describe("kumactl apply", func() {
	var rootCtx *kumactl_cmd.RootContext
	var rootCmd *cobra.Command
	var store core_store.ResourceStore
	BeforeEach(func() {
		rootCtx = test_kumactl.MakeMinimalRootContext()
		rootCtx.Runtime.Registry = registry.Global()
		rootCtx.Runtime.NewResourceStore = func(util_http.Client) core_store.ResourceStore {
			return store
		}
		store = core_store.NewPaginationStore(memory_resources.NewStore())
		rootCmd = cmd.NewRootCmd(rootCtx)
	})

	ValidatePersistedResource := func() {
		resource := mesh.NewDataplaneResource()
		err := store.Get(context.Background(), resource, core_store.GetByKey("sample", "default"))
		Expect(err).ToNot(HaveOccurred())

		// then
		Expect(resource.Meta.GetName()).To(Equal("sample"))
		Expect(resource.Meta.GetMesh()).To(Equal("default"))
		// and
		Expect(resource.Spec.Networking.Address).To(Equal("2.2.2.2"))
		// and
		Expect(resource.Spec.Networking.Inbound).To(HaveLen(1))
		Expect(resource.Spec.Networking.Inbound[0].Address).To(Equal("1.1.1.1"))
		Expect(resource.Spec.Networking.Inbound[0].Port).To(Equal(uint32(80)))
		Expect(resource.Spec.Networking.Inbound[0].ServicePort).To(Equal(uint32(8080)))
		Expect(resource.Spec.Networking.Inbound[0].Tags).To(HaveKeyWithValue("service", "web"))
		Expect(resource.Spec.Networking.Inbound[0].Tags).To(HaveKeyWithValue("version", "1.0"))
		Expect(resource.Spec.Networking.Inbound[0].Tags).To(HaveKeyWithValue("env", "production"))
		// and
		Expect(resource.Spec.Networking.Outbound).To(HaveLen(1))
		Expect(resource.Spec.Networking.Outbound[0].Port).To(Equal(uint32(3000)))
		// nolint:staticcheck
		Expect(resource.Spec.Networking.Outbound[0].GetService()).To(Equal("postgres"))
	}

	It("should require -f arg", func() {
		// setup
		rootCmd.SetArgs([]string{
			"--config-file", filepath.Join("..", "testdata", "sample-kumactl.config.yaml"),
			"apply",
		})

		// when
		err := rootCmd.Execute()

		// then
		Expect(err).To(MatchError(`required flag(s) "file" not set`))
	})

	It("should read configuration from stdin (-f - arg)", func() {
		// setup
		mockStdin, err := os.Open(filepath.Join("testdata", "golden", "apply-dataplane.input.yaml"))
		Expect(err).ToNot(HaveOccurred())

		// given
		rootCmd.SetArgs([]string{
			"--config-file", filepath.Join("..", "testdata", "sample-kumactl.config.yaml"),
			"apply", "-f", "-",
		})
		rootCmd.SetIn(mockStdin)

		// when
		err = rootCmd.Execute()
		// then
		Expect(err).ToNot(HaveOccurred())

		// and
		ValidatePersistedResource()
	})

	It("should apply a new Dataplane resource", func() {
		// given
		rootCmd.SetArgs([]string{
			"--config-file", filepath.Join("..", "testdata", "sample-kumactl.config.yaml"),
			"apply", "-f", filepath.Join("testdata", "golden", "apply-dataplane.input.yaml"),
		})

		// when
		err := rootCmd.Execute()
		// then
		Expect(err).ToNot(HaveOccurred())

		// and
		ValidatePersistedResource()
	})

	It("should apply an updated Dataplane resource", func() {
		// setup
		newResource := mesh.DataplaneResource{
			Spec: &v1alpha1.Dataplane{
				Networking: &v1alpha1.Dataplane_Networking{
					Address: "8.8.8.8",
					Inbound: []*v1alpha1.Dataplane_Networking_Inbound{
						{
							Port:        443,
							ServicePort: 8443,
							Tags: map[string]string{
								"service": "default",
								"version": "default",
								"env":     "default",
							},
						},
					},
				},
			},
		}
		err := store.Create(context.Background(), &newResource, core_store.CreateByKey("sample", "default"))
		Expect(err).ToNot(HaveOccurred())

		// given
		rootCmd.SetArgs([]string{
			"--config-file", filepath.Join("..", "testdata", "sample-kumactl.config.yaml"),
			"apply", "-f", filepath.Join("testdata", "golden", "apply-dataplane.input.yaml"),
		})

		// when
		err = rootCmd.Execute()
		// then
		Expect(err).ToNot(HaveOccurred())

		// and
		ValidatePersistedResource()
	})

	It("should apply a new Dataplane resource from URL", func() {
		// setup http server
		mux := http.NewServeMux()
		mux.Handle("/testdata/", http.StripPrefix("/testdata", http.FileServer(http.Dir("./testdata/golden"))))

		server := httptest.NewServer(mux)
		defer server.Close()

		port, err := strconv.Atoi(strings.Split(server.Listener.Addr().String(), ":")[1])
		testurl := fmt.Sprintf("http://localhost:%v/testdata/apply-dataplane.input.yaml", port)

		Expect(err).ToNot(HaveOccurred())

		// then
		Eventually(func() bool {
			resp, err := http.Get(testurl) // #nosec G107 -- reused in different places
			if err != nil {
				return false
			}
			if resp.StatusCode != 200 {
				return false
			}
			Expect(resp.Body.Close()).To(Succeed())
			return true
		}, "5s", "100ms").Should(BeTrue())

		// given
		rootCmd.SetArgs([]string{
			"apply", "-f", testurl,
		})

		// when
		err = rootCmd.Execute()

		// then
		Expect(err).ToNot(HaveOccurred())

		// and
		ValidatePersistedResource()
	})

	It("should fill in template (multiple variables)", func() {
		// given
		rootCmd.SetArgs([]string{
			"--config-file", filepath.Join("..", "testdata", "sample-kumactl.config.yaml"),
			"apply", "-f", filepath.Join("testdata", "apply-mesh-template.yaml"),
			"-v", "name=meshinit", "-v", "type=Mesh",
		})

		// when
		err := rootCmd.Execute()
		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		resource := mesh.NewMeshResource()
		// with production code, the mesh is not required for zone store. API Server then infer mesh from the name
		err = store.Get(context.Background(), resource, core_store.GetByKey("meshinit", ""))
		Expect(err).ToNot(HaveOccurred())

		// then
		Expect(resource.Meta.GetName()).To(Equal("meshinit"))
		Expect(resource.Meta.GetMesh()).To(Equal(core_model.NoMesh))
	})

	It("should return kuma api server error", func() {
		// setup
		rootCtx.Runtime.NewResourceStore = func(util_http.Client) core_store.ResourceStore {
			kumaErr := &types.Error{
				Title:  "Could not process resource",
				Detail: "Resource is not valid",
				InvalidParameters: []types.InvalidParameter{
					{
						Field:  "path",
						Reason: "cannot be empty",
					},
					{
						Field:  "mesh",
						Reason: "cannot be empty",
					},
				},
			}
			store := test_store.FailingStore{
				Err: kumaErr,
			}
			return &store
		}

		// given
		rootCmd.SetArgs([]string{
			"--config-file", filepath.Join("..", "testdata", "sample-kumactl.config.yaml"),
			"apply", "-f", filepath.Join("testdata", "golden", "apply-mesh.input.yaml"),
		})
		buf := &bytes.Buffer{}
		rootCmd.SetOut(buf)
		rootCmd.SetErr(buf)

		// when
		err := rootCmd.Execute()
		// then
		Expect(err).To(HaveOccurred())

		// then
		Expect(buf.String()).To(Equal(
			`Error: resource type="Mesh" mesh="" name="sample": failed server side: Could not process resource (Resource is not valid);path=cannot be empty ;mesh=cannot be empty 
`))
	})

	It("should print configuration with resolved variable without applying", func() {
		// setup
		err := store.Create(context.Background(), &mesh.DataplaneResource{
			Meta: &model.ResourceMeta{
				Name: "sample",
				Mesh: "default",
			},
			Spec: &v1alpha1.Dataplane{
				Networking: &v1alpha1.Dataplane_Networking{
					Address: "1.1.1.1",
				},
			},
		}, core_store.CreateByKey("sample", "default"))
		Expect(err).ToNot(HaveOccurred())

		// given
		rootCmd.SetArgs([]string{
			"apply", "-f", filepath.Join("testdata", "apply-dataplane-template.yaml"),
			"-v", "address=2.2.2.2", "--dry-run",
		})
		buf := &bytes.Buffer{}
		rootCmd.SetOut(buf)
		rootCmd.SetErr(buf)

		// when
		err = rootCmd.Execute()
		// then
		Expect(err).ToNot(HaveOccurred())
		// then
		resource := mesh.NewDataplaneResource()
		err = store.Get(context.Background(), resource, core_store.GetByKey("sample", "default"))
		Expect(err).ToNot(HaveOccurred())
		Expect(resource.Spec.Networking.Address).To(Equal("1.1.1.1"))

		// then
		Expect(buf.String()).To(MatchYAML(
			`creationTime: "0001-01-01T00:00:00Z"
mesh: default
modificationTime: "0001-01-01T00:00:00Z"
name: sample
type: Dataplane
networking:
  address: 2.2.2.2
  inbound:
    - port: 80
      tags:
        "kuma.io/service": "web"
`))
	})

	It("should support variable names that include dot character", func() {
		// given
		rootCmd.SetArgs([]string{
			"apply", "-f", filepath.Join("testdata", "apply-dataplane-template-dots.yaml"),
			"-v", "var.with.dots.in.the.name=2.2.2.2",
		})

		// when
		err := rootCmd.Execute()
		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		resource := mesh.NewDataplaneResource()
		err = store.Get(context.Background(), resource, core_store.GetByKey("sample", "default"))
		Expect(err).ToNot(HaveOccurred())

		// then
		Expect(resource.Meta.GetName()).To(Equal("sample"))
		Expect(resource.Meta.GetMesh()).To(Equal("default"))
		Expect(resource.Spec.Networking.Address).To(Equal("2.2.2.2"))
	})

	It("dry-run outputs templated resources", func() {
		// given
		rootCmd.SetArgs([]string{
			"apply", "-f", filepath.Join("testdata", "apply-many-dataplane-template.yaml"),
			"-v", "address=2.2.2.2", "--dry-run",
		})

		// when
		buf := &bytes.Buffer{}
		rootCmd.SetOut(buf)
		rootCmd.SetErr(buf)
		err := rootCmd.Execute()

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(buf.String()).To(matchers.MatchGoldenEqual(filepath.Join("testdata", "apply-many-dataplane-template.golden.yaml")))
	})

	type testCase struct {
		resource string
		err      string
	}

	DescribeTable("should fail on invalid resource",
		func(given testCase) {
			// given
			rootCmd.SetArgs([]string{
				"--config-file", filepath.Join("..", "testdata", "sample-kumactl.config.yaml"),
				"apply", "-f", "-",
			})
			rootCmd.SetIn(strings.NewReader(given.resource))
			stderr := &bytes.Buffer{}
			rootCmd.SetErr(stderr)

			// when
			err := rootCmd.Execute()

			// then
			Expect(err).To(HaveOccurred())
			Expect(err).To(HaveOccurred())
			Expect(stderr.String()).To(ContainSubstring(given.err))
		},
		Entry("no mesh", testCase{
			resource: `
type: Dataplane
name: dp-1
` + defaultNetworkingSection,
			err: "mesh: cannot be empty",
		}),
		Entry("no name", testCase{
			resource: `
type: Dataplane
mesh: default
` + defaultNetworkingSection,
			err: "name: cannot be empty",
		}),
		Entry("invalid data", testCase{
			resource: `
type: Dataplane
name: dp-1
mesh: default
networking:
  inbound: 0 # should be a string
`,
			err: `resource[0]: failed to parse resource: invalid Dataplane object: "error unmarshaling JSON: while decoding JSON: json: cannot unmarshal number into Go value of type []json.RawMessage"`,
		}),
		Entry("no resource", testCase{
			resource: ``,
			err:      "no resource(s) passed to apply",
		}),
		Entry("data not passing schema validation", testCase{
			resource: `
type: MeshTrafficPermission
mesh: mesh-1
name: mtp-1
spec:
  targetRef:
    kind: Mesh
  from:
    - targetRef:
        kind: Mesh
      default:
        action: foo
`,
			err: `resource[0]: failed to parse resource: spec.from[0].default.action: in body should be one of [Allow Deny AllowWithShadowDeny]`,
		}),
	)

	DescribeTable("apply test",
		func(ctx SpecContext, inputFile string) {
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}
			rootCmd.SetArgs([]string{
				"apply", "-f", inputFile,
			})
			rootCmd.SetOut(stdout)
			rootCmd.SetErr(stderr)

			// when
			_ = rootCmd.Execute()

			Expect(stdout.String()).To(matchers.MatchGoldenEqual(strings.Replace(inputFile, ".input.yaml", ".golden.stdout.txt", 1)))
			Expect(stderr.String()).To(matchers.MatchGoldenEqual(strings.Replace(inputFile, ".input.yaml", ".golden.stderr.txt", 1)))
			storeOut, err := test_store.ExtractResources(ctx, store)
			Expect(err).To(Succeed())
			Expect(storeOut).To(matchers.MatchGoldenEqual(strings.Replace(inputFile, ".input.yaml", ".golden.store.yaml", 1)))
		},
		test.EntriesForFolder("golden"),
	)
})
