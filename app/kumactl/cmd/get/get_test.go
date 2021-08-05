package get_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"

	"github.com/kumahq/kuma/app/kumactl/cmd"
	kumactl_cmd "github.com/kumahq/kuma/app/kumactl/pkg/cmd"
	"github.com/kumahq/kuma/app/kumactl/pkg/entities"
	config_proto "github.com/kumahq/kuma/pkg/config/app/kumactl/v1alpha1"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	memory_resources "github.com/kumahq/kuma/pkg/plugins/resources/memory"
)

func hasSubCommand(cmd *cobra.Command, sub string) bool {
	for _, c := range cmd.Commands() {
		if c.Use == sub {
			return true
		}
	}

	return false
}

var _ = Describe("kumactl get ", func() {
	Describe("Get Command", func() {
		var rootCtx *kumactl_cmd.RootContext
		var rootCmd, getCmd *cobra.Command
		var store core_store.ResourceStore

		BeforeEach(func() {
			// setup
			rootCtx = kumactl_cmd.DefaultRootContext()
			rootCtx.Runtime.NewResourceStore = func(*config_proto.ControlPlaneCoordinates_ApiServer) (core_store.ResourceStore, error) {
				return store, nil
			}
			store = core_store.NewPaginationStore(memory_resources.NewStore())

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
			Expect(len(getCmd.Commands()) > len(entities.Names)).To(BeTrue())

			// then
			for _, sub := range entities.All {
				Expect(hasSubCommand(getCmd, sub.Singular+" NAME")).To(BeTrue(), "failing to find "+sub.Singular)
			}
		})
	})
})
