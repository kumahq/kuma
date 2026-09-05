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

	api_types "github.com/kumahq/kuma/v3/api/openapi/types"
	"github.com/kumahq/kuma/v3/app/kumactl/cmd"
	"github.com/kumahq/kuma/v3/app/kumactl/pkg/resources"
	test_kumactl "github.com/kumahq/kuma/v3/app/kumactl/pkg/test"
	"github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/test/matchers"
	util_http "github.com/kumahq/kuma/v3/pkg/util/http"
)

type testPolicyInspectClient struct {
	ensureMesh string
	ensureName string
	ensureSize int
	ensurePage string
	dpResponse api_types.InspectDataplanesForPolicyResponse
}

func (t *testPolicyInspectClient) DataplanesForPolicy(ctx context.Context, desc model.ResourceTypeDescriptor, mesh, name string, size int, offset string) (api_types.InspectDataplanesForPolicyResponse, error) {
	if t.ensureMesh != "" {
		Expect(mesh).To(Equal(t.ensureMesh))
	}
	Expect(name).To(Equal(t.ensureName))
	Expect(size).To(Equal(t.ensureSize))
	Expect(offset).To(Equal(t.ensurePage))
	return t.dpResponse, nil
}

var _ resources.PolicyInspectClient = &testPolicyInspectClient{}

var _ = Describe("kumactl inspect POLICY", func() {
	type testCase struct {
		goldenFile         string
		serverResponseFile string
		mesh               string
		cmdArgs            []string
		size               int
		offset             string
	}
	DescribeTable("kumactl inspect dataplane",
		func(given testCase) {
			// given
			rawResponse, err := os.ReadFile(path.Join("testdata", given.serverResponseFile))
			Expect(err).ToNot(HaveOccurred())

			entryList := api_types.InspectDataplanesForPolicyResponse{}
			Expect(json.Unmarshal(rawResponse, &entryList)).To(Succeed())
			client := &testPolicyInspectClient{
				ensureMesh: given.mesh,
				ensureName: "tt1",
				ensureSize: given.size,
				ensurePage: given.offset,
				dpResponse: entryList,
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
		Entry("mtp", testCase{
			goldenFile:         "inspect-mtp-dp.golden.txt",
			serverResponseFile: "inspect-mtp-dp.server-response.json",
			cmdArgs:            []string{"inspect", "meshtrafficpermission", "tt1", "--size", "25", "--offset", "next-page"},
			size:               25,
			offset:             "next-page",
		}),
		Entry("mtp with deprecated new-api flag", testCase{
			goldenFile:         "inspect-mtp-dp.golden.txt",
			serverResponseFile: "inspect-mtp-dp.server-response.json",
			cmdArgs:            []string{"inspect", "meshtrafficpermission", "tt1", "--new-api"},
		}),
	)
})
