package kri

import (
	"fmt"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	"strings"
	"github.com/pkg/errors"
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

func NoSectionName(id Identifier) Identifier {
	idCopy := id
	idCopy.SectionName = ""
	return idCopy
}
