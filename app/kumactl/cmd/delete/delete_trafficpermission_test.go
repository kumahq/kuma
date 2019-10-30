package delete_test

import (
	"bytes"
	"context"
	"github.com/Kong/kuma/app/kumactl/cmd"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"
	"path/filepath"
	"time"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	kumactl_cmd "github.com/Kong/kuma/app/kumactl/pkg/cmd"
	config_proto "github.com/Kong/kuma/pkg/config/app/kumactl/v1alpha1"
	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/Kong/kuma/pkg/core/resources/model"
	core_store "github.com/Kong/kuma/pkg/core/resources/store"
	memory_resources "github.com/Kong/kuma/pkg/plugins/resources/memory"

	test_model "github.com/Kong/kuma/pkg/test/resources/model"
)

var _ = Describe("kumactl delete traffic permission", func() {

	trafficPermissionResources := []*mesh_core.TrafficPermissionResource{
		{
			Spec: mesh_proto.TrafficPermission{
				Rules: []*mesh_proto.TrafficPermission_Rule{
					{
						Sources: []*mesh_proto.Selector{
							{
								Match: map[string]string{
									"service": "web1",
									"version": "1.0",
								},
							},
						},
						Destinations: []*mesh_proto.Selector{
							{
								Match: map[string]string{
									"service": "backend1",
									"env":     "dev",
								},
							},
						},
					},
				},
			},
			Meta: &test_model.ResourceMeta{
				Mesh:      "Mesh1",
				Name:      "web1-to-backend1",
				Namespace: "default",
			},
		},
		{
			Spec: mesh_proto.TrafficPermission{
				Rules: []*mesh_proto.TrafficPermission_Rule{
					{
						Sources: []*mesh_proto.Selector{
							{
								Match: map[string]string{
									"service": "web2",
									"version": "1.0",
								},
							},
						},
						Destinations: []*mesh_proto.Selector{
							{
								Match: map[string]string{
									"service": "backend2",
									"env":     "dev",
								},
							},
						},
					},
				},
			},
			Meta: &test_model.ResourceMeta{
				Mesh:      "Mesh2",
				Name:      "web2-to-backend2",
				Namespace: "default",
			},
		},
	}

	Describe("Delete TrafficPermission", func() {

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

			for _, ds := range trafficPermissionResources {
				err := store.Create(context.Background(), ds, core_store.CreateBy(core_model.MetaToResourceKey(ds.GetMeta())))
				Expect(err).ToNot(HaveOccurred())
			}

			rootCmd = cmd.NewRootCmd(rootCtx)
			outbuf = &bytes.Buffer{}
			errbuf = &bytes.Buffer{}
			rootCmd.SetOut(outbuf)
			rootCmd.SetErr(errbuf)
		})

		It("should throw an error in case of a non existing traffic permission", func() {
			// given
			rootCmd.SetArgs([]string{
				"--config-file", filepath.Join("..", "testdata", "sample-kumactl.config.yaml"),
				"delete", "traffic-permission", "some-non-existing-trafficpermission"})

			// when
			err := rootCmd.Execute()

			// then
			Expect(err).To(HaveOccurred())
			// and
			Expect(err.Error()).To(Equal("there is no TrafficPermission with name \"some-non-existing-trafficpermission\""))
			// and
			Expect(outbuf.String()).To(Equal("Error: there is no TrafficPermission with name \"some-non-existing-trafficpermission\"\n"))
			// and
			Expect(errbuf.Bytes()).To(BeEmpty())
		})

		It("should delete the traffic permission if exists", func() {

			// given
			rootCmd.SetArgs([]string{
				"--config-file", filepath.Join("..", "testdata", "sample-kumactl.config.yaml"),
				"delete", "traffic-permission", "web2-to-backend2", "--mesh", "Mesh2"})

			// when
			err := rootCmd.Execute()

			// then
			Expect(err).ToNot(HaveOccurred())

			// and
			list := &mesh_core.TrafficPermissionResourceList{}
			err = store.List(context.Background(), list, core_store.ListByNamespace("default"))
			Expect(err).ToNot(HaveOccurred())
			Expect(list.Items).To(HaveLen(1))
			// and
			Expect(errbuf.String()).To(BeEmpty())
			Expect(errbuf.String()).To(BeEmpty())
			// and
			Expect(outbuf.String()).To(Equal("deleted TrafficPermission \"web2-to-backend2\"\n"))
		})

	})
})
