package cmd

import (
	"io/ioutil"
	"strings"

	"github.com/hoisie/mustache"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	kuma_dp "github.com/kumahq/kuma/pkg/config/app/kuma-dp"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
)

func readDataplaneResource(cmd *cobra.Command, cfg *kuma_dp.Config) (*core_mesh.DataplaneResource, error) {
	var b []byte
	var err error
	switch cfg.DataplaneRuntime.DataplaneTemplate {
	case "":
		return nil, nil
	case "-":
		if b, err = ioutil.ReadAll(cmd.InOrStdin()); err != nil {
			return nil, err
		}
	default:
		if b, err = ioutil.ReadFile(cfg.DataplaneRuntime.DataplaneTemplate); err != nil {
			return nil, errors.Wrap(err, "error while reading provided file")
		}
	}
	b, err = processDataplaneTemplate(b, cfg.DataplaneRuntime.DataplaneTemplateVars)
	runLog.Info("rendered template:", cfg.DataplaneRuntime.DataplaneTemplate, string(b))
	if err != nil {
		return nil, err
	}
	dp, err := core_mesh.ParseDataplaneYAML(b)
	if err != nil {
		return nil, err
	}

	if err = dp.Validate(); err != nil {
		return nil, err
	}

	cfg.Dataplane.Mesh = dp.Meta.GetMesh()
	cfg.Dataplane.Name = dp.Meta.GetName()

	err = cfg.Dataplane.Validate()
	if err != nil {
		return nil, err
	}

	return dp, nil
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

func processDataplaneTemplate(template []byte, values map[string]string) ([]byte, error) {
	// TODO error checking -- match number of placeholders with number of
	// passed values
	ctx := contextMap{}
	for k, v := range values {
		ctx.merge(newContextMap(k, v))
	}
	data := mustache.Render(string(template), ctx)
	return []byte(data), nil
}
