package generate

import (
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	kumactl_cmd "github.com/kumahq/kuma/app/kumactl/pkg/cmd"
)

type generateZoneEgressTokenContext struct {
	*kumactl_cmd.RootContext

	args struct {
		zone     string
		validFor time.Duration
	}
}

func NewGenerateZoneEgressTokenCmd(pctx *kumactl_cmd.RootContext) *cobra.Command {
	ctx := &generateZoneEgressTokenContext{RootContext: pctx}
	cmd := &cobra.Command{
		Use:   "zone-egress-token",
		Short: "Generate Zone Egress Token",
		Long:  `Generate Zone Egress Token that is used to prove Zone Egress identity.`,
		Example: `
Generate token bound by zone
$ kumactl generate zone-egress-token --zone zone-1 --valid-for 24h
`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := pctx.CurrentZoneEgressTokenClient()
			if err != nil {
				return errors.Wrap(err, "failed to create zone egress token client")
			}

			token, err := client.Generate(ctx.args.zone, ctx.args.validFor)
			if err != nil {
				return errors.Wrap(err, "failed to generate a zone egress token")
			}
			_, err = cmd.OutOrStdout().Write([]byte(token))
			return err
		},
	}
	cmd.Flags().StringVar(&ctx.args.zone, "zone", "", "name of the zone where egress resides")

	cmd.Flags().DurationVar(&ctx.args.validFor, "valid-for", 0, `how long the token will be valid (for example "24h")`)

	_ = cmd.MarkFlagRequired("valid-for")

	return cmd
}
