package cmd

import (
	"os"
	"path/filepath"
	"text/template"

	"github.com/spf13/cobra"

	"github.com/kumahq/kuma/tools/policy-gen/generator/pkg/parse"
	"github.com/kumahq/kuma/tools/policy-gen/generator/pkg/save"
)

func newOpenAPI(rootArgs *args) *cobra.Command {
	localArgs := struct {
		openAPITemplate string
	}{}
	cmd := &cobra.Command{
		Use:   "openapi",
		Short: "Generate an OpenAPI schema for the policy REST",
		Long:  "Generate an OpenAPI schema for the policy REST.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			policyName := filepath.Base(rootArgs.pluginDir)
			policyPath := filepath.Join(rootArgs.pluginDir, "api", rootArgs.version, policyName+".go")
			if _, err := os.Stat(policyPath); err != nil {
				return err
			}

			pconfig, err := parse.Policy(policyPath)
			if err != nil {
				return err
			}
			if pconfig.SkipRegistration {
				return nil
			}

			tmpl, err := template.ParseFiles(localArgs.openAPITemplate)
			if err != nil {
				return err
			}

			outPath := filepath.Join(filepath.Dir(policyPath), "rest.yaml")
			return save.PlainTemplate(tmpl, pconfig, outPath)
		},
	}

	cmd.Flags().StringVar(&localArgs.openAPITemplate, "openapi-template-path", "", "path to the OpenAPI template file")

	return cmd
}
