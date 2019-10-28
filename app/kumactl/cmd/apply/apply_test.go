package apply_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/app/kumactl/cmd"
	kumactl_cmd "github.com/Kong/kuma/app/kumactl/pkg/cmd"
	"github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"

	config_proto "github.com/Kong/kuma/pkg/config/app/kumactl/v1alpha1"
	core_store "github.com/Kong/kuma/pkg/core/resources/store"
	memory_resources "github.com/Kong/kuma/pkg/plugins/resources/memory"
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
			},
		}
		store = memory_resources.NewStore()
		rootCmd = cmd.NewRootCmd(rootCtx)
	})

	ValidatePersistedResource := func() {
		resource := mesh.DataplaneResource{}
		err := store.Get(context.Background(), &resource, core_store.GetByKey("default", "sample", "default"))
		Expect(err).ToNot(HaveOccurred())

		// then
		Expect(resource.Meta.GetName()).To(Equal("sample"))
		Expect(resource.Meta.GetMesh()).To(Equal("default"))
		Expect(resource.Meta.GetNamespace()).To(Equal("default"))
		// and
		Expect(resource.Spec.Networking.Inbound).To(HaveLen(1))
		Expect(resource.Spec.Networking.Inbound[0].Interface).To(Equal("1.1.1.1:80:8080"))
		Expect(resource.Spec.Networking.Inbound[0].Tags).To(HaveKeyWithValue("service", "web"))
		Expect(resource.Spec.Networking.Inbound[0].Tags).To(HaveKeyWithValue("version", "1.0"))
		Expect(resource.Spec.Networking.Inbound[0].Tags).To(HaveKeyWithValue("env", "production"))
		// and
		Expect(resource.Spec.Networking.Outbound).To(HaveLen(1))
		Expect(resource.Spec.Networking.Outbound[0].Interface).To(Equal(":30000"))
		Expect(resource.Spec.Networking.Outbound[0].Service).To(Equal("postgres"))
		Expect(resource.Spec.Networking.Outbound[0].ServicePort).To(Equal(uint32(5432)))
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
					Inbound: []*v1alpha1.Dataplane_Networking_Inbound{
						{
							Interface: "8.8.8.8:443:8443",
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
		err := store.Create(context.Background(), &newResource, core_store.CreateByKey("default", "sample", "default"))
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
		err = store.Get(context.Background(), &resource, core_store.GetByKey("default", "sample", ""))
		Expect(err).ToNot(HaveOccurred())

		// then
		Expect(resource.Meta.GetName()).To(Equal("sample"))
		Expect(resource.Meta.GetMesh()).To(Equal(""))
		Expect(resource.Meta.GetNamespace()).To(Equal("default"))
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
		err = store.Get(context.Background(), &resource, core_store.GetByKey("default", "meshinit", ""))
		Expect(err).ToNot(HaveOccurred())

		// then
		Expect(resource.Meta.GetName()).To(Equal("meshinit"))
		Expect(resource.Meta.GetMesh()).To(Equal(""))
		Expect(resource.Meta.GetNamespace()).To(Equal("default"))
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
