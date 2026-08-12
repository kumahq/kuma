package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/spf13/cobra"

	commontemplate "github.com/kumahq/kuma/v3/tools/common/template"
	"github.com/kumahq/kuma/v3/tools/policy-gen/generator/pkg/parse"
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

			if pconfig.IsPolicy {
				for _, version := range versions {
					pluginDir := filepath.Join(rootArgs.pluginDir, "plugin", version)
					if _, err := os.Stat(pluginDir); err != nil {
						if os.IsNotExist(err) {
							continue
						}
						return err
					}

					matcherData := struct {
						Package     string
						Name        string
						GoModule    string
						ResourceDir string
						Version     string
					}{
						Package:     version,
						Name:        pconfig.Name,
						GoModule:    rootArgs.goModule,
						ResourceDir: rootArgs.pluginDir,
						Version:     version,
					}

					matcherOutPath := filepath.Join(pluginDir, "zz_generated.matcher.go")
					if err := commontemplate.GoTemplate(pluginMatcherTemplate, matcherData, matcherOutPath); err != nil {
						return err
					}
				}
			}

			if pconfig.SkipRegistration {
				return nil
			}

			data := struct {
				Package           string
				Versions          []string
				Name              string
				GoModule          string
				ResourceDir       string
				IsPolicy          bool
				RegisterGenerator bool
			}{
				Package:           strings.ToLower(pconfig.Name),
				Name:              pconfig.Name,
				Versions:          versions,
				GoModule:          rootArgs.goModule,
				ResourceDir:       rootArgs.pluginDir,
				IsPolicy:          pconfig.IsPolicy,
				RegisterGenerator: pconfig.RegisterGenerator,
			}

			outPath := filepath.Join(rootArgs.pluginDir, "zz_generated.plugin.go")
			if err := commontemplate.GoTemplate(pluginGoTemplate, data, outPath); err != nil {
				return err
			}

			return nil
		},
	}

	return cmd
}

var pluginGoTemplate = template.Must(template.New("plugin-go").Parse(`
package {{ .Package }}

{{ $pkg := .ResourceDir }}
{{ $name := .Name }}
{{ $gomodule := .GoModule }}
{{ $isPolicy := .IsPolicy }}
{{ $registerGenerator := .RegisterGenerator }}

import (
{{- if or $isPolicy $registerGenerator }}
	"github.com/kumahq/kuma/v3/pkg/core/plugins"
{{- end}}
	"github.com/kumahq/kuma/v3/pkg/core/resources/registry"
{{- range $idx, $version := .Versions}}
	api_{{ $version }} "{{ $gomodule }}/{{ $pkg }}/api/{{ $version }}"
	k8s_{{ $version }} "{{ $gomodule }}/{{ $pkg }}/k8s/{{ $version }}"
{{- if $isPolicy }}
	plugin_{{ $version }} "{{ $gomodule }}/{{ $pkg }}/plugin/{{ $version }}"
{{- end }}
{{- if $registerGenerator }}
	generator_{{ $version }} "{{ $gomodule }}/{{ $pkg }}/generator/{{ $version }}"
{{- end }}
{{- end}}
)

func InitPlugin() {
	{{- range $idx, $version := .Versions}}
	registry.AddKubeScheme(k8s_{{ $version }}.AddToScheme)
	registry.RegisterType(api_{{ $version }}.{{ $name }}ResourceTypeDescriptor)
{{- if $isPolicy }}
	plugins.Register(plugins.PluginName(api_{{ $version }}.{{ $name }}ResourceTypeDescriptor.KumactlArg), plugin_{{ $version }}.NewPlugin())
{{- end }}
{{- if $registerGenerator }}
	plugins.Register(plugins.PluginName(api_{{ $version }}.{{ $name }}ResourceTypeDescriptor.KumactlArg), generator_{{ $version }}.NewPlugin())
{{- end }}
	{{- end}}
}
`))

var pluginMatcherTemplate = template.Must(template.New("plugin-matcher-go").Parse(`
package {{ .Package }}

import (
	core_plugins "github.com/kumahq/kuma/v3/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/v3/pkg/core/xds"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/matchers"
	api "{{ .GoModule }}/{{ .ResourceDir }}/api/{{ .Version }}"
	xds_context "github.com/kumahq/kuma/v3/pkg/xds/context"
)

func (p plugin) MatchedPolicies(dataplane *core_mesh.DataplaneResource, resources xds_context.Resources, opts ...core_plugins.MatchedPoliciesOption) (core_xds.TypedMatchingPolicies, error) {
	return matchers.MatchedPolicies(api.{{ .Name }}Type, dataplane, resources, opts...)
}
`))
