package api_server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/emicklei/go-restful/v3"
	. "github.com/onsi/gomega"

	api_types "github.com/kumahq/kuma/v3/api/openapi/types"
	"github.com/kumahq/kuma/v3/pkg/api-server/types"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
)

func TestCatalogReadOnlyParity(t *testing.T) {
	descriptors := []core_model.ResourceTypeDescriptor{
		catalogPolicy("GlobalProvided", core_model.GlobalToZonesFlag, false),
		catalogPolicy("ZoneProvided", core_model.ZoneToGlobalFlag, false),
		catalogPolicy("ProvidedByBoth", core_model.GlobalToZonesFlag|core_model.ZoneToGlobalFlag, false),
		catalogPolicy("NoProvider", core_model.KDSDisabledFlag, false),
		catalogPolicy("IntrinsicReadOnly", core_model.GlobalToZonesFlag|core_model.ZoneToGlobalFlag, true),
	}
	tests := []struct {
		name          string
		isGlobal      bool
		federatedZone bool
		readOnly      bool
		expected      map[string]bool
	}{
		{
			name: "standalone Zone",
			expected: map[string]bool{
				"GlobalProvided":    false,
				"ZoneProvided":      false,
				"ProvidedByBoth":    false,
				"NoProvider":        false,
				"IntrinsicReadOnly": true,
			},
		},
		{
			name:     "Global",
			isGlobal: true,
			expected: map[string]bool{
				"GlobalProvided":    false,
				"ZoneProvided":      true,
				"ProvidedByBoth":    false,
				"NoProvider":        true,
				"IntrinsicReadOnly": true,
			},
		},
		{
			name:          "federated Zone",
			federatedZone: true,
			expected: map[string]bool{
				"GlobalProvided":    true,
				"ZoneProvided":      false,
				"ProvidedByBoth":    false,
				"NoProvider":        true,
				"IntrinsicReadOnly": true,
			},
		},
		{
			name:     "API read-only",
			readOnly: true,
			expected: map[string]bool{
				"GlobalProvided":    true,
				"ZoneProvided":      true,
				"ProvidedByBoth":    true,
				"NoProvider":        true,
				"IntrinsicReadOnly": true,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewWithT(t)
			policies, resources := catalogResponses(g, descriptors, test.isGlobal, test.federatedZone, test.readOnly)

			g.Expect(policies.Policies).To(HaveLen(len(test.expected)))
			g.Expect(resources.Resources).To(HaveLen(len(test.expected)))
			resourceReadOnly := map[string]bool{}
			for _, resource := range resources.Resources {
				resourceReadOnly[resource.Name] = resource.ReadOnly
			}
			for _, policy := range policies.Policies {
				g.Expect(policy.ReadOnly).To(Equal(test.expected[policy.Name]), policy.Name)
				g.Expect(resourceReadOnly).To(HaveKeyWithValue(policy.Name, policy.ReadOnly))
			}
		})
	}
}

func catalogPolicy(name core_model.ResourceType, flags core_model.KDSFlagType, readOnly bool) core_model.ResourceTypeDescriptor {
	return core_model.ResourceTypeDescriptor{
		Name:     name,
		WsPath:   string(name),
		Scope:    core_model.ScopeMesh,
		KDSFlags: flags,
		ReadOnly: readOnly,
		IsPolicy: true,
	}
}

func catalogResponses(
	g *WithT,
	descriptors []core_model.ResourceTypeDescriptor,
	isGlobal bool,
	federatedZone bool,
	readOnly bool,
) (types.PoliciesResponse, api_types.ResourceTypeDescriptionList) {
	g.THelper()
	ws := new(restful.WebService)
	addPoliciesWsEndpoints(ws, isGlobal, federatedZone, readOnly, descriptors)
	container := restful.NewContainer()
	container.Add(ws)

	get := func(path string, target any) {
		request := httptest.NewRequestWithContext(context.Background(), http.MethodGet, path, http.NoBody)
		response := httptest.NewRecorder()
		container.ServeHTTP(response, request)
		g.Expect(response.Code).To(Equal(http.StatusOK))
		g.Expect(json.Unmarshal(response.Body.Bytes(), target)).To(Succeed())
	}

	policies := types.PoliciesResponse{}
	resources := api_types.ResourceTypeDescriptionList{}
	get("/policies", &policies)
	get("/_resources", &resources)
	return policies, resources
}
