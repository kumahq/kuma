package cmd

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/spf13/cobra"

	"github.com/kumahq/kuma/tools/openapi/gotemplates"
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
			openApiOutPath := filepath.Join(filepath.Dir(policyPath), "rest.yaml")
			if err := save.PlainFileTemplate(localArgs.openAPITemplate, openApiOutPath, pconfig); err != nil {
				return err
			}
			schemaOutPath := filepath.Join(filepath.Dir(policyPath), "schema.yaml")
			if err := save.PlainFileTemplate(localArgs.jsonSchemaTemplate, schemaOutPath, pconfig); err != nil {
				return err
			}

			yqExec := exec.CommandContext(cmd.Context(), //nolint:gosec
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

// newKriPolicies generates a KRI endpoint for all policies found under pkg/plugins/policies
func newKriPolicies(rootArgs *args) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kri-policies",
		Short: "Generate KRI OpenAPI fragment for all policies",
		Long:  "Collect all policies and render the KRI endpoint OpenAPI fragment for them.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			// locate policy plugin dirs under pkg/plugins/policies
			base := filepath.Join("pkg", "plugins", "policies")
			entries, err := os.ReadDir(base)
			if err != nil {
				return fmt.Errorf("failed to read policies directory %s: %w", base, err)
			}

			type resource struct {
				ResourceType string
			}
			var resources []resource

			for _, e := range entries {
				if !e.IsDir() {
					continue
				}
				policyDir := filepath.Join(base, e.Name())
				// assume api/<version>/<policyName>.go
				policyPath := filepath.Join(policyDir, "api", rootArgs.version, e.Name()+".go")
				if _, err := os.Stat(policyPath); err != nil {
					// skip missing policy files
					continue
				}
				pconfig, err := parse.Policy(policyPath)
				if err != nil {
					return fmt.Errorf("failed to parse %s: %w", policyPath, err)
				}
				if pconfig.SkipRegistration {
					continue
				}
				resources = append(resources, resource{ResourceType: pconfig.Name})
			}

			// prepare template data matching the template expectations
			data := struct {
				Resources []resource
			}{
				Resources: resources,
			}

			// render template
			tmpl := template.Must(template.New("kri").Funcs(save.FuncMap).Parse(gotemplates.KriEndpointTemplate))
			var outBuf bytes.Buffer
			if err := tmpl.Execute(&outBuf, data); err != nil {
				return fmt.Errorf("failed to execute KRI template: %w", err)
			}

			outDir := filepath.Join("api", "openapi", "specs", "common")
			if err := os.MkdirAll(outDir, 0o755); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", outDir, err)
			}
			outPath := filepath.Join(outDir, "kri-policies.yaml")
			if err := os.WriteFile(outPath, outBuf.Bytes(), 0o600); err != nil {
				return fmt.Errorf("failed to write %s: %w", outPath, err)
			}

			return nil
		},
	}

	return cmd
}
