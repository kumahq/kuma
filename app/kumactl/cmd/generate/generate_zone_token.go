package generate

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	kumactl_cmd "github.com/kumahq/kuma/app/kumactl/pkg/cmd"
	"github.com/kumahq/kuma/pkg/core/tokens"
	"github.com/kumahq/kuma/pkg/tokens/builtin/zone"
)

type generateZoneTokenContext struct {
	*kumactl_cmd.RootContext

	args *generateZoneTokenContextArgs
}

type generateZoneTokenContextArgs struct {
	zone           string
	scope          []string
	validFor       time.Duration
	kid            string
	signingKeyPath string
}

func NewGenerateZoneTokenCmd(pctx *kumactl_cmd.RootContext) *cobra.Command {
	args := &generateZoneTokenContextArgs{}
	ctx := &generateZoneTokenContext{RootContext: pctx, args: args}

	cmd := &cobra.Command{
		Use:   "zone-token",
		Short: "Generate Zone Token",
		Long:  `Generate Zone Token that is used to prove identity of zone components (Zone Ingress, Zone Egress).`,
		Example: `
Generate token using a control plane
$ kumactl generate zone-token --zone zone-1 --valid-for 24h
$ kumactl generate zone-token --zone zone-1 --valid-for 24h --scope egress
$ kumactl generate zone-token --zone zone-1 --valid-for 24h --scope ingress
$ kumactl generate zone-token --zone zone-1 --valid-for 24h --scope ingress --scope egress

Generate token using offline signing key
$ kumactl generate zone-token --zone zone-1 --valid-for 24h --signing-key-path /keys/key.pem --kid 1`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := validateArgs(ctx.args); err != nil {
				return err
			}

			var token string

			if args.signingKeyPath != "" {
				if args.kid == "" {
					return errors.New("--kid is required when --signing-key-path is used")
				}
				issuer := zone.NewTokenIssuer(tokens.NewTokenIssuer(tokens.NewFileSigningKeyManager(args.signingKeyPath, args.kid)))
				t, err := issuer.Generate(cmd.Context(), zone.Identity{
					Zone:  args.zone,
					Scope: args.scope,
				}, args.validFor)
				if err != nil {
					return err
				}
				token = t
			} else {
				if args.kid != "" {
					return errors.New("--kid cannot be used when --signing-key-path is used")
				}
				client, err := pctx.CurrentZoneTokenClient()
				if err != nil {
					return errors.Wrap(err, "failed to create zone token client")
				}

				token, err = client.Generate(ctx.args.zone, ctx.args.scope, ctx.args.validFor)
				if err != nil {
					return errors.Wrap(err, "failed to generate a zone token")
				}
			}

			_, err := cmd.OutOrStdout().Write([]byte(token))
			return err
		},
	}

	cmd.Flags().StringVar(&ctx.args.zone, "zone", "", "name of the zone where resides")
	cmd.Flags().StringSliceVar(&ctx.args.scope, "scope", zone.FullScope, fmt.Sprintf("scope of resources which the token will be able to identify (can be: %v)", zone.FullScope))
	cmd.Flags().DurationVar(&ctx.args.validFor, "valid-for", 0, `how long the token will be valid (for example "24h")`)
	cmd.Flags().StringVar(&args.signingKeyPath, "signing-key-path", "", "path to a file that contains private signing key. When specified, control plane won't be used to issue the token.")
	cmd.Flags().StringVar(&args.kid, "kid", "", "ID of the key that is used to issue a token. Required when --signing-key-path is used.")

	_ = cmd.MarkFlagRequired("zone")
	_ = cmd.MarkFlagRequired("valid-for")

	return cmd
}

func validateArgs(args *generateZoneTokenContextArgs) error {
	var unsupportedScopes []string

	for _, s := range args.scope {
		if !zone.InScope(zone.FullScope, s) {
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
