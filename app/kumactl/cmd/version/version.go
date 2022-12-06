package version

import (
	"context"
	"encoding/json"

	"github.com/spf13/cobra"

	kumactl_cmd "github.com/kumahq/kuma/app/kumactl/pkg/cmd"
	"github.com/kumahq/kuma/pkg/api-server/types"
	kuma_version "github.com/kumahq/kuma/pkg/version"
)

func NewCmd(pctx *kumactl_cmd.RootContext) *cobra.Command {
	args := struct {
		detailed bool
	}{}

	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print version",
		Long:  `Print version.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if args.detailed {
				cmd.Println(kuma_version.Build.FormatDetailedProductInfo())
			} else {
				cmd.Printf("Client: %s %s\n", kuma_version.Product, kuma_version.Build.Version)
			}

			var kumaCPInfo *types.IndexResponse

			client, err := pctx.CurrentApiClient()
			if err == nil {
				kumaCPInfo, err = client.GetVersion(context.Background())
			}

			if kumaCPInfo != nil {
				if args.detailed {
					s, err := json.MarshalIndent(kumaCPInfo, "", "   ")
					if err != nil {
						return err
					}
					cmd.Printf("Server: %s\n", string(s))
				} else {
					cmd.Printf("Server: %s %s\n", kumaCPInfo.Tagline, kumaCPInfo.Version)
				}
			} else {
				cmd.PrintErrf("Unable to connect to control plane: %v\n", err)
			}

			return nil
		},
	}

	cmd.PersistentFlags().BoolVarP(&args.detailed, "detailed", "a", false, "Print detailed version")

	return cmd
}
