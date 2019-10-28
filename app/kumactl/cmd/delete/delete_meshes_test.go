package delete_test

import (
	"bytes"
	"context"
	"path/filepath"
	"time"

	"github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/app/kumactl/cmd"
	config_proto "github.com/Kong/kuma/pkg/config/app/kumactl/v1alpha1"
	"github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/Kong/kuma/pkg/core/resources/model"
	core_store "github.com/Kong/kuma/pkg/core/resources/store"
	memory_resources "github.com/Kong/kuma/pkg/plugins/resources/memory"
	test_model "github.com/Kong/kuma/pkg/test/resources/model"

	kumactl_cmd "github.com/Kong/kuma/app/kumactl/pkg/cmd"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"
)

var _ = Describe("kumactl delete mesh", func() {

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
				Logging: &v1alpha1.Logging{
					AccessLogs: &v1alpha1.Logging_AccessLogs{
						Enabled:  true,
						FilePath: "/tmp/access.log",
					},
				},
			},
			Meta: &test_model.ResourceMeta{
				Mesh:      "mesh1",
				Name:      "mesh1",
				Namespace: "default",
			},
		},
		{
			Spec: v1alpha1.Mesh{
				Mtls: &v1alpha1.Mesh_Mtls{
					Enabled: false,
					Ca: &v1alpha1.CertificateAuthority{
						Type: &v1alpha1.CertificateAuthority_Builtin_{
							Builtin: &v1alpha1.CertificateAuthority_Builtin{},
						},
					},
				},
			},
			Meta: &test_model.ResourceMeta{
				Mesh:      "mesh2",
				Name:      "mesh2",
				Namespace: "default",
			},
		},
	}

	Describe("Delete Mesh", func() {

		var rootCtx *kumactl_cmd.RootContext
		var rootCmd *cobra.Command
		var outbuf, errbuf *bytes.Buffer
		var store core_store.ResourceStore

		BeforeEach(func() {
			// setup
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

			// given
			rootCmd.SetArgs([]string{
				"--config-file", filepath.Join("..", "testdata", "sample-kumactl.config.yaml"),
				"delete", "mesh", "mesh2"})

			// when
			err := rootCmd.Execute()

			// then
			Expect(err).ToNot(HaveOccurred())

			// and
			list := &mesh.MeshResourceList{}
			err = store.List(context.Background(), list, core_store.ListByNamespace("default"))
			Expect(err).ToNot(HaveOccurred())
			Expect(list.Items).To(HaveLen(1))
			// and
			Expect(errbuf.String()).To(BeEmpty())
			// and
			Expect(outbuf.String()).To(Equal("deleted Mesh \"mesh2\"\n"))
		})
	})

})
