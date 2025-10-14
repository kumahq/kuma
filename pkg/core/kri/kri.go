package kri

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/exp/constraints"

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
	if i.IsEmpty() {
		return ""
	}

	desc, err := registry.Global().DescriptorFor(i.ResourceType)
	if err != nil {
		panic(err)
	}

	return fmt.Sprintf("kri_%s_%s_%s_%s_%s_%s", desc.ShortName, i.Mesh, i.Zone, i.Namespace, i.Name, i.SectionName)
}

func (i Identifier) IsEmpty() bool {
	return i == (Identifier{})
}

func (i Identifier) IsLocallyOriginated(isGlobal bool, zone string) bool {
	switch isGlobal {
	case true:
		// In Global CP, resources without a zone are considered locally originated.
		return i.Zone == ""
	case false:
		// In Zone CP, resources are treated as locally originated if KRI zone matches the current CP zone.
		return i.Zone == zone
	default:
		return true
	}
}

func From(r core_model.Resource) Identifier {
	return FromResourceMeta(r.GetMeta(), r.Descriptor().Name)
}

func FromResourceMeta(rm core_model.ResourceMeta, resourceType core_model.ResourceType) Identifier {
	if rm == nil {
		return Identifier{}
	}

	return Identifier{
		ResourceType: resourceType,
		Mesh:         rm.GetMesh(),
		Zone:         rm.GetLabels()[mesh_proto.ZoneTag],
		Namespace:    rm.GetLabels()[mesh_proto.KubeNamespaceTag],
		Name:         core_model.GetDisplayName(rm),
	}
}

func FromString(s string) (Identifier, error) {
	parts := strings.Split(s, "_")
	if len(parts) < 7 {
		return Identifier{}, errors.Errorf("invalid identifier string: %q", s)
	}
	if parts[0] != "kri" {
		return Identifier{}, errors.Errorf("identifier must start with 'kri': %q", s)
	}
	ds := registry.Global().ObjectDescriptors(core_model.TypeFilterFn(func(d core_model.ResourceTypeDescriptor) bool {
		return d.ShortName == parts[1] && d.ShortName != ""
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
		SectionName:  strings.Join(parts[6:], ""),
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
	if len(parts) < 7 {
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

func WithSectionName[T ~string | constraints.Unsigned](id Identifier, sectionName T) Identifier {
	// cannot add section name to empty identifier
	if id.IsEmpty() {
		return id
	}
	id.SectionName = fmt.Sprintf("%v", sectionName)
	return id
}
