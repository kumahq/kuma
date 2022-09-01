package main

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
	"text/template"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var cfg = config{}

type config struct {
	name              string
	skipValidator     bool
	force             bool
	basePath          string
	gomodule          string
	version           string
	useCustomMatching bool
}

func (c config) policyPath() string {
	return path.Join(c.basePath, c.lowercase())
}

func (c config) lowercase() string {
	return strings.ToLower(c.name)
}

var rootCmd = &cobra.Command{
	Use:   "bootstrap",
	Short: "Bootstrap a new policy",
	RunE: func(cmd *cobra.Command, args []string) error {
		if cfg.name == "" {
			return errors.New("-name is required")
		}
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Bootstraping policy: %s at path %s\n", cfg.name, cfg.policyPath())
		if !cfg.force {
			_, err := os.Stat(cfg.policyPath())
			if err == nil {
				return fmt.Errorf("path %s already exists use -force to overwrite it", cfg.policyPath())
			}
		} else {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deleting old policy code\n")
			if err := os.RemoveAll(cfg.policyPath()); err != nil {
				return err
			}
		}
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Generating proto file\n")
		if err := generateProto(cfg); err != nil {
			return err
		}
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Generating plugin file\n")
		if err := generatePlugin(cfg); err != nil {
			return err
		}
		if err := exec.Command("make", "generate/policy/"+cfg.lowercase()).Run(); err != nil {
			return err
		}
		_, _ = cmd.OutOrStdout().Write([]byte(fmt.Sprintf(`
Successfully bootstrapped policy
regenerate auto generated files with: make generate/policy/%s

Useful files:
  - %s the proto definition
  - %s the validator
  - %s the plugin implementation
`,
			cfg.lowercase(),
			fmt.Sprintf("%s/api/%s/%s.proto", cfg.policyPath(), cfg.version, cfg.lowercase()),
			fmt.Sprintf("%s/api/%s/validator.go", cfg.policyPath(), cfg.version),
			fmt.Sprintf("%s/plugin/%s/plugin.go", cfg.policyPath(), cfg.version),
		)))
		return nil
	},
}

func generateProto(c config) error {
	apiPath := path.Join(c.policyPath(), "api", c.version)
	if err := os.MkdirAll(apiPath, os.ModePerm); err != nil {
		return err
	}
	f, err := os.Create(path.Join(apiPath, c.lowercase()+".proto"))
	if err != nil {
		return err
	}
	err = protoTemplate.Execute(f, map[string]interface{}{
		"name":              c.name,
		"nameLower":         c.lowercase(),
		"module":            path.Join(c.gomodule, c.basePath),
		"useCustomMatching": c.useCustomMatching,
		"version":           c.version,
	})
	if err != nil {
		return err
	}
	if c.skipValidator {
		return nil
	}
	f, err = os.Create(path.Join(apiPath, "validator.go"))
	if err != nil {
		return err
	}
	return validatorTemplate.Execute(f, map[string]interface{}{
		"name":              c.name,
		"version":           c.version,
		"useCustomMatching": c.useCustomMatching,
	})
}

func generatePlugin(c config) error {
	dir := path.Join(c.policyPath(), "plugin", c.version)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}
	f, err := os.Create(path.Join(dir, "plugin.go"))
	if err != nil {
		return err
	}
	return pluginTemplate.Execute(f, map[string]interface{}{
		"name":              c.name,
		"useCustomMatching": c.useCustomMatching,
		"version":           c.version,
		"package":           fmt.Sprintf("%s/%s/%s/api/%s", c.gomodule, c.basePath, c.lowercase(), c.version),
	})
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVar(&cfg.name, "name", "", "The name of the policy (UpperCamlCase)")
	rootCmd.Flags().StringVar(&cfg.basePath, "path", "pkg/plugins/policies", "Where to put the generated code")
	rootCmd.Flags().StringVar(&cfg.gomodule, "gomodule", "github.com/kumahq/kuma", "Where to put the generated code")
	rootCmd.Flags().StringVar(&cfg.version, "version", "v1alpha1", "The version to use")
	rootCmd.Flags().BoolVar(&cfg.skipValidator, "skip-validator", false, "don't generator a validator empty file")
	rootCmd.Flags().BoolVar(&cfg.force, "force", false, "Overwrite any existing code")
	rootCmd.Flags().BoolVar(&cfg.useCustomMatching, "use-custom-matching", false, "Whether to use standard policy type (regular matchers etc)")
}

var protoTemplate = template.Must(template.New("").Option("missingkey=error").Parse(
	`syntax = "proto3";

package kuma.plugins.policies.{{ .nameLower }}.{{ .version }};

import "mesh/options.proto";
option go_package = "{{ .module }}/{{.nameLower}}/api/{{ .version }}";

import "config.proto";

option (doc.config) = {
  type : Policy,
  name : "{{ .name }}",
  file_name : "{{ .nameLower }}"
};

// {{ .name }}
message {{ .name }} {
  option (kuma.mesh.policy) = {
    // Toggle this to have the policy registered or not in Kuma
    skip_registration : false,
  };

  // User defined fields
}
`))

var pluginTemplate = template.Must(template.New("").Option("missingkey=error").Parse(
	`package {{ .version }}

import (
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	{{- if not .useCustomMatching }}
	"github.com/kumahq/kuma/pkg/plugins/policies/matchers"
	api "{{ .package }}"
	{{- end}}
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
)

var _ core_plugins.PolicyPlugin = &plugin{}

type plugin struct {
}

func NewPlugin() core_plugins.Plugin {
	return &plugin{}
}

func (p plugin) MatchedPolicies(dataplane *core_mesh.DataplaneResource, resources xds_context.Resources) (core_xds.TypedMatchingPolicies, error) {
	{{- if .useCustomMatching }}
	panic("implement me")
	{{- else }}
	return matchers.MatchedPolicies(api.{{ .name }}Type, dataplane, resources)
	{{- end }}
}

func (p plugin) Apply(rs *core_xds.ResourceSet, ctx xds_context.Context, proxy *core_xds.Proxy) error {
	panic("implement me")
}
`))

var validatorTemplate = template.Must(template.New("").Option("missingkey=error").Parse(
	`package {{.version}}
import (
	"github.com/kumahq/kuma/pkg/core/validators"
)

func (r *{{.name}}Resource) validate() error {
	var err validators.ValidationError
	err.AddViolationAt(validators.RootedAt(""), "Not implemented!")
	return err.OrNil()
}
`))
