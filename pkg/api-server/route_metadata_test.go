package api_server

import (
	"net/http"

	"github.com/emicklei/go-restful/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/v2/pkg/api-server/authn"
	"github.com/kumahq/kuma/v2/pkg/core/resources/model"
	"github.com/kumahq/kuma/v2/pkg/core/runtime"
)

var _ = Describe("route metadata provider", func() {
	descriptor := model.ResourceTypeDescriptor{
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
		provider := func(model.ResourceTypeDescriptor, string) map[string]string {
			return map[string]string{"x-test": "v"}
		}
		endpoints := resourceEndpoints{
			descriptor:            descriptor,
			routeMetadataProvider: provider,
		}
		ws := new(restful.WebService)

		endpoints.addCreateOrUpdateEndpoint(ws, "/meshes/{mesh}/"+descriptor.WsPath)

		put := putRoute(ws)
		Expect(put).ToNot(BeNil())
		Expect(put.Metadata).To(HaveKeyWithValue("x-test", "v"))
	})

	It("attaches no metadata when the provider is nil", func() {
		endpoints := resourceEndpoints{
			descriptor:            descriptor,
			routeMetadataProvider: nil,
		}
		ws := new(restful.WebService)

		endpoints.addCreateOrUpdateEndpoint(ws, "/meshes/{mesh}/"+descriptor.WsPath)

		put := putRoute(ws)
		Expect(put).ToNot(BeNil())
		Expect(put.Metadata).ToNot(HaveKey("x-test"))
	})

	It("drops reserved keys but keeps the rest", func() {
		provider := func(model.ResourceTypeDescriptor, string) map[string]string {
			return map[string]string{authn.MetadataAuthKey: authn.MetadataAuthSkip, "x-ok": "v"}
		}
		endpoints := resourceEndpoints{
			descriptor:            descriptor,
			routeMetadataProvider: provider,
		}
		ws := new(restful.WebService)

		endpoints.addCreateOrUpdateEndpoint(ws, "/meshes/{mesh}/"+descriptor.WsPath)

		put := putRoute(ws)
		Expect(put).ToNot(BeNil())
		Expect(put.Metadata).To(HaveKeyWithValue("x-ok", "v"))
		Expect(put.Metadata).ToNot(HaveKey(authn.MetadataAuthKey))
	})

	It("is satisfied by the runtime RouteMetadataProvider type", func() {
		var _ runtime.RouteMetadataProvider = func(model.ResourceTypeDescriptor, string) map[string]string {
			return nil
		}
	})
})
