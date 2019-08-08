package cmd

import (
	"context"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/apis/mesh"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"
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

	It("should apply a Dataplane resource", func() {
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
