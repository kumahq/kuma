package apply

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	kumactl_cmd "github.com/kumahq/kuma/v2/app/kumactl/pkg/cmd"
	yaml_output "github.com/kumahq/kuma/v2/app/kumactl/pkg/output/yaml"
	"github.com/kumahq/kuma/v2/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/v2/pkg/core/resources/model"
	rest_types "github.com/kumahq/kuma/v2/pkg/core/resources/model/rest"
	"github.com/kumahq/kuma/v2/pkg/core/resources/registry"
	"github.com/kumahq/kuma/v2/pkg/core/resources/store"
	"github.com/kumahq/kuma/v2/pkg/plugins/resources/remote"
	"github.com/kumahq/kuma/v2/pkg/util/template"
	"github.com/kumahq/kuma/v2/pkg/util/yaml"
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

Apply all resources from a directory
$ kumactl apply -f resources/
`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			_ = kumactl_cmd.CheckCompatibility(pctx.FetchServerVersion, cmd.ErrOrStderr())

			b, err := readContent(ctx.args.file, cmd)
			if err != nil {
				return err
			}
			if len(b) == 0 {
				return fmt.Errorf("no resource(s) passed to apply")
			}
			var resources []model.Resource
			rawResources := yaml.SplitYAML(string(b))
			var hasErrors bool
			for i, rawResource := range rawResources {
				if rawResource == "" {
					continue
				}
				bytes := []byte(rawResource)
				if len(ctx.args.vars) > 0 {
					bytes = template.Render(rawResource, ctx.args.vars)
				}
				res, pErr := rest_types.YAML.UnmarshalCore(bytes)
				if pErr != nil {
					_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "resource[%d]: failed to parse resource: %+v\n", i, pErr)
					hasErrors = true
					continue
				}
				if vErr := mesh.ValidateMeta(res.GetMeta(), res.Descriptor().Scope); vErr.HasViolations() {
					_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "resource[%d]: failed to read meta: %+v\n", i, vErr.OrNil())
					hasErrors = true
					continue
				}
				resources = append(resources, res)
			}
			if hasErrors {
				return errors.New("failed to validate some resources")
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
					warnings, isUpdate, err := upsert(cmd.Context(), pctx.Runtime.Registry, rs, resource)
					if err != nil {
						return fmt.Errorf("resource type=%q mesh=%q name=%q: failed server side: %s", resource.Descriptor().Name, resource.GetMeta().GetMesh(), resource.GetMeta().GetName(), err.Error())
					}
					action := "Created"
					if isUpdate {
						action = "Updated"
					}
					_, _ = fmt.Fprintf(cmd.OutOrStdout(), "resource type=%q mesh=%q name=%q %s\n", resource.Descriptor().Name, resource.GetMeta().GetMesh(), resource.GetMeta().GetName(), action)
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
	cmd.Flags().StringVarP(&ctx.args.file, "file", "f", "", "Path to file or directory to apply. When a directory is provided, all .yaml, .yml, and .json files are applied. Pass `-` to read from stdin")
	_ = cmd.MarkFlagRequired("file")
	cmd.Flags().StringToStringVarP(&ctx.args.vars, "var", "v", map[string]string{}, "Variable to replace in configuration")
	cmd.Flags().BoolVar(&ctx.args.dryRun, "dry-run", false, "Resolve variable and prints result out without actual applying")
	return cmd
}

var supportedExtensions = map[string]bool{
	".yaml": true,
	".yml":  true,
	".json": true,
}

func readContent(file string, cmd *cobra.Command) ([]byte, error) {
	if file == "-" {
		return io.ReadAll(cmd.InOrStdin())
	}
	if strings.HasPrefix(file, "http://") || strings.HasPrefix(file, "https://") {
		return readURL(file)
	}
	info, err := os.Stat(file)
	if err != nil {
		return nil, errors.Wrap(err, "error while reading provided file")
	}
	if info.IsDir() {
		return readDirectory(file)
	}
	b, err := os.ReadFile(file)
	if err != nil {
		return nil, errors.Wrap(err, "error while reading provided file")
	}
	return b, nil
}

func readURL(file string) ([]byte, error) {
	if _, err := url.ParseRequestURI(file); err != nil {
		return nil, errors.Wrap(err, "invalid URL for --file")
	}
	client := &http.Client{
		Timeout: timeout,
	}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, file, http.NoBody)
	if err != nil {
		return nil, errors.Wrap(err, "error creating new http request")
	}
	resp, err := client.Do(req) // #nosec G704 -- URL validated with ParseRequestURI above
	if err != nil {
		return nil, errors.Wrap(err, "error with GET http request")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("error while retrieving URL")
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "error while reading provided file")
	}
	return b, nil
}

func readDirectory(dir string) ([]byte, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, errors.Wrap(err, "error reading directory")
	}
	var files []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		// Skip symlinks that point to directories
		if entry.Type()&os.ModeSymlink != 0 {
			resolved, err := os.Stat(filepath.Join(dir, entry.Name()))
			if err != nil {
				continue
			}
			if resolved.IsDir() {
				continue
			}
		}
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if supportedExtensions[ext] {
			files = append(files, filepath.Join(dir, entry.Name()))
		}
	}
	sort.Strings(files)
	if len(files) == 0 {
		return nil, nil
	}
	var combined []byte
	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			return nil, errors.Wrapf(err, "error reading file %s", f)
		}
		if len(combined) > 0 {
			combined = append(combined, []byte("\n---\n")...)
		}
		combined = append(combined, b...)
	}
	return combined, nil
}

func upsert(ctx context.Context, typeRegistry registry.TypeRegistry, rs store.ResourceStore, res model.Resource) ([]string, bool, error) {
	newRes, err := typeRegistry.NewObject(res.Descriptor().Name)
	if err != nil {
		return nil, false, err
	}

	var warnings []string
	warnContext := context.WithValue(ctx, remote.WarningsCallback, func(s []string) {
		warnings = s
	})

	meta := res.GetMeta()
	if err := rs.Get(ctx, newRes, store.GetByKey(meta.GetName(), meta.GetMesh())); err != nil {
		if store.IsNotFound(err) {
			cerr := rs.Create(warnContext, res, store.CreateByKey(meta.GetName(), meta.GetMesh()), store.CreateWithLabels(meta.GetLabels()))
			return warnings, false, cerr
		} else {
			return nil, false, err
		}
	}
	if err := newRes.SetSpec(res.GetSpec()); err != nil {
		return nil, false, err
	}
	uerr := rs.Update(warnContext, newRes, store.UpdateWithLabels(meta.GetLabels()))
	return warnings, true, uerr
}
