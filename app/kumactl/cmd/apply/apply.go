package apply

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
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

			if ctx.args.file == "-" {
				b, err = io.ReadAll(cmd.InOrStdin())
				if err != nil {
					return err
				}
			} else {
				if strings.HasPrefix(ctx.args.file, "http://") || strings.HasPrefix(ctx.args.file, "https://") {
					client := &http.Client{
						Timeout: timeout,
					}
					req, err := http.NewRequest("GET", ctx.args.file, http.NoBody)
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
				} else {
					b, err = os.ReadFile(ctx.args.file)
					if err != nil {
						return errors.Wrap(err, "error while reading provided file")
					}
				}
			}
			if len(b) == 0 {
				return fmt.Errorf("no resource(s) passed to apply")
			}
			var resources []model.Resource
			rawResources := yaml.SplitYAML(string(b))
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
					return errors.Wrap(err, "YAML contains invalid resource")
				}
				if err := mesh.ValidateMeta(res.GetMeta(), res.Descriptor().Scope); err.HasViolations() {
					return err.OrNil()
				}
				resources = append(resources, res)
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

func upsert(ctx context.Context, typeRegistry registry.TypeRegistry, rs store.ResourceStore, res model.Resource) ([]string, error) {
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
		if store.IsNotFound(err) {
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
