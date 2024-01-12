package mesh

import (
	"fmt"
	"strconv"
	"strings"
	"text/template"

	"github.com/asaskevich/govalidator"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
)

func tagKeyOrName(parameter *mesh_proto.VirtualOutbound_Conf_TemplateParameter) string {
	if parameter.TagKey == "" {
		return parameter.Name
	}
	return parameter.TagKey
}

func (t *VirtualOutboundResource) evalTemplate(tmplStr string, tags map[string]string) (string, error) {
	entries := map[string]string{}
	for _, v := range t.Spec.Conf.Parameters {
		tagKey := tagKeyOrName(v)
		val, ok := tags[tagKey]
		if ok {
			entries[v.Name] = val
		}
	}
	sb := strings.Builder{}
	tmpl, err := template.New("").Parse(tmplStr)
	if err != nil {
		return "", fmt.Errorf("failed compiling gotemplate error=%q", err.Error())
	}
	err = tmpl.Execute(&sb, entries)
	if err != nil {
		return "", fmt.Errorf("pre evaluation of template with parameters failed with error=%q", err.Error())
	}
	return sb.String(), nil
}

func (t *VirtualOutboundResource) EvalPort(tags map[string]string) (uint32, error) {
	s, err := t.evalTemplate(t.Spec.Conf.Port, tags)
	if err != nil {
		return 0, err
	}
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("evaluation of template with parameters didn't evaluate to a parsable number result=%q", s)
	}
	if i <= 0 || i > 65535 {
		return 0, fmt.Errorf("evaluation of template returned a port outside of the range [1..65535] result=%d", i)
	}
	return uint32(i), nil
}

func (t *VirtualOutboundResource) EvalHost(tags map[string]string) (string, error) {
	s, err := t.evalTemplate(t.Spec.Conf.Host, tags)
	if err != nil {
		return "", err
	}
	if !govalidator.IsDNSName(s) {
		return "", fmt.Errorf("evaluation of template with parameters didn't return a valid dns name result=%q", s)
	}
	return s, nil
}

func (t *VirtualOutboundResource) FilterTags(tags map[string]string) map[string]string {
	out := map[string]string{}
	for _, v := range t.Spec.Conf.Parameters {
		tagKey := tagKeyOrName(v)
		if tagValue, exists := tags[tagKey]; exists {
			out[tagKey] = tagValue
		}
	}
	return out
}
