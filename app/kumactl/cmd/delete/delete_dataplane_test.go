package delete_test

import (
	"bytes"
	"context"
	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/app/kumactl/cmd"
	kumactl_cmd "github.com/Kong/kuma/app/kumactl/pkg/cmd"
	config_proto "github.com/Kong/kuma/pkg/config/app/kumactl/v1alpha1"
	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/Kong/kuma/pkg/core/resources/model"
	core_store "github.com/Kong/kuma/pkg/core/resources/store"
	memory_resources "github.com/Kong/kuma/pkg/plugins/resources/memory"
	test_model "github.com/Kong/kuma/pkg/test/resources/model"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"
	"path/filepath"
)

var _ = Describe("kumactl delete dataplane", func() {
	var dataplanes []*mesh_core.DataplaneResource

	BeforeEach(func() {
		dataplanes = []*mesh_core.DataplaneResource{
			{
				Meta: &test_model.ResourceMeta{
					Namespace: "default",
					Mesh:      "Mesh1",
					Name:      "Mesh1",
				},
				Spec: mesh_proto.Dataplane{
					Networking: &mesh_proto.Dataplane_Networking{
						Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
							{
								Interface: "127.0.0.1:8080:80",
								Tags: map[string]string{
									"service": "mobile",
									"version": "v1",
								},
							},
							{
								Interface: "127.0.0.1:8090:90",
								Tags: map[string]string{
									"service": "metrics",
									"version": "v1",
								},
							},
						},
					},
				},
			},
			{
				Meta: &test_model.ResourceMeta{
					Namespace: "default",
					Mesh:      "Mesh2",
					Name:      "Mesh2",
				},
				Spec: mesh_proto.Dataplane{
					Networking: &mesh_proto.Dataplane_Networking{
						Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
							{
								Interface: "127.0.0.2:8080:80",
								Tags: map[string]string{
									"service": "web",
									"version": "v2",
								},
							},
						},
					},
				},
			},
		}
	})

	Describe("Delete dataplane", func() {
		var rootCtx *kumactl_cmd.RootContext
		var rootCmd *cobra.Command
		var outbuf, errbuf *bytes.Buffer
		var store core_store.ResourceStore

		BeforeEach(func() {
			// setup

			rootCtx = &kumactl_cmd.RootContext{
				Runtime: kumactl_cmd.RootRuntime{
					NewResourceStore: func(*config_proto.ControlPlaneCoordinates_ApiServer) (core_store.ResourceStore, error) {
						return store, nil
					},
				},
			}

			store = memory_resources.NewStore()

			for _, pt := range dataplanes {
				key := core_model.MetaToResourceKey(pt.Meta)
				err := store.Create(context.Background(), pt, core_store.CreateBy(key))
				Expect(err).ToNot(HaveOccurred())
			}

			rootCmd = cmd.NewRootCmd(rootCtx)
			outbuf = &bytes.Buffer{}
			errbuf = &bytes.Buffer{}
			rootCmd.SetOut(outbuf)
			rootCmd.SetErr(errbuf)
		})

		It("should throw an error in case of a non existing dataplane", func() {
			// given
			rootCmd.SetArgs([]string{
				"--config-file", filepath.Join("..", "testdata", "sample-kumactl.config.yaml"),
				"delete", "dataplane", "some-non-existing-dataplane"})

			// when
			err := rootCmd.Execute()

			// then
			Expect(err).To(HaveOccurred())
			// and
			Expect(err.Error()).To(Equal("there is no Dataplane with name \"some-non-existing-dataplane\""))
			// and
			Expect(outbuf.String()).To(Equal("Error: there is no Dataplane with name \"some-non-existing-dataplane\"\n"))
			// and
			Expect(errbuf.Bytes()).To(BeEmpty())
		})

		It("should delete the dataplane if exists", func() {

			// given
			rootCmd.SetArgs([]string{
				"--config-file", filepath.Join("..", "testdata", "sample-kumactl.config.yaml"),
				"delete", "dataplane", "Mesh2"})

			// when
			err := rootCmd.Execute()

			// then
			Expect(err).ToNot(HaveOccurred())

			// and
			list := &mesh_core.DataplaneResourceList{}
			err = store.List(context.Background(), list, core_store.ListByNamespace("default"))
			Expect(err).ToNot(HaveOccurred())
			Expect(list.Items).To(HaveLen(1))
			// and
			Expect(errbuf.String()).To(BeEmpty())
			// and
			Expect(outbuf.String()).To(Equal("deleted Dataplane \"Mesh2\"\n"))
		})
	})
})
