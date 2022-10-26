package cmd

import (
	"os"
	"path/filepath"
	"text/template"

	"github.com/spf13/cobra"

	"github.com/kumahq/kuma/tools/policy-gen/generator/pkg/parse"
	"github.com/kumahq/kuma/tools/policy-gen/generator/pkg/save"
)

func newGettersCmd(rootArgs *args) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "getters",
		Short: "Generate Get*() methods for each field",
		Long:  "Generate Get*() methods for each field.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			policyName := filepath.Base(rootArgs.pluginDir)
			policyPath := filepath.Join(rootArgs.pluginDir, "api", rootArgs.version, policyName+".go")
			if _, err := os.Stat(policyPath); err != nil {
				return err
			}

			getters, imports, err := parse.GettersAndImports(policyPath)
			if err != nil {
				return err
			}

			outPath := filepath.Join(filepath.Dir(policyPath), "zz_generated.getters.go")
			return save.GoTemplate(gettersTemplate, struct {
				Imports map[string]string
				Getters []*parse.Getter
				Package string
			}{
				Imports: imports,
				Getters: getters,
				Package: "v1alpha1",
			}, outPath)
		},
	}

	return cmd
}

var gettersTemplate = template.Must(template.New("getters").Parse(
	`package {{.Package}}
{{with .Imports}}
import (
  {{- range $key, $value := . -}}
  {{$key}} "{{$value}}"
  {{end -}}
)
{{end}}
{{range .Getters}}
{{if .PtrField}}
func ({{.ReceiverVar}} *{{.ReceiverType}}) Get{{.FieldName}}() *{{.FieldType}} {
  if {{.ReceiverVar}} == nil {
    return {{.ZeroValue}}
  }
  return {{.ReceiverVar}}.{{.FieldName}}
}
{{else}}
func ({{.ReceiverVar}} *{{.ReceiverType}}) Get{{.FieldName}}() {{.FieldType}} {
  if {{.ReceiverVar}} == nil {
    return {{.ZeroValue}}
  }
  return {{.ReceiverVar}}.{{.FieldName}}
}
{{end}}
{{end}}`))
