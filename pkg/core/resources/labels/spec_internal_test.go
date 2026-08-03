package labels

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	config_core "github.com/kumahq/kuma/v3/pkg/config/core"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
)

var _ = Describe("requiredOnOverlap", func() {
	It("treats identical predicates as overlapping", func() {
		r := RequiredOn{Environments: []config_core.EnvironmentType{config_core.KubernetesEnvironment}}
		Expect(requiredOnOverlap(r, r)).To(BeTrue())
	})

	It("treats disjoint environments as non-overlapping", func() {
		a := RequiredOn{Environments: []config_core.EnvironmentType{config_core.KubernetesEnvironment}}
		b := RequiredOn{Environments: []config_core.EnvironmentType{config_core.UniversalEnvironment}}
		Expect(requiredOnOverlap(a, b)).To(BeFalse())
	})

	It("treats an empty dimension as 'any', so it overlaps a constrained one", func() {
		a := RequiredOn{}
		b := RequiredOn{ResourceTypes: []core_model.ResourceType{core_mesh.DataplaneType}}
		Expect(requiredOnOverlap(a, b)).To(BeTrue())
	})

	It("is disjoint when any single-valued dimension is disjoint", func() {
		a := RequiredOn{
			Modes:        []config_core.CpMode{config_core.Zone},
			Environments: []config_core.EnvironmentType{config_core.KubernetesEnvironment},
		}
		b := RequiredOn{
			Modes:        []config_core.CpMode{config_core.Global},
			Environments: []config_core.EnvironmentType{config_core.KubernetesEnvironment},
		}
		Expect(requiredOnOverlap(a, b)).To(BeFalse())
	})
})
