package mesh

import (
	"errors"
	"fmt"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/core/resources/registry"
)

const (
	DataplaneType model.ResourceType = "Dataplane"
)

var _ model.Resource = &DataplaneResource{}

type DataplaneResource struct {
	Meta model.ResourceMeta
	Spec *mesh_proto.Dataplane
}

func NewDataplaneResource() *DataplaneResource {
	return &DataplaneResource{
		Spec: &mesh_proto.Dataplane{},
	}
}

func (t *DataplaneResource) GetMeta() model.ResourceMeta {
	return t.Meta
}

func (t *DataplaneResource) SetMeta(m model.ResourceMeta) {
	t.Meta = m
}

func (t *DataplaneResource) GetSpec() model.ResourceSpec {
	return t.Spec
}

func (t *DataplaneResource) SetSpec(spec model.ResourceSpec) error {
	protoType, ok := spec.(*mesh_proto.Dataplane)
	if !ok {
		return fmt.Errorf("invalid type %T for Spec", spec)
	} else {
		if protoType == nil {
			t.Spec = &mesh_proto.Dataplane{}
		} else {
			t.Spec = protoType
		}
		return nil
	}
}

func (t *DataplaneResource) GetStatus() model.ResourceStatus {
	return nil
}

func (t *DataplaneResource) SetStatus(_ model.ResourceStatus) error {
	return errors.New("status not supported")
}

func (t *DataplaneResource) Descriptor() model.ResourceTypeDescriptor {
	return DataplaneResourceTypeDescriptor
}

var _ model.ResourceList = &DataplaneResourceList{}

type DataplaneResourceList struct {
	Items      []*DataplaneResource
	Pagination model.Pagination
}

func (l *DataplaneResourceList) GetItems() []model.Resource {
	res := make([]model.Resource, len(l.Items))
	for i, elem := range l.Items {
		res[i] = elem
	}
	return res
}

func (l *DataplaneResourceList) GetItemType() model.ResourceType {
	return DataplaneType
}

func (l *DataplaneResourceList) NewItem() model.Resource {
	return NewDataplaneResource()
}

func (l *DataplaneResourceList) AddItem(r model.Resource) error {
	if trr, ok := r.(*DataplaneResource); ok {
		l.Items = append(l.Items, trr)
		return nil
	} else {
		return model.ErrorInvalidItemType((*DataplaneResource)(nil), r)
	}
}

func (l *DataplaneResourceList) GetPagination() *model.Pagination {
	return &l.Pagination
}

func (l *DataplaneResourceList) SetPagination(p model.Pagination) {
	l.Pagination = p
}

var DataplaneResourceTypeDescriptor = model.ResourceTypeDescriptor{
	Name:                  DataplaneType,
	Resource:              NewDataplaneResource(),
	ResourceList:          &DataplaneResourceList{},
	ReadOnly:              false,
	AdminOnly:             false,
	Scope:                 model.ScopeMesh,
	KDSFlags:              model.ZoneToGlobalFlag,
	WsPath:                "dataplanes",
	KumactlArg:            "dataplane",
	KumactlListArg:        "dataplanes",
	AllowToInspect:        false,
	IsPolicy:              false,
	SingularDisplayName:   "Dataplane",
	PluralDisplayName:     "Dataplanes",
	ShortName:             "dp",
	IsExperimental:        false,
	IsProxy:               true,
	AffectsPolicyMatching: false,
	Insight:               NewDataplaneInsightResource(),
	Overview:              NewDataplaneOverviewResource(),
}

func init() {
	registry.RegisterType(DataplaneResourceTypeDescriptor)
}

const (
	DataplaneInsightType model.ResourceType = "DataplaneInsight"
)

var _ model.Resource = &DataplaneInsightResource{}

type DataplaneInsightResource struct {
	Meta model.ResourceMeta
	Spec *mesh_proto.DataplaneInsight
}

func NewDataplaneInsightResource() *DataplaneInsightResource {
	return &DataplaneInsightResource{
		Spec: &mesh_proto.DataplaneInsight{},
	}
}

func (t *DataplaneInsightResource) GetMeta() model.ResourceMeta {
	return t.Meta
}

func (t *DataplaneInsightResource) SetMeta(m model.ResourceMeta) {
	t.Meta = m
}

func (t *DataplaneInsightResource) GetSpec() model.ResourceSpec {
	return t.Spec
}

func (t *DataplaneInsightResource) SetSpec(spec model.ResourceSpec) error {
	protoType, ok := spec.(*mesh_proto.DataplaneInsight)
	if !ok {
		return fmt.Errorf("invalid type %T for Spec", spec)
	} else {
		if protoType == nil {
			t.Spec = &mesh_proto.DataplaneInsight{}
		} else {
			t.Spec = protoType
		}
		return nil
	}
}

func (t *DataplaneInsightResource) GetStatus() model.ResourceStatus {
	return nil
}

func (t *DataplaneInsightResource) SetStatus(_ model.ResourceStatus) error {
	return errors.New("status not supported")
}

func (t *DataplaneInsightResource) Descriptor() model.ResourceTypeDescriptor {
	return DataplaneInsightResourceTypeDescriptor
}

var _ model.ResourceList = &DataplaneInsightResourceList{}

type DataplaneInsightResourceList struct {
	Items      []*DataplaneInsightResource
	Pagination model.Pagination
}

func (l *DataplaneInsightResourceList) GetItems() []model.Resource {
	res := make([]model.Resource, len(l.Items))
	for i, elem := range l.Items {
		res[i] = elem
	}
	return res
}

func (l *DataplaneInsightResourceList) GetItemType() model.ResourceType {
	return DataplaneInsightType
}

func (l *DataplaneInsightResourceList) NewItem() model.Resource {
	return NewDataplaneInsightResource()
}

func (l *DataplaneInsightResourceList) AddItem(r model.Resource) error {
	if trr, ok := r.(*DataplaneInsightResource); ok {
		l.Items = append(l.Items, trr)
		return nil
	} else {
		return model.ErrorInvalidItemType((*DataplaneInsightResource)(nil), r)
	}
}

func (l *DataplaneInsightResourceList) GetPagination() *model.Pagination {
	return &l.Pagination
}

func (l *DataplaneInsightResourceList) SetPagination(p model.Pagination) {
	l.Pagination = p
}

var DataplaneInsightResourceTypeDescriptor = model.ResourceTypeDescriptor{
	Name:                  DataplaneInsightType,
	Resource:              NewDataplaneInsightResource(),
	ResourceList:          &DataplaneInsightResourceList{},
	ReadOnly:              true,
	AdminOnly:             false,
	Scope:                 model.ScopeMesh,
	KDSFlags:              model.ZoneToGlobalFlag,
	WsPath:                "dataplane-insights",
	KumactlArg:            "",
	KumactlListArg:        "",
	AllowToInspect:        false,
	IsPolicy:              false,
	SingularDisplayName:   "Dataplane Insight",
	PluralDisplayName:     "Dataplane Insights",
	IsExperimental:        false,
	IsProxy:               false,
	AffectsPolicyMatching: false,
}

func init() {
	registry.RegisterType(DataplaneInsightResourceTypeDescriptor)
}

const (
	DataplaneOverviewType model.ResourceType = "DataplaneOverview"
)

var _ model.Resource = &DataplaneOverviewResource{}

type DataplaneOverviewResource struct {
	Meta model.ResourceMeta
	Spec *mesh_proto.DataplaneOverview
}

func NewDataplaneOverviewResource() *DataplaneOverviewResource {
	return &DataplaneOverviewResource{
		Spec: &mesh_proto.DataplaneOverview{},
	}
}

func (t *DataplaneOverviewResource) GetMeta() model.ResourceMeta {
	return t.Meta
}

func (t *DataplaneOverviewResource) SetMeta(m model.ResourceMeta) {
	t.Meta = m
}

func (t *DataplaneOverviewResource) GetSpec() model.ResourceSpec {
	return t.Spec
}

func (t *DataplaneOverviewResource) SetSpec(spec model.ResourceSpec) error {
	protoType, ok := spec.(*mesh_proto.DataplaneOverview)
	if !ok {
		return fmt.Errorf("invalid type %T for Spec", spec)
	} else {
		if protoType == nil {
			t.Spec = &mesh_proto.DataplaneOverview{}
		} else {
			t.Spec = protoType
		}
		return nil
	}
}

func (t *DataplaneOverviewResource) GetStatus() model.ResourceStatus {
	return nil
}

func (t *DataplaneOverviewResource) SetStatus(_ model.ResourceStatus) error {
	return errors.New("status not supported")
}

func (t *DataplaneOverviewResource) Descriptor() model.ResourceTypeDescriptor {
	return DataplaneOverviewResourceTypeDescriptor
}

// newDataplaneOverviewResourceSpec normalizes a nil spec the way SetSpec does. A nil one would
// drop the field from the response without any error: the marshaller omits a nil message
// and then throws the whole spec away once what is left renders as an empty object.
func newDataplaneOverviewResourceSpec(spec *mesh_proto.Dataplane) *mesh_proto.DataplaneOverview {
	if spec == nil {
		spec = &mesh_proto.Dataplane{}
	}
	return &mesh_proto.DataplaneOverview{
		Dataplane: spec,
	}
}

func (t *DataplaneOverviewResource) SetOverviewSpec(resource model.Resource, insight model.Resource) error {
	t.SetMeta(resource.GetMeta())
	overview := newDataplaneOverviewResourceSpec(resource.GetSpec().(*mesh_proto.Dataplane))
	if insight != nil {
		ins, ok := insight.GetSpec().(*mesh_proto.DataplaneInsight)
		if !ok {
			return errors.New("failed to convert to insight type 'DataplaneInsight'")
		}
		overview.DataplaneInsight = ins
	}
	return t.SetSpec(overview)
}

var _ model.ResourceList = &DataplaneOverviewResourceList{}

type DataplaneOverviewResourceList struct {
	Items      []*DataplaneOverviewResource
	Pagination model.Pagination
}

func (l *DataplaneOverviewResourceList) GetItems() []model.Resource {
	res := make([]model.Resource, len(l.Items))
	for i, elem := range l.Items {
		res[i] = elem
	}
	return res
}

func (l *DataplaneOverviewResourceList) GetItemType() model.ResourceType {
	return DataplaneOverviewType
}

func (l *DataplaneOverviewResourceList) NewItem() model.Resource {
	return NewDataplaneOverviewResource()
}

func (l *DataplaneOverviewResourceList) AddItem(r model.Resource) error {
	if trr, ok := r.(*DataplaneOverviewResource); ok {
		l.Items = append(l.Items, trr)
		return nil
	} else {
		return model.ErrorInvalidItemType((*DataplaneOverviewResource)(nil), r)
	}
}

func (l *DataplaneOverviewResourceList) GetPagination() *model.Pagination {
	return &l.Pagination
}

func (l *DataplaneOverviewResourceList) SetPagination(p model.Pagination) {
	l.Pagination = p
}

var DataplaneOverviewResourceTypeDescriptor = model.ResourceTypeDescriptor{
	Name:                  DataplaneOverviewType,
	Resource:              NewDataplaneOverviewResource(),
	ResourceList:          &DataplaneOverviewResourceList{},
	ReadOnly:              false,
	AdminOnly:             false,
	Scope:                 model.ScopeMesh,
	WsPath:                "",
	KumactlArg:            "",
	KumactlListArg:        "",
	AllowToInspect:        false,
	IsPolicy:              false,
	SingularDisplayName:   "Dataplane Overview",
	PluralDisplayName:     "Dataplane Overviews",
	IsExperimental:        false,
	IsProxy:               false,
	AffectsPolicyMatching: false,
}
