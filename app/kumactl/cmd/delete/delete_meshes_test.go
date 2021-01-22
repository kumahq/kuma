package delete_test

import (
	"bytes"
	"context"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/app/kumactl/cmd"
	kumactl_cmd "github.com/kumahq/kuma/app/kumactl/pkg/cmd"
	config_proto "github.com/kumahq/kuma/pkg/config/app/kumactl/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	memory_resources "github.com/kumahq/kuma/pkg/plugins/resources/memory"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
)

var _ = Describe("kumactl delete mesh", func() {

	sampleMeshes := []*mesh.MeshResource{
		{
			Meta: &test_model.ResourceMeta{
				Name: "mesh1",
			},
			Spec: &mesh_proto.Mesh{},
		},
		{
			Meta: &test_model.ResourceMeta{
				Name: "mesh2",
			},
			Spec: &mesh_proto.Mesh{},
		},
	}

	Describe("Delete Mesh", func() {

		var rootCtx *kumactl_cmd.RootContext
		var rootCmd *cobra.Command
		var outbuf, errbuf *bytes.Buffer
		var store core_store.ResourceStore

		BeforeEach(func() {
			// setup
			rootCtx = kumactl_cmd.DefaultRootContext()
			rootCtx.Runtime.NewResourceStore = func(*config_proto.ControlPlaneCoordinates_ApiServer) (core_store.ResourceStore, error) {
				return store, nil
			}

			store = core_store.NewPaginationStore(memory_resources.NewStore())

			for _, ds := range sampleMeshes {
				key := core_model.MetaToResourceKey(ds.Meta)
				err := store.Create(context.Background(), ds, core_store.CreateBy(key))
				Expect(err).ToNot(HaveOccurred())
			}

			rootCmd = cmd.NewRootCmd(rootCtx)
			outbuf = &bytes.Buffer{}
			errbuf = &bytes.Buffer{}
			rootCmd.SetOut(outbuf)
			rootCmd.SetErr(errbuf)
		})

		It("should throw an error in case of no args", func() {
			// given
			rootCmd.SetArgs([]string{
				"--config-file", filepath.Join("..", "testdata", "sample-kumactl.config.yaml"),
				"delete", "mesh"})

			// when
			err := rootCmd.Execute()

			// then
			Expect(err).To(HaveOccurred())
			// and
			Expect(err.Error()).To(Equal("accepts 2 arg(s), received 1"))
			// and
			Expect(outbuf.String()).To(MatchRegexp(`Error: accepts 2 arg\(s\), received 1`))
			// and
			Expect(errbuf.Bytes()).To(BeEmpty())
		})

		It("should throw an error in case of a non existing mesh", func() {
			// given
			rootCmd.SetArgs([]string{
				"--config-file", filepath.Join("..", "testdata", "sample-kumactl.config.yaml"),
				"delete", "mesh", "some-non-existing-mesh"})

			// when
			err := rootCmd.Execute()

			// then
			Expect(err).To(HaveOccurred())
			// and
			Expect(err.Error()).To(Equal("there is no Mesh with name \"some-non-existing-mesh\""))
			// and
			Expect(outbuf.String()).To(Equal("Error: there is no Mesh with name \"some-non-existing-mesh\"\n"))
			// and
			Expect(errbuf.Bytes()).To(BeEmpty())
		})

		It("should delete the mesh if exists", func() {
			By("running delete command")
			// given
			rootCmd.SetArgs([]string{
				"--config-file", filepath.Join("..", "testdata", "sample-kumactl.config.yaml"),
				"delete", "mesh", "mesh2"})

			// when
			err := rootCmd.Execute()
			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			// and
			Expect(errbuf.String()).To(BeEmpty())
			// and
			Expect(outbuf.String()).To(Equal("deleted Mesh \"mesh2\"\n"))

			By("verifying that resource under test was actually deleted")
			// when
			err = store.Get(context.Background(), mesh.NewMeshResource(), core_store.GetBy(core_model.ResourceKey{Name: "mesh2"}))
			// then
			Expect(core_store.IsResourceNotFound(err)).To(BeTrue())

			By("verifying that another mesh wasn't affected")
			// when
			err = store.Get(context.Background(), mesh.NewMeshResource(), core_store.GetBy(core_model.ResourceKey{Name: "mesh1"}))
			// then
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
