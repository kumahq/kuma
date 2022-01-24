package generate

import (
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/thediveo/enumflag"

	kumactl_cmd "github.com/kumahq/kuma/app/kumactl/pkg/cmd"
	"github.com/kumahq/kuma/pkg/tokens/builtin/zone"
)

type generateZoneTokenContext struct {
	*kumactl_cmd.RootContext

	args *generateZoneTokenContextArgs
}

type generateZoneTokenContextArgs struct {
	zone     string
	scope    zone.Scope
	validFor time.Duration
}

func NewGenerateZoneTokenCmd(pctx *kumactl_cmd.RootContext) *cobra.Command {
	args := &generateZoneTokenContextArgs{scope: zone.FullScope}
	ctx := &generateZoneTokenContext{RootContext: pctx, args: args}

	cmd := &cobra.Command{
		Use:   "zone-token",
		Short: "Generate Zone Token",
		// TODO (bartsmykla): update descriptions when this token will be able to
		//  be used to prove identities of zone dataplanes and ingresses as well
		Long: `Generate Zone Token that is used to prove identity of Zone egresses.`,
		Example: `Generate token bound by zone
$ kumactl generate zone-token --zone zone-1 --valid-for 24h
$ kumactl generate zone-token --zone zone-1 --valid-for 24h --scope egress`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := pctx.CurrentZoneTokenClient()
			if err != nil {
				return errors.Wrap(err, "failed to create zone token client")
			}

			token, err := client.Generate(ctx.args.zone, ctx.args.scope, ctx.args.validFor)
			if err != nil {
				return errors.Wrap(err, "failed to generate a zone token")
			}

			_, err = cmd.OutOrStdout().Write([]byte(token))

			return err
		},
	}

	cmd.Flags().StringVar(&ctx.args.zone, "zone", "", "name of the zone where resides")
	cmd.Flags().Var(
		enumflag.NewSlice(&ctx.args.scope, "scope...", zone.ScopeItemsIds, enumflag.EnumCaseInsensitive),
		"scope",
		// TODO (bartsmykla): update usage when this token will be able to
		//  be used to prove identities of zone dataplanes and ingresses as well
		"scope of the token; can be 'egress'")
	cmd.Flags().DurationVar(&ctx.args.validFor, "valid-for", 0, `how long the token will be valid (for example "24h")`)

	_ = cmd.MarkFlagRequired("valid-for")

	return cmd
}
