package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/kumahq/kuma/v2/tools/common/template"
	"github.com/kumahq/kuma/v2/tools/policy-gen/generator/pkg/parse"
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
			openApiOutPath := filepath.Join(filepath.Dir(policyPath), "rest.yaml")
			if err := template.PlainFileTemplate(localArgs.openAPITemplate, openApiOutPath, pconfig); err != nil {
				return err
			}
			schemaOutPath := filepath.Join(filepath.Dir(policyPath), "schema.yaml")
			if err := template.PlainFileTemplate(localArgs.jsonSchemaTemplate, schemaOutPath, pconfig); err != nil {
				return err
			}

			yqExec := exec.CommandContext(cmd.Context(), //nolint:gosec
				localArgs.yqBin, "e", "-i",
				fmt.Sprintf(`load(%q) as $crd
  | .properties *= (
      $crd.spec.versions[0].schema.openAPIV3Schema.properties
      | del(.apiVersion, .metadata, .kind)
    ) * {"type": {"enum": [$crd.spec.names.kind]}}
  | .description = $crd.spec.versions[0].schema.openAPIV3Schema.description
  | (.properties | select(has("status")).status) |= . + {"readOnly": true}`, crdPath),
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
