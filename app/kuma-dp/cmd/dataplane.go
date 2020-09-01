package cmd

import (
	"io/ioutil"
	"strings"

	"github.com/hoisie/mustache"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	kuma_dp "github.com/kumahq/kuma/pkg/config/app/kuma-dp"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/model/rest"
)

func readDataplaneResource(cmd *cobra.Command, cfg *kuma_dp.Config) (*rest.Resource, error) {
	var b []byte
	var err error
	// load from file first
	if cfg.DataplaneRuntime.ResourcePath == "-" {
		if b, err = ioutil.ReadAll(cmd.InOrStdin()); err != nil {
			return nil, err
		}
	} else if cfg.DataplaneRuntime.ResourcePath != "" {
		if b, err = ioutil.ReadFile(cfg.DataplaneRuntime.ResourcePath); err != nil {
			return nil, errors.Wrap(err, "error while reading provided file")
		}
	}
	// override with inline resource
	if cfg.DataplaneRuntime.Resource != "" {
		b = []byte(cfg.DataplaneRuntime.Resource)
	}

	if len(b) == 0 {
		return nil, nil
	}

	b = processDataplaneTemplate(b, cfg.DataplaneRuntime.ResourceVars)
	runLog.Info("rendered dataplane", "dataplane", string(b))

	res, err := rest.Unmarshall(b)
	if err != nil {
		return nil, err
	}
	if res.Meta.Type != string(core_mesh.DataplaneType) {
		return nil, errors.Errorf("invalid resource of type: %s. Expected: Dataplane", res.Meta.Type)
	}
	if err := core_mesh.ValidateMeta(res.Meta.GetName(), res.Meta.GetMesh(), model.ScopeMesh); err.HasViolations() {
		return nil, &err
	}

	return res, nil
}

type contextMap map[string]interface{}

func (cm contextMap) merge(other contextMap) {
	for k, v := range other {
		cm[k] = v
	}
}

func newContextMap(key, value string) contextMap {
	if !strings.Contains(key, ".") {
		return map[string]interface{}{
			key: value,
		}
	}

	parts := strings.SplitAfterN(key, ".", 2)
	return map[string]interface{}{
		parts[0][:len(parts[0])-1]: newContextMap(parts[1], value),
	}
}

func processDataplaneTemplate(template []byte, values map[string]string) []byte {
	ctx := contextMap{}
	for k, v := range values {
		ctx.merge(newContextMap(k, v))
	}
	data := mustache.Render(string(template), ctx)
	return []byte(data)
}
