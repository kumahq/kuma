package inspect_test

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	gomega_types "github.com/onsi/gomega/types"
	"github.com/spf13/cobra"

	"github.com/kumahq/kuma/app/kumactl/cmd"
	"github.com/kumahq/kuma/app/kumactl/pkg/resources"
	api_server_types "github.com/kumahq/kuma/pkg/api-server/types"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	test_kumactl "github.com/kumahq/kuma/pkg/test/kumactl"
	"github.com/kumahq/kuma/pkg/test/matchers"
	util_http "github.com/kumahq/kuma/pkg/util/http"
)

type testDataplaneInspectClient struct {
	response *api_server_types.DataplaneInspectEntryList
}

func (t *testDataplaneInspectClient) InspectPolicies(ctx context.Context, mesh, name string) (*api_server_types.DataplaneInspectEntryList, error) {
	return t.response, nil
}

func (t *testDataplaneInspectClient) InspectConfigDump(ctx context.Context, mesh, name string) ([]byte, error) {
	return nil, nil
}

var _ resources.DataplaneInspectClient = &testDataplaneInspectClient{}

var _ = Describe("kumactl inspect dataplane", func() {

	var rootCmd *cobra.Command
	var buf *bytes.Buffer

	BeforeEach(func() {
		rawResponse, err := os.ReadFile(path.Join("testdata", "inspect-dataplane.server-response.json"))
		Expect(err).ToNot(HaveOccurred())

		receiver := &api_server_types.DataplaneInspectEntryListReceiver{
			NewResource: registry.Global().NewObject,
		}
		Expect(json.Unmarshal(rawResponse, receiver)).To(Succeed())

		testClient := &testDataplaneInspectClient{
			response: &receiver.DataplaneInspectEntryList,
		}

		rootCtx, err := test_kumactl.MakeRootContext(time.Now(), nil)
		Expect(err).ToNot(HaveOccurred())

		rootCtx.Runtime.NewDataplaneInspectClient = func(client util_http.Client) resources.DataplaneInspectClient {
			return testClient
		}

		rootCmd = cmd.NewRootCmd(rootCtx)
		buf = &bytes.Buffer{}
		rootCmd.SetOut(buf)
	})

	type testCase struct {
		goldenFile string
		matcher    func(path ...string) gomega_types.GomegaMatcher
	}
	DescribeTable("kumactl inspect dataplane",
		func(given testCase) {
			// given
			rootCmd.SetArgs([]string{
				"--config-file", filepath.Join("..", "testdata", "sample-kumactl.config.yaml"),
				"inspect", "dataplane", "backend-1"})

			// when
			err := rootCmd.Execute()

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(buf.String()).To(given.matcher("testdata", given.goldenFile))
		},
		Entry("default output", testCase{
			goldenFile: "inspect-dataplane.golden.txt",
			matcher:    matchers.MatchGoldenEqual,
		}),
	)
})
