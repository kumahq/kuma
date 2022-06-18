package generate

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	kumactl_cmd "github.com/kumahq/kuma/app/kumactl/pkg/cmd"
	"github.com/kumahq/kuma/pkg/plugins/authn/api-server/tokens/ws/client"
)

var NewHTTPUserTokenClient = client.NewHTTPUserTokenClient

type generateUserTokenCmd struct {
	name     string
	groups   []string
	validFor time.Duration
}

func NewGenerateUserTokenCmd(pctx *kumactl_cmd.RootContext) *cobra.Command {
	var args generateUserTokenCmd
	cmd := &cobra.Command{
		Use:   "user-token",
		Short: "Generate User Token",
		Long:  `Generate User Token that is used to prove User identity.`,
		Example: `
Generate token
$ kumactl generate user-token --name john.doe@example.com --group users --valid-for 24h
`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := pctx.BaseAPIServerClient()
			if err != nil {
				return err
			}

			tokenClient := NewHTTPUserTokenClient(client)
			token, err := tokenClient.Generate(args.name, args.groups, args.validFor)
			if err != nil {
				return fmt.Errorf("failed to generate a user token: %w", err)
			}
			_, err = cmd.OutOrStdout().Write([]byte(token))
			return err
		},
	}
	cmd.Flags().StringVar(&args.name, "name", "", "name of the user")
	_ = cmd.MarkFlagRequired("name")
	cmd.Flags().StringSliceVar(&args.groups, "group", nil, "group of the user")
	cmd.Flags().DurationVar(&args.validFor, "valid-for", 0, `how long the token will be valid (for example "24h")`)
	_ = cmd.MarkFlagRequired("valid-for")
	return cmd
}
