package generate

import (
	kumactl_cmd "github.com/Kong/kuma/app/kumactl/pkg/cmd"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type generateDpTokenContext struct {
	*kumactl_cmd.RootContext

	args struct {
		name string
		mesh string
	}
}

func NewGenerateDpTokenCmd(pctx *kumactl_cmd.RootContext) *cobra.Command {
	ctx := &generateDpTokenContext{RootContext: pctx}
	cmd := &cobra.Command{
		Use:   "dp-token",
		Short: "Generate Dataplane Token",
		Long:  `Generate Dataplane Token that is used to prove Dataplane identity.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := pctx.CurrentDpTokenClient()
			if err != nil {
				return errors.Wrap(err, "failed to create dp token client")
			}

			token, err := client.Generate(ctx.args.name, ctx.args.mesh)
			if err != nil {
				return errors.Wrap(err, "failed to generate a dataplane token")
			}
			_, err = cmd.OutOrStdout().Write([]byte(token))
			return err
		},
	}
	cmd.Flags().StringVarP(&ctx.args.name, "name", "", "", "name of the Dataplane")
	_ = cmd.MarkFlagRequired("name")
	cmd.Flags().StringVarP(&ctx.args.mesh, "mesh", "", "", "mesh to which the Dataplane belongs to")
	_ = cmd.MarkFlagRequired("mesh")
	return cmd
}
