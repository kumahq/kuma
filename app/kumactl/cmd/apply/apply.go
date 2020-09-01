package apply

import (
	"context"
	"crypto/tls"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/util/template"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	kumactl_cmd "github.com/kumahq/kuma/app/kumactl/pkg/cmd"
	"github.com/kumahq/kuma/app/kumactl/pkg/output"
	"github.com/kumahq/kuma/app/kumactl/pkg/output/printers"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/model/rest"
	rest_types "github.com/kumahq/kuma/pkg/core/resources/model/rest"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	"github.com/kumahq/kuma/pkg/core/resources/store"
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
		RunE: func(cmd *cobra.Command, args []string) error {
			var b []byte
			var err error

			if ctx.args.file == "" || ctx.args.file == "-" {
				b, err = ioutil.ReadAll(cmd.InOrStdin())
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
					b, err = ioutil.ReadAll(resp.Body)
					if err != nil {
						return errors.Wrap(err, "error while reading provided file")
					}
				} else {
					b, err = ioutil.ReadFile(ctx.args.file)
					if err != nil {
						return errors.Wrap(err, "error while reading provided file")
					}
				}
			}

			bytes := template.Render(string(b), ctx.args.vars)
			res, err := rest.UnmarshallToCore(bytes)
			if err != nil {
				return errors.Wrap(err, "YAML contains invalid resource")
			}
			if err := mesh.ValidateMeta(res.GetMeta().GetName(), res.GetMeta().GetMesh(), res.Scope()); err.HasViolations() {
				return err.OrNil()
			}

			if ctx.args.dryRun {
				p, err := printers.NewGenericPrinter(output.YAMLFormat)
				if err != nil {
					return err
				}
				if err := p.Print(rest_types.From.Resource(res), cmd.OutOrStdout()); err != nil {
					return err
				}
				return nil
			}

			var rs store.ResourceStore
			if res.GetType() == system.SecretType { // Secret is exposed via Admin Server. It will be merged into API Server eventually.
				rs, err = pctx.CurrentAdminResourceStore()
			} else {
				rs, err = pctx.CurrentResourceStore()
			}
			if err != nil {
				return err
			}

			if err := upsert(rs, res); err != nil {
				return err
			}
			return nil
		},
	}
	cmd.PersistentFlags().StringVarP(&ctx.args.file, "file", "f", "", "Path to file to apply")
	cmd.PersistentFlags().StringToStringVarP(&ctx.args.vars, "var", "v", map[string]string{}, "Variable to replace in configuration")
	cmd.PersistentFlags().BoolVar(&ctx.args.dryRun, "dry-run", false, "Resolve variable and prints result out without actual applying")
	return cmd
}

func upsert(rs store.ResourceStore, res model.Resource) error {
	newRes, err := registry.Global().NewObject(res.GetType())
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
