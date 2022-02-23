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
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"

	"github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/app/kumactl/cmd"
	kumactl_cmd "github.com/kumahq/kuma/app/kumactl/pkg/cmd"
	config_proto "github.com/kumahq/kuma/pkg/config/app/kumactl/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/rest/errors/types"
	memory_resources "github.com/kumahq/kuma/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/pkg/test/resources/model"
	test_store "github.com/kumahq/kuma/pkg/test/store"
	util_http "github.com/kumahq/kuma/pkg/util/http"
	"github.com/kumahq/kuma/pkg/util/test"
)

var _ = Describe("kumactl apply", func() {

	var rootCtx *kumactl_cmd.RootContext
	var rootCmd *cobra.Command
	var store core_store.ResourceStore
	BeforeEach(func() {
		rootCtx = &kumactl_cmd.RootContext{
			Runtime: kumactl_cmd.RootRuntime{
				Registry: registry.Global(),
				NewBaseAPIServerClient: func(server *config_proto.ControlPlaneCoordinates_ApiServer, _ time.Duration) (util_http.Client, error) {
					return nil, nil
				},
				NewResourceStore: func(util_http.Client) core_store.ResourceStore {
					return store
				},
				NewAPIServerClient: test.GetMockNewAPIServerClient(),
			},
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
		Expect(resource.Spec.Networking.Outbound[0].Service).To(Equal("postgres"))
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
		mockStdin, err := os.Open(filepath.Join("testdata", "apply-dataplane.yaml"))
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
			"apply", "-f", filepath.Join("testdata", "apply-dataplane.yaml")},
		)

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
			"apply", "-f", filepath.Join("testdata", "apply-dataplane.yaml")},
		)

		// when
		err = rootCmd.Execute()
		// then
		Expect(err).ToNot(HaveOccurred())

		// and
		ValidatePersistedResource()
	})

	It("should apply a Mesh resource", func() {
		// given
		rootCmd.SetArgs([]string{
			"--config-file", filepath.Join("..", "testdata", "sample-kumactl.config.yaml"),
			"apply", "-f", filepath.Join("testdata", "apply-mesh.yaml")},
		)

		// when
		err := rootCmd.Execute()
		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		resource := mesh.NewMeshResource()
		// with production code, the mesh is not required for zone store. API Server then infer mesh from the name
		err = store.Get(context.Background(), resource, core_store.GetByKey("sample", ""))
		Expect(err).ToNot(HaveOccurred())

		// then
		Expect(resource.Meta.GetName()).To(Equal("sample"))
		Expect(resource.Meta.GetMesh()).To(Equal(core_model.NoMesh))
	})

	It("should apply a Secret resource", func() {
		// given
		rootCmd.SetArgs([]string{
			"--config-file", filepath.Join("..", "testdata", "sample-kumactl.config.yaml"),
			"apply", "-f", filepath.Join("testdata", "apply-secret.yaml")},
		)

		// when
		err := rootCmd.Execute()
		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		secret := system.NewSecretResource()
		err = store.Get(context.Background(), secret, core_store.GetByKey("sample", "default"))
		Expect(err).ToNot(HaveOccurred())

		// then
		Expect(secret.Meta.GetName()).To(Equal("sample"))
		Expect(secret.Meta.GetMesh()).To(Equal("default"))
	})

	It("should apply a new Dataplane resource from URL", func() {
		// setup http server
		mux := http.NewServeMux()
		mux.Handle("/testdata/", http.StripPrefix("/testdata/", http.FileServer(http.Dir("./testdata"))))

		server := httptest.NewServer(mux)
		defer server.Close()

		port, err := strconv.Atoi(strings.Split(server.Listener.Addr().String(), ":")[1])
		testurl := fmt.Sprintf("http://localhost:%v/testdata/apply-dataplane.yaml", port)

		Expect(err).ToNot(HaveOccurred())

		// then
		Eventually(func() bool {
			resp, err := http.Get(testurl)
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
			"apply", "-f", testurl},
		)

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
			"-v", "name=meshinit", "-v", "type=Mesh"},
		)

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

	It("should apply multiple resources of same type", func() {
		// given
		rootCmd.SetArgs([]string{
			"--config-file", filepath.Join("..", "testdata", "sample-kumactl.config.yaml"),
			"apply", "-f", filepath.Join("testdata", "apply-multiple-resource-same-type.yaml")},
		)

		// when
		err := rootCmd.Execute()
		// then
		Expect(err).ToNot(HaveOccurred())

		// and
		resource := mesh.DataplaneResourceList{}
		err = store.List(context.Background(), &resource, core_store.ListByMesh("default"))
		Expect(err).ToNot(HaveOccurred())

		Expect(resource.Items[0].Meta.GetName()).To(Equal("sample1"))
		Expect(resource.Items[1].Meta.GetName()).To(Equal("sample2"))
	})

	It("should apply multiple resources of different type", func() {
		// given
		rootCmd.SetArgs([]string{
			"--config-file", filepath.Join("..", "testdata", "sample-kumactl.config.yaml"),
			"apply", "-f", filepath.Join("testdata", "apply-multiple-resource-different-type.yaml")},
		)

		// when
		err := rootCmd.Execute()
		// then
		Expect(err).ToNot(HaveOccurred())

		// and
		dataplaneResource := mesh.NewDataplaneResource()
		err = store.Get(context.Background(), dataplaneResource, core_store.GetByKey("sample1", "default"))
		Expect(err).ToNot(HaveOccurred())

		Expect(dataplaneResource.Meta.GetName()).To(Equal("sample1"))
		Expect(dataplaneResource.Meta.GetMesh()).To(Equal("default"))

		secret := system.NewSecretResource()
		err = store.Get(context.Background(), secret, core_store.GetByKey("sample", "default"))
		Expect(err).ToNot(HaveOccurred())

		Expect(secret.Meta.GetName()).To(Equal("sample"))
		Expect(secret.Meta.GetMesh()).To(Equal("default"))

	})

	It("should return kuma api server error", func() {
		// setup
		rootCtx.Runtime.NewResourceStore = func(util_http.Client) core_store.ResourceStore {
			kumaErr := &types.Error{
				Title:   "Could not process resource",
				Details: "Resource is not valid",
				Causes: []types.Cause{
					{
						Field:   "path",
						Message: "cannot be empty",
					},
					{
						Field:   "mesh",
						Message: "cannot be empty",
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
			"apply", "-f", filepath.Join("testdata", "apply-mesh.yaml")},
		)
		buf := &bytes.Buffer{}
		rootCmd.SetOut(buf)
		rootCmd.SetErr(buf)

		// when
		err := rootCmd.Execute()
		// then
		Expect(err).To(HaveOccurred())

		// then
		Expect(buf.String()).To(Equal(
			`Error: Could not process resource (Resource is not valid)
* path: cannot be empty
* mesh: cannot be empty
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
		Expect(buf.String()).To(Equal(
			`creationTime: "0001-01-01T00:00:00Z"
mesh: default
modificationTime: "0001-01-01T00:00:00Z"
name: sample
type: Dataplane
networking:
  address: 2.2.2.2
`))
	})

	It("should support variable names that include dot character", func() {
		// given
		rootCmd.SetArgs([]string{
			"apply", "-f", filepath.Join("testdata", "apply-dataplane-template-dots.yaml"),
			"-v", "var.with.dots.in.the.name=2.2.2.2"},
		)

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

	type testCase struct {
		resource string
		err      string
	}

	DescribeTable("should fail on invalid resource",
		func(given testCase) {
			// given
			rootCmd.SetArgs([]string{
				"--config-file", filepath.Join("..", "testdata", "sample-kumactl.config.yaml"),
				"apply", "-f", "-"},
			)
			rootCmd.SetIn(strings.NewReader(given.resource))

			// when
			err := rootCmd.Execute()

			// then
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(given.err))
		},
		Entry("no mesh", testCase{
			resource: `
type: Dataplane
name: dp-1
`,
			err: "mesh: cannot be empty",
		}),
		Entry("no name", testCase{
			resource: `
type: Dataplane
mesh: default
`,
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
			err: `YAML contains invalid resource: invalid Dataplane object "dp-1"`,
		}),
		Entry("no resource", testCase{
			resource: ``,
			err:      "no resource(s) passed to apply",
		}),
	)
})
