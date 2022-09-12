package apply

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	kumactl_cmd "github.com/kumahq/kuma/app/kumactl/pkg/cmd"
	"github.com/kumahq/kuma/app/kumactl/pkg/output"
	"github.com/kumahq/kuma/app/kumactl/pkg/output/printers"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	rest_types "github.com/kumahq/kuma/pkg/core/resources/model/rest"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	"github.com/kumahq/kuma/pkg/core/resources/store"
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
			if err := pctx.CheckServerVersionCompatibility(); err != nil {
				cmd.PrintErrln(err)
			}

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
						Timeout:   timeout,
						Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
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
				if err := mesh.ValidateMeta(res.GetMeta().GetName(), res.GetMeta().GetMesh(), res.Descriptor().Scope); err.HasViolations() {
					return err.OrNil()
				}
				resources = append(resources, res)
			}
			for _, resource := range resources {
				if ctx.args.dryRun {
					p, err := printers.NewGenericPrinter(output.YAMLFormat)
					if err != nil {
						return err
					}
					if err := p.Print(rest_types.From.Resource(resource), cmd.OutOrStdout()); err != nil {
						return err
					}
				} else {
					rs, err := pctx.CurrentResourceStore()
					if err != nil {
						return err
					}

					if err := upsert(pctx.Runtime.Registry, rs, resource); err != nil {
						return err
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

func upsert(typeRegistry registry.TypeRegistry, rs store.ResourceStore, res model.Resource) error {
	newRes, err := typeRegistry.NewObject(res.Descriptor().Name)
	if err != nil {
		return err
	}
	meta := res.GetMeta()
	if err := rs.Get(context.Background(), newRes, store.GetByKey(meta.GetName(), meta.GetMesh())); err != nil {
		if store.IsResourceNotFound(err) {
			return rs.Create(context.Background(), res, store.CreateByKey(meta.GetName(), meta.GetMesh()))
		} else {
			return err
		}
	}
	if err := newRes.SetSpec(res.GetSpec()); err != nil {
		return err
	}
	return rs.Update(context.Background(), newRes)
}
