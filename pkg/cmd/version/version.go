package version

import (
	"github.com/spf13/cobra"

	kuma_version "github.com/kumahq/kuma/pkg/version"
)

func NewVersionCmd() *cobra.Command {
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
				cmd.Printf("%s: %s\n", kuma_version.Product, kuma_version.Build.Version)
			}

			return nil
		},
	}
	// flags
	cmd.PersistentFlags().BoolVarP(&args.detailed, "detailed", "a", false, "Print detailed version")
	return cmd
}
