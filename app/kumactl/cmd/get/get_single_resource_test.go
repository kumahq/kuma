package get_test

import (
	"bytes"
	"fmt"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"

	"github.com/kumahq/kuma/app/kumactl/cmd"
	kumactl_cmd "github.com/kumahq/kuma/app/kumactl/pkg/cmd"
	"github.com/kumahq/kuma/app/kumactl/pkg/resources"
	"github.com/kumahq/kuma/pkg/api-server/types"
	config_proto "github.com/kumahq/kuma/pkg/config/app/kumactl/v1alpha1"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	memory_resources "github.com/kumahq/kuma/pkg/plugins/resources/memory"
	. "github.com/kumahq/kuma/pkg/test/matchers"
	kuma_version "github.com/kumahq/kuma/pkg/version"
)

type testApiServerClient struct {
}

func (c *testApiServerClient) GetVersion() (*types.IndexResponse, error) {
	return &types.IndexResponse{
		Version: kuma_version.Build.Version,
	}, nil
}

var _ = Describe("kumactl get [resource] NAME", func() {
	var rootCtx *kumactl_cmd.RootContext
	var rootCmd *cobra.Command
	var outbuf, errbuf *bytes.Buffer
	var store core_store.ResourceStore
	var testClient *testApiServerClient
	rootTime, _ := time.Parse(time.RFC3339, "2008-04-01T16:05:36.995Z")
	var _ resources.ApiServerClient = &testApiServerClient{}
	BeforeEach(func() {
		rootCtx = &kumactl_cmd.RootContext{
			Runtime: kumactl_cmd.RootRuntime{
				Now: func() time.Time { return rootTime },
				NewResourceStore: func(*config_proto.ControlPlaneCoordinates_ApiServer) (core_store.ResourceStore, error) {
					return store, nil
				},
				NewAPIServerClient: func(*config_proto.ControlPlaneCoordinates_ApiServer) (resources.ApiServerClient, error) {
					return testClient, nil
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

	entries := []TableEntry{
		Entry("circuit-breaker", "circuit-breaker"),
		Entry("fault-injection", "fault-injection"),
		Entry("dataplane", "dataplane"),
		Entry("mesh", "mesh"),
		Entry("healthcheck", "healthcheck"),
		Entry("proxytemplate", "proxytemplate"),
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
			Expect(err.Error()).To(Equal("requires at least 1 arg(s), only received 0"))
			Expect(outbuf.String()).To(MatchRegexp(`Error: requires at least 1 arg\(s\), only received 0`))
			Expect(errbuf.Bytes()).To(BeEmpty())
		},
		entries...,
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
			// and
			Expect(errbuf.Bytes()).To(BeEmpty())

		},
		entries...,
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
			Expect(outbuf.String()).To(MatchGoldenEqual(filepath.Join("testdata", resourceTable)))
		},
		entries...,
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
			Expect(outbuf.String()).To(MatchGoldenEqual(filepath.Join("testdata", resourceJSON)))
		},
		entries...,
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
			Expect(outbuf.String()).To(MatchGoldenEqual(filepath.Join("testdata", resourceYAML)))
		},
		entries...,
	)
})
