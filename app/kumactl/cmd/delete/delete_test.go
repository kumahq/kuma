package delete_test

import (
	"bytes"
	"github.com/Kong/kuma/app/kumactl/cmd"
	kumactl_cmd "github.com/Kong/kuma/app/kumactl/pkg/cmd"
	config_proto "github.com/Kong/kuma/pkg/config/app/kumactl/v1alpha1"
	core_store "github.com/Kong/kuma/pkg/core/resources/store"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"
	"path/filepath"
	"time"
)

var _ = Describe("kumactl delete ", func() {
	Describe("Delete Command", func() {
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
			rootCmd = cmd.NewRootCmd(rootCtx)
			outbuf = &bytes.Buffer{}
			errbuf = &bytes.Buffer{}
			rootCmd.SetOut(outbuf)
			rootCmd.SetErr(errbuf)
		})

		It("should throw an error in case of no args", func() {
			// given
			rootCmd.SetArgs([]string{
				"--config-file", filepath.Join("..", "testdata", "sample-kumactl.config.yaml"),
				"delete"})

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
				"--config-file", filepath.Join("..", "testdata", "sample-kumactl.config.yaml"),
				"delete", "some-type", "some-name"})

			// when
			err := rootCmd.Execute()

			// then
			Expect(err).To(HaveOccurred())
			// and
			Expect(err.Error()).To(Equal("unknown TYPE: some-type. Allowed values: mesh, dataplane, proxytemplate, traffic-log, traffic-permission"))
			// and
			Expect(outbuf.String()).To(MatchRegexp(`unknown TYPE: some-type. Allowed values: mesh, dataplane, proxytemplate, traffic-log, traffic-permission`))
			// and
			Expect(errbuf.Bytes()).To(BeEmpty())
		})
	})
})
