package cli

import (
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	kumactl_client "github.com/kumahq/kuma/app/kumactl/pkg/client"
	kumactl_cmd "github.com/kumahq/kuma/app/kumactl/pkg/cmd"
	"github.com/kumahq/kuma/pkg/plugins/authn/api-server/tokens/ws/client"
)

var NewHTTPUserTokenClient = client.NewHTTPUserTokenClient

type generateUserTokenCmd struct {
	name     string
	group    string
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
$ kumactl generate user-token --name john.doe@acme.org --group users 
`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cp, err := pctx.CurrentControlPlane()
			if err != nil {
				return err
			}
			apiClient, err := kumactl_client.ApiServerClient(cp.Coordinates.ApiServer)
			if err != nil {
				return err
			}

			tokenClient := NewHTTPUserTokenClient(apiClient)
			token, err := tokenClient.Generate(args.name, args.group, args.validFor)
			if err != nil {
				return errors.Wrap(err, "failed to generate a user token")
			}
			_, err = cmd.OutOrStdout().Write([]byte(token))
			return err
		},
	}
	cmd.Flags().StringVar(&args.name, "name", "", "name of the user")
	_ = cmd.MarkFlagRequired("name")
	cmd.Flags().StringVar(&args.group, "group", "", "group of the user")
	cmd.Flags().DurationVar(&args.validFor, "valid-for", 0, `how long the token will be valid (for example "24h"). If 0, then token has no expiration time`)
	return cmd
}
