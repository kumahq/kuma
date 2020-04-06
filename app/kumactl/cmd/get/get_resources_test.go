package get_test

import (
	"bytes"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/spf13/cobra"

	"github.com/Kong/kuma/app/kumactl/cmd"
	kumactl_cmd "github.com/Kong/kuma/app/kumactl/pkg/cmd"
	config_proto "github.com/Kong/kuma/pkg/config/app/kumactl/v1alpha1"
	core_store "github.com/Kong/kuma/pkg/core/resources/store"
	memory_resources "github.com/Kong/kuma/pkg/plugins/resources/memory"
)

var _ = Describe("kumactl get TYPE NAME", func() {
	var rootCtx *kumactl_cmd.RootContext
	var rootCmd *cobra.Command
	var outbuf, errbuf *bytes.Buffer
	var store core_store.ResourceStore

	BeforeEach(func() {
		// setup
		rootCtx = &kumactl_cmd.RootContext{
			Runtime: kumactl_cmd.RootRuntime{
				Now: time.Now,
				NewResourceStore: func(*config_proto.ControlPlaneCoordinates_ApiServer) (core_store.ResourceStore, error) {
					return store, nil
				},
			},
		}
		store = memory_resources.NewStore()

		rootCmd = cmd.NewRootCmd(rootCtx)
		outbuf = &bytes.Buffer{}
		errbuf = &bytes.Buffer{}
		rootCmd.SetOut(outbuf)
		rootCmd.SetErr(errbuf)
	})
	It("should throw an error in case of no args", func() {
		// given
		rootCmd.SetArgs([]string{
			"get", "resource"})

		// when
		err := rootCmd.Execute()

		// then
		Expect(err).To(HaveOccurred())
		// and
		Expect(err.Error()).To(Equal("accepts 2 arg(s), received 0"))
		// and
		Expect(outbuf.String()).To(MatchRegexp(`Error: accepts 2 arg\(s\), received 0`))
		// and
		Expect(errbuf.Bytes()).To(BeEmpty())
	})
	It("should throw an error in case of unsupported resource type", func() {
		// given
		rootCmd.SetArgs([]string{
			"get", "resource", "some-type", "some-name"})

		// when
		err := rootCmd.Execute()

		// then
		Expect(err).To(HaveOccurred())
		// and
		Expect(err.Error()).To(Equal("unknown TYPE: some-type. Allowed values: mesh, dataplane, healthcheck, proxytemplate, traffic-log, traffic-permission, traffic-route, traffic-trace, fault-injection"))
		// and
		Expect(outbuf.String()).To(MatchRegexp(`unknown TYPE: some-type. Allowed values: mesh, dataplane, healthcheck, proxytemplate, traffic-log, traffic-permission, traffic-route, traffic-trace, fault-injection`))
		// and
		Expect(errbuf.Bytes()).To(BeEmpty())
	})
})
