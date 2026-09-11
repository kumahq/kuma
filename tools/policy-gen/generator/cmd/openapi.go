package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/kumahq/kuma/v3/tools/common/template"
	"github.com/kumahq/kuma/v3/tools/openapi/unions"
	"github.com/kumahq/kuma/v3/tools/policy-gen/generator/pkg/parse"
)

func newOpenAPI(rootArgs *args) *cobra.Command {
	localArgs := struct {
		openAPITemplate    string
		jsonSchemaTemplate string
		yqBin              string
		errorSchema        string
	}{}
	cmd := &cobra.Command{
		Use:   "openapi",
		Short: "Generate an OpenAPI schema for the policy REST",
		Long:  "Generate an OpenAPI schema for the policy REST.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if localArgs.errorSchema == "" {
				return errors.New("--error-schema must not be empty")
			}
			pluginDir := filepath.Clean(rootArgs.pluginDir)
			policyName := filepath.Base(pluginDir)
			policyPath := filepath.Join(pluginDir, "api", rootArgs.version, policyName+".go")
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

			crdPath := filepath.Join(pluginDir, "k8s", "crd", "kuma.io_"+strings.ToLower(pconfig.Plural)+".yaml")
			// A resource that ships an opaque Kubernetes definition keeps a typed schema
			// beside it purely so the REST API reference can describe its spec.
			if schemaPath := filepath.Join(pluginDir, "k8s", "schema", "kuma.io_"+strings.ToLower(pconfig.Plural)+".yaml"); fileExists(schemaPath) {
				crdPath = schemaPath
			}

			// Generate temporary files
			tmpRestPath := filepath.Join(tmpDir, "rest.yaml")
			restData := struct {
				parse.PolicyConfig
				ErrorSchema string
			}{PolicyConfig: pconfig, ErrorSchema: localArgs.errorSchema}
			if err := template.PlainFileTemplate(localArgs.openAPITemplate, tmpRestPath, restData); err != nil {
				return err
			}
			tmpSchemaPath := filepath.Join(tmpDir, "schema.yaml")
			if err := template.PlainFileTemplate(localArgs.jsonSchemaTemplate, tmpSchemaPath, pconfig); err != nil {
				return err
			}

			// Describe discriminated unions with a oneOf, which controller-gen
			// cannot express, so consumers do not have to infer which property a
			// given `type` selects. Appended to the enrichment expression so the
			// generated file keeps its key order.
			crdProperties, err := unions.CRDProperties(crdPath)
			if err != nil {
				return err
			}
			// The enrichment merges the CRD properties into `.properties`, so a
			// union at `spec.foo` in the CRD lands at `.properties.spec.foo`.
			unionAssignments, err := unions.Assignments(crdProperties, []string{"properties"})
			if err != nil {
				return err
			}
			if unionAssignments != "" {
				unionAssignments = "\n  | " + unionAssignments
			}

			// A plugin originated resource nests its spec under "spec" in the REST API,
			// a core resource inlines it, so the spec's own properties are hoisted to
			// the top level and the "spec" key itself is dropped.
			specProperties := "$crd.spec.versions[0].schema.openAPIV3Schema.properties"
			if !pconfig.PluginOriginated {
				specProperties = `(($crd.spec.versions[0].schema.openAPIV3Schema.properties.spec.properties // {})` +
					` * ($crd.spec.versions[0].schema.openAPIV3Schema.properties | del(.spec)))`
			}

			// Enrich schema with CRD information
			yqEnrichSchema := exec.CommandContext(cmd.Context(), //nolint:gosec
				localArgs.yqBin, "e", "-i",
				fmt.Sprintf(`load(%q) as $crd
  | .properties *= (
      %s
      | del(.apiVersion, .metadata, .kind)
    ) * {"type": {"enum": [$crd.spec.names.kind]}}
  | .description = $crd.spec.versions[0].schema.openAPIV3Schema.description
  | (.properties | select(has("status")).status) |= . + {"readOnly": true}%s`, crdPath, specProperties, unionAssignments),
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
			if err := os.WriteFile(finalOutputPath, content, 0o600); err != nil { //nolint:gosec // G703: path is derived from validated plugin directory and controlled generation output
				return err
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&localArgs.openAPITemplate, "openapi-template-path", "", "path to the OpenAPI template file")
	cmd.Flags().StringVar(&localArgs.jsonSchemaTemplate, "jsonschema-template-path", "", "path to the jsonschema template file")
	cmd.Flags().StringVar(&localArgs.yqBin, "yq-bin", "", "path to a binary of yq")
	cmd.Flags().StringVar(&localArgs.errorSchema, "error-schema", template.DefaultOpenAPIErrorSchema, "OpenAPI document with the shared error responses, relative to the specs root")

	return cmd
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
