package entities

import (
	core_system "github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/model"
)

type Definition struct {
	Singular     string
	Plural       string
	ResourceType model.ResourceType
	ReadOnly     bool
}

var All []Definition
var Names []string
var ByName = map[string]Definition{}

func init() {
	All = append(All, meshEntities...)
	All = append(All, systemEntities...)
	All = append(All, Definition{Singular: "global-secret", Plural: "global-secrets", ResourceType: core_system.GlobalSecretType})

	for i := range All {
		if All[i].Singular == "health-check" {
			// Preserving incoherency between kumactl and the web-service.
			All[i].Singular = "healthcheck"
			All[i].Plural = "healthchecks"
		}
		ByName[All[i].Singular] = All[i]
		Names = append(Names, All[i].Singular)
	}
}
