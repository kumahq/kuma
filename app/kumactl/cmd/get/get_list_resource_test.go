package get_test

import (
	"bytes"
	"fmt"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"

	"github.com/kumahq/kuma/app/kumactl/cmd"
	test_kumactl "github.com/kumahq/kuma/app/kumactl/pkg/test"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	memory_resources "github.com/kumahq/kuma/pkg/plugins/resources/memory"
	. "github.com/kumahq/kuma/pkg/test/matchers"
)

var _ = Describe("kumactl get [resource]", func() {
	var rootCmd *cobra.Command
	var outbuf *bytes.Buffer
	var store core_store.ResourceStore
	rootTime, _ := time.Parse(time.RFC3339, "2008-04-01T16:05:36.995Z")
	BeforeEach(func() {
		store = core_store.NewPaginationStore(memory_resources.NewStore())
		rootCtx, _ := test_kumactl.MakeRootContext(rootTime, store)
		rootCtx.Runtime.Registry = registry.Global()
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
		Entry("circuit-breaker", "circuit-breakers"),
		Entry("fault-injection", "fault-injections"),
		Entry("dataplane", "dataplanes"),
		Entry("external-service", "external-services"),
		Entry("fault-injection", "fault-injections"),
		Entry("global-secret", "global-secrets"),
		Entry("healthcheck", "healthchecks"),
		Entry("mesh", "meshes"),
		Entry("proxytemplate", "proxytemplates"),
		Entry("rate-limit", "rate-limits"),
		Entry("retry", "retries"),
		Entry("secret", "secrets"),
		Entry("traffic-log", "traffic-logs"),
		Entry("traffic-permission", "traffic-permissions"),
		Entry("traffic-route", "traffic-routes"),
		Entry("traffic-trace", "traffic-traces"),
		Entry("zone-ingress", "zone-ingresses"),
		Entry("zoneegress", "zoneegresses"),
		Entry("zone", "zones"),
		Entry("meshtimeout", "meshtimeouts"),
	}

	DescribeTable("kumactl get [resource] -otable",
		func(resource string) {
			// setup - add resource to store
			resourceYAML := fmt.Sprintf("get-%s.input.yaml", resource)
			rootCmd.SetArgs([]string{"apply", "-f", filepath.Join("testdata/list", resourceYAML)})
			Expect(rootCmd.Execute()).To(Succeed())

			// given
			resourceTable := fmt.Sprintf("get-%s.golden.txt", resource)

			// when
			rootCmd.SetArgs([]string{"get", resource, "-otable"})
			outbuf.Reset()
			err := rootCmd.Execute()

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(outbuf.String()).To(MatchGoldenEqual("testdata/list", resourceTable))
		},
		entries,
	)

	DescribeTable("kumactl get [resource] -otable --size 1",
		func(resource string) {
			// setup - add resource to store
			resourceYAML := fmt.Sprintf("get-%s.input.yaml", resource)
			rootCmd.SetArgs([]string{"apply", "-f", filepath.Join("testdata/list", resourceYAML)})
			Expect(rootCmd.Execute()).To(Succeed())

			// given
			resourceTable := fmt.Sprintf("get-%s.pagination.golden.txt", resource)

			// when
			rootCmd.SetArgs([]string{"get", resource, "-otable", "--size", "1"})
			outbuf.Reset()
			err := rootCmd.Execute()

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(outbuf.String()).To(MatchGoldenEqual("testdata/list", resourceTable))
		},
		entries,
	)

	DescribeTable("kumactl get [resource] -ojson",
		func(resource string) {
			// setup - add resource to store
			resourceYAML := fmt.Sprintf("get-%s.input.yaml", resource)
			rootCmd.SetArgs([]string{"apply", "-f", filepath.Join("testdata/list", resourceYAML)})
			Expect(rootCmd.Execute()).To(Succeed())

			// given
			resourceJSON := fmt.Sprintf("get-%s.golden.json", resource)

			// when
			rootCmd.SetArgs([]string{"get", resource, "-ojson"})
			outbuf.Reset()
			err := rootCmd.Execute()

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(outbuf.String()).To(MatchGoldenEqual("testdata/list", resourceJSON))
		},
		entries,
	)

	DescribeTable("kumactl get [resource] -oyaml",
		func(resource string) {
			// setup - add resource to store
			resourceYAML := fmt.Sprintf("get-%s.input.yaml", resource)
			rootCmd.SetArgs([]string{"apply", "-f", filepath.Join("testdata/list", resourceYAML)})
			Expect(rootCmd.Execute()).To(Succeed())

			// when
			rootCmd.SetArgs([]string{"get", resource, "-oyaml"})
			outbuf.Reset()
			err := rootCmd.Execute()

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(outbuf.String()).To(MatchGoldenEqual("testdata/list", fmt.Sprintf("get-%s.golden.yaml", resource)))
		},
		entries,
	)
}, Ordered)
