package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"text/template"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	commontemplate "github.com/kumahq/kuma/v3/tools/common/template"
	"github.com/kumahq/kuma/v3/tools/openapi/gotemplates"
	"github.com/kumahq/kuma/v3/tools/policy-gen/generator/pkg/parse"
)

var ProcessProtoResources = true

type resource struct {
	ResourceType string
	Path         string
}

// coreResources lists the resources of the mesh API that publish a rest.yaml outside the
// policy directories. They were discovered by walking the protobuf registry until their
// specs became Go structs.
var coreResources = []resource{
	{ResourceType: "Dataplane", Path: "/specs/protoresources/dataplane/rest.yaml"},
	{ResourceType: "Mesh", Path: "/specs/protoresources/mesh/rest.yaml"},
}

func newKriPolicies(rootArgs *args) *cobra.Command {
	var errorSchema string
	cmd := &cobra.Command{
		Use:   "kri",
		Short: "Generate KRI OpenAPI fragment",
		Long:  "Collect all policies and resources to render the KRI endpoint OpenAPI fragment for them.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if errorSchema == "" {
				return errors.New("--error-schema must not be empty")
			}
			resources, err := gatherPlugins(rootArgs)
			if err != nil {
				return err
			}

			if ProcessProtoResources {
				resources = slices.Concat(resources, coreResources)
			}

			// sort resources deterministically by ResourceType
			sort.Slice(resources, func(i, j int) bool {
				return resources[i].ResourceType < resources[j].ResourceType
			})

			data := struct {
				Resources   []resource
				ErrorSchema string
			}{
				Resources:   resources,
				ErrorSchema: errorSchema,
			}

			// render template
			tmpl := template.Must(template.New("kri").Funcs(commontemplate.FuncMap).Parse(gotemplates.KriEndpointTemplate))
			var outBuf bytes.Buffer
			if err := tmpl.Execute(&outBuf, data); err != nil {
				return errors.Wrapf(err, "failed to execute KRI template")
			}

			outDir := filepath.Join("api", "openapi", "specs", "kri")
			if err := os.MkdirAll(outDir, 0o755); err != nil {
				return errors.Wrapf(err, "failed to create directory %s", outDir)
			}
			outPath := filepath.Join(outDir, "kri.yaml")
			if err := os.WriteFile(outPath, outBuf.Bytes(), 0o600); err != nil {
				return errors.Wrapf(err, "failed to write %s", outPath)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&errorSchema, "error-schema", commontemplate.DefaultOpenAPIErrorSchema, "OpenAPI document with the shared error responses, relative to the specs root")

	return cmd
}

func gatherPlugins(rootArgs *args) ([]resource, error) {
	var resources []resource
	// locate policy plugin dirs under pkg/plugins/policies
	base := filepath.Join("pkg", "plugins", "policies")
	entries, err := os.ReadDir(base)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read policies directory %s", base)
	}

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
			return nil, errors.Wrapf(err, "failed to parse %s", policyPath)
		}
		if pconfig.SkipRegistration || pconfig.ShortName == "" {
			continue
		}
		resources = append(resources, resource{
			ResourceType: pconfig.Name,
			Path:         "/specs/policies/" + strings.ToLower(pconfig.Name) + "/rest.yaml",
		})
	}
	return resources, nil
}
