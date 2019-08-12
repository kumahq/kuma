package cmd

import (
	"context"
	"github.com/Kong/konvoy/components/konvoy-control-plane/api/mesh/v1alpha1"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/apis/mesh"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"

	config_proto "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/config/app/konvoyctl/v1alpha1"
	core_store "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/store"
	memory_resources "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/plugins/resources/memory"
)

var _ = Describe("konvoyctl apply", func() {

	var rootCtx *rootContext
	var rootCmd *cobra.Command
	var store core_store.ResourceStore

	BeforeEach(func() {
		rootCtx = &rootContext{
			runtime: rootRuntime{
				newResourceStore: func(controlPlane *config_proto.ControlPlane) (core_store.ResourceStore, error) {
					return store, nil
				},
			},
		}
		store = memory_resources.NewStore()
		rootCmd = newRootCmd(rootCtx)
	})

	It("should read configuration from stdin (no -f arg)", func() {
		// setup
		mockStdin, err := os.Open(filepath.Join("testdata", "apply-dataplane.yaml"))
		Expect(err).ToNot(HaveOccurred())

		oldStdin := os.Stdin
		defer func() { rootCmd.SetIn(oldStdin) }()
		rootCmd.SetIn(mockStdin)

		// given
		rootCmd.SetArgs([]string{
			"--config-file", filepath.Join("testdata", "sample-konvoyctl.config.yaml"),
			"apply",
		})

		// when
		err = rootCmd.Execute()
		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		resource := mesh.DataplaneResource{}
		err = store.Get(context.Background(), &resource, core_store.GetByKey("default", "sample", "default"))
		Expect(err).ToNot(HaveOccurred())

		// then
		Expect(resource.Spec.Tags).To(HaveKeyWithValue("service", "web"))
		Expect(resource.Spec.Tags).To(HaveKeyWithValue("version", "1.0"))
		Expect(resource.Spec.Tags).To(HaveKeyWithValue("env", "production"))
		Expect(resource.Meta.GetName()).To(Equal("sample"))
		Expect(resource.Meta.GetMesh()).To(Equal("default"))
		Expect(resource.Meta.GetNamespace()).To(Equal("default"))
	})

	It("should read configuration from stdin (-f - arg)", func() {
		// setup
		mockStdin, err := os.Open(filepath.Join("testdata", "apply-dataplane.yaml"))

		oldStdin := os.Stdin
		defer func() { rootCmd.SetIn(oldStdin) }()
		rootCmd.SetIn(mockStdin)

		// given
		rootCmd.SetArgs([]string{
			"--config-file", filepath.Join("testdata", "sample-konvoyctl.config.yaml"),
			"apply", "-f", "-",
		})

		// when
		err = rootCmd.Execute()
		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		resource := mesh.DataplaneResource{}
		err = store.Get(context.Background(), &resource, core_store.GetByKey("default", "sample", "default"))
		Expect(err).ToNot(HaveOccurred())

		// then
		Expect(resource.Spec.Tags).To(HaveKeyWithValue("service", "web"))
		Expect(resource.Spec.Tags).To(HaveKeyWithValue("version", "1.0"))
		Expect(resource.Spec.Tags).To(HaveKeyWithValue("env", "production"))
		Expect(resource.Meta.GetName()).To(Equal("sample"))
		Expect(resource.Meta.GetMesh()).To(Equal("default"))
		Expect(resource.Meta.GetNamespace()).To(Equal("default"))
	})

	It("should apply a new Dataplane resource", func() {
		// given
		rootCmd.SetArgs([]string{
			"--config-file", filepath.Join("testdata", "sample-konvoyctl.config.yaml"),
			"apply", "-f", filepath.Join("testdata", "apply-dataplane.yaml")},
		)

		// when
		err := rootCmd.Execute()
		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		resource := mesh.DataplaneResource{}
		err = store.Get(context.Background(), &resource, core_store.GetByKey("default", "sample", "default"))
		Expect(err).ToNot(HaveOccurred())

		// then
		Expect(resource.Spec.Tags).To(HaveKeyWithValue("service", "web"))
		Expect(resource.Spec.Tags).To(HaveKeyWithValue("version", "1.0"))
		Expect(resource.Spec.Tags).To(HaveKeyWithValue("env", "production"))
		Expect(resource.Meta.GetName()).To(Equal("sample"))
		Expect(resource.Meta.GetMesh()).To(Equal("default"))
		Expect(resource.Meta.GetNamespace()).To(Equal("default"))
	})

	It("should apply an updated Dataplane resource", func() {
		// setup
		newResource := mesh.DataplaneResource{
			Spec: v1alpha1.Dataplane{
				Tags: map[string]string{
					"service": "default",
					"version": "default",
					"env":     "default",
				},
			},
		}
		err := store.Create(context.Background(), &newResource, core_store.CreateByKey("default", "sample", "default"))
		Expect(err).ToNot(HaveOccurred())

		// given
		rootCmd.SetArgs([]string{
			"--config-file", filepath.Join("testdata", "sample-konvoyctl.config.yaml"),
			"apply", "-f", filepath.Join("testdata", "apply-dataplane.yaml")},
		)

		// when
		err = rootCmd.Execute()
		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		resource := mesh.DataplaneResource{}
		err = store.Get(context.Background(), &resource, core_store.GetByKey("default", "sample", "default"))
		Expect(err).ToNot(HaveOccurred())

		// then
		Expect(resource.Spec.Tags).To(HaveKeyWithValue("service", "web"))
		Expect(resource.Spec.Tags).To(HaveKeyWithValue("version", "1.0"))
		Expect(resource.Spec.Tags).To(HaveKeyWithValue("env", "production"))
		Expect(resource.Meta.GetName()).To(Equal("sample"))
		Expect(resource.Meta.GetMesh()).To(Equal("default"))
		Expect(resource.Meta.GetNamespace()).To(Equal("default"))
	})

	It("should apply a Mesh resource", func() {
		// given
		rootCmd.SetArgs([]string{
			"--config-file", filepath.Join("testdata", "sample-konvoyctl.config.yaml"),
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
})
