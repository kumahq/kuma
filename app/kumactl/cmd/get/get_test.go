package get_test

import (
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"

	"github.com/kumahq/kuma/app/kumactl/cmd"
	kumactl_cmd "github.com/kumahq/kuma/app/kumactl/pkg/cmd"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	memory_resources "github.com/kumahq/kuma/pkg/plugins/resources/memory"
	util_http "github.com/kumahq/kuma/pkg/util/http"
)

func hasSubCommand(cmd *cobra.Command, sub string) bool {
	for _, c := range cmd.Commands() {
		if c.Use == sub {
			return true
		}
	}

	return false
}

func ExecuteRootCommand(cmd *cobra.Command, resourceName string, formatOpt string, pageOpt string) error {
	args := []string{
		"--config-file",
		filepath.Join("..", "testdata", "sample-kumactl.config.yaml"),
		"get",
		resourceName,
	}

	if formatOpt != "" {
		args = append(args, formatOpt)
	}

	if pageOpt != "" {
		args = append(args, pageOpt)
	}

	cmd.SetArgs(args)
	return cmd.Execute()
}

var _ = Describe("kumactl get ", func() {
	Describe("Get Command", func() {
		var rootCtx *kumactl_cmd.RootContext
		var rootCmd, getCmd *cobra.Command
		var store core_store.ResourceStore

		BeforeEach(func() {
			// setup
			store = core_store.NewPaginationStore(memory_resources.NewStore())

			rootCtx = kumactl_cmd.DefaultRootContext()
			rootCtx.Runtime.NewResourceStore = func(util_http.Client) core_store.ResourceStore {
				return store
			}
			rootCmd = cmd.NewRootCmd(rootCtx)
			for _, cmd := range rootCmd.Commands() {
				if cmd.Use == "get" {
					getCmd = cmd
					break
				}
			}
			Expect(getCmd).ToNot(BeNil())
		})

		It("should have get commands for all defined types", func() {
			// when
			all := registry.Global().ObjectDescriptors(model.HasKumactlEnabled())
			Expect(len(getCmd.Commands()) > len(all)).To(BeTrue())

			// then
			for _, sub := range all {
				Expect(hasSubCommand(getCmd, sub.KumactlArg+" NAME")).To(BeTrue(), "failing to find "+sub.KumactlArg)
			}
		})
	})
})
