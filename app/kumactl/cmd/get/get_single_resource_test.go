package get_test

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"

	"github.com/kumahq/kuma/app/kumactl/cmd"
	kumactl_cmd "github.com/kumahq/kuma/app/kumactl/pkg/cmd"
	"github.com/kumahq/kuma/app/kumactl/pkg/resources"
	"github.com/kumahq/kuma/pkg/api-server/types"
	config_proto "github.com/kumahq/kuma/pkg/config/app/kumactl/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	memory_resources "github.com/kumahq/kuma/pkg/plugins/resources/memory"
	. "github.com/kumahq/kuma/pkg/test/matchers"
	util_http "github.com/kumahq/kuma/pkg/util/http"
	kuma_version "github.com/kumahq/kuma/pkg/version"
)

type testApiServerClient struct {
}

func (c *testApiServerClient) GetVersion(_ context.Context) (*types.IndexResponse, error) {
	return &types.IndexResponse{
		Version: kuma_version.Build.Version,
		Tagline: kuma_version.Product,
	}, nil
}

var _ = Describe("kumactl get [resource] NAME", func() {
	var rootCmd *cobra.Command
	var outbuf *bytes.Buffer
	var store core_store.ResourceStore
	var testClient *testApiServerClient
	rootTime, _ := time.Parse(time.RFC3339, "2008-04-01T16:05:36.995Z")
	var _ resources.ApiServerClient = &testApiServerClient{}
	BeforeEach(func() {
		rootCtx := &kumactl_cmd.RootContext{
			Runtime: kumactl_cmd.RootRuntime{
				Registry: registry.Global(),
				Now:      func() time.Time { return rootTime },
				NewBaseAPIServerClient: func(server *config_proto.ControlPlaneCoordinates_ApiServer, _ time.Duration) (util_http.Client, error) {
					return nil, nil
				},
				NewResourceStore: func(util_http.Client) core_store.ResourceStore {
					return store
				},
				NewAPIServerClient: func(util_http.Client) resources.ApiServerClient {
					return testClient
				},
			},
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

	entries := []TableEntry{
		Entry("circuit-breaker", "circuit-breaker"),
		Entry("fault-injection", "fault-injection"),
		Entry("dataplane", "dataplane"),
		Entry("mesh", "mesh"),
		Entry("healthcheck", "healthcheck"),
		Entry("proxytemplate", "proxytemplate"),
		Entry("rate-limit", "rate-limit"),
		Entry("traffic-log", "traffic-log"),
		Entry("traffic-permission", "traffic-permission"),
		Entry("traffic-route", "traffic-route"),
		Entry("traffic-trace", "traffic-trace"),
		Entry("secret", "secret"),
		Entry("global-secret", "global-secret"),
		Entry("retry", "retry"),
	}

	DescribeTable("should throw an error in case of no args",
		func(resource string) {
			// given
			rootCmd.SetArgs([]string{
				"get", resource})

			// when
			err := rootCmd.Execute()

			// then
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("accepts 1 arg(s), received 0"))
			Expect(outbuf.String()).To(MatchRegexp(`Error: accepts 1 arg\(s\), received 0`))
		},
		entries,
	)

	DescribeTable("should return error message if doesn't exist",
		func(resource string) {
			// given
			rootCmd.SetArgs([]string{
				"get", resource, "unknown-resource"})

			// when
			err := rootCmd.Execute()

			// then
			Expect(err).To(HaveOccurred())
			// and
			if resource == "mesh" || resource == "global-secret" {
				Expect(outbuf.String()).To(Equal("Error: No resources found\n"))
			} else {
				Expect(outbuf.String()).To(Equal("Error: No resources found in default mesh\n"))
			}
		},
		entries,
	)

	DescribeTable("kumactl get [resource] [name] -otable",
		func(resource string) {
			// setup - add resource to store
			resourceYAML := fmt.Sprintf("get-%s.golden.yaml", resource)
			rootCmd.SetArgs([]string{"apply", "-f", filepath.Join("testdata", resourceYAML)})
			Expect(rootCmd.Execute()).To(Succeed())

			// given
			resourceTable := fmt.Sprintf("get-%s.golden.txt", resource)

			// when
			resourceName := fmt.Sprintf("%s-1", resource)
			rootCmd.SetArgs([]string{"get", resource, resourceName, "-otable"})
			outbuf.Reset()
			err := rootCmd.Execute()

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(outbuf.String()).To(MatchGoldenEqual("testdata", resourceTable))
		},
		entries,
	)

	DescribeTable("kumactl get [resource] [name] -ojson",
		func(resource string) {
			// setup - add resource to store
			resourceYAML := fmt.Sprintf("get-%s.golden.yaml", resource)
			rootCmd.SetArgs([]string{"apply", "-f", filepath.Join("testdata", resourceYAML)})
			Expect(rootCmd.Execute()).To(Succeed())

			// given
			resourceJSON := fmt.Sprintf("get-%s.golden.json", resource)

			// when
			resourceName := fmt.Sprintf("%s-1", resource)
			rootCmd.SetArgs([]string{"get", resource, resourceName, "-ojson"})
			outbuf.Reset()
			err := rootCmd.Execute()

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(outbuf.String()).To(MatchGoldenEqual("testdata", resourceJSON))
		},
		entries,
	)

	DescribeTable("kumactl get [resource] [name] -oyaml",
		func(resource string) {
			// setup - add resource to store
			resourceYAML := fmt.Sprintf("get-%s.golden.yaml", resource)
			rootCmd.SetArgs([]string{"apply", "-f", filepath.Join("testdata", resourceYAML)})
			Expect(rootCmd.Execute()).To(Succeed())

			// when
			resourceName := fmt.Sprintf("%s-1", resource)
			rootCmd.SetArgs([]string{"get", resource, resourceName, "-oyaml"})
			outbuf.Reset()
			err := rootCmd.Execute()

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(outbuf.String()).To(MatchGoldenEqual("testdata", resourceYAML))
		},
		entries,
	)
})
