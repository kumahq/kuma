package export_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"

	api_types "github.com/kumahq/kuma/api/openapi/types"
	"github.com/kumahq/kuma/app/kumactl/cmd"
	"github.com/kumahq/kuma/app/kumactl/pkg/client"
	"github.com/kumahq/kuma/app/kumactl/pkg/resources"
	test_kumactl "github.com/kumahq/kuma/app/kumactl/pkg/test"
	"github.com/kumahq/kuma/pkg/api-server/mappers"
	core_system "github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	memory_resources "github.com/kumahq/kuma/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/pkg/test/matchers"
	"github.com/kumahq/kuma/pkg/test/resources/samples"
	util_http "github.com/kumahq/kuma/pkg/util/http"
)

var _ = Describe("kumactl export", func() {
	var rootCmd *cobra.Command
	var store core_store.ResourceStore
	var buf *bytes.Buffer

	rootTime, _ := time.Parse(time.RFC3339, "2008-04-27T16:05:36.995Z")

	BeforeEach(func() {
		store = core_store.NewPaginationStore(memory_resources.NewStore())
		defs := registry.Global().ObjectDescriptors()
		rootCtx, err := test_kumactl.MakeRootContext(rootTime, store, defs...)
		Expect(err).ToNot(HaveOccurred())
		res := test_kumactl.DummyIndexResponse
		res.Version = "0.0.0-testversion"
		rootCtx.Runtime.NewAPIServerClient = func(client util_http.Client) resources.ApiServerClient {
			return resources.NewStaticApiServiceClient(res)
		}
		rootCtx.Runtime.NewKubernetesResourcesClient = func(client util_http.Client) client.KubernetesResourcesClient {
			return fileBasedKubernetesResourcesClient{}
		}
		rootCtx.Runtime.NewResourcesListClient = func(u util_http.Client) client.ResourcesListClient {
			return staticResourcesListClient{}
		}
		rootCtx.InstallCpContext.Args.Namespace = "kuma-system"

		rootCmd = cmd.NewRootCmd(rootCtx)
		buf = &bytes.Buffer{}
		rootCmd.SetOut(buf)
	})

	type testCase struct {
		args       []string
		goldenFile string
		resources  []model.Resource
	}
	DescribeTable("should succeed",
		func(given testCase) {
			for _, res := range given.resources {
				err := store.Create(context.Background(), res, core_store.CreateByKey(res.GetMeta().GetName(), res.GetMeta().GetMesh()))
				Expect(err).ToNot(HaveOccurred())
			}

			args := append([]string{
				"--config-file",
				filepath.Join("..", "testdata", "sample-kumactl.config.yaml"),
				"export",
			}, given.args...)
			rootCmd.SetArgs(args)

			// when
			err := rootCmd.Execute()

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(buf.String()).To(matchers.MatchGoldenEqual("testdata", given.goldenFile))
		},
		Entry("no args", testCase{
			resources: []model.Resource{
				samples.MeshDefault(),
				samples.SampleSigningKeyGlobalSecret(),
				samples.SampleSigningKeySecret(),
				samples.MeshDefaultBuilder().WithName("another-mesh").Build(),
				samples.SampleSigningKeySecretBuilder().WithMesh("another-mesh").Build(),
				samples.ServiceInsight().WithMesh("another-mesh").Build(),
				samples.SampleGlobalSecretAdminCa(),
			},
			goldenFile: "export.golden.yaml",
		}),
		Entry("kubernetes profile=all", testCase{
			resources: []model.Resource{
				samples.MeshDefault(),
				samples.SampleSigningKeyGlobalSecret(),
				samples.MeshAccessLogWithFileBackend(),
				samples.Retry(),
				samples.MeshTimeoutInCustomNamespace(),
				samples.MeshDefaultBuilder().WithName("mesh-with-mtls").WithBuiltinMTLSBackend("ca-1").Build(),
				samples.SampleSecretBuilder().WithMesh("mesh-with-mtls").WithName(core_system.BuiltinCertSecretName("mesh-with-mtls", "ca-1")).Build(),
				samples.SampleSecretBuilder().WithMesh("mesh-with-mtls").WithName(core_system.BuiltinKeySecretName("mesh-with-mtls", "ca-1")).Build(),
			},
			args: []string{
				"--format=kubernetes",
				"--profile", "all",
				"-a",
			},
			goldenFile: "export-kube.golden.yaml",
		}),
		Entry("universal profile=all", testCase{
			resources: []model.Resource{
				samples.MeshDefault(),
				samples.SampleSigningKeyGlobalSecret(),
				samples.MeshAccessLogWithFileBackend(),
				samples.Retry(),
				samples.MeshTimeoutInCustomNamespace(),
				samples.MeshDefaultBuilder().WithName("mesh-with-mtls").WithBuiltinMTLSBackend("ca-1").Build(),
				samples.SampleSecretBuilder().WithMesh("mesh-with-mtls").WithName(core_system.BuiltinCertSecretName("mesh-with-mtls", "ca-1")).Build(),
				samples.SampleSecretBuilder().WithMesh("mesh-with-mtls").WithName(core_system.BuiltinKeySecretName("mesh-with-mtls", "ca-1")).Build(),
			},
			args: []string{
				"--format=universal",
				"--profile", "all",
				"-a",
			},
			goldenFile: "export-uni.golden.yaml",
		}),
		Entry("kubernetes profile=no-dataplanes without secrets", testCase{
			resources: []model.Resource{
				samples.MeshDefault(),
				samples.SampleSigningKeyGlobalSecret(),
				samples.MeshAccessLogWithFileBackend(),
				samples.Retry(),
				samples.DataplaneBackend(),
			},
			args: []string{
				"--profile", "no-dataplanes",
			},
			goldenFile: "export-no-dp.golden.yaml",
		}),
		Entry("kubernetes profile=federation-with-policies but no namespaced", testCase{
			resources: []model.Resource{
				samples.MeshDefault(),
				samples.SampleSigningKeyGlobalSecret(),
				samples.MeshAccessLogWithFileBackend(),
				samples.Retry(),
				samples.DataplaneBackend(),
				samples.MeshTimeoutInCustomNamespace(),
			},
			args: []string{
				"--profile", "federation-with-policies",
				"--format=kubernetes",
			},
			goldenFile: "export-federation-with-policies.golden.yaml",
		}),
	)

	type invalidTestCase struct {
		args []string
		err  string
	}
	DescribeTable("should fail on invalid resource",
		func(given invalidTestCase) {
			// given
			args := []string{
				"--config-file", filepath.Join("..", "testdata", "sample-kumactl.config.yaml"),
				"export",
			}
			args = append(args, given.args...)
			rootCmd.SetArgs(args)

			// when
			err := rootCmd.Execute()

			// then
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(given.err))
		},
		Entry("invalid profile", invalidTestCase{
			args: []string{"--profile", "something"},
			err:  "invalid profile",
		}),
		Entry("invalid format", invalidTestCase{
			args: []string{"--format", "something"},
			err:  "invalid format",
		}),
	)
})

type fileBasedKubernetesResourcesClient struct{}

var _ client.KubernetesResourcesClient = &fileBasedKubernetesResourcesClient{}

func (f fileBasedKubernetesResourcesClient) Get(_ context.Context, descriptor model.ResourceTypeDescriptor, name, mesh string) (map[string]interface{}, error) {
	inputBytes, err := os.ReadFile(filepath.Join("testdata", fmt.Sprintf("kube-input.%s.%s.json", descriptor.Name, name)))
	if err != nil {
		return nil, err
	}
	res := map[string]interface{}{}
	if err := json.Unmarshal(inputBytes, &res); err != nil {
		return nil, err
	}
	return res, nil
}

type staticResourcesListClient struct{}

var _ client.ResourcesListClient = &staticResourcesListClient{}

func (s staticResourcesListClient) List(ctx context.Context) (api_types.ResourceTypeDescriptionList, error) {
	defs := registry.Global().ObjectDescriptors()
	return mappers.MapResourceTypeDescription(defs, false, true), nil
}
