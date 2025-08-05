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

	api_types "github.com/kumahq/kuma/api/openapi/types"
	"github.com/kumahq/kuma/app/kumactl/cmd"
	"github.com/kumahq/kuma/app/kumactl/pkg/resources"
	test_kumactl "github.com/kumahq/kuma/app/kumactl/pkg/test"
	api_server_types "github.com/kumahq/kuma/pkg/api-server/types"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/test/matchers"
	util_http "github.com/kumahq/kuma/pkg/util/http"
)

type testPolicyInspectClient struct {
	ensureMesh string
	response   *api_server_types.PolicyInspectEntryList
	dpResponse api_types.InspectDataplanesForPolicyResponse
}

func (t *testPolicyInspectClient) DataplanesForPolicy(ctx context.Context, desc model.ResourceTypeDescriptor, mesh, name string) (api_types.InspectDataplanesForPolicyResponse, error) {
	if t.ensureMesh != "" {
		Expect(mesh).To(Equal(t.ensureMesh))
	}
	return t.dpResponse, nil
}

func (t *testPolicyInspectClient) Inspect(ctx context.Context, policyDesc model.ResourceTypeDescriptor, mesh, name string) (*api_server_types.PolicyInspectEntryList, error) {
	if t.ensureMesh != "" {
		Expect(mesh).To(Equal(t.ensureMesh))
	}
	return t.response, nil
}

var _ resources.PolicyInspectClient = &testPolicyInspectClient{}

var _ = Describe("kumactl inspect POLICY", func() {
	type testCase struct {
		goldenFile         string
		serverResponseFile string
		mesh               string
		cmdArgs            []string
	}
	DescribeTable("kumactl inspect dataplane",
		func(given testCase) {
			// given
			rawResponse, err := os.ReadFile(path.Join("testdata", given.serverResponseFile))
			Expect(err).ToNot(HaveOccurred())

			newApi := false
			for _, arg := range given.cmdArgs {
				if arg == "--new-api" {
					newApi = true
					break
				}
			}
			var client *testPolicyInspectClient
			if newApi {
				entryList := api_types.InspectDataplanesForPolicyResponse{}
				Expect(json.Unmarshal(rawResponse, &entryList)).To(Succeed())
				client = &testPolicyInspectClient{
					ensureMesh: given.mesh,
					dpResponse: entryList,
				}

			} else {
				entryList := &api_server_types.PolicyInspectEntryList{}
				Expect(json.Unmarshal(rawResponse, entryList)).To(Succeed())
				client = &testPolicyInspectClient{
					ensureMesh: given.mesh,
					response:   entryList,
				}
			}

			rootCtx := test_kumactl.MakeMinimalRootContext()
			rootCtx.Runtime.NewPolicyInspectClient = func(_ util_http.Client) resources.PolicyInspectClient {
				return client
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
			mesh:               "default",
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
		Entry("service policy (no kind in response)", testCase{
			goldenFile:         "inspect-health-check-1.5.golden.txt",
			serverResponseFile: "inspect-health-check-1.5.server-response.json",
			cmdArgs:            []string{"inspect", "healthcheck", "hc1"},
		}),
		Entry("dataplane policy", testCase{
			goldenFile:         "inspect-traffic-trace.golden.txt",
			serverResponseFile: "inspect-traffic-trace.server-response.json",
			cmdArgs:            []string{"inspect", "traffic-trace", "tt1"},
		}),
		Entry("new-api mtp", testCase{
			goldenFile:         "inspect-mtp-dp.golden.txt",
			serverResponseFile: "inspect-mtp-dp.server-response.json",
			cmdArgs:            []string{"inspect", "meshtrafficpermission", "tt1", "--new-api"},
		}),
		Entry("other-mesh", testCase{
			goldenFile:         "inspect-traffic-trace-other-mesh.golden.txt",
			serverResponseFile: "inspect-traffic-trace-other-mesh.server-response.json",
			mesh:               "other-mesh",
			cmdArgs:            []string{"inspect", "traffic-trace", "tt1", "--mesh", "other-mesh"},
		}),
	)
})
