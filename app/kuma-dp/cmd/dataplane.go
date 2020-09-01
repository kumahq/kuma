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
	switch cfg.DataplaneRuntime.ResourcePath {
	case "":
		return nil, nil
	case "-":
		if b, err = ioutil.ReadAll(cmd.InOrStdin()); err != nil {
			return nil, err
		}
	default:
		if b, err = ioutil.ReadFile(cfg.DataplaneRuntime.ResourcePath); err != nil {
			return nil, errors.Wrap(err, "error while reading provided file")
		}
	}
	if cfg.DataplaneRuntime.Resource != "" {
		b = []byte(cfg.DataplaneRuntime.Resource)
	}
	b = processDataplaneTemplate(b, cfg.DataplaneRuntime.ResourceVars)
	runLog.Info("rendered dataplane", "dataplane", string(b))

	dp, err := core_mesh.ParseDataplaneYAML(b)
	if err != nil {
		return nil, err
	}

	if err = dp.Validate(); err != nil {
		return nil, err
	}

	if err := core_mesh.ValidateMeta(dp.Meta.GetName(), dp.Meta.GetMesh()); err.HasViolations() {
		return nil, &err
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

func processDataplaneTemplate(template []byte, values map[string]string) []byte {
	ctx := contextMap{}
	for k, v := range values {
		ctx.merge(newContextMap(k, v))
	}
	data := mustache.Render(string(template), ctx)
	return []byte(data)
}
