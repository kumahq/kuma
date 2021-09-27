package version

import (
	"context"
	"fmt"

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
			buildInfo := kuma_version.Build

			if args.detailed {
				cmd.Println(fmt.Sprintf("Product:    %s", kuma_version.Product))
				cmd.Println(fmt.Sprintf("Version:    %s", buildInfo.Version))
				cmd.Println(fmt.Sprintf("Git Tag:    %s", buildInfo.GitTag))
				cmd.Println(fmt.Sprintf("Git Commit: %s", buildInfo.GitCommit))
				cmd.Println(fmt.Sprintf("Build Date: %s", buildInfo.BuildDate))
			} else {
				cmd.Printf("Client: %s %s\n", kuma_version.Product, buildInfo.Version)
			}

			var kumaCPInfo *types.IndexResponse

			client, err := pctx.CurrentApiClient()
			if err == nil {
				kumaCPInfo, err = client.GetVersion(context.Background())
			}

			if kumaCPInfo != nil {
				cmd.Printf("Server: %s %s\n", kumaCPInfo.Tagline, kumaCPInfo.Version)
			} else {
				cmd.PrintErrf("Unable to connect to control plane: %v\n", err)
			}

			return nil
		},
	}

	cmd.PersistentFlags().BoolVarP(&args.detailed, "detailed", "a", false, "Print detailed version")

	return cmd
}
