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

			// Create temp directory for intermediate files
			tmpDir, err := os.MkdirTemp("", "openapi-gen-*")
			if err != nil {
				return err
			}
			defer os.RemoveAll(tmpDir)

			crdPath := filepath.Join(rootArgs.pluginDir, "k8s", "crd", "kuma.io_"+strings.ToLower(pconfig.Plural)+".yaml")

			// Generate temporary files
			tmpRestPath := filepath.Join(tmpDir, "rest.yaml")
			if err := template.PlainFileTemplate(localArgs.openAPITemplate, tmpRestPath, pconfig); err != nil {
				return err
			}
			tmpSchemaPath := filepath.Join(tmpDir, "schema.yaml")
			if err := template.PlainFileTemplate(localArgs.jsonSchemaTemplate, tmpSchemaPath, pconfig); err != nil {
				return err
			}

			// Enrich schema with CRD information
			yqEnrichSchema := exec.CommandContext(cmd.Context(), //nolint:gosec
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
  | (.properties | select(has("status")).status) |= . + {"readOnly": true}`, crdPath),
				tmpSchemaPath,
			)
			yqEnrichSchema.Stderr = cmd.ErrOrStderr()
			if err := yqEnrichSchema.Run(); err != nil {
				return err
			}

			// Merge schema.yaml into rest.yaml by replacing the $ref
			yqMerge := exec.CommandContext(cmd.Context(), //nolint:gosec
				localArgs.yqBin, "e", "-i",
				fmt.Sprintf(`.components.schemas.%sItem = load(%q)`, pconfig.Name, tmpSchemaPath),
				tmpRestPath,
			)
			yqMerge.Stderr = cmd.ErrOrStderr()
			if err := yqMerge.Run(); err != nil {
				return err
			}

			// Write the merged file back to the original location as rest.yaml
			finalOutputPath := filepath.Join(filepath.Dir(policyPath), "rest.yaml")
			content, err := os.ReadFile(tmpRestPath)
			if err != nil {
				return err
			}
			if err := os.WriteFile(finalOutputPath, content, 0o600); err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&localArgs.openAPITemplate, "openapi-template-path", "", "path to the OpenAPI template file")
	cmd.Flags().StringVar(&localArgs.jsonSchemaTemplate, "jsonschema-template-path", "", "path to the jsonschema template file")
	cmd.Flags().StringVar(&localArgs.yqBin, "yq-bin", "", "path to a binary of yq")

	return cmd
}
