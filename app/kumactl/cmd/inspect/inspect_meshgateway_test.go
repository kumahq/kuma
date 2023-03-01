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

	"github.com/kumahq/kuma/app/kumactl/cmd"
	"github.com/kumahq/kuma/app/kumactl/pkg/resources"
	api_server_types "github.com/kumahq/kuma/pkg/api-server/types"
	test_kumactl "github.com/kumahq/kuma/pkg/test/kumactl"
	"github.com/kumahq/kuma/pkg/test/matchers"
	util_http "github.com/kumahq/kuma/pkg/util/http"
)

type testMeshGatewayInspectClient struct {
	response api_server_types.GatewayDataplanesInspectEntryList
}

func (t *testMeshGatewayInspectClient) InspectDataplanes(ctx context.Context, mesh, name string) (api_server_types.GatewayDataplanesInspectEntryList, error) {
	return t.response, nil
}

var _ resources.MeshGatewayInspectClient = &testMeshGatewayInspectClient{}

type testCase struct {
	serverOutput string
	goldenFile   string
	matcher      func(path ...string) gomega_types.GomegaMatcher
}

var _ = DescribeTable("kumactl inspect meshgateway",
	func(given testCase) {
		rawResponse, err := os.ReadFile(path.Join("testdata", given.serverOutput))
		Expect(err).ToNot(HaveOccurred())

		response := api_server_types.GatewayDataplanesInspectEntryList{}
		Expect(json.Unmarshal(rawResponse, &response)).To(Succeed())

		testClient := &testMeshGatewayInspectClient{
			response: response,
		}

		rootCtx := test_kumactl.MakeMinimalRootContext()
		rootCtx.Runtime.NewMeshGatewayInspectClient = func(client util_http.Client) resources.MeshGatewayInspectClient {
			return testClient
		}

		rootCmd := cmd.NewRootCmd(rootCtx)
		buf := &bytes.Buffer{}
		rootCmd.SetOut(buf)
		// given
		rootCmd.SetArgs([]string{
			"--config-file", filepath.Join("..", "testdata", "sample-kumactl.config.yaml"),
			"inspect", "meshgateway", "meshgateway-1",
		})

		// when
		err = rootCmd.Execute()

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(buf.String()).To(given.matcher("testdata", given.goldenFile))
	},
	Entry("default output", testCase{
		serverOutput: "inspect-meshgateway-dataplanes.server-response.json",
		goldenFile:   "inspect-meshgateway.golden.txt",
		matcher:      matchers.MatchGoldenEqual,
	}),
)
