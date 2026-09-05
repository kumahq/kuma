package inspect

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/kumahq/kuma/v3/api/openapi/types"
	"github.com/kumahq/kuma/v3/app/kumactl/pkg/cmd"
	"github.com/kumahq/kuma/v3/app/kumactl/pkg/output"
	"github.com/kumahq/kuma/v3/app/kumactl/pkg/output/printers"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
)

func newInspectPolicyCmd(policyDesc core_model.ResourceTypeDescriptor, pctx *cmd.RootContext) *cobra.Command {
	var size int
	var offset string
	cmd := &cobra.Command{
		Use:   fmt.Sprintf("%s NAME", policyDesc.KumactlArg),
		Short: fmt.Sprintf("Inspect %s", policyDesc.Name),
		Long:  fmt.Sprintf("Inspect %s.", policyDesc.Name),
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := pctx.CurrentPolicyInspectClient()
			if err != nil {
				return errors.Wrap(err, "failed to create a policy inspect client")
			}
			name := args[0]
			res, err := client.DataplanesForPolicy(cmd.Context(), policyDesc, pctx.CurrentMesh(), name, size, offset)
			if err != nil {
				return err
			}
			format := output.Format(pctx.InspectContext.Args.OutputFormat)
			return printers.GenericPrint(format, res, printers.Table{
				Headers: []string{"Type", "Mesh", "Name"},
				FooterFn: func(container any) string {
					return fmt.Sprintf("Total: %d", container.(types.InspectDataplanesForPolicyResponse).Total)
				},
				RowForItem: func(i int, container any) ([]string, error) {
					items := container.(types.InspectDataplanesForPolicyResponse).Items
					if i >= len(items) {
						return nil, nil
					}
					itm := items[i]
					return []string{itm.Type, itm.Mesh, itm.Name}, nil
				},
			}, cmd.OutOrStdout())
		},
	}
	cmd.PersistentFlags().StringVarP(&pctx.Args.Mesh, "mesh", "m", "default", "mesh to use")
	cmd.PersistentFlags().IntVar(&size, "size", 0, "maximum number of elements to return")
	cmd.PersistentFlags().StringVar(&offset, "offset", "", "the offset that indicates starting element of the dataplane list to retrieve")
	cmd.PersistentFlags().Bool("new-api", false, "deprecated; the current inspect API is always used")
	return cmd
}
