// Generated by tools/policy-gen.
// Run "make generate" to update this file.

// nolint:whitespace
package v1alpha1

import (
	_ "embed"
	"errors"
	"fmt"

	"k8s.io/kube-openapi/pkg/validation/spec"
	"sigs.k8s.io/yaml"

	"github.com/kumahq/kuma/pkg/core/resources/model"
)

//go:embed schema.yaml
var rawSchema []byte

func init() {
	var schema spec.Schema
	if rawSchema != nil {
		if err := yaml.Unmarshal(rawSchema, &schema); err != nil {
			panic(err)
		}
	}
	rawSchema = nil
	MeshHTTPRouteResourceTypeDescriptor.Schema = &schema
}

const (
	MeshHTTPRouteType model.ResourceType = "MeshHTTPRoute"
)

var _ model.Resource = &MeshHTTPRouteResource{}

type MeshHTTPRouteResource struct {
	Meta model.ResourceMeta
	Spec *MeshHTTPRoute
}

func NewMeshHTTPRouteResource() *MeshHTTPRouteResource {
	return &MeshHTTPRouteResource{
		Spec: &MeshHTTPRoute{},
	}
}

func (t *MeshHTTPRouteResource) GetMeta() model.ResourceMeta {
	return t.Meta
}

func (t *MeshHTTPRouteResource) SetMeta(m model.ResourceMeta) {
	t.Meta = m
}

func (t *MeshHTTPRouteResource) GetSpec() model.ResourceSpec {
	return t.Spec
}

func (t *MeshHTTPRouteResource) SetSpec(spec model.ResourceSpec) error {
	protoType, ok := spec.(*MeshHTTPRoute)
	if !ok {
		return fmt.Errorf("invalid type %T for Spec", spec)
	} else {
		if protoType == nil {
			t.Spec = &MeshHTTPRoute{}
		} else {
			t.Spec = protoType
		}
		return nil
	}
}

func (t *MeshHTTPRouteResource) GetStatus() model.ResourceStatus {
	return nil
}

func (t *MeshHTTPRouteResource) SetStatus(_ model.ResourceStatus) error {
	return errors.New("status not supported")
}

func (t *MeshHTTPRouteResource) Descriptor() model.ResourceTypeDescriptor {
	return MeshHTTPRouteResourceTypeDescriptor
}

func (t *MeshHTTPRouteResource) Validate() error {
	if v, ok := interface{}(t).(interface{ validate() error }); !ok {
		return nil
	} else {
		return v.validate()
	}
}

var _ model.ResourceList = &MeshHTTPRouteResourceList{}

type MeshHTTPRouteResourceList struct {
	Items      []*MeshHTTPRouteResource
	Pagination model.Pagination
}

func (l *MeshHTTPRouteResourceList) GetItems() []model.Resource {
	res := make([]model.Resource, len(l.Items))
	for i, elem := range l.Items {
		res[i] = elem
	}
	return res
}

func (l *MeshHTTPRouteResourceList) GetItemType() model.ResourceType {
	return MeshHTTPRouteType
}

func (l *MeshHTTPRouteResourceList) NewItem() model.Resource {
	return NewMeshHTTPRouteResource()
}

func (l *MeshHTTPRouteResourceList) AddItem(r model.Resource) error {
	if trr, ok := r.(*MeshHTTPRouteResource); ok {
		l.Items = append(l.Items, trr)
		return nil
	} else {
		return model.ErrorInvalidItemType((*MeshHTTPRouteResource)(nil), r)
	}
}

func (l *MeshHTTPRouteResourceList) GetPagination() *model.Pagination {
	return &l.Pagination
}

func (l *MeshHTTPRouteResourceList) SetPagination(p model.Pagination) {
	l.Pagination = p
}

var MeshHTTPRouteResourceTypeDescriptor = model.ResourceTypeDescriptor{
	Name:                         MeshHTTPRouteType,
	Resource:                     NewMeshHTTPRouteResource(),
	ResourceList:                 &MeshHTTPRouteResourceList{},
	Scope:                        model.ScopeMesh,
	KDSFlags:                     model.GlobalToAllZonesFlag | model.ZoneToGlobalFlag | model.GlobalToAllButOriginalZoneFlag,
	WsPath:                       "meshhttproutes",
	KumactlArg:                   "meshhttproute",
	KumactlListArg:               "meshhttproutes",
	AllowToInspect:               true,
	IsPolicy:                     true,
	IsExperimental:               false,
	SingularDisplayName:          "Mesh HTTP Route",
	PluralDisplayName:            "Mesh HTTP Routes",
	IsPluginOriginated:           true,
	IsTargetRefBased:             true,
	HasToTargetRef:               true,
	HasFromTargetRef:             false,
	HasRulesTargetRef:            false,
	HasStatus:                    false,
	AllowedOnSystemNamespaceOnly: false,
	IsReferenceableInTo:          false,
	ShortName:                    "mhttpr",
	InterpretFromEntriesAsRules:  false,
}
