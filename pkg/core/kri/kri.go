package kri

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
)

type Identifier struct {
	ResourceType core_model.ResourceType
	Mesh         string
	Zone         string
	Namespace    string
	Name         string
	SectionName  string
}

func (i Identifier) MarshalText() ([]byte, error) {
	return []byte(i.String()), nil
}

func (i Identifier) String() string {
	desc, err := registry.Global().DescriptorFor(i.ResourceType)
	if err != nil {
		panic(err)
	}

	return fmt.Sprintf("kri_%s_%s_%s_%s_%s_%s", desc.ShortName, i.Mesh, i.Zone, i.Namespace, i.Name, i.SectionName)
}

func From(r core_model.Resource, sectionName string) Identifier {
	return FromResourceMeta(r.GetMeta(), r.Descriptor().Name, sectionName)
}

func FromResourceMeta(rm core_model.ResourceMeta, resourceType core_model.ResourceType, sectionName string) Identifier {
	return Identifier{
		ResourceType: resourceType,
		Mesh:         rm.GetMesh(),
		Zone:         rm.GetLabels()[mesh_proto.ZoneTag],
		Namespace:    rm.GetLabels()[mesh_proto.KubeNamespaceTag],
		Name:         core_model.GetDisplayName(rm),
		SectionName:  sectionName,
	}
}

func FromString(s string) (Identifier, error) {
	parts := strings.Split(s, "_")
	if len(parts) != 7 {
		return Identifier{}, errors.Errorf("invalid identifier string: %q", s)
	}
	if parts[0] != "kri" {
		return Identifier{}, errors.Errorf("identifier must start with 'kri': %q", s)
	}
	ds := registry.Global().ObjectDescriptors(core_model.TypeFilterFn(func(d core_model.ResourceTypeDescriptor) bool {
		return d.ShortName == parts[1]
	}))
	if len(ds) == 0 {
		return Identifier{}, errors.Errorf("unknown short name of resource type: %q", parts[1])
	}
	return Identifier{
		ResourceType: ds[0].Name,
		Mesh:         parts[2],
		Zone:         parts[3],
		Namespace:    parts[4],
		Name:         parts[5],
		SectionName:  parts[6],
	}, nil
}

func MustFromString(s string) Identifier {
	id, err := FromString(s)
	if err != nil {
		panic(err)
	}
	return id
}

func NoSectionName(id Identifier) Identifier {
	idCopy := id
	idCopy.SectionName = ""
	return idCopy
}

func IsValid(s string) bool {
	parts := strings.Split(s, "_")
	if len(parts) != 7 {
		return false
	}
	if parts[0] != "kri" {
		return false
	}
	return true
}

func Compare(a, b Identifier) int {
	return strings.Compare(a.String(), b.String())
}

func (i Identifier) HasSectionName() bool {
	return i.SectionName != ""
}

func WithSectionName(id Identifier, sectionName string) Identifier {
	idCopy := id
	idCopy.SectionName = sectionName
	return idCopy
}
