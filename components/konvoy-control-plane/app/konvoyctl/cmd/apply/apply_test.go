package apply_test

import (
	"context"
	"os"
	"path/filepath"

	"github.com/Kong/konvoy/components/konvoy-control-plane/api/mesh/v1alpha1"
	"github.com/Kong/konvoy/components/konvoy-control-plane/app/konvoyctl/cmd"
	konvoyctl_cmd "github.com/Kong/konvoy/components/konvoy-control-plane/app/konvoyctl/pkg/cmd"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/apis/mesh"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"

	config_proto "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/config/app/konvoyctl/v1alpha1"
	core_store "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/store"
	memory_resources "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/plugins/resources/memory"
)

var _ = Describe("konvoyctl apply", func() {

	var rootCtx *konvoyctl_cmd.RootContext
	var rootCmd *cobra.Command
	var store core_store.ResourceStore

	BeforeEach(func() {
		rootCtx = &konvoyctl_cmd.RootContext{
			Runtime: konvoyctl_cmd.RootRuntime{
				NewResourceStore: func(controlPlane *config_proto.ControlPlane) (core_store.ResourceStore, error) {
					return store, nil
				},
			},
		}
		store = memory_resources.NewStore()
		rootCmd = cmd.NewRootCmd(rootCtx)
	})

	ValidatePersistedResource := func() error {
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

		return nil
	}

	It("should read configuration from stdin (no -f arg)", func() {
		// setup
		mockStdin, err := os.Open(filepath.Join("testdata", "apply-dataplane.yaml"))
		Expect(err).ToNot(HaveOccurred())

		// given
		rootCmd.SetArgs([]string{
			"--config-file", filepath.Join("..", "testdata", "sample-konvoyctl.config.yaml"),
			"apply",
		})
		rootCmd.SetIn(mockStdin)

		// when
		err = rootCmd.Execute()
		// then
		Expect(err).ToNot(HaveOccurred())

		// and
		Expect(ValidatePersistedResource()).To(Succeed())
	})

	It("should read configuration from stdin (-f - arg)", func() {
		// setup
		mockStdin, err := os.Open(filepath.Join("testdata", "apply-dataplane.yaml"))
		Expect(err).ToNot(HaveOccurred())

		// given
		rootCmd.SetArgs([]string{
			"--config-file", filepath.Join("..", "testdata", "sample-konvoyctl.config.yaml"),
			"apply", "-f", "-",
		})
		rootCmd.SetIn(mockStdin)

		// when
		err = rootCmd.Execute()
		// then
		Expect(err).ToNot(HaveOccurred())

		// and
		Expect(ValidatePersistedResource()).To(Succeed())
	})

	It("should apply a new Dataplane resource", func() {
		// given
		rootCmd.SetArgs([]string{
			"--config-file", filepath.Join("..", "testdata", "sample-konvoyctl.config.yaml"),
			"apply", "-f", filepath.Join("testdata", "apply-dataplane.yaml")},
		)

		// when
		err := rootCmd.Execute()
		// then
		Expect(err).ToNot(HaveOccurred())

		// and
		Expect(ValidatePersistedResource()).To(Succeed())
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
			"--config-file", filepath.Join("..", "testdata", "sample-konvoyctl.config.yaml"),
			"apply", "-f", filepath.Join("testdata", "apply-dataplane.yaml")},
		)

		// when
		err = rootCmd.Execute()
		// then
		Expect(err).ToNot(HaveOccurred())

		// and
		Expect(ValidatePersistedResource()).To(Succeed())
	})

	It("should apply a Mesh resource", func() {
		// given
		rootCmd.SetArgs([]string{
			"--config-file", filepath.Join("..", "testdata", "sample-konvoyctl.config.yaml"),
			"apply", "-f", filepath.Join("testdata", "apply-mesh.yaml")},
		)

		// when
		err := rootCmd.Execute()
		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		resource := mesh.MeshResource{}
		err = store.Get(context.Background(), &resource, core_store.GetByKey("default", "sample", "sample"))
		Expect(err).ToNot(HaveOccurred())

		// then
		Expect(resource.Meta.GetName()).To(Equal("sample"))
		Expect(resource.Meta.GetMesh()).To(Equal("sample"))
		Expect(resource.Meta.GetNamespace()).To(Equal("default"))
	})

	It("should fill in template (single variable)", func() {
		// given
		rootCmd.SetArgs([]string{
			"--config-file", filepath.Join("..", "testdata", "sample-konvoyctl.config.yaml"),
			"apply", "-f", filepath.Join("testdata", "apply-mesh-template-repeated-placeholder.yaml"),
			"-a", "name=meshinit"},
		)

		// when
		err := rootCmd.Execute()
		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		resource := mesh.MeshResource{}
		err = store.Get(context.Background(), &resource, core_store.GetByKey("default", "meshinit", "meshinit"))
		Expect(err).ToNot(HaveOccurred())

		// then
		Expect(resource.Meta.GetName()).To(Equal("meshinit"))
		Expect(resource.Meta.GetMesh()).To(Equal("meshinit"))
		Expect(resource.Meta.GetNamespace()).To(Equal("default"))
	})

	It("should fill in template (multiple variables)", func() {
		// given
		rootCmd.SetArgs([]string{
			"--config-file", filepath.Join("..", "testdata", "sample-konvoyctl.config.yaml"),
			"apply", "-f", filepath.Join("testdata", "apply-mesh-template.yaml"),
			"-a", "name=meshinit", "-a", "mesh=meshinit", "-a", "type=Mesh"},
		)

		// when
		err := rootCmd.Execute()
		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		resource := mesh.MeshResource{}
		err = store.Get(context.Background(), &resource, core_store.GetByKey("default", "meshinit", "meshinit"))
		Expect(err).ToNot(HaveOccurred())

		// then
		Expect(resource.Meta.GetName()).To(Equal("meshinit"))
		Expect(resource.Meta.GetMesh()).To(Equal("meshinit"))
		Expect(resource.Meta.GetNamespace()).To(Equal("default"))
	})
})
