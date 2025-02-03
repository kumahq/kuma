package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/spf13/cobra"

	"github.com/kumahq/kuma/tools/policy-gen/generator/pkg/parse"
	"github.com/kumahq/kuma/tools/policy-gen/generator/pkg/save"
)

func newOpenAPI(rootArgs *args) *cobra.Command {
	localArgs := struct {
		openAPITemplate    string
		jsonSchemaTemplate string
		yqBin              string
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
			crdPath := filepath.Join(rootArgs.pluginDir, "k8s", "crd", "kuma.io_"+strings.ToLower(pconfig.Plural)+".yaml")

			tmpl, err := template.ParseFiles(localArgs.openAPITemplate)
			if err != nil {
				return err
			}

			openApiOutPath := filepath.Join(filepath.Dir(policyPath), "rest.yaml")
			err = save.PlainTemplate(tmpl, pconfig, openApiOutPath)
			if err != nil {
				return err
			}
			schemaTmpl, err := template.ParseFiles(localArgs.jsonSchemaTemplate)
			if err != nil {
				return err
			}
			schemaOutPath := filepath.Join(filepath.Dir(policyPath), "schema.yaml")
			err = save.PlainTemplate(schemaTmpl, pconfig, schemaOutPath)
			if err != nil {
				return err
			}

			yqExec := exec.CommandContext(cmd.Context(), // nolint: gosec
				localArgs.yqBin, "e", "-i",
				fmt.Sprintf(`.properties *= (
    load(%q)
    | (
        .spec.versions[0]
        | .schema.openAPIV3Schema.properties
        | del(.apiVersion)
        | del(.metadata)
        | del(.kind)
      ) * {"type": {"enum": [.spec.names.kind]}}
  )
  | (.properties.status? ) |= . + {"readOnly": true}`, crdPath),
				schemaOutPath,
			)
			yqExec.Stderr = cmd.ErrOrStderr()
			return yqExec.Run()
		},
	}

	cmd.Flags().StringVar(&localArgs.openAPITemplate, "openapi-template-path", "", "path to the OpenAPI template file")
	cmd.Flags().StringVar(&localArgs.jsonSchemaTemplate, "jsonschema-template-path", "", "path to the jsonschema template file")
	cmd.Flags().StringVar(&localArgs.yqBin, "yq-bin", "", "path to a binary of yq")

	return cmd
}
