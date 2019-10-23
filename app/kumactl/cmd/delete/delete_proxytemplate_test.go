package delete_test

import (
	"bytes"
	"context"
	"github.com/Kong/kuma/app/kumactl/cmd"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"
	"path/filepath"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	kumactl_cmd "github.com/Kong/kuma/app/kumactl/pkg/cmd"
	config_proto "github.com/Kong/kuma/pkg/config/app/kumactl/v1alpha1"
	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/Kong/kuma/pkg/core/resources/model"
	core_store "github.com/Kong/kuma/pkg/core/resources/store"
	memory_resources "github.com/Kong/kuma/pkg/plugins/resources/memory"

	test_model "github.com/Kong/kuma/pkg/test/resources/model"
)

var _ = Describe("kumactl delete proxytemplates", func() {

	var sampleProxyTemplates []*mesh_core.ProxyTemplateResource

	BeforeEach(func() {
		sampleProxyTemplates = []*mesh_core.ProxyTemplateResource{
			{
				Meta: &test_model.ResourceMeta{
					Namespace: "default",
					Mesh:      "Mesh1",
					Name:      "custom-template",
				},
				Spec: mesh_proto.ProxyTemplate{},
			},
			{
				Meta: &test_model.ResourceMeta{
					Namespace: "default",
					Mesh:      "Mesh2",
					Name:      "another-template",
				},
				Spec: mesh_proto.ProxyTemplate{},
			},
			{
				Meta: &test_model.ResourceMeta{
					Namespace: "default",
					Mesh:      "Mesh3",
					Name:      "simple-template",
				},
				Spec: mesh_proto.ProxyTemplate{},
			},
		}
	})

	Describe("Delete ProxyTemplate", func() {

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

			for _, pt := range sampleProxyTemplates {
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

		It("should throw an error in case of a non existing proxytemplate", func() {
			// given
			rootCmd.SetArgs([]string{
				"--config-file", filepath.Join("..", "testdata", "sample-kumactl.config.yaml"),
				"delete", "proxytemplate", "some-non-existing-proxytemplate"})

			// when
			err := rootCmd.Execute()

			// then
			Expect(err).To(HaveOccurred())
			// and
			Expect(err.Error()).To(Equal("there is no ProxyTemplate with name \"some-non-existing-proxytemplate\""))
			// and
			Expect(outbuf.String()).To(Equal("Error: there is no ProxyTemplate with name \"some-non-existing-proxytemplate\"\n"))
			// and
			Expect(errbuf.Bytes()).To(BeEmpty())
		})

		It("should delete the proxytemplate if exists", func() {

			// given
			rootCmd.SetArgs([]string{
				"--config-file", filepath.Join("..", "testdata", "sample-kumactl.config.yaml"),
				"delete", "proxytemplate", "custom-template", "--mesh", "Mesh1"})

			// when
			err := rootCmd.Execute()

			// then
			Expect(err).ToNot(HaveOccurred())

			// and
			list := &mesh_core.ProxyTemplateResourceList{}
			err = store.List(context.Background(), list, core_store.ListByNamespace("default"))
			Expect(err).ToNot(HaveOccurred())
			Expect(list.Items).To(HaveLen(2))
			// and
			Expect(errbuf.String()).To(BeEmpty())
			Expect(errbuf.String()).To(BeEmpty())
			// and
			Expect(outbuf.String()).To(Equal("deleted ProxyTemplate \"custom-template\"\n"))
		})

	})
})
