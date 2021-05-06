package mesh

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"text/template"

	"github.com/asaskevich/govalidator"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
)

const (
	VirtualOutboundType model.ResourceType = "VirtualOutbound"
)

var _ model.Resource = &VirtualOutboundResource{}

type VirtualOutboundResource struct {
	Meta model.ResourceMeta
	Spec *mesh_proto.VirtualOutbound
}

func NewVirtualOutboundResource() *VirtualOutboundResource {
	return &VirtualOutboundResource{
		Spec: &mesh_proto.VirtualOutbound{},
	}
}

func (r *VirtualOutboundResource) GetType() model.ResourceType {
	return VirtualOutboundType
}
func (r *VirtualOutboundResource) GetMeta() model.ResourceMeta {
	return r.Meta
}
func (r *VirtualOutboundResource) SetMeta(m model.ResourceMeta) {
	r.Meta = m
}
func (r *VirtualOutboundResource) GetSpec() model.ResourceSpec {
	return r.Spec
}
func (r *VirtualOutboundResource) SetSpec(spec model.ResourceSpec) error {
	status, ok := spec.(*mesh_proto.VirtualOutbound)
	if !ok {
		return errors.New("invalid type of spec")
	} else {
		r.Spec = status
		return nil
	}
}
func (r *VirtualOutboundResource) Scope() model.ResourceScope {
	return model.ScopeMesh
}

func (r *VirtualOutboundResource) Selectors() []*mesh_proto.Selector {
	return r.Spec.GetSelectors()
}

func (r *VirtualOutboundResource) evalTemplate(tmplStr string, tags map[string]string) (string, error) {
	entries := map[string]string{}
	for k, v := range r.Spec.Conf.Parameters {
		val, ok := tags[v]
		if ok {
			entries[k] = val
		}
	}
	sb := strings.Builder{}
	tmpl, err := template.New("").Parse(tmplStr)
	if err != nil {
		return "", fmt.Errorf("failed compiling gotemplate error='%s'", err.Error())
	}
	err = tmpl.Execute(&sb, entries)
	if err != nil {
		return "", fmt.Errorf("pre evaluation of template with parameters failed with error='%s'", err.Error())
	}
	return sb.String(), nil
}

func (r *VirtualOutboundResource) EvalPort(tags map[string]string) (uint32, error) {
	s, err := r.evalTemplate(r.Spec.Conf.Port, tags)
	if err != nil {
		return 0, err
	}
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("evaluation of template with parameters didn't evaluate to a parsable number result='%s'", s)
	}
	if i <= 0 || i > 65535 {
		return 0, fmt.Errorf("evaluation of template returned a port outside of the range [1..65535] result='%d'", i)
	}
	return uint32(i), nil
}

func (r *VirtualOutboundResource) EvalHost(tags map[string]string) (string, error) {
	s, err := r.evalTemplate(r.Spec.Conf.Host, tags)
	if err != nil {
		return "", err
	}
	if !govalidator.IsDNSName(s) {
		return "", fmt.Errorf("evaluation of template with parameters didn't return a valid dns name result='%s'", s)
	}
	return s, nil
}

func (r *VirtualOutboundResource) FilterTags(tags map[string]string) map[string]string {
	out := map[string]string{
		mesh_proto.ServiceTag: tags[mesh_proto.ServiceTag],
	}

	for _, v := range r.Spec.Conf.Parameters {
		out[v] = tags[v]
	}
	return out
}

var _ model.ResourceList = &VirtualOutboundResourceList{}

type VirtualOutboundResourceList struct {
	Items      []*VirtualOutboundResource
	Pagination model.Pagination
}

func (l *VirtualOutboundResourceList) GetItems() []model.Resource {
	res := make([]model.Resource, len(l.Items))
	for i, elem := range l.Items {
		res[i] = elem
	}
	return res
}
func (l *VirtualOutboundResourceList) GetItemType() model.ResourceType {
	return VirtualOutboundType
}
func (l *VirtualOutboundResourceList) NewItem() model.Resource {
	return NewVirtualOutboundResource()
}
func (l *VirtualOutboundResourceList) AddItem(r model.Resource) error {
	if trr, ok := r.(*VirtualOutboundResource); ok {
		l.Items = append(l.Items, trr)
		return nil
	} else {
		return model.ErrorInvalidItemType((*VirtualOutboundResource)(nil), r)
	}
}
func (l *VirtualOutboundResourceList) GetPagination() *model.Pagination {
	return &l.Pagination
}

func init() {
	registry.RegisterType(NewVirtualOutboundResource())
	registry.RegistryListType(&VirtualOutboundResourceList{})
}
