package mesh

import (
	"errors"
	"fmt"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/core/resources/registry"
)

const (
	MeshType model.ResourceType = "Mesh"
)

var _ model.Resource = &MeshResource{}

type MeshResource struct {
	Meta model.ResourceMeta
	Spec *mesh_proto.Mesh
}

func NewMeshResource() *MeshResource {
	return &MeshResource{
		Spec: &mesh_proto.Mesh{},
	}
}

func (t *MeshResource) GetMeta() model.ResourceMeta {
	return t.Meta
}

func (t *MeshResource) SetMeta(m model.ResourceMeta) {
	t.Meta = m
}

func (t *MeshResource) GetSpec() model.ResourceSpec {
	return t.Spec
}

func (t *MeshResource) SetSpec(spec model.ResourceSpec) error {
	protoType, ok := spec.(*mesh_proto.Mesh)
	if !ok {
		return fmt.Errorf("invalid type %T for Spec", spec)
	} else {
		if protoType == nil {
			t.Spec = &mesh_proto.Mesh{}
		} else {
			t.Spec = protoType
		}
		return nil
	}
}

func (t *MeshResource) GetStatus() model.ResourceStatus {
	return nil
}

func (t *MeshResource) SetStatus(_ model.ResourceStatus) error {
	return errors.New("status not supported")
}

func (t *MeshResource) Descriptor() model.ResourceTypeDescriptor {
	return MeshResourceTypeDescriptor
}

var _ model.ResourceList = &MeshResourceList{}

type MeshResourceList struct {
	Items      []*MeshResource
	Pagination model.Pagination
}

func (l *MeshResourceList) GetItems() []model.Resource {
	res := make([]model.Resource, len(l.Items))
	for i, elem := range l.Items {
		res[i] = elem
	}
	return res
}

func (l *MeshResourceList) GetItemType() model.ResourceType {
	return MeshType
}

func (l *MeshResourceList) NewItem() model.Resource {
	return NewMeshResource()
}

func (l *MeshResourceList) AddItem(r model.Resource) error {
	if trr, ok := r.(*MeshResource); ok {
		l.Items = append(l.Items, trr)
		return nil
	} else {
		return model.ErrorInvalidItemType((*MeshResource)(nil), r)
	}
}

func (l *MeshResourceList) GetPagination() *model.Pagination {
	return &l.Pagination
}

func (l *MeshResourceList) SetPagination(p model.Pagination) {
	l.Pagination = p
}

var MeshResourceTypeDescriptor = model.ResourceTypeDescriptor{
	Name:                  MeshType,
	Resource:              NewMeshResource(),
	ResourceList:          &MeshResourceList{},
	ReadOnly:              false,
	AdminOnly:             false,
	Scope:                 model.ScopeGlobal,
	KDSFlags:              model.GlobalToZonesFlag,
	SkipKDSHash:           true,
	WsPath:                "meshes",
	KumactlArg:            "mesh",
	KumactlListArg:        "meshes",
	AllowToInspect:        false,
	IsPolicy:              false,
	SingularDisplayName:   "Mesh",
	PluralDisplayName:     "Meshes",
	ShortName:             "m",
	IsExperimental:        false,
	IsProxy:               false,
	AffectsPolicyMatching: true,
	Insight:               NewMeshInsightResource(),
	Overview:              NewMeshOverviewResource(),
}

func init() {
	registry.RegisterType(MeshResourceTypeDescriptor)
}

const (
	MeshInsightType model.ResourceType = "MeshInsight"
)

var _ model.Resource = &MeshInsightResource{}

type MeshInsightResource struct {
	Meta model.ResourceMeta
	Spec *mesh_proto.MeshInsight
}

func NewMeshInsightResource() *MeshInsightResource {
	return &MeshInsightResource{
		Spec: &mesh_proto.MeshInsight{},
	}
}

func (t *MeshInsightResource) GetMeta() model.ResourceMeta {
	return t.Meta
}

func (t *MeshInsightResource) SetMeta(m model.ResourceMeta) {
	t.Meta = m
}

func (t *MeshInsightResource) GetSpec() model.ResourceSpec {
	return t.Spec
}

func (t *MeshInsightResource) SetSpec(spec model.ResourceSpec) error {
	protoType, ok := spec.(*mesh_proto.MeshInsight)
	if !ok {
		return fmt.Errorf("invalid type %T for Spec", spec)
	} else {
		if protoType == nil {
			t.Spec = &mesh_proto.MeshInsight{}
		} else {
			t.Spec = protoType
		}
		return nil
	}
}

func (t *MeshInsightResource) GetStatus() model.ResourceStatus {
	return nil
}

func (t *MeshInsightResource) SetStatus(_ model.ResourceStatus) error {
	return errors.New("status not supported")
}

func (t *MeshInsightResource) Descriptor() model.ResourceTypeDescriptor {
	return MeshInsightResourceTypeDescriptor
}

var _ model.ResourceList = &MeshInsightResourceList{}

type MeshInsightResourceList struct {
	Items      []*MeshInsightResource
	Pagination model.Pagination
}

func (l *MeshInsightResourceList) GetItems() []model.Resource {
	res := make([]model.Resource, len(l.Items))
	for i, elem := range l.Items {
		res[i] = elem
	}
	return res
}

func (l *MeshInsightResourceList) GetItemType() model.ResourceType {
	return MeshInsightType
}

func (l *MeshInsightResourceList) NewItem() model.Resource {
	return NewMeshInsightResource()
}

func (l *MeshInsightResourceList) AddItem(r model.Resource) error {
	if trr, ok := r.(*MeshInsightResource); ok {
		l.Items = append(l.Items, trr)
		return nil
	} else {
		return model.ErrorInvalidItemType((*MeshInsightResource)(nil), r)
	}
}

func (l *MeshInsightResourceList) GetPagination() *model.Pagination {
	return &l.Pagination
}

func (l *MeshInsightResourceList) SetPagination(p model.Pagination) {
	l.Pagination = p
}

var MeshInsightResourceTypeDescriptor = model.ResourceTypeDescriptor{
	Name:                  MeshInsightType,
	Resource:              NewMeshInsightResource(),
	ResourceList:          &MeshInsightResourceList{},
	ReadOnly:              true,
	AdminOnly:             false,
	Scope:                 model.ScopeGlobal,
	KDSFlags:              model.ProvidedByGlobalFlag,
	WsPath:                "mesh-insights",
	KumactlArg:            "",
	KumactlListArg:        "",
	AllowToInspect:        false,
	IsPolicy:              false,
	SingularDisplayName:   "Mesh Insight",
	PluralDisplayName:     "Mesh Insights",
	IsExperimental:        false,
	IsProxy:               false,
	AffectsPolicyMatching: false,
}

func init() {
	registry.RegisterType(MeshInsightResourceTypeDescriptor)
}

const (
	MeshOverviewType model.ResourceType = "MeshOverview"
)

var _ model.Resource = &MeshOverviewResource{}

type MeshOverviewResource struct {
	Meta model.ResourceMeta
	Spec *mesh_proto.MeshOverview
}

func NewMeshOverviewResource() *MeshOverviewResource {
	return &MeshOverviewResource{
		Spec: &mesh_proto.MeshOverview{},
	}
}

func (t *MeshOverviewResource) GetMeta() model.ResourceMeta {
	return t.Meta
}

func (t *MeshOverviewResource) SetMeta(m model.ResourceMeta) {
	t.Meta = m
}

func (t *MeshOverviewResource) GetSpec() model.ResourceSpec {
	return t.Spec
}

func (t *MeshOverviewResource) SetSpec(spec model.ResourceSpec) error {
	protoType, ok := spec.(*mesh_proto.MeshOverview)
	if !ok {
		return fmt.Errorf("invalid type %T for Spec", spec)
	} else {
		if protoType == nil {
			t.Spec = &mesh_proto.MeshOverview{}
		} else {
			t.Spec = protoType
		}
		return nil
	}
}

func (t *MeshOverviewResource) GetStatus() model.ResourceStatus {
	return nil
}

func (t *MeshOverviewResource) SetStatus(_ model.ResourceStatus) error {
	return errors.New("status not supported")
}

func (t *MeshOverviewResource) Descriptor() model.ResourceTypeDescriptor {
	return MeshOverviewResourceTypeDescriptor
}

// newMeshOverviewSpec normalizes a nil spec the way SetSpec does. A nil one would drop
// the field from the response without any error: the marshaller omits a nil message and
// then throws the whole spec away once what is left renders as an empty object.
func newMeshOverviewSpec(spec *mesh_proto.Mesh) *mesh_proto.MeshOverview {
	if spec == nil {
		spec = &mesh_proto.Mesh{}
	}
	return &mesh_proto.MeshOverview{
		Mesh: spec,
	}
}

func (t *MeshOverviewResource) SetOverviewSpec(resource model.Resource, insight model.Resource) error {
	t.SetMeta(resource.GetMeta())
	overview := newMeshOverviewSpec(resource.GetSpec().(*mesh_proto.Mesh))
	if insight != nil {
		ins, ok := insight.GetSpec().(*mesh_proto.MeshInsight)
		if !ok {
			return errors.New("failed to convert to insight type 'MeshInsight'")
		}
		overview.MeshInsight = ins
	}
	return t.SetSpec(overview)
}

var _ model.ResourceList = &MeshOverviewResourceList{}

type MeshOverviewResourceList struct {
	Items      []*MeshOverviewResource
	Pagination model.Pagination
}

func (l *MeshOverviewResourceList) GetItems() []model.Resource {
	res := make([]model.Resource, len(l.Items))
	for i, elem := range l.Items {
		res[i] = elem
	}
	return res
}

func (l *MeshOverviewResourceList) GetItemType() model.ResourceType {
	return MeshOverviewType
}

func (l *MeshOverviewResourceList) NewItem() model.Resource {
	return NewMeshOverviewResource()
}

func (l *MeshOverviewResourceList) AddItem(r model.Resource) error {
	if trr, ok := r.(*MeshOverviewResource); ok {
		l.Items = append(l.Items, trr)
		return nil
	} else {
		return model.ErrorInvalidItemType((*MeshOverviewResource)(nil), r)
	}
}

func (l *MeshOverviewResourceList) GetPagination() *model.Pagination {
	return &l.Pagination
}

func (l *MeshOverviewResourceList) SetPagination(p model.Pagination) {
	l.Pagination = p
}

var MeshOverviewResourceTypeDescriptor = model.ResourceTypeDescriptor{
	Name:                  MeshOverviewType,
	Resource:              NewMeshOverviewResource(),
	ResourceList:          &MeshOverviewResourceList{},
	ReadOnly:              false,
	AdminOnly:             false,
	Scope:                 model.ScopeGlobal,
	WsPath:                "",
	KumactlArg:            "",
	KumactlListArg:        "",
	AllowToInspect:        false,
	IsPolicy:              false,
	SingularDisplayName:   "Mesh Overview",
	PluralDisplayName:     "Mesh Overviews",
	IsExperimental:        false,
	IsProxy:               false,
	AffectsPolicyMatching: false,
}
