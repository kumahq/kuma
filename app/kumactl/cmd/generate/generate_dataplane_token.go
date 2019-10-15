package generate

import (
	kumactl_cmd "github.com/Kong/kuma/app/kumactl/pkg/cmd"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type generateDataplaneTokenContext struct {
	*kumactl_cmd.RootContext

	args struct {
		name string
		mesh string
	}
}

func NewGenerateDataplaneTokenCmd(pctx *kumactl_cmd.RootContext) *cobra.Command {
	ctx := &generateDataplaneTokenContext{RootContext: pctx}
	cmd := &cobra.Command{
		Use:   "dataplane-token",
		Short: "Generate Dataplane Token",
		Long:  `Generate Dataplane Token that is used to prove Dataplane identity.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := pctx.CurrentDataplaneTokenClient()
			if err != nil {
				return errors.Wrap(err, "failed to create dataplane token client")
			}

			token, err := client.Generate(ctx.args.name, pctx.Args.Mesh)
			if err != nil {
				return errors.Wrap(err, "failed to generate a dataplane token")
			}
			_, err = cmd.OutOrStdout().Write([]byte(token))
			return err
		},
	}
	cmd.Flags().StringVar(&ctx.args.name, "name", "", "name of the Dataplane")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}
