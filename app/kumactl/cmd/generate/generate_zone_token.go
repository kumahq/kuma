package generate

import (
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	kumactl_cmd "github.com/kumahq/kuma/app/kumactl/pkg/cmd"
	"github.com/kumahq/kuma/pkg/tokens/builtin/zone"
)

type generateZoneTokenContext struct {
	*kumactl_cmd.RootContext

	args *generateZoneTokenContextArgs
}

type generateZoneTokenContextArgs struct {
	zone     string
	scope    []string
	validFor time.Duration
}

func NewGenerateZoneTokenCmd(pctx *kumactl_cmd.RootContext) *cobra.Command {
	args := &generateZoneTokenContextArgs{}
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
			if err := validateArgs(ctx.args); err != nil {
				return err
			}

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
	// TODO (bartsmykla): update when Zone Token will be available for dataplanes and ingresses
	cmd.Flags().StringSliceVar(&ctx.args.scope, "scope", zone.FullScope, "scope of resources which the token will be able to identify (can be 'egress')")
	cmd.Flags().DurationVar(&ctx.args.validFor, "valid-for", 0, `how long the token will be valid (for example "24h")`)

	_ = cmd.MarkFlagRequired("valid-for")

	return cmd
}

func validateArgs(args *generateZoneTokenContextArgs) error {
	var unsupportedScopes []string

	for _, s := range args.scope {
		// TODO (bartsmykla): update when Zone Token will be available for dataplanes and ingresses
		if s != zone.EgressScope {
			unsupportedScopes = append(unsupportedScopes, s)
		}
	}

	if len(unsupportedScopes) > 0 {
		return errors.Errorf(
			"invalid --scope values: %+v (supported scopes: %+v)",
			unsupportedScopes,
			zone.FullScope,
		)
	}

	return nil
}
