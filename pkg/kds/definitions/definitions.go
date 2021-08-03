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
	Direction KdsTypeFlag
}

type KdsTypeFlag uint32

const (
	ConsumedByZone   = KdsTypeFlag(1)
	ConsumedByGlobal = KdsTypeFlag(1 << 2)
	ProvidedByZone   = KdsTypeFlag(1 << 3)
	ProvidedByGlobal = KdsTypeFlag(1 << 4)
	SendEverywhere   = ConsumedByZone | ConsumedByGlobal | ProvidedByZone | ProvidedByGlobal
	FromZoneToGlobal = ConsumedByGlobal | ProvidedByZone
	FromGlobalToZone = ProvidedByGlobal | ConsumedByZone
)

type kdsTypes []KdsDefinition

func (kt kdsTypes) Get(flag KdsTypeFlag) []model.ResourceType {
	var res []model.ResourceType
	for i := range kt {
		if flag == 0 || kt[i].Direction&flag != 0 {
			res = append(res, kt[i].Type)
		}
	}
	return res
}

func (kt kdsTypes) TypeHasFlag(resourceType model.ResourceType, flag KdsTypeFlag) bool {
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
