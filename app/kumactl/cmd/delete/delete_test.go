package delete_test

import (
	"bytes"
	"context"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/core/resources/apis/system"

	"github.com/spf13/cobra"

	"github.com/kumahq/kuma/app/kumactl/cmd"
	kumactl_cmd "github.com/kumahq/kuma/app/kumactl/pkg/cmd"
	config_proto "github.com/kumahq/kuma/pkg/config/app/kumactl/v1alpha1"
	mesh_core "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	memory_resources "github.com/kumahq/kuma/pkg/plugins/resources/memory"
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
				},
			}
			store = core_store.NewPaginationStore(memory_resources.NewStore())

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
			Expect(err.Error()).To(Equal("unknown TYPE: some-type. Allowed values: mesh, " +
				"dataplane, healthcheck, proxytemplate, traffic-log, traffic-permission, " +
				"traffic-route, traffic-trace, fault-injection, circuit-breaker, retry, secret, " +
				"zone"))
			// and
			Expect(outbuf.String()).To(MatchRegexp("unknown TYPE: some-type. " +
				"Allowed values: mesh, dataplane, healthcheck, proxytemplate, traffic-log, " +
				"traffic-permission, traffic-route, traffic-trace, fault-injection, " +
				"circuit-breaker, retry, secret, zone"))
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
					resource:        func() core_model.Resource { return mesh_core.NewDataplaneResource() },
					expectedMessage: "deleted Dataplane \"web\"\n",
				}),
				Entry("healthchecks", testCase{
					typ:             "healthcheck",
					name:            "web-to-backend",
					resource:        func() core_model.Resource { return mesh_core.NewHealthCheckResource() },
					expectedMessage: "deleted HealthCheck \"web-to-backend\"\n",
				}),
				Entry("retries", testCase{
					typ:             "retry",
					name:            "web-to-backend",
					resource:        func() core_model.Resource { return mesh_core.NewRetryResource() },
					expectedMessage: "deleted Retry \"web-to-backend\"\n",
				}),
				Entry("traffic-permissions", testCase{
					typ:             "traffic-permission",
					name:            "everyone-to-everyone",
					resource:        func() core_model.Resource { return mesh_core.NewTrafficPermissionResource() },
					expectedMessage: "deleted TrafficPermission \"everyone-to-everyone\"\n",
				}),
				Entry("traffic-logs", testCase{
					typ:             "traffic-log",
					name:            "all-requests",
					resource:        func() core_model.Resource { return mesh_core.NewTrafficLogResource() },
					expectedMessage: "deleted TrafficLog \"all-requests\"\n",
				}),
				Entry("traffic-routes", testCase{
					typ:             "traffic-route",
					name:            "web-to-backend",
					resource:        func() core_model.Resource { return mesh_core.NewTrafficRouteResource() },
					expectedMessage: "deleted TrafficRoute \"web-to-backend\"\n",
				}),
				Entry("traffic-traces", testCase{
					typ:             "traffic-trace",
					name:            "web",
					resource:        func() core_model.Resource { return mesh_core.NewTrafficTraceResource() },
					expectedMessage: "deleted TrafficTrace \"web\"\n",
				}),
				Entry("fault-injections", testCase{
					typ:             "fault-injection",
					name:            "web",
					resource:        func() core_model.Resource { return mesh_core.NewFaultInjectionResource() },
					expectedMessage: "deleted FaultInjection \"web\"\n",
				}),
				Entry("circuit-breaker", testCase{
					typ:             "circuit-breaker",
					name:            "web",
					resource:        func() core_model.Resource { return mesh_core.NewCircuitBreakerResource() },
					expectedMessage: "deleted CircuitBreaker \"web\"\n",
				}),
				Entry("secret", testCase{
					typ:             "secret",
					name:            "web",
					resource:        func() core_model.Resource { return system.NewSecretResource() },
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
					resource:        func() core_model.Resource { return mesh_core.NewDataplaneResource() },
					expectedMessage: "Error: there is no Dataplane with name \"web\"\n",
				}),
				Entry("healthchecks", testCase{
					typ:             "healthcheck",
					name:            "web-to-backend",
					resource:        func() core_model.Resource { return mesh_core.NewHealthCheckResource() },
					expectedMessage: "Error: there is no HealthCheck with name \"web-to-backend\"\n",
				}),
				Entry("retries", testCase{
					typ:             "retry",
					name:            "web-to-backend",
					resource:        func() core_model.Resource { return mesh_core.NewRetryResource() },
					expectedMessage: "Error: there is no Retry with name \"web-to-backend\"\n",
				}),
				Entry("traffic-permissions", testCase{
					typ:             "traffic-permission",
					name:            "everyone-to-everyone",
					resource:        func() core_model.Resource { return mesh_core.NewTrafficPermissionResource() },
					expectedMessage: "Error: there is no TrafficPermission with name \"everyone-to-everyone\"\n",
				}),
				Entry("traffic-logs", testCase{
					typ:             "traffic-log",
					name:            "all-requests",
					resource:        func() core_model.Resource { return mesh_core.NewTrafficLogResource() },
					expectedMessage: "Error: there is no TrafficLog with name \"all-requests\"\n",
				}),
				Entry("traffic-routes", testCase{
					typ:             "traffic-route",
					name:            "web-to-backend",
					resource:        func() core_model.Resource { return mesh_core.NewTrafficRouteResource() },
					expectedMessage: "Error: there is no TrafficRoute with name \"web-to-backend\"\n",
				}),
				Entry("traffic-traces", testCase{
					typ:             "traffic-trace",
					name:            "web",
					resource:        func() core_model.Resource { return mesh_core.NewTrafficRouteResource() },
					expectedMessage: "Error: there is no TrafficTrace with name \"web\"\n",
				}),
				Entry("fault-injections", testCase{
					typ:             "fault-injection",
					name:            "web",
					resource:        func() core_model.Resource { return mesh_core.NewFaultInjectionResource() },
					expectedMessage: "Error: there is no FaultInjection with name \"web\"\n",
				}),
				Entry("fault-injections", testCase{
					typ:             "circuit-breaker",
					name:            "web",
					resource:        func() core_model.Resource { return mesh_core.NewCircuitBreakerResource() },
					expectedMessage: "Error: there is no CircuitBreaker with name \"web\"\n",
				}),
				Entry("secret", testCase{
					typ:             "secret",
					name:            "web",
					resource:        func() core_model.Resource { return system.NewSecretResource() },
					expectedMessage: "Error: there is no Secret with name \"web\"\n",
				}),
			)
		})
	})
})
