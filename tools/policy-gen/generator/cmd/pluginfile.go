package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/spf13/cobra"

	"github.com/kumahq/kuma/tools/policy-gen/generator/pkg/parse"
	"github.com/kumahq/kuma/tools/policy-gen/generator/pkg/save"
)

func newPluginFile(rootArgs *args) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plugin-file",
		Short: "Generate a plugin.go file for the policy",
		Long:  "Generate a plugin.go file for the policy.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			policyName := filepath.Base(rootArgs.pluginDir)
			apiDir := filepath.Join(rootArgs.pluginDir, "api")
			policyPath := filepath.Join(apiDir, rootArgs.version, policyName+".go")
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

			files, err := os.ReadDir(apiDir)
			if err != nil {
				return err
			}

			versions := []string{}
			for _, file := range files {
				if file.IsDir() {
					versions = append(versions, file.Name())
				}
			}

			outPath := filepath.Join(rootArgs.pluginDir, "zz_generated.plugin.go")
			return save.GoTemplate(pluginGoTemplate, struct {
				Package  string
				Versions []string
				Name     string
				GoModule string
			}{
				Package:  strings.ToLower(pconfig.Name),
				Name:     pconfig.Name,
				Versions: versions,
				GoModule: rootArgs.goModule,
			}, outPath)
		},
	}

	return cmd
}

var pluginGoTemplate = template.Must(template.New("plugin-go").Parse(`
package {{ .Package }}

{{ $pkg := .Package }}
{{ $name := .Name }}
{{ $gomodule := .GoModule }}

import (
	"github.com/kumahq/kuma/pkg/plugins/policies/core"
{{- range $idx, $version := .Versions}}
	api_{{ $version }} "{{ $gomodule }}/pkg/plugins/policies/{{ $pkg }}/api/{{ $version }}"
	k8s_{{ $version }} "{{ $gomodule }}/pkg/plugins/policies/{{ $pkg }}/k8s/{{ $version }}"
	plugin_{{ $version }} "{{ $gomodule }}/pkg/plugins/policies/{{ $pkg }}/plugin/{{ $version }}"
{{- end}}
)

func init() {
	{{- range $idx, $version := .Versions}}
	core.Register(
		api_{{ $version }}.{{ $name }}ResourceTypeDescriptor,
		k8s_{{ $version }}.AddToScheme,
		plugin_{{ $version }}.NewPlugin(),
	)
	{{- end}}
}
`))
