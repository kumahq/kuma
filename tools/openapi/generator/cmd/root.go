package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

type args struct {
	version string
}

func newRootCmd() *cobra.Command {
	rootArgs := &args{}

	cmd := &cobra.Command{
		Use:   "oapi-gen",
		Short: "Tool to generate OpenAPI for Kuma",
		Long:  "Tool to generate OpenAPI for Kuma",
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			// once command line flags have been parsed,
			// avoid printing usage instructions
			cmd.SilenceUsage = true
			return nil
		},
	}

	cmd.AddCommand(newKriPolicies(rootArgs))

	cmd.PersistentFlags().StringVar(&rootArgs.version, "version", "v1alpha1", "policy version")

	return cmd
}

func DefaultRootCmd() *cobra.Command {
	return newRootCmd()
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := DefaultRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}
