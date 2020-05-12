package delete_test

import (
	"bytes"
	"context"
	"path/filepath"
	"time"

	"github.com/Kong/kuma/pkg/catalog"
	catalog_client "github.com/Kong/kuma/pkg/catalog/client"
	"github.com/Kong/kuma/pkg/core/resources/apis/system"
	test_catalog "github.com/Kong/kuma/pkg/test/catalog"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/spf13/cobra"

	"github.com/Kong/kuma/app/kumactl/cmd"
	kumactl_cmd "github.com/Kong/kuma/app/kumactl/pkg/cmd"
	config_proto "github.com/Kong/kuma/pkg/config/app/kumactl/v1alpha1"
	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/Kong/kuma/pkg/core/resources/model"
	core_store "github.com/Kong/kuma/pkg/core/resources/store"
	memory_resources "github.com/Kong/kuma/pkg/plugins/resources/memory"
)

var _ = Describe("kumactl delete ", func() {
	Describe("Delete Command", func() {
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
					NewAdminResourceStore: func(string, *config_proto.Context_AdminApiCredentials) (core_store.ResourceStore, error) {
						return store, nil
					},
					NewCatalogClient: func(s string) (catalog_client.CatalogClient, error) {
						return &test_catalog.StaticCatalogClient{
							Resp: catalog.Catalog{
								Apis: catalog.Apis{
									DataplaneToken: catalog.DataplaneTokenApi{
										LocalUrl: "http://localhost:1234",
									},
								},
							},
						}, nil
					},
				},
			}
			store = memory_resources.NewStore()

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
				"delete"})

			// when
			err := rootCmd.Execute()

			// then
			Expect(err).To(HaveOccurred())
			// and
			Expect(err.Error()).To(Equal("accepts 2 arg(s), received 0"))
			// and
			Expect(outbuf.String()).To(MatchRegexp(`Error: accepts 2 arg\(s\), received 0`))
			// and
			Expect(errbuf.Bytes()).To(BeEmpty())
		})

		It("should throw an error in case of unsupported resource type", func() {
			// given
			rootCmd.SetArgs([]string{
				"--config-file", filepath.Join("..", "testdata", "sample-kumactl.config.yaml"),
				"delete", "some-type", "some-name"})

			// when
			err := rootCmd.Execute()

			// then
			Expect(err).To(HaveOccurred())
			// and
			Expect(err.Error()).To(Equal("unknown TYPE: some-type. Allowed values: mesh, dataplane, healthcheck, proxytemplate, traffic-log, traffic-permission, traffic-route, traffic-trace, fault-injection, secret"))
			// and
			Expect(outbuf.String()).To(MatchRegexp(`unknown TYPE: some-type. Allowed values: mesh, dataplane, healthcheck, proxytemplate, traffic-log, traffic-permission, traffic-route, traffic-trace, fault-injection, secret`))
			// and
			Expect(errbuf.Bytes()).To(BeEmpty())
		})

		Describe("kumactl delete TYPE NAME", func() {

			type testCase struct {
				typ             string // TYPE
				name            string // NAME
				resource        func() core_model.Resource
				expectedMessage string // output
			}

			DescribeTable("should succeed if resource exists",
				func(given testCase) {
					key1 := core_model.ResourceKey{Mesh: "demo", Name: given.name}    // resource under test
					key2 := core_model.ResourceKey{Mesh: "default", Name: given.name} // resource with the same name but in a differrent mesh
					key3 := core_model.ResourceKey{Mesh: "demo", Name: "example"}     // another resource in the same mesh

					By("creating resources necessary for the test")
					// setup
					for _, key := range []core_model.ResourceKey{key1, key2, key3} {
						// when
						err := store.Create(context.Background(), given.resource(), core_store.CreateBy(key))
						// then
						Expect(err).ToNot(HaveOccurred())
					}

					By("running delete command")
					// given
					rootCmd.SetArgs([]string{
						"--config-file", filepath.Join("..", "testdata", "sample-kumactl.config.yaml"),
						"delete", given.typ, given.name, "--mesh", "demo"})

					// when
					err := rootCmd.Execute()
					// then
					Expect(err).ToNot(HaveOccurred())
					// and
					Expect(errbuf.String()).To(BeEmpty())
					Expect(outbuf.String()).To(Equal(given.expectedMessage))

					By("verifying that resource under test was actually deleted")
					// when
					err = store.Get(context.Background(), given.resource(), core_store.GetBy(key1))
					// then
					Expect(core_store.IsResourceNotFound(err)).To(BeTrue())

					By("verifying that resource with the same name but in a differrent mesh wasn't affected")
					// when
					err = store.Get(context.Background(), given.resource(), core_store.GetBy(key2))
					// then
					Expect(err).ToNot(HaveOccurred())

					By("verifying that another resource in the same mesh wasn't affected")
					// when
					err = store.Get(context.Background(), given.resource(), core_store.GetBy(key3))
					// then
					Expect(err).ToNot(HaveOccurred())
				},
				Entry("dataplanes", testCase{
					typ:             "dataplane",
					name:            "web",
					resource:        func() core_model.Resource { return &mesh_core.DataplaneResource{} },
					expectedMessage: "deleted Dataplane \"web\"\n",
				}),
				Entry("healthchecks", testCase{
					typ:             "healthcheck",
					name:            "web-to-backend",
					resource:        func() core_model.Resource { return &mesh_core.HealthCheckResource{} },
					expectedMessage: "deleted HealthCheck \"web-to-backend\"\n",
				}),
				Entry("traffic-permissions", testCase{
					typ:             "traffic-permission",
					name:            "everyone-to-everyone",
					resource:        func() core_model.Resource { return &mesh_core.TrafficPermissionResource{} },
					expectedMessage: "deleted TrafficPermission \"everyone-to-everyone\"\n",
				}),
				Entry("traffic-logs", testCase{
					typ:             "traffic-log",
					name:            "all-requests",
					resource:        func() core_model.Resource { return &mesh_core.TrafficLogResource{} },
					expectedMessage: "deleted TrafficLog \"all-requests\"\n",
				}),
				Entry("traffic-routes", testCase{
					typ:             "traffic-route",
					name:            "web-to-backend",
					resource:        func() core_model.Resource { return &mesh_core.TrafficRouteResource{} },
					expectedMessage: "deleted TrafficRoute \"web-to-backend\"\n",
				}),
				Entry("traffic-traces", testCase{
					typ:             "traffic-trace",
					name:            "web",
					resource:        func() core_model.Resource { return &mesh_core.TrafficTraceResource{} },
					expectedMessage: "deleted TrafficTrace \"web\"\n",
				}),
				Entry("fault-injections", testCase{
					typ:             "fault-injection",
					name:            "web",
					resource:        func() core_model.Resource { return &mesh_core.FaultInjectionResource{} },
					expectedMessage: "deleted FaultInjection \"web\"\n",
				}),
				Entry("secret", testCase{
					typ:             "secret",
					name:            "web",
					resource:        func() core_model.Resource { return &system.SecretResource{} },
					expectedMessage: "deleted Secret \"web\"\n",
				}),
			)

			DescribeTable("should fail if resource doesn't exist",
				func(given testCase) {
					By("running delete command")
					// given
					rootCmd.SetArgs([]string{
						"--config-file", filepath.Join("..", "testdata", "sample-kumactl.config.yaml"),
						"delete", given.typ, given.name})

					// when
					err := rootCmd.Execute()
					// then
					Expect(err).To(HaveOccurred())
					// and
					Expect(outbuf.String()).To(Equal(given.expectedMessage))
					// and
					Expect(errbuf.Bytes()).To(BeEmpty())
				},
				Entry("dataplanes", testCase{
					typ:             "dataplane",
					name:            "web",
					resource:        func() core_model.Resource { return &mesh_core.DataplaneResource{} },
					expectedMessage: "Error: there is no Dataplane with name \"web\"\n",
				}),
				Entry("healthchecks", testCase{
					typ:             "healthcheck",
					name:            "web-to-backend",
					resource:        func() core_model.Resource { return &mesh_core.HealthCheckResource{} },
					expectedMessage: "Error: there is no HealthCheck with name \"web-to-backend\"\n",
				}),
				Entry("traffic-permissions", testCase{
					typ:             "traffic-permission",
					name:            "everyone-to-everyone",
					resource:        func() core_model.Resource { return &mesh_core.TrafficPermissionResource{} },
					expectedMessage: "Error: there is no TrafficPermission with name \"everyone-to-everyone\"\n",
				}),
				Entry("traffic-logs", testCase{
					typ:             "traffic-log",
					name:            "all-requests",
					resource:        func() core_model.Resource { return &mesh_core.TrafficLogResource{} },
					expectedMessage: "Error: there is no TrafficLog with name \"all-requests\"\n",
				}),
				Entry("traffic-routes", testCase{
					typ:             "traffic-route",
					name:            "web-to-backend",
					resource:        func() core_model.Resource { return &mesh_core.TrafficRouteResource{} },
					expectedMessage: "Error: there is no TrafficRoute with name \"web-to-backend\"\n",
				}),
				Entry("traffic-traces", testCase{
					typ:             "traffic-trace",
					name:            "web",
					resource:        func() core_model.Resource { return &mesh_core.TrafficRouteResource{} },
					expectedMessage: "Error: there is no TrafficTrace with name \"web\"\n",
				}),
				Entry("fault-injections", testCase{
					typ:             "fault-injection",
					name:            "web",
					resource:        func() core_model.Resource { return &mesh_core.FaultInjectionResource{} },
					expectedMessage: "Error: there is no FaultInjection with name \"web\"\n",
				}),
				Entry("secret", testCase{
					typ:             "secret",
					name:            "web",
					resource:        func() core_model.Resource { return &system.SecretResource{} },
					expectedMessage: "Error: there is no Secret with name \"web\"\n",
				}),
			)
		})
	})
})
