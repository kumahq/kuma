package kri

import (
	"fmt"
	"strings"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
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
	var pairs []string
	if i.ResourceType != "" {
		pairs = append(pairs, strings.ToLower(string(i.ResourceType)))
	}
	if i.Mesh != "" {
		pairs = append(pairs, fmt.Sprintf("mesh/%s", i.Mesh))
	}
	if i.Zone != "" {
		pairs = append(pairs, fmt.Sprintf("zone/%s", i.Zone))
	}
	if i.Namespace != "" {
		pairs = append(pairs, fmt.Sprintf("namespace/%s", i.Namespace))
	}
	if i.Name != "" {
		pairs = append(pairs, fmt.Sprintf("name/%s", i.Name))
	}
	if i.SectionName != "" {
		pairs = append(pairs, fmt.Sprintf("section/%s", i.SectionName))
	}
	return strings.Join(pairs, ":")
}

func From(r core_model.Resource, sectionName string) Identifier {
	return Identifier{
		ResourceType: r.Descriptor().Name,
		Mesh:         r.GetMeta().GetMesh(),
		Zone:         r.GetMeta().GetLabels()[mesh_proto.ZoneTag],
		Namespace:    r.GetMeta().GetLabels()[mesh_proto.KubeNamespaceTag],
		Name:         core_model.GetDisplayName(r.GetMeta()),
		SectionName:  sectionName,
	}
}

func NoSectionName(id Identifier) Identifier {
	idCopy := id
	idCopy.SectionName = ""
	return idCopy
}
