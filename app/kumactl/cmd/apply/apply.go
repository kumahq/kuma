package apply

import (
	"context"
	"fmt"
	"golang.org/x/exp/slices"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	kumactl_cmd "github.com/kumahq/kuma/app/kumactl/pkg/cmd"
	yaml_output "github.com/kumahq/kuma/app/kumactl/pkg/output/yaml"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	rest_types "github.com/kumahq/kuma/pkg/core/resources/model/rest"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/plugins/resources/remote"
	"github.com/kumahq/kuma/pkg/util/template"
	"github.com/kumahq/kuma/pkg/util/yaml"
)

const (
	timeout = 10 * time.Second
)

var yamlExt = []string{".yaml", ".yml"}

type applyContext struct {
	*kumactl_cmd.RootContext

	args struct {
		file   string
		vars   map[string]string
		dryRun bool
	}
}

func NewApplyCmd(pctx *kumactl_cmd.RootContext) *cobra.Command {
	ctx := &applyContext{RootContext: pctx}
	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Create or modify Kuma resources",
		Long:  `Create or modify Kuma resources.`,
		Example: `
Apply a resource from file
$ kumactl apply -f resource.yaml

Apply a resource from stdin
$ echo "
type: Mesh
name: demo
" | kumactl apply -f -

Apply a resource from external URL
$ kumactl apply -f https://example.com/resource.yaml
`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			_ = kumactl_cmd.CheckCompatibility(pctx.FetchServerVersion, cmd.ErrOrStderr())

			var b []byte
			var err error
			var resources []model.Resource

			if ctx.args.file == "-" {
				b, err = io.ReadAll(cmd.InOrStdin())
				if err != nil {
					return err
				}
				if len(b) == 0 {
					return fmt.Errorf("no resource(s) passed to apply")
				}
				r, err := bytesToResources(ctx, cmd, b)
				if err != nil {
					return errors.Wrap(err, "error parsing file to resources")
				}
				resources = append(resources, r...)
			} else if strings.HasPrefix(ctx.args.file, "http://") || strings.HasPrefix(ctx.args.file, "https://") {
				client := &http.Client{
					Timeout: timeout,
				}
				req, err := http.NewRequest("GET", ctx.args.file, nil)
				if err != nil {
					return errors.Wrap(err, "error creating new http request")
				}
				resp, err := client.Do(req)
				if err != nil {
					return errors.Wrap(err, "error with GET http request")
				}
				if resp.StatusCode != http.StatusOK {
					return errors.Wrap(err, "error while retrieving URL")
				}
				defer resp.Body.Close()
				b, err = io.ReadAll(resp.Body)
				if err != nil {
					return errors.Wrap(err, "error while reading provided file")
				}
				r, err := bytesToResources(ctx, cmd, b)
				if err != nil {
					return errors.Wrap(err, "error parsing file to resources")
				}
				resources = append(resources, r...)
			} else {
				// Process local yaml files
				r, err := localFileToResources(ctx, cmd)
				if err != nil {
					return errors.Wrap(err, "error processing file")
				}
				resources = append(resources, r...)
			}

			var rs store.ResourceStore
			if !ctx.args.dryRun {
				rs, err = pctx.CurrentResourceStore()
				if err != nil {
					return err
				}
			}
			p := yaml_output.NewPrinter()
			for _, resource := range resources {
				if rs == nil {
					if err := p.Print(rest_types.From.Resource(resource), cmd.OutOrStdout()); err != nil {
						return err
					}
				} else {
					warnings, err := upsert(cmd.Context(), pctx.Runtime.Registry, rs, resource)
					if err != nil {
						return err
					}
					for _, w := range warnings {
						if _, err := fmt.Fprintf(cmd.ErrOrStderr(), "Warning: %v\n", w); err != nil {
							return err
						}
					}
				}
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&ctx.args.file, "file", "f", "", "Path to file to apply. Pass `-` to read from stdin")
	_ = cmd.MarkFlagRequired("file")
	cmd.Flags().StringToStringVarP(&ctx.args.vars, "var", "v", map[string]string{}, "Variable to replace in configuration")
	cmd.Flags().BoolVar(&ctx.args.dryRun, "dry-run", false, "Resolve variable and prints result out without actual applying")
	return cmd
}

// localFileToResources reads and converts a local file into a list of model.Resource
// the local file could be a directory, in which case it processes all the yaml files in the directory
func localFileToResources(ctx *applyContext, cmd *cobra.Command) ([]model.Resource, error) {
	var resources []model.Resource
	file, err := os.Open(ctx.args.file)
	if err != nil {
		return nil, errors.Wrap(err, "error while opening provided file")
	}
	defer file.Close()
	orgDir, _ := filepath.Split(ctx.args.file)

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, errors.Wrap(err, "error getting stats for the provided file")
	}

	var yamlFiles []string
	if fileInfo.IsDir() {
		for {
			names, err := file.Readdirnames(10)
			if err != nil {
				if err == io.EOF {
					break
				} else {
					return nil, errors.Wrap(err, "error reading file names in directory")
				}
			}
			for _, n := range names {
				if slices.Contains(yamlExt, filepath.Ext(n)) {
					yamlFiles = append(yamlFiles, n)
				}
			}
		}
	} else {
		if slices.Contains(yamlExt, filepath.Ext(fileInfo.Name())) {
			yamlFiles = append(yamlFiles, fileInfo.Name())
		}
		// TODO should this check be added?
		//else {
		//	return nil, fmt.Errorf("error the specified input file extension isn't yaml")
		//}
	}
	var b []byte
	for _, f := range yamlFiles {
		joined := filepath.Join(orgDir, f)
		b, err = os.ReadFile(joined)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("error while reading the provided file [%s]", f))
		}
		r, err := bytesToResources(ctx, cmd, b)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("error parsing file [%s] to resources", f))
		}
		resources = append(resources, r...)
	}
	if len(resources) == 0 {
		return nil, fmt.Errorf("no resource(s) passed to apply")
	}
	return resources, nil
}

// bytesToResources converts a slice of bytes into a slice of model.Resource
func bytesToResources(ctx *applyContext, cmd *cobra.Command, fileBytes []byte) ([]model.Resource, error) {
	var resources []model.Resource
	rawResources := yaml.SplitYAML(string(fileBytes))
	for _, rawResource := range rawResources {
		if len(rawResource) == 0 {
			continue
		}
		bytes := []byte(rawResource)
		if len(ctx.args.vars) > 0 {
			bytes = template.Render(rawResource, ctx.args.vars)
		}
		res, err := rest_types.YAML.UnmarshalCore(bytes)
		if err != nil {
			return nil, errors.Wrap(err, "YAML contains invalid resource")
		}
		if err, msg := mesh.ValidateMetaBackwardsCompatible(res.GetMeta(), res.Descriptor().Scope); err.HasViolations() {
			return nil, err.OrNil()
		} else if msg != "" {
			if _, printErr := fmt.Fprintln(cmd.ErrOrStderr(), msg); printErr != nil {
				return nil, printErr
			}
		}
		resources = append(resources, res)
	}
	return resources, nil
}

func upsert(ctx context.Context, typeRegistry registry.TypeRegistry, rs store.ResourceStore, res model.Resource) error {
	newRes, err := typeRegistry.NewObject(res.Descriptor().Name)
	if err != nil {
		return nil, err
	}

	var warnings []string
	warnContext := context.WithValue(ctx, remote.WarningsCallback, func(s []string) {
		warnings = s
	})

	meta := res.GetMeta()
	if err := rs.Get(ctx, newRes, store.GetByKey(meta.GetName(), meta.GetMesh())); err != nil {
		if store.IsResourceNotFound(err) {
			cerr := rs.Create(warnContext, res, store.CreateByKey(meta.GetName(), meta.GetMesh()), store.CreateWithLabels(meta.GetLabels()))
			return warnings, cerr
		} else {
			return nil, err
		}
	}
	if err := newRes.SetSpec(res.GetSpec()); err != nil {
		return nil, err
	}
	uerr := rs.Update(warnContext, newRes, store.UpdateWithLabels(meta.GetLabels()))
	return warnings, uerr
}
