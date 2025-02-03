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
	generateTargetRef bool
	generateTo        bool
	generateFrom      bool
	isPolicy          bool
	hasStatus         bool
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
		if err := generateType(cfg); err != nil {
			return err
		}
		if cfg.isPolicy {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Generating plugin file\n")
			if err := generatePlugin(cfg); err != nil {
				return err
			}
		}
		path := fmt.Sprintf("generate/policy/%s", cfg.lowercase())
		if err := exec.Command("make", path, "POLICIES_DIR="+cfg.basePath).Run(); err != nil {
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

func generateType(c config) error {
	apiPath := path.Join(c.policyPath(), "api", c.version)
	if err := os.MkdirAll(apiPath, os.ModePerm); err != nil {
		return err
	}
	f, err := os.Create(path.Join(apiPath, c.lowercase()+".go"))
	if err != nil {
		return err
	}
	err = typeTemplate.Execute(f, map[string]interface{}{
		"name":              c.name,
		"nameLower":         c.lowercase(),
		"module":            path.Join(c.gomodule, c.basePath),
		"version":           c.version,
		"generateTargetRef": c.generateTargetRef,
		"generateTo":        c.generateTo,
		"generateFrom":      c.generateFrom,
		"isPolicy":          c.isPolicy,
		"hasStatus":         c.hasStatus,
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
		"generateTargetRef": c.generateTargetRef,
		"generateTo":        c.generateTo,
		"generateFrom":      c.generateFrom,
		"isPolicy":          c.isPolicy,
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
		"version":           c.version,
		"package":           fmt.Sprintf("%s/%s/%s/api/%s", c.gomodule, c.basePath, c.lowercase(), c.version),
		"generateTargetRef": c.generateTargetRef,
		"generateTo":        c.generateTo,
		"generateFrom":      c.generateFrom,
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
	rootCmd.Flags().BoolVar(&cfg.generateTargetRef, "generate-target-ref", true, "Generate top-level TargetRef for dataplane proxy matching")
	rootCmd.Flags().BoolVar(&cfg.generateTo, "generate-to", false, "Generate 'to' array for outgoing traffic configuration")
	rootCmd.Flags().BoolVar(&cfg.generateFrom, "generate-from", false, "Generate 'from' array for incoming traffic configuration")
	rootCmd.Flags().BoolVar(&cfg.isPolicy, "is-policy", false, "Resource is a policy")
	rootCmd.Flags().BoolVar(&cfg.hasStatus, "has-status", false, "Resource has a status field")
}

var typeTemplate = template.Must(template.New("").Option("missingkey=error").Parse(
	`// +kubebuilder:object:generate=true
package {{ .version }}
{{- if or .generateTargetRef (or .generateTo .generateFrom) }}

import (
	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
)
{{- end}}

// {{ .name }}
// +kuma:policy:skip_registration=true
// +kuma:policy:is_policy={{ .isPolicy }}
type {{ .name }} struct {
	{{- if .generateTargetRef }}
	// TargetRef is a reference to the resource the policy takes an effect on.
	// The resource could be either a real store object or virtual resource
	// defined in-place.
	TargetRef common_api.TargetRef` + " `json:\"targetRef\"`" + `
	{{- end }}
	{{- if .generateTo }}
	// To list makes a match between the consumed services and corresponding configurations
	To []To` + " `json:\"to,omitempty\"`" + `
	{{- end}}
	{{- if .generateFrom }}
	// From list makes a match between clients and corresponding configurations
	From []From` + " `json:\"from,omitempty\"`" + `
	{{- end}}
}
{{- if .generateTo }}

type To struct {
	// TargetRef is a reference to the resource that represents a group of
	// destinations.
	TargetRef common_api.TargetRef` + " `json:\"targetRef\"`" + `
	// Default is a configuration specific to the group of destinations referenced in
	// 'targetRef'
	Default Conf` + " `json:\"default,omitempty\"`" + `
}

{{- end}}
{{- if .generateFrom }}

type From struct {
	// TargetRef is a reference to the resource that represents a group of
	// clients.
	TargetRef common_api.TargetRef` + " `json:\"targetRef\"`" + `
	// Default is a configuration specific to the group of clients referenced in
	// 'targetRef'
	Default Conf` + " `json:\"default,omitempty\"`" + `
}

{{- end}}

{{- if .isPolicy }}
type Conf struct {
	// TODO add configuration fields
}
{{- end}}

{{- if .hasStatus }}
type {{ .name }}Status struct {
	// TODO add status fields
}
{{- end}}
`))

var pluginTemplate = template.Must(template.New("").Option("missingkey=error").Parse(
	`package {{ .version }}

import (
	"github.com/kumahq/kuma/pkg/core"

	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	{{- if .generateTargetRef }}
	"github.com/kumahq/kuma/pkg/plugins/policies/core/matchers"
	api "{{ .package }}"
	{{- end}}
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
)

var _ core_plugins.PolicyPlugin = &plugin{}
var log = core.Log.WithName("{{.name}}")

type plugin struct {
}

func NewPlugin() core_plugins.Plugin {
	return &plugin{}
}

func (p plugin) MatchedPolicies(dataplane *core_mesh.DataplaneResource, resources xds_context.Resources, opts ...core_plugins.MatchedPoliciesOption) (core_xds.TypedMatchingPolicies, error) {	{{- if not .generateTargetRef }}
	panic("implement me")
	{{- else }}
	return matchers.MatchedPolicies(api.{{ .name }}Type, dataplane, resources, opts...), nil
	{{- end }}
}

func (p plugin) Apply(rs *core_xds.ResourceSet, ctx xds_context.Context, proxy *core_xds.Proxy) error {
	log.Info("apply is not implemented")
	return nil
}
`))

var validatorTemplate = template.Must(template.New("").Option("missingkey=error").Parse(
	`package {{.version}}

import (
	{{- if or .generateTargetRef (or .generateTo .generateFrom) }}
	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	{{- end}}
	"github.com/kumahq/kuma/pkg/core/validators"
)

func (r *{{.name}}Resource) validate() error {
	var verr validators.ValidationError
	path := validators.RootedAt("spec")
	{{- if .generateTargetRef }}
	verr.AddErrorAt(path.Field("targetRef"), validateTop(r.Spec.TargetRef))
	{{- end }}
	{{- if and .generateTo .generateFrom }}
	if len(r.Spec.To) == 0 && len(r.Spec.From) == 0 {
		verr.AddViolationAt(path, "at least one of 'from', 'to' has to be defined")
	}
	{{- else if .generateTo }}
	if len(r.Spec.To) == 0 {
		verr.AddViolationAt(path.Field("to"), "needs at least one item")
	}
	{{- else if .generateFrom }}
	if len(r.Spec.From) == 0 {
		verr.AddViolationAt(path.Field("from"), "needs at least one item")
	}
	{{- end }}
	{{- if .generateTo }}
	verr.AddErrorAt(path, validateTo(r.Spec.To))
	{{- end }}
	{{- if .generateFrom }}
	verr.AddErrorAt(path, validateFrom(r.Spec.From))
	{{- end }}
	return verr.OrNil()
}
{{- if .generateTargetRef }}

func validateTop(targetRef common_api.TargetRef) validators.ValidationError {
	targetRefErr := mesh.ValidateTargetRef(targetRef, &mesh.ValidateTargetRefOpts{
		SupportedKinds: []common_api.TargetRefKind{
			// TODO add supported TargetRef kinds for this policy
		},
	})
	return targetRefErr
}

{{- end }}
{{- if .generateFrom }}

func validateFrom(from []From) validators.ValidationError {
	var verr validators.ValidationError
	for idx, fromItem := range from {
		path := validators.RootedAt("from").Index(idx)
		verr.AddErrorAt(path.Field("targetRef"), mesh.ValidateTargetRef(fromItem.TargetRef, &mesh.ValidateTargetRefOpts{
			SupportedKinds: []common_api.TargetRefKind{
				// TODO add supported TargetRef for 'from' item
			},
		}))
		verr.AddErrorAt(path.Field("default"), validateDefault(fromItem.Default))
	}
	return verr
}

{{- end }}
{{- if .generateTo }}

func validateTo(to []To) validators.ValidationError {
	var verr validators.ValidationError
	for idx, toItem := range to {
		path := validators.RootedAt("to").Index(idx)
		verr.AddErrorAt(path.Field("targetRef"), mesh.ValidateTargetRef(toItem.TargetRef, &mesh.ValidateTargetRefOpts{
			SupportedKinds: []common_api.TargetRefKind{
				// TODO add supported TargetRef for 'to' item
			},
		}))
		verr.AddErrorAt(path.Field("default"), validateDefault(toItem.Default))
	}
	return verr
}
{{- end }}

{{- if .isPolicy }}
func validateDefault(conf Conf) validators.ValidationError {
	var verr validators.ValidationError
	// TODO add default conf validation
	return verr
}
{{- end }}
`))
