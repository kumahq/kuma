package cmd

import (
	"bytes"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"

	"github.com/kumahq/kuma/tools/common/save"
	commontypes "github.com/kumahq/kuma/tools/common/types"
	"github.com/kumahq/kuma/tools/openapi/gotemplates"
	"github.com/kumahq/kuma/tools/policy-gen/generator/pkg/parse"
	"github.com/kumahq/kuma/tools/resource-gen/genutils"
)

var ProcessProtoResources = true

type resource struct {
	ResourceType string
	Path         string
}

func newKriPolicies(rootArgs *args) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kri",
		Short: "Generate KRI OpenAPI fragment",
		Long:  "Collect all policies and resources to render the KRI endpoint OpenAPI fragment for them.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			resources, err := gatherPlugins(rootArgs)
			if err != nil {
				return err
			}

			if ProcessProtoResources {
				protoResources := gatherProtoResources()
				resources = append(resources, protoResources...)
			}

			data := struct {
				Resources []resource
			}{
				Resources: resources,
			}

			// render template
			tmpl := template.Must(template.New("kri").Funcs(save.FuncMap).Parse(gotemplates.KriEndpointTemplate))
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

	return cmd
}

func gatherProtoResources() []resource {
	var resources []resource
	var types []protoreflect.MessageType
	protoregistry.GlobalTypes.RangeMessages(
		genutils.OnKumaResourceMessage("mesh", func(m protoreflect.MessageType) bool {
			types = append(types, m)
			return true
		}))

	for _, t := range types {
		resourceInfo := genutils.ToResourceInfo(t.Descriptor())
		if resourceInfo.ShortName != "" {
			log.Printf("Skipping %s because it does not have shortName", resourceInfo.ResourceType)
		}
		_, exists := commontypes.ProtoTypeToType[resourceInfo.ResourceType]
		if !exists {
			log.Printf("Skipping %s because it does not have mapping defined in tools/common/types/proto.go. Please add it there.", resourceInfo.ResourceType)
		}
		if resourceInfo.ShortName != "" {
			resources = append(resources, resource{
				ResourceType: resourceInfo.ResourceType,
				Path:         "/specs/protoresources/" + strings.ToLower(resourceInfo.ResourceType) + "/rest.yaml",
			})
		}
	}
	return resources
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
