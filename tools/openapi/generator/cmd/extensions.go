package cmd

import (
	"github.com/spf13/cobra"

	"github.com/kumahq/kuma/v3/tools/openapi/extensions"
)

func newExtensions(_ *args) *cobra.Command {
	opts := extensions.Options{}
	cmd := &cobra.Command{
		Use:   "extensions",
		Short: "Document the registered resource extensions in an OpenAPI spec",
		Long: "Document the registered resource extensions in an OpenAPI spec.\n\n" +
			"What gets documented is decided by the packages this binary imports: an\n" +
			"extension is described where it is registered, and nowhere else. A build\n" +
			"that registers none leaves the document untouched.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			opts.Stderr = cmd.ErrOrStderr()
			return extensions.Generate(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringVar(&opts.Spec, "spec", "", "path to the OpenAPI document to patch in place")
	cmd.Flags().StringVar(&opts.WorkDir, "work-dir", "build/openapi-extensions",
		"directory, relative to the module root, for the throwaway package controller-gen reads")
	cmd.Flags().StringVar(&opts.ControllerGenBin, "controller-gen-bin", "controller-gen", "path to a controller-gen binary")
	cmd.Flags().StringVar(&opts.YqBin, "yq-bin", "yq", "path to a yq binary")
	_ = cmd.MarkFlagRequired("spec")

	return cmd
}
