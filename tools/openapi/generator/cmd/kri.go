package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"

	"github.com/kumahq/kuma/tools/common/save"
	"github.com/kumahq/kuma/tools/openapi/gotemplates"
	"github.com/kumahq/kuma/tools/policy-gen/generator/pkg/parse"
	"github.com/kumahq/kuma/tools/resource-gen/genutils"
)

func newKriPolicies(rootArgs *args) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kri",
		Short: "Generate KRI OpenAPI fragment",
		Long:  "Collect all policies and resources to render the KRI endpoint OpenAPI fragment for them.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			// locate policy plugin dirs under pkg/plugins/policies
			base := filepath.Join("pkg", "plugins", "policies")
			entries, err := os.ReadDir(base)
			if err != nil {
				return fmt.Errorf("failed to read policies directory %s: %w", base, err)
			}

			type resource struct {
				ResourceType string
				Path         string
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
				resources = append(resources, resource{
					ResourceType: pconfig.Name,
					Path:         "/specs/policies/" + strings.ToLower(pconfig.Name) + "/rest.yaml",
				})
			}

			KriResources := map[string]bool{
				"Dataplane":   true,
				"MeshGateway": true,
			}

			var types []protoreflect.MessageType
			protoregistry.GlobalTypes.RangeMessages(
				genutils.OnKumaResourceMessage("mesh", func(m protoreflect.MessageType) bool {
					types = append(types, m)
					return true
				}))

			// prepare template data matching the template expectations
			data := struct {
				Resources []resource
			}{
				Resources: resources,
			}

			for _, t := range types {
				resourceInfo := genutils.ToResourceInfo(t.Descriptor())
				_, ok := KriResources[resourceInfo.ResourceType]
				if ok {
					data.Resources = append(data.Resources, resource{
						ResourceType: resourceInfo.ResourceType,
						Path:         "/specs/protoresources/" + strings.ToLower(resourceInfo.ResourceType) + "/rest.yaml",
					})
				}
			}

			// render template
			tmpl := template.Must(template.New("kri").Funcs(save.FuncMap).Parse(gotemplates.KriEndpointTemplate))
			var outBuf bytes.Buffer
			if err := tmpl.Execute(&outBuf, data); err != nil {
				return fmt.Errorf("failed to execute KRI template: %w", err)
			}

			outDir := filepath.Join("api", "openapi", "specs", "kri")
			if err := os.MkdirAll(outDir, 0o755); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", outDir, err)
			}
			outPath := filepath.Join(outDir, "kri.yaml")
			if err := os.WriteFile(outPath, outBuf.Bytes(), 0o600); err != nil {
				return fmt.Errorf("failed to write %s: %w", outPath, err)
			}

			return nil
		},
	}

	return cmd
}
