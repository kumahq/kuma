// Generated by tools/resource-gen.
// Run "make generate" to update this file.

// nolint:whitespace
package system

import (
	"fmt"

	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
)

const (
	ConfigType model.ResourceType = "Config"
)

var _ model.Resource = &ConfigResource{}

type ConfigResource struct {
	Meta model.ResourceMeta
	Spec *system_proto.Config
}

func NewConfigResource() *ConfigResource {
	return &ConfigResource{
		Spec: &system_proto.Config{},
	}
}

func (t *ConfigResource) GetMeta() model.ResourceMeta {
	return t.Meta
}

func (t *ConfigResource) SetMeta(m model.ResourceMeta) {
	t.Meta = m
}

func (t *ConfigResource) GetSpec() model.ResourceSpec {
	return t.Spec
}

func (t *ConfigResource) SetSpec(spec model.ResourceSpec) error {
	protoType, ok := spec.(*system_proto.Config)
	if !ok {
		return fmt.Errorf("invalid type %T for Spec", spec)
	} else {
		if protoType == nil {
			t.Spec = &system_proto.Config{}
		} else {
			t.Spec = protoType
		}
		return nil
	}
}

func (t *ConfigResource) Descriptor() model.ResourceTypeDescriptor {
	return ConfigResourceTypeDescriptor
}

var _ model.ResourceList = &ConfigResourceList{}

type ConfigResourceList struct {
	Items      []*ConfigResource
	Pagination model.Pagination
}

func (l *ConfigResourceList) Descriptor() model.ResourceTypeDescriptor {
	return ConfigResourceTypeDescriptor
}

func (l *ConfigResourceList) GetItems() []model.Resource {
	res := make([]model.Resource, len(l.Items))
	for i, elem := range l.Items {
		res[i] = elem
	}
	return res
}

func (l *ConfigResourceList) AddItem(r model.Resource) error {
	if trr, ok := r.(*ConfigResource); ok {
		l.Items = append(l.Items, trr)
		return nil
	} else {
		return model.ErrorInvalidItemType((*ConfigResource)(nil), r)
	}
}

func (l *ConfigResourceList) GetPagination() *model.Pagination {
	return &l.Pagination
}

var ConfigResourceTypeDescriptor = model.ResourceTypeDescriptor{
	Name:                ConfigType,
	Resource:            NewConfigResource(),
	ResourceList:        &ConfigResourceList{},
	ReadOnly:            false,
	AdminOnly:           false,
	Scope:               model.ScopeGlobal,
	KDSFlags:            model.FromGlobalToZone,
	WsPath:              "",
	KumactlArg:          "",
	KumactlListArg:      "",
	AllowToInspect:      false,
	IsPolicy:            false,
	SingularDisplayName: "Config",
	PluralDisplayName:   "Configs",
	IsExperimental:      false,
}

func init() {
	registry.RegisterType(ConfigResourceTypeDescriptor)
}

const (
	SecretType model.ResourceType = "Secret"
)

var _ model.Resource = &SecretResource{}

type SecretResource struct {
	Meta model.ResourceMeta
	Spec *system_proto.Secret
}

func NewSecretResource() *SecretResource {
	return &SecretResource{
		Spec: &system_proto.Secret{},
	}
}

func (t *SecretResource) GetMeta() model.ResourceMeta {
	return t.Meta
}

func (t *SecretResource) SetMeta(m model.ResourceMeta) {
	t.Meta = m
}

func (t *SecretResource) GetSpec() model.ResourceSpec {
	return t.Spec
}

func (t *SecretResource) SetSpec(spec model.ResourceSpec) error {
	protoType, ok := spec.(*system_proto.Secret)
	if !ok {
		return fmt.Errorf("invalid type %T for Spec", spec)
	} else {
		if protoType == nil {
			t.Spec = &system_proto.Secret{}
		} else {
			t.Spec = protoType
		}
		return nil
	}
}

func (t *SecretResource) Descriptor() model.ResourceTypeDescriptor {
	return SecretResourceTypeDescriptor
}

var _ model.ResourceList = &SecretResourceList{}

type SecretResourceList struct {
	Items      []*SecretResource
	Pagination model.Pagination
}

func (l *SecretResourceList) Descriptor() model.ResourceTypeDescriptor {
	return SecretResourceTypeDescriptor
}

func (l *SecretResourceList) GetItems() []model.Resource {
	res := make([]model.Resource, len(l.Items))
	for i, elem := range l.Items {
		res[i] = elem
	}
	return res
}

func (l *SecretResourceList) AddItem(r model.Resource) error {
	if trr, ok := r.(*SecretResource); ok {
		l.Items = append(l.Items, trr)
		return nil
	} else {
		return model.ErrorInvalidItemType((*SecretResource)(nil), r)
	}
}

func (l *SecretResourceList) GetPagination() *model.Pagination {
	return &l.Pagination
}

var SecretResourceTypeDescriptor = model.ResourceTypeDescriptor{
	Name:                SecretType,
	Resource:            NewSecretResource(),
	ResourceList:        &SecretResourceList{},
	ReadOnly:            false,
	AdminOnly:           true,
	Scope:               model.ScopeMesh,
	KDSFlags:            model.FromGlobalToZone,
	WsPath:              "secrets",
	KumactlArg:          "secret",
	KumactlListArg:      "secrets",
	AllowToInspect:      false,
	IsPolicy:            false,
	SingularDisplayName: "Secret",
	PluralDisplayName:   "Secrets",
	IsExperimental:      false,
}

func init() {
	registry.RegisterType(SecretResourceTypeDescriptor)
}

const (
	ZoneType model.ResourceType = "Zone"
)

var _ model.Resource = &ZoneResource{}

type ZoneResource struct {
	Meta model.ResourceMeta
	Spec *system_proto.Zone
}

func NewZoneResource() *ZoneResource {
	return &ZoneResource{
		Spec: &system_proto.Zone{},
	}
}

func (t *ZoneResource) GetMeta() model.ResourceMeta {
	return t.Meta
}

func (t *ZoneResource) SetMeta(m model.ResourceMeta) {
	t.Meta = m
}

func (t *ZoneResource) GetSpec() model.ResourceSpec {
	return t.Spec
}

func (t *ZoneResource) SetSpec(spec model.ResourceSpec) error {
	protoType, ok := spec.(*system_proto.Zone)
	if !ok {
		return fmt.Errorf("invalid type %T for Spec", spec)
	} else {
		if protoType == nil {
			t.Spec = &system_proto.Zone{}
		} else {
			t.Spec = protoType
		}
		return nil
	}
}

func (t *ZoneResource) Descriptor() model.ResourceTypeDescriptor {
	return ZoneResourceTypeDescriptor
}

var _ model.ResourceList = &ZoneResourceList{}

type ZoneResourceList struct {
	Items      []*ZoneResource
	Pagination model.Pagination
}

func (l *ZoneResourceList) Descriptor() model.ResourceTypeDescriptor {
	return ZoneResourceTypeDescriptor
}

func (l *ZoneResourceList) GetItems() []model.Resource {
	res := make([]model.Resource, len(l.Items))
	for i, elem := range l.Items {
		res[i] = elem
	}
	return res
}

func (l *ZoneResourceList) AddItem(r model.Resource) error {
	if trr, ok := r.(*ZoneResource); ok {
		l.Items = append(l.Items, trr)
		return nil
	} else {
		return model.ErrorInvalidItemType((*ZoneResource)(nil), r)
	}
}

func (l *ZoneResourceList) GetPagination() *model.Pagination {
	return &l.Pagination
}

var ZoneResourceTypeDescriptor = model.ResourceTypeDescriptor{
	Name:                ZoneType,
	Resource:            NewZoneResource(),
	ResourceList:        &ZoneResourceList{},
	ReadOnly:            false,
	AdminOnly:           false,
	Scope:               model.ScopeGlobal,
	WsPath:              "zones",
	KumactlArg:          "zone",
	KumactlListArg:      "zones",
	AllowToInspect:      false,
	IsPolicy:            false,
	SingularDisplayName: "Zone",
	PluralDisplayName:   "Zones",
	IsExperimental:      false,
}

func init() {
	registry.RegisterType(ZoneResourceTypeDescriptor)
}

const (
	ZoneInsightType model.ResourceType = "ZoneInsight"
)

var _ model.Resource = &ZoneInsightResource{}

type ZoneInsightResource struct {
	Meta model.ResourceMeta
	Spec *system_proto.ZoneInsight
}

func NewZoneInsightResource() *ZoneInsightResource {
	return &ZoneInsightResource{
		Spec: &system_proto.ZoneInsight{},
	}
}

func (t *ZoneInsightResource) GetMeta() model.ResourceMeta {
	return t.Meta
}

func (t *ZoneInsightResource) SetMeta(m model.ResourceMeta) {
	t.Meta = m
}

func (t *ZoneInsightResource) GetSpec() model.ResourceSpec {
	return t.Spec
}

func (t *ZoneInsightResource) SetSpec(spec model.ResourceSpec) error {
	protoType, ok := spec.(*system_proto.ZoneInsight)
	if !ok {
		return fmt.Errorf("invalid type %T for Spec", spec)
	} else {
		if protoType == nil {
			t.Spec = &system_proto.ZoneInsight{}
		} else {
			t.Spec = protoType
		}
		return nil
	}
}

func (t *ZoneInsightResource) Descriptor() model.ResourceTypeDescriptor {
	return ZoneInsightResourceTypeDescriptor
}

var _ model.ResourceList = &ZoneInsightResourceList{}

type ZoneInsightResourceList struct {
	Items      []*ZoneInsightResource
	Pagination model.Pagination
}

func (l *ZoneInsightResourceList) Descriptor() model.ResourceTypeDescriptor {
	return ZoneInsightResourceTypeDescriptor
}

func (l *ZoneInsightResourceList) GetItems() []model.Resource {
	res := make([]model.Resource, len(l.Items))
	for i, elem := range l.Items {
		res[i] = elem
	}
	return res
}

func (l *ZoneInsightResourceList) AddItem(r model.Resource) error {
	if trr, ok := r.(*ZoneInsightResource); ok {
		l.Items = append(l.Items, trr)
		return nil
	} else {
		return model.ErrorInvalidItemType((*ZoneInsightResource)(nil), r)
	}
}

func (l *ZoneInsightResourceList) GetPagination() *model.Pagination {
	return &l.Pagination
}

var ZoneInsightResourceTypeDescriptor = model.ResourceTypeDescriptor{
	Name:                ZoneInsightType,
	Resource:            NewZoneInsightResource(),
	ResourceList:        &ZoneInsightResourceList{},
	ReadOnly:            true,
	AdminOnly:           false,
	Scope:               model.ScopeGlobal,
	WsPath:              "zone-insights",
	KumactlArg:          "",
	KumactlListArg:      "",
	AllowToInspect:      false,
	IsPolicy:            false,
	SingularDisplayName: "Zone Insight",
	PluralDisplayName:   "Zone Insights",
	IsExperimental:      false,
}

func init() {
	registry.RegisterType(ZoneInsightResourceTypeDescriptor)
}

const (
	ZoneOverviewType model.ResourceType = "ZoneOverview"
)

var _ model.Resource = &ZoneOverviewResource{}

type ZoneOverviewResource struct {
	Meta model.ResourceMeta
	Spec *system_proto.ZoneOverview
}

func NewZoneOverviewResource() *ZoneOverviewResource {
	return &ZoneOverviewResource{
		Spec: &system_proto.ZoneOverview{},
	}
}

func (t *ZoneOverviewResource) GetMeta() model.ResourceMeta {
	return t.Meta
}

func (t *ZoneOverviewResource) SetMeta(m model.ResourceMeta) {
	t.Meta = m
}

func (t *ZoneOverviewResource) GetSpec() model.ResourceSpec {
	return t.Spec
}

func (t *ZoneOverviewResource) SetSpec(spec model.ResourceSpec) error {
	protoType, ok := spec.(*system_proto.ZoneOverview)
	if !ok {
		return fmt.Errorf("invalid type %T for Spec", spec)
	} else {
		if protoType == nil {
			t.Spec = &system_proto.ZoneOverview{}
		} else {
			t.Spec = protoType
		}
		return nil
	}
}

func (t *ZoneOverviewResource) Descriptor() model.ResourceTypeDescriptor {
	return ZoneOverviewResourceTypeDescriptor
}

var _ model.ResourceList = &ZoneOverviewResourceList{}

type ZoneOverviewResourceList struct {
	Items      []*ZoneOverviewResource
	Pagination model.Pagination
}

func (l *ZoneOverviewResourceList) Descriptor() model.ResourceTypeDescriptor {
	return ZoneOverviewResourceTypeDescriptor
}

func (l *ZoneOverviewResourceList) GetItems() []model.Resource {
	res := make([]model.Resource, len(l.Items))
	for i, elem := range l.Items {
		res[i] = elem
	}
	return res
}

func (l *ZoneOverviewResourceList) AddItem(r model.Resource) error {
	if trr, ok := r.(*ZoneOverviewResource); ok {
		l.Items = append(l.Items, trr)
		return nil
	} else {
		return model.ErrorInvalidItemType((*ZoneOverviewResource)(nil), r)
	}
}

func (l *ZoneOverviewResourceList) GetPagination() *model.Pagination {
	return &l.Pagination
}

var ZoneOverviewResourceTypeDescriptor = model.ResourceTypeDescriptor{
	Name:                ZoneOverviewType,
	Resource:            NewZoneOverviewResource(),
	ResourceList:        &ZoneOverviewResourceList{},
	ReadOnly:            false,
	AdminOnly:           false,
	Scope:               model.ScopeGlobal,
	WsPath:              "",
	KumactlArg:          "",
	KumactlListArg:      "",
	AllowToInspect:      false,
	IsPolicy:            false,
	SingularDisplayName: "Zone Overview",
	PluralDisplayName:   "Zone Overviews",
	IsExperimental:      false,
}
