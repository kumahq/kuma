package cmd

import (
	"os"
	"path/filepath"
	"text/template"

	"github.com/spf13/cobra"

	"github.com/kumahq/kuma/tools/policy-gen/generator/pkg/parse"
	"github.com/kumahq/kuma/tools/policy-gen/generator/pkg/save"
)

func newHelpers(rootArgs *args) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "helpers",
		Short: "Generate helper funcs for the policy",
		Long:  "Generate helper funcs for the policy.",
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

			outPath := filepath.Join(filepath.Dir(policyPath), "zz_generated.helpers.go")
			return save.GoTemplate(helpersTemplate, map[string]interface{}{
				"name":                  pconfig.Name,
				"version":               pconfig.Package,
				"generateTo":            pconfig.HasTo,
				"generateFrom":          pconfig.HasFrom,
				"generateGetPolicyItem": !pconfig.HasFrom && !pconfig.HasTo,
			}, outPath)
		},
	}

	return cmd
}

var helpersTemplate = template.Must(template.New("missingkey=error").Parse(
	`
// Generated by tools/resource-gen.
// Run "make generate" to update this file.

// nolint:whitespace
package {{.version}}

import (
{{ if .generateGetPolicyItem}}
	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
{{- end}}
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
)

{{ if .generateFrom }}

func (x *From) GetDefaultAsInterface() interface{} {
	return x.Default
}

func (x *{{.name}}) GetFromList() []core_xds.PolicyItem {
	var result []core_xds.PolicyItem
	for _, item := range x.From {
		result = append(result, item)
	}
	return result
}
{{- end }}
{{ if .generateTo }}

func (x *To) GetDefaultAsInterface() interface{} {
	return x.Default
}

func (x *{{.name}}) GetToList() []core_xds.PolicyItem {
	var result []core_xds.PolicyItem
	for _, item := range x.To {
		result = append(result, item)
	}
	return result
}
{{- end }}

{{ if .generateGetPolicyItem}}

func (x *{{.name}}) GetDefaultAsInterface() interface{} {
	return x.Default
}

func (x *{{.name}}) GetPolicyItem() core_xds.PolicyItem {
	return &policyItem{
		{{.name}}: x,
	}
}

// policyItem is an auxiliary struct with the implementation of the GetTargetRef() to always return empty result
type policyItem struct {
	*{{.name}}
}

var _ core_xds.PolicyItem = &policyItem{}

func (p *policyItem) GetTargetRef() *common_api.TargetRef {
	return &common_api.TargetRef{Kind: common_api.Mesh}
}
{{- end }}
`))
