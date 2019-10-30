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

var _ = Describe("kumactl delete trafficlog", func() {

	trafficLoggingResources := []*mesh_core.TrafficLogResource{
		{
			Spec: mesh_proto.TrafficLog{
				Rules: []*mesh_proto.TrafficLog_Rule{
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
						Conf: &mesh_proto.TrafficLog_Rule_Conf{
							Backend: "file",
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
			Spec: mesh_proto.TrafficLog{
				Rules: []*mesh_proto.TrafficLog_Rule{
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
						Conf: &mesh_proto.TrafficLog_Rule_Conf{
							Backend: "logstash",
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

	Describe("delete trafficlog", func() {

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

			for _, ds := range trafficLoggingResources {
				err := store.Create(context.Background(), ds, core_store.CreateBy(core_model.MetaToResourceKey(ds.GetMeta())))
				Expect(err).ToNot(HaveOccurred())
			}

			rootCmd = cmd.NewRootCmd(rootCtx)
			outbuf = &bytes.Buffer{}
			errbuf = &bytes.Buffer{}
			rootCmd.SetOut(outbuf)
			rootCmd.SetErr(errbuf)
		})

		Describe("Delete TrafficLog", func() {

			It("should throw an error in case of a non existing trafficlog", func() {
				// given
				rootCmd.SetArgs([]string{
					"--config-file", filepath.Join("..", "testdata", "sample-kumactl.config.yaml"),
					"delete", "traffic-log", "some-non-existing-trafficlog"})

				// when
				err := rootCmd.Execute()

				// then
				Expect(err).To(HaveOccurred())
				// and
				Expect(err.Error()).To(Equal("there is no TrafficLog with name \"some-non-existing-trafficlog\""))
				// and
				Expect(outbuf.String()).To(Equal("Error: there is no TrafficLog with name \"some-non-existing-trafficlog\"\n"))
				// and
				Expect(errbuf.Bytes()).To(BeEmpty())
			})

			It("should delete the trafficlog if exists", func() {

				// given
				rootCmd.SetArgs([]string{
					"--config-file", filepath.Join("..", "testdata", "sample-kumactl.config.yaml"),
					"delete", "traffic-log", "web1-to-backend1", "--mesh", "Mesh1"})

				// when
				err := rootCmd.Execute()

				// then
				Expect(err).ToNot(HaveOccurred())

				// and
				list := &mesh_core.TrafficLogResourceList{}
				err = store.List(context.Background(), list, core_store.ListByNamespace("default"))
				Expect(err).ToNot(HaveOccurred())
				Expect(list.Items).To(HaveLen(1))
				// and
				Expect(errbuf.String()).To(BeEmpty())
				Expect(errbuf.String()).To(BeEmpty())
				// and
				Expect(outbuf.String()).To(Equal("deleted TrafficLog \"web1-to-backend1\"\n"))
			})

		})
	})

})
