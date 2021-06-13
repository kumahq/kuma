package generate

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	kumactl_cmd "github.com/kumahq/kuma/app/kumactl/pkg/cmd"
)

type generateZoneIngressTokenContext struct {
	*kumactl_cmd.RootContext

	args struct {
		zone string
	}
}

func NewGenerateZoneIngressTokenCmd(pctx *kumactl_cmd.RootContext) *cobra.Command {
	ctx := &generateZoneIngressTokenContext{RootContext: pctx}
	cmd := &cobra.Command{
		Use:   "zone-ingress-token",
		Short: "Generate Zone Ingress Token",
		Long:  `Generate Zone Ingress Token that is used to prove Zone Ingress identity.`,
		Example: `
Generate token bound by zone
$ kumactl generate zone-ingress-token --zone zone-1
`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := pctx.CurrentZoneIngressTokenClient()
			if err != nil {
				return errors.Wrap(err, "failed to create zone ingress token client")
			}

			token, err := client.Generate(ctx.args.zone)
			if err != nil {
				return errors.Wrap(err, "failed to generate a zone ingress token")
			}
			_, err = cmd.OutOrStdout().Write([]byte(token))
			return err
		},
	}
	cmd.Flags().StringVar(&ctx.args.zone, "zone", "", "name of the zone where ingress resides")
	return cmd
}
