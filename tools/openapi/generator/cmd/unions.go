package cmd

import (
	"github.com/spf13/cobra"

	"github.com/kumahq/kuma/v3/tools/openapi/unions"
)

func newUnions(_ *args) *cobra.Command {
	var spec, yqBin string
	cmd := &cobra.Command{
		Use:   "unions",
		Short: "Describe the discriminated unions of an OpenAPI spec",
		Long: "Describe the discriminated unions of an OpenAPI spec in place, as a oneOf of\n" +
			"named member schemas with a discriminator. Meant for the copies of the\n" +
			"generated specs that are bundled for API consumers.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return unions.Patch(cmd.Context(), spec, yqBin, cmd.ErrOrStderr())
		},
	}

	cmd.Flags().StringVar(&spec, "spec", "", "path to the OpenAPI document to patch in place")
	cmd.Flags().StringVar(&yqBin, "yq-bin", "yq", "path to a yq binary")
	_ = cmd.MarkFlagRequired("spec")

	return cmd
}
