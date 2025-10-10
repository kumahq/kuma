package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/kumahq/kuma/tools/common/save"
	"github.com/kumahq/kuma/tools/openapi/gotemplates"
	"github.com/kumahq/kuma/tools/policy-gen/generator/pkg/parse"
	"github.com/spf13/cobra"
)

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
