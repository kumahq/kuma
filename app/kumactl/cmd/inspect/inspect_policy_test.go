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

	"github.com/kumahq/kuma/app/kumactl/cmd"
	"github.com/kumahq/kuma/app/kumactl/pkg/resources"
	api_server_types "github.com/kumahq/kuma/pkg/api-server/types"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	test_kumactl "github.com/kumahq/kuma/pkg/test/kumactl"
	"github.com/kumahq/kuma/pkg/test/matchers"
	util_http "github.com/kumahq/kuma/pkg/util/http"
)

type testPolicyInspectClient struct {
	response *api_server_types.PolicyInspectEntryList
}

func (t *testPolicyInspectClient) Inspect(ctx context.Context, policyDesc model.ResourceTypeDescriptor, mesh, name string) (*api_server_types.PolicyInspectEntryList, error) {
	return t.response, nil
}

var _ resources.PolicyInspectClient = &testPolicyInspectClient{}

var _ = Describe("kumactl inspect POLICY", func() {

	type testCase struct {
		goldenFile         string
		serverResponseFile string
		cmdArgs            []string
	}
	DescribeTable("kumactl inspect dataplane",
		func(given testCase) {
			// given
			rawResponse, err := os.ReadFile(path.Join("testdata", given.serverResponseFile))
			Expect(err).ToNot(HaveOccurred())

			entryList := &api_server_types.PolicyInspectEntryList{}
			Expect(json.Unmarshal(rawResponse, entryList)).To(Succeed())

			rootCtx, err := test_kumactl.MakeRootContext(time.Now(), nil)
			Expect(err).ToNot(HaveOccurred())

			rootCtx.Runtime.NewPolicyInspectClient = func(client util_http.Client) resources.PolicyInspectClient {
				return &testPolicyInspectClient{
					response: entryList,
				}
			}

			rootCmd := cmd.NewRootCmd(rootCtx)
			buf := &bytes.Buffer{}
			rootCmd.SetOut(buf)

			rootCmd.SetArgs(append([]string{"--config-file", filepath.Join("..", "testdata", "sample-kumactl.config.yaml")},
				given.cmdArgs...))

			// when
			err = rootCmd.Execute()

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(buf.String()).To(matchers.MatchGoldenEqual("testdata", given.goldenFile))
		},
		Entry("inbound policy", testCase{
			goldenFile:         "inspect-traffic-permission.golden.txt",
			serverResponseFile: "inspect-traffic-permission.server-response.json",
			cmdArgs:            []string{"inspect", "traffic-permission", "tp1"},
		}),
		Entry("outbound policy", testCase{
			goldenFile:         "inspect-timeout.golden.txt",
			serverResponseFile: "inspect-timeout.server-response.json",
			cmdArgs:            []string{"inspect", "timeout", "t1"},
		}),
		Entry("service policy", testCase{
			goldenFile:         "inspect-health-check.golden.txt",
			serverResponseFile: "inspect-health-check.server-response.json",
			cmdArgs:            []string{"inspect", "healthcheck", "hc1"},
		}),
		Entry("dataplane policy", testCase{
			goldenFile:         "inspect-traffic-trace.golden.txt",
			serverResponseFile: "inspect-traffic-trace.server-response.json",
			cmdArgs:            []string{"inspect", "traffic-trace", "tt1"},
		}),
	)
})
