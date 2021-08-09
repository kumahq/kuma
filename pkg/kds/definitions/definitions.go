// Package kds provides support of Kuma Discovery Service, extension of xDS
package definitions

import (
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/model"
)

const (
	googleApis = "type.googleapis.com/"

	// KumaResource is the type URL of the KumaResource protobuf.
	KumaResource = googleApis + "kuma.mesh.v1alpha1.KumaResource"
)

type KdsDefinition struct {
	Type      model.ResourceType
	Direction KDSFlagType
}

type KDSFlagType uint32

const (
	ConsumedByZone   = KDSFlagType(1)
	ConsumedByGlobal = KDSFlagType(1 << 2)
	ProvidedByZone   = KDSFlagType(1 << 3)
	ProvidedByGlobal = KDSFlagType(1 << 4)
	FromZoneToGlobal = ConsumedByGlobal | ProvidedByZone
	FromGlobalToZone = ProvidedByGlobal | ConsumedByZone
)

type kdsTypes []KdsDefinition

// Get return a list of all model.ResourceType
func (kt kdsTypes) Get() []model.ResourceType {
	var res []model.ResourceType
	for i := range kt {
		res = append(res, kt[i].Type)
	}
	return res
}

// Select return a list of all model.ResourceType that have this flag
func (kt kdsTypes) Select(flag KDSFlagType) []model.ResourceType {
	var res []model.ResourceType
	for i := range kt {
		if kt[i].Direction&flag != 0 {
			res = append(res, kt[i].Type)
		}
	}
	return res
}

func (kt kdsTypes) TypeHasFlag(resourceType model.ResourceType, flag KDSFlagType) bool {
	for _, consumedTyp := range kt {
		if consumedTyp.Type == resourceType {
			return 0 != (consumedTyp.Direction & flag)
		}
	}
	return false
}

var All kdsTypes = append(append([]KdsDefinition{
	{
		Type:      system.GlobalSecretType,
		Direction: FromGlobalToZone,
	},
}, meshDefinitions...), systemDefinitions...)
