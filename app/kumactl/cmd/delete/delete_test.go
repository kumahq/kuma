package delete_test

import (
	"bytes"
	"context"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"

	"github.com/kumahq/kuma/app/kumactl/cmd"
	kumactl_cmd "github.com/kumahq/kuma/app/kumactl/pkg/cmd"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	memory_resources "github.com/kumahq/kuma/pkg/plugins/resources/memory"
	util_http "github.com/kumahq/kuma/pkg/util/http"
	"github.com/kumahq/kuma/pkg/util/test"
)

var _ = Describe("kumactl delete ", func() {
	Describe("Delete Command", func() {
		var rootCtx *kumactl_cmd.RootContext
		var rootCmd *cobra.Command
		var outbuf *bytes.Buffer
		var store core_store.ResourceStore

		BeforeEach(func() {
			// setup
			rootCtx = kumactl_cmd.DefaultRootContext()
			rootCtx.Runtime.NewAPIServerClient = test.GetMockNewAPIServerClient()
			rootCtx.Runtime.NewResourceStore = func(util_http.Client) core_store.ResourceStore {
				return store
			}
			store = core_store.NewPaginationStore(memory_resources.NewStore())

			rootCmd = cmd.NewRootCmd(rootCtx)

			// Different versions of cobra might emit errors to stdout
			// or stderr. It's too fragile to depend on precidely what
			// it does, and that's not something that needs to be tested
			// within Kuma anyway. So we just combine all the output
			// and validate the aggregate.
			outbuf = &bytes.Buffer{}
			rootCmd.SetOut(outbuf)
			rootCmd.SetErr(outbuf)
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
			Expect(err.Error()).To(ContainSubstring("unknown TYPE: some-type. Allowed values:"))
			// and
			Expect(outbuf.String()).To(ContainSubstring("unknown TYPE: some-type. Allowed values:"))
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
				Entry("circuit-breaker", testCase{
					typ:             "circuit-breaker",
					name:            "web",
					resource:        func() core_model.Resource { return core_mesh.NewCircuitBreakerResource() },
					expectedMessage: "deleted CircuitBreaker \"web\"\n",
				}),
				Entry("dataplanes", testCase{
					typ:             "dataplane",
					name:            "web",
					resource:        func() core_model.Resource { return core_mesh.NewDataplaneResource() },
					expectedMessage: "deleted Dataplane \"web\"\n",
				}),
				Entry("external-services", testCase{
					typ:             "external-service",
					name:            "httpbin",
					resource:        func() core_model.Resource { return core_mesh.NewExternalServiceResource() },
					expectedMessage: "deleted ExternalService \"httpbin\"\n",
				}),
				Entry("fault-injections", testCase{
					typ:             "fault-injection",
					name:            "web",
					resource:        func() core_model.Resource { return core_mesh.NewFaultInjectionResource() },
					expectedMessage: "deleted FaultInjection \"web\"\n",
				}),
				Entry("healthchecks", testCase{
					typ:             "healthcheck",
					name:            "web-to-backend",
					resource:        func() core_model.Resource { return core_mesh.NewHealthCheckResource() },
					expectedMessage: "deleted HealthCheck \"web-to-backend\"\n",
				}),
				Entry("proxytemplate", testCase{
					typ:             "proxytemplate",
					name:            "test-pt",
					resource:        func() core_model.Resource { return core_mesh.NewProxyTemplateResource() },
					expectedMessage: "deleted ProxyTemplate \"test-pt\"\n",
				}),
				Entry("retries", testCase{
					typ:             "retry",
					name:            "web-to-backend",
					resource:        func() core_model.Resource { return core_mesh.NewRetryResource() },
					expectedMessage: "deleted Retry \"web-to-backend\"\n",
				}),
				Entry("rate-limits", testCase{
					typ:             "rate-limit",
					name:            "100-rps",
					resource:        func() core_model.Resource { return core_mesh.NewRateLimitResource() },
					expectedMessage: "deleted RateLimit \"100-rps\"\n",
				}),
				Entry("timeouts", testCase{
					typ:             "timeout",
					name:            "web",
					resource:        func() core_model.Resource { return core_mesh.NewTimeoutResource() },
					expectedMessage: "deleted Timeout \"web\"\n",
				}),
				Entry("traffic-logs", testCase{
					typ:             "traffic-log",
					name:            "all-requests",
					resource:        func() core_model.Resource { return core_mesh.NewTrafficLogResource() },
					expectedMessage: "deleted TrafficLog \"all-requests\"\n",
				}),
				Entry("traffic-permissions", testCase{
					typ:             "traffic-permission",
					name:            "everyone-to-everyone",
					resource:        func() core_model.Resource { return core_mesh.NewTrafficPermissionResource() },
					expectedMessage: "deleted TrafficPermission \"everyone-to-everyone\"\n",
				}),
				Entry("traffic-routes", testCase{
					typ:             "traffic-route",
					name:            "web-to-backend",
					resource:        func() core_model.Resource { return core_mesh.NewTrafficRouteResource() },
					expectedMessage: "deleted TrafficRoute \"web-to-backend\"\n",
				}),
				Entry("traffic-traces", testCase{
					typ:             "traffic-trace",
					name:            "web",
					resource:        func() core_model.Resource { return core_mesh.NewTrafficTraceResource() },
					expectedMessage: "deleted TrafficTrace \"web\"\n",
				}),
				Entry("secrets", testCase{
					typ:             "secret",
					name:            "web",
					resource:        func() core_model.Resource { return system.NewSecretResource() },
					expectedMessage: "deleted Secret \"web\"\n",
				}),
			)

			DescribeTable("should succeed if resource exists",
				func(given testCase) {
					key := core_model.ResourceKey{Name: given.name}

					By("creating resources necessary for the test")
					// setup

					// when
					err := store.Create(context.Background(), given.resource(), core_store.CreateBy(key))
					// then
					Expect(err).ToNot(HaveOccurred())

					By("running delete command")
					// given
					rootCmd.SetArgs([]string{
						"--config-file", filepath.Join("..", "testdata", "sample-kumactl.config.yaml"),
						"delete", given.typ, given.name})

					// when
					err = rootCmd.Execute()
					// then
					Expect(err).ToNot(HaveOccurred())
					// and
					Expect(outbuf.String()).To(Equal(given.expectedMessage))

					By("verifying that resource under test was actually deleted")
					// when
					err = store.Get(context.Background(), given.resource(), core_store.GetBy(key))
					// then
					Expect(core_store.IsResourceNotFound(err)).To(BeTrue())
				},
				Entry("meshes", testCase{
					typ:             "mesh",
					name:            "test-mesh",
					resource:        func() core_model.Resource { return core_mesh.NewMeshResource() },
					expectedMessage: "deleted Mesh \"test-mesh\"\n",
				}),
				Entry("global-secrets", testCase{
					typ:             "global-secret",
					name:            "test-secret",
					resource:        func() core_model.Resource { return system.NewGlobalSecretResource() },
					expectedMessage: "deleted GlobalSecret \"test-secret\"\n",
				}),
				Entry("zones", testCase{
					typ:             "zone",
					name:            "eu-north",
					resource:        func() core_model.Resource { return system.NewZoneResource() },
					expectedMessage: "deleted Zone \"eu-north\"\n",
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
				},
				Entry("dataplanes", testCase{
					typ:             "dataplane",
					name:            "web",
					resource:        func() core_model.Resource { return core_mesh.NewDataplaneResource() },
					expectedMessage: "Error: there is no Dataplane with name \"web\"\n",
				}),
				Entry("healthchecks", testCase{
					typ:             "healthcheck",
					name:            "web-to-backend",
					resource:        func() core_model.Resource { return core_mesh.NewHealthCheckResource() },
					expectedMessage: "Error: there is no HealthCheck with name \"web-to-backend\"\n",
				}),
				Entry("retries", testCase{
					typ:             "retry",
					name:            "web-to-backend",
					resource:        func() core_model.Resource { return core_mesh.NewRetryResource() },
					expectedMessage: "Error: there is no Retry with name \"web-to-backend\"\n",
				}),
				Entry("traffic-permissions", testCase{
					typ:             "traffic-permission",
					name:            "everyone-to-everyone",
					resource:        func() core_model.Resource { return core_mesh.NewTrafficPermissionResource() },
					expectedMessage: "Error: there is no TrafficPermission with name \"everyone-to-everyone\"\n",
				}),
				Entry("traffic-logs", testCase{
					typ:             "traffic-log",
					name:            "all-requests",
					resource:        func() core_model.Resource { return core_mesh.NewTrafficLogResource() },
					expectedMessage: "Error: there is no TrafficLog with name \"all-requests\"\n",
				}),
				Entry("traffic-routes", testCase{
					typ:             "traffic-route",
					name:            "web-to-backend",
					resource:        func() core_model.Resource { return core_mesh.NewTrafficRouteResource() },
					expectedMessage: "Error: there is no TrafficRoute with name \"web-to-backend\"\n",
				}),
				Entry("traffic-traces", testCase{
					typ:             "traffic-trace",
					name:            "web",
					resource:        func() core_model.Resource { return core_mesh.NewTrafficRouteResource() },
					expectedMessage: "Error: there is no TrafficTrace with name \"web\"\n",
				}),
				Entry("fault-injections", testCase{
					typ:             "fault-injection",
					name:            "web",
					resource:        func() core_model.Resource { return core_mesh.NewFaultInjectionResource() },
					expectedMessage: "Error: there is no FaultInjection with name \"web\"\n",
				}),
				Entry("fault-injections", testCase{
					typ:             "circuit-breaker",
					name:            "web",
					resource:        func() core_model.Resource { return core_mesh.NewCircuitBreakerResource() },
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
