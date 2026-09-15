package api_server

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/emicklei/go-restful/v3"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/v3/pkg/core/resources/apis/system"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
)

func TestResourceRoutes(t *testing.T) {
	tests := []struct {
		name       string
		descriptor core_model.ResourceTypeDescriptor
		expected   []resourceRoute
	}{
		{
			name: "mesh primary path",
			descriptor: core_model.ResourceTypeDescriptor{
				Scope:  core_model.ScopeMesh,
				WsPath: "resources",
			},
			expected: []resourceRoute{
				{role: meshCRUDListRoute, pathRole: primaryResourcePath},
				{role: crossMeshListRoute, pathRole: primaryResourcePath},
			},
		},
		{
			name: "mesh primary and alias paths",
			descriptor: core_model.ResourceTypeDescriptor{
				Scope:             core_model.ScopeMesh,
				WsPath:            "resources",
				AlternativeWsPath: "resource-aliases",
			},
			expected: []resourceRoute{
				{role: meshCRUDListRoute, pathRole: primaryResourcePath},
				{role: crossMeshListRoute, pathRole: primaryResourcePath},
				{role: meshCRUDListRoute, pathRole: aliasResourcePath},
				{role: crossMeshListRoute, pathRole: aliasResourcePath},
			},
		},
		{
			name:       "global primary and alias paths",
			descriptor: system.GlobalSecretResourceTypeDescriptor,
			expected: []resourceRoute{
				{role: globalCRUDListRoute, pathRole: primaryResourcePath},
				{role: globalCRUDListRoute, pathRole: aliasResourcePath},
			},
		},
		{
			name: "unknown scope",
			descriptor: core_model.ResourceTypeDescriptor{
				Scope:  "Unknown",
				WsPath: "resources",
			},
			expected: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewWithT(t)

			g.Expect(resourceRoutes(test.descriptor)).To(Equal(test.expected))
		})
	}
}

func TestRegisterResourceRoutes(t *testing.T) {
	t.Run("preserves mesh route and metadata order", func(t *testing.T) {
		for _, readOnly := range []bool{false, true} {
			t.Run(fmt.Sprintf("read-only=%t", readOnly), func(t *testing.T) {
				g := NewWithT(t)
				descriptor := core_model.ResourceTypeDescriptor{
					Name:              "Resource",
					Scope:             core_model.ScopeMesh,
					WsPath:            "resources",
					AlternativeWsPath: "resource-aliases",
					ReadOnly:          readOnly,
				}
				var metadataCalls []string
				endpoints := resourceEndpoints{
					resourceCrudHandler: &resourceCrudHandler{
						descriptor: descriptor,
						routeMetadataProvider: func(actual core_model.ResourceTypeDescriptor, method string) map[string]string {
							g.Expect(actual).To(Equal(descriptor))
							metadataCalls = append(metadataCalls, method)
							return nil
						},
					},
				}
				ws := new(restful.WebService)

				registerResourceRoutes(ws, endpoints)

				g.Expect(routeMethodPaths(ws)).To(Equal([]string{
					http.MethodPut + " /meshes/{mesh}/resources/{name}",
					http.MethodDelete + " /meshes/{mesh}/resources/{name}",
					http.MethodGet + " /meshes/{mesh}/resources/{name}",
					http.MethodGet + " /meshes/{mesh}/resources",
					http.MethodGet + " /resources",
					http.MethodPut + " /meshes/{mesh}/resource-aliases/{name}",
					http.MethodDelete + " /meshes/{mesh}/resource-aliases/{name}",
					http.MethodGet + " /meshes/{mesh}/resource-aliases/{name}",
					http.MethodGet + " /meshes/{mesh}/resource-aliases",
					http.MethodGet + " /resource-aliases",
				}))
				g.Expect(metadataCalls).To(Equal([]string{
					http.MethodPut,
					http.MethodDelete,
					http.MethodGet,
					http.MethodGet,
					http.MethodGet,
					http.MethodPut,
					http.MethodDelete,
					http.MethodGet,
					http.MethodGet,
					http.MethodGet,
				}))
			})
		}
	})

	t.Run("preserves GlobalSecret primary and alias paths", func(t *testing.T) {
		g := NewWithT(t)
		endpoints := resourceEndpoints{
			resourceCrudHandler: &resourceCrudHandler{
				descriptor: system.GlobalSecretResourceTypeDescriptor,
			},
		}
		ws := new(restful.WebService)

		registerResourceRoutes(ws, endpoints)

		g.Expect(routeMethodPaths(ws)).To(Equal([]string{
			http.MethodPut + " /globalsecrets/{name}",
			http.MethodDelete + " /globalsecrets/{name}",
			http.MethodGet + " /globalsecrets/{name}",
			http.MethodGet + " /globalsecrets",
			http.MethodPut + " /global-secrets/{name}",
			http.MethodDelete + " /global-secrets/{name}",
			http.MethodGet + " /global-secrets/{name}",
			http.MethodGet + " /global-secrets",
		}))
	})
}

func routeMethodPaths(ws *restful.WebService) []string {
	routes := ws.Routes()
	result := make([]string, len(routes))
	for i, route := range routes {
		result[i] = route.Method + " " + route.Path
	}
	return result
}
