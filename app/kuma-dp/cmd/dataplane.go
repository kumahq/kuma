package cmd

import (
	"io"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	kuma_dp "github.com/kumahq/kuma/pkg/config/app/kuma-dp"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/model/rest"
	"github.com/kumahq/kuma/pkg/util/template"
)

func readResource(cmd *cobra.Command, r *kuma_dp.DataplaneRuntime) (model.Resource, error) {
	var b []byte
	var err error

	// Load from file first.
	switch r.ResourcePath {
	case "":
		if r.Resource != "" {
			b = []byte(r.Resource)
		}
	case "-":
		if b, err = io.ReadAll(cmd.InOrStdin()); err != nil {
			return nil, err
		}
	default:
		if b, err = os.ReadFile(r.ResourcePath); err != nil {
			return nil, errors.Wrap(err, "error while reading provided file")
		}
	}

	if len(b) == 0 {
		return nil, nil
	}

	b = template.Render(string(b), r.ResourceVars)
	runLog.Info("rendered resource", "resource", string(b))

	res, err := rest.YAML.UnmarshalCore(b)
	if err != nil {
		return nil, err
	}

	if err := core_mesh.ValidateMeta(res.GetMeta().GetName(), res.GetMeta().GetMesh(), res.Descriptor().Scope); err.HasViolations() {
		return nil, &err
	}

	return res, nil
}
