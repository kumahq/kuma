package inspect_test

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	gomega_types "github.com/onsi/gomega/types"
	"github.com/spf13/cobra"

	"github.com/kumahq/kuma/app/kumactl/cmd"
	"github.com/kumahq/kuma/app/kumactl/pkg/resources"
	api_server_types "github.com/kumahq/kuma/pkg/api-server/types"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	test_kumactl "github.com/kumahq/kuma/pkg/test/kumactl"
	"github.com/kumahq/kuma/pkg/test/matchers"
	util_http "github.com/kumahq/kuma/pkg/util/http"
)

type testDataplaneInspectClient struct {
	response api_server_types.DataplaneInspectResponse
}

func (t *testDataplaneInspectClient) InspectPolicies(ctx context.Context, mesh, name string) (api_server_types.DataplaneInspectResponse, error) {
	return t.response, nil
}

func (t *testDataplaneInspectClient) InspectConfigDump(ctx context.Context, mesh, name string) ([]byte, error) {
	return nil, nil
}

var _ resources.DataplaneInspectClient = &testDataplaneInspectClient{}

var _ = Describe("kumactl inspect dataplane", func() {
	var rootCmd *cobra.Command
	var buf *bytes.Buffer

	type testCase struct {
		serverOutput string
		goldenFile   string
		matcher      func(path ...string) gomega_types.GomegaMatcher
	}
	DescribeTable("kumactl inspect dataplane",
		func(given testCase) {
			rawResponse, err := os.ReadFile(path.Join("testdata", given.serverOutput))
			Expect(err).ToNot(HaveOccurred())

			response := api_server_types.DataplaneInspectResponse{}
			Expect(json.Unmarshal(rawResponse, &response)).To(Succeed())

			testClient := &testDataplaneInspectClient{
				response: response,
			}

			rootCtx := test_kumactl.MakeMinimalRootContext()

			rootCtx.Runtime.NewDataplaneInspectClient = func(client util_http.Client) resources.DataplaneInspectClient {
				return testClient
			}
			rootCtx.Runtime.NewInspectEnvoyProxyClient = func(descriptor model.ResourceTypeDescriptor, client util_http.Client) resources.InspectEnvoyProxyClient {
				return nil
			}

			rootCmd = cmd.NewRootCmd(rootCtx)
			buf = &bytes.Buffer{}
			rootCmd.SetOut(buf)
			// given
			rootCmd.SetArgs([]string{
				"--config-file", filepath.Join("..", "testdata", "sample-kumactl.config.yaml"),
				"inspect", "dataplane", "backend-1",
			})

			// when
			err = rootCmd.Execute()

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(buf.String()).To(given.matcher("testdata", given.goldenFile))
		},
		Entry("default output", testCase{
			serverOutput: "inspect-dataplane.server-response.json",
			goldenFile:   "inspect-dataplane.golden.txt",
			matcher:      matchers.MatchGoldenEqual,
		}),
		Entry("default output (no kind in response)", testCase{
			serverOutput: "inspect-dataplane-1.5.server-response.json",
			goldenFile:   "inspect-dataplane.golden.txt",
			matcher:      matchers.MatchGoldenEqual,
		}),
		Entry("builtin gateway dataplane", testCase{
			serverOutput: "inspect-gateway-dataplane.server-response.json",
			goldenFile:   "inspect-gateway-dataplane.golden.txt",
			matcher:      matchers.MatchGoldenEqual,
		}),
	)
})
