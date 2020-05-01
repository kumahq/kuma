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

	"github.com/Kong/kuma/pkg/catalog"
	catalog_client "github.com/Kong/kuma/pkg/catalog/client"
	"github.com/Kong/kuma/pkg/core/resources/apis/system"
	test_catalog "github.com/Kong/kuma/pkg/test/catalog"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"

	"github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/app/kumactl/cmd"
	kumactl_cmd "github.com/Kong/kuma/app/kumactl/pkg/cmd"
	config_proto "github.com/Kong/kuma/pkg/config/app/kumactl/v1alpha1"
	"github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	core_store "github.com/Kong/kuma/pkg/core/resources/store"
	"github.com/Kong/kuma/pkg/core/rest/errors/types"
	memory_resources "github.com/Kong/kuma/pkg/plugins/resources/memory"
	"github.com/Kong/kuma/pkg/test/resources/model"
	test_store "github.com/Kong/kuma/pkg/test/store"
)

var _ = Describe("kumactl apply", func() {

	var rootCtx *kumactl_cmd.RootContext
	var rootCmd *cobra.Command
	var store core_store.ResourceStore
	BeforeEach(func() {
		rootCtx = &kumactl_cmd.RootContext{
			Runtime: kumactl_cmd.RootRuntime{
				NewResourceStore: func(*config_proto.ControlPlaneCoordinates_ApiServer) (core_store.ResourceStore, error) {
					return store, nil
				},
				NewAdminResourceStore: func(string, *config_proto.Context_AdminApiCredentials) (core_store.ResourceStore, error) {
					return store, nil
				},
				NewCatalogClient: func(s string) (catalog_client.CatalogClient, error) {
					return &test_catalog.StaticCatalogClient{
						Resp: catalog.Catalog{
							Apis: catalog.Apis{
								DataplaneToken: catalog.DataplaneTokenApi{
									LocalUrl: "http://localhost:1234",
								},
							},
						},
					}, nil
				},
			},
		}
		store = memory_resources.NewStore()
		rootCmd = cmd.NewRootCmd(rootCtx)
	})

	ValidatePersistedResource := func() {
		resource := mesh.DataplaneResource{}
		err := store.Get(context.Background(), &resource, core_store.GetByKey("sample", "default"))
		Expect(err).ToNot(HaveOccurred())

		// then
		Expect(resource.Meta.GetName()).To(Equal("sample"))
		Expect(resource.Meta.GetMesh()).To(Equal("default"))
		// and
		Expect(resource.Spec.Networking.Address).To(Equal("2.2.2.2"))
		//and
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
		Expect(resource.Spec.Networking.Outbound[0].Service).To(Equal("postgres"))
	}

	It("should read configuration from stdin (no -f arg)", func() {
		// setup
		mockStdin, err := os.Open(filepath.Join("testdata", "apply-dataplane.yaml"))
		Expect(err).ToNot(HaveOccurred())

		// given
		rootCmd.SetArgs([]string{
			"--config-file", filepath.Join("..", "testdata", "sample-kumactl.config.yaml"),
			"apply",
		})
		rootCmd.SetIn(mockStdin)

		// when
		err = rootCmd.Execute()
		// then
		Expect(err).ToNot(HaveOccurred())

		// and
		ValidatePersistedResource()
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
			Spec: v1alpha1.Dataplane{
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
		resource := mesh.MeshResource{}
		// with production code, the mesh is not required for remote store. API Server then infer mesh from the name
		err = store.Get(context.Background(), &resource, core_store.GetByKey("sample", ""))
		Expect(err).ToNot(HaveOccurred())

		// then
		Expect(resource.Meta.GetName()).To(Equal("sample"))
		Expect(resource.Meta.GetMesh()).To(Equal("sample"))
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
		secret := system.SecretResource{}
		err = store.Get(context.Background(), &secret, core_store.GetByKey("sample", "default"))
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
		resource := mesh.MeshResource{}
		// with production code, the mesh is not required for remote store. API Server then infer mesh from the name
		err = store.Get(context.Background(), &resource, core_store.GetByKey("meshinit", ""))
		Expect(err).ToNot(HaveOccurred())

		// then
		Expect(resource.Meta.GetName()).To(Equal("meshinit"))
		Expect(resource.Meta.GetMesh()).To(Equal("meshinit"))
	})

	It("should return kuma api server error", func() {
		// setup
		rootCtx.Runtime.NewResourceStore = func(*config_proto.ControlPlaneCoordinates_ApiServer) (core_store.ResourceStore, error) {
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
			return &store, nil
		}

		// given
		rootCmd.SetArgs([]string{
			"--config-file", filepath.Join("..", "testdata", "sample-kumactl.config.yaml"),
			"apply", "-f", filepath.Join("testdata", "apply-mesh.yaml")},
		)
		buf := &bytes.Buffer{}
		rootCmd.SetOut(buf)

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
			Spec: v1alpha1.Dataplane{
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

		// when
		err = rootCmd.Execute()
		// then
		Expect(err).ToNot(HaveOccurred())
		// then
		var resource mesh.DataplaneResource
		err = store.Get(context.Background(), &resource, core_store.GetByKey("sample", "default"))
		Expect(err).ToNot(HaveOccurred())
		Expect(resource.Spec.Networking.Address).To(Equal("1.1.1.1"))

		// then
		Expect(buf.String()).To(Equal(
			`creationTime: "1970-01-01T00:00:00Z"
mesh: default
modificationTime: "1970-01-01T00:00:00Z"
name: sample
networking:
  address: 2.2.2.2
type: Dataplane
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
		resource := mesh.DataplaneResource{}
		err = store.Get(context.Background(), &resource, core_store.GetByKey("sample", "default"))
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
			Expect(err).To(MatchError(given.err))
		},
		Entry("no mesh", testCase{
			resource: `
type: Dataplane
name: dp-1
`,
			err: "YAML contains invalid resource: Mesh field cannot be empty",
		}),
		Entry("no name", testCase{
			resource: `
type: Dataplane
mesh: default
`,
			err: "YAML contains invalid resource: Name field cannot be empty",
		}),
		Entry("invalid data", testCase{
			resource: `
type: Dataplane
name: dp-1
mesh: default
networking:
  inbound: 0 # should be a string
`,
			err: "YAML contains invalid resource: json: cannot unmarshal number into Go value of type []json.RawMessage",
		}),
	)
})
