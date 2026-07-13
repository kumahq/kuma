package api_server

import (
	"net/http"

	"github.com/emicklei/go-restful/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/core/runtime"
)

var _ = Describe("route metadata provider", func() {
	descriptor := core_model.ResourceTypeDescriptor{
		Name:   "MeshTrafficPermission",
		WsPath: "meshtrafficpermissions",
	}

	putRoute := func(ws *restful.WebService) *restful.Route {
		for i := range ws.Routes() {
			if ws.Routes()[i].Method == http.MethodPut {
				return &ws.Routes()[i]
			}
		}
		return nil
	}

	It("attaches provider metadata to the generated PUT route", func() {
		provider := func(core_model.ResourceTypeDescriptor, string) map[string]string {
			return map[string]string{"x-test": "v"}
		}
		endpoints := resourceEndpoints{
			resourceEndpointsContext: resourceEndpointsContext{
				descriptor:            descriptor,
				routeMetadataProvider: provider,
			},
		}
		ws := new(restful.WebService)

		endpoints.addCreateOrUpdateEndpoint(ws, "/meshes/{mesh}/"+descriptor.WsPath)

		put := putRoute(ws)
		Expect(put).ToNot(BeNil())
		Expect(put.Metadata).To(HaveKeyWithValue("x-test", "v"))
	})

	It("attaches no metadata when the provider is nil", func() {
		endpoints := resourceEndpoints{
			resourceEndpointsContext: resourceEndpointsContext{
				descriptor:            descriptor,
				routeMetadataProvider: nil,
			},
		}
		ws := new(restful.WebService)

		endpoints.addCreateOrUpdateEndpoint(ws, "/meshes/{mesh}/"+descriptor.WsPath)

		put := putRoute(ws)
		Expect(put).ToNot(BeNil())
		Expect(put.Metadata).ToNot(HaveKey("x-test"))
	})

	It("is satisfied by the runtime RouteMetadataProvider type", func() {
		var _ runtime.RouteMetadataProvider = func(core_model.ResourceTypeDescriptor, string) map[string]string {
			return nil
		}
	})
})
