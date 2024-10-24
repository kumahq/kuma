package generate

import (
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	kumactl_cmd "github.com/kumahq/kuma/app/kumactl/pkg/cmd"
	"github.com/kumahq/kuma/pkg/core/tokens"
	"github.com/kumahq/kuma/pkg/core/user"
	"github.com/kumahq/kuma/pkg/plugins/authn/api-server/tokens/issuer"
	"github.com/kumahq/kuma/pkg/plugins/authn/api-server/tokens/ws/client"
)

var NewHTTPUserTokenClient = client.NewHTTPUserTokenClient

type generateUserTokenCmd struct {
	name           string
	groups         []string
	validFor       time.Duration
	kid            string
	signingKeyPath string
}

func NewGenerateUserTokenCmd(pctx *kumactl_cmd.RootContext) *cobra.Command {
	var args generateUserTokenCmd
	cmd := &cobra.Command{
		Use:   "user-token",
		Short: "Generate User Token",
		Long:  `Generate User Token that is used to prove User identity.`,
		Example: `
Generate token using a control plane
$ kumactl generate user-token --name john.doe@example.com --group users --valid-for 24h

Generate token using offline signing key
$ kumactl generate user-token --name john.doe@example.com --group users --valid-for 24h --signing-key-path /keys/key.pem --kid 1
`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := pctx.BaseAPIServerClient()
			if err != nil {
				return err
			}

			var token string

			if args.signingKeyPath != "" {
				if args.kid == "" {
					return errors.New("--kid is required when --signing-key-path is used")
				}
				userTokenIssuer := issuer.NewUserTokenIssuer(tokens.NewTokenIssuer(tokens.NewFileSigningKeyManager(args.signingKeyPath, args.kid)))
				token, err = userTokenIssuer.Generate(cmd.Context(), user.User{
					Name:   args.name,
					Groups: args.groups,
				}, args.validFor)
				if err != nil {
					return err
				}
			} else {
				if args.kid != "" {
					return errors.New("--kid cannot be used when --signing-key-path is used")
				}
				tokenClient := NewHTTPUserTokenClient(client)
				token, err = tokenClient.Generate(args.name, args.groups, args.validFor)
				if err != nil {
					return errors.Wrap(err, "failed to generate a user token")
				}
			}

			_, err = cmd.OutOrStdout().Write([]byte(token))
			return err
		},
	}
	cmd.Flags().StringVar(&args.name, "name", "", "name of the user")
	cmd.Flags().StringSliceVar(&args.groups, "group", nil, "group of the user")
	cmd.Flags().DurationVar(&args.validFor, "valid-for", 0, `how long the token will be valid (for example "24h")`)
	cmd.Flags().StringVar(&args.signingKeyPath, "signing-key-path", "", "path to a file that contains private signing key. When specified, control plane won't be used to issue the token.")
	cmd.Flags().StringVar(&args.kid, "kid", "", "ID of the key that is used to issue a token. Required when --signing-key-path is used.")

	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("valid-for")
	return cmd
}
