package api_server

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/emicklei/go-restful/v3"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	config_core "github.com/kumahq/kuma/v3/pkg/config/core"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/core/resources/store"
	core_user "github.com/kumahq/kuma/v3/pkg/core/user"
	test_model "github.com/kumahq/kuma/v3/pkg/test/resources/model"
)

func TestCreateOrUpdateResourcePrecedence(t *testing.T) {
	const validBody = `{"type":"Mesh","name":"mesh-1"}`
	lookupErr := errors.New("lookup failed")
	accessErr := errors.New("access denied")

	t.Run("parent Mesh lookup precedes decoding", func(t *testing.T) {
		g := NewWithT(t)
		events := []string{}
		resManager := &recordingResourceManager{events: &events, getErr: lookupErr}
		handler := newContractCrudHandler(resManager, &recordingResourceAccess{events: &events})
		handler.descriptor = core_mesh.DataplaneResourceTypeDescriptor
		request := newCrudRequest(http.MethodPut, "/meshes/missing/dataplanes/dp-1", "dp-1", "{")
		request.PathParameters()["mesh"] = "missing"

		_, err := handler.createOrUpdateResource(request)

		expectTitledError(g, err, "Failed to retrieve Mesh")
		g.Expect(errors.Is(err, lookupErr)).To(BeTrue())
		g.Expect(events).To(Equal([]string{"lookup"}))
	})

	t.Run("decoding precedes resource lookup", func(t *testing.T) {
		g := NewWithT(t)
		events := []string{}
		resManager := &recordingResourceManager{events: &events, getErr: lookupErr}
		handler := newContractCrudHandler(resManager, &recordingResourceAccess{events: &events})

		_, err := handler.createOrUpdateResource(newCrudRequest(http.MethodPut, "/meshes/mesh-1", "mesh-1", "{"))

		expectTitledError(g, err, "Could not process a resource")
		g.Expect(events).To(BeEmpty())
	})

	t.Run("lookup failure precedes request validation", func(t *testing.T) {
		g := NewWithT(t)
		events := []string{}
		resManager := &recordingResourceManager{events: &events, getErr: lookupErr}
		handler := newContractCrudHandler(resManager, &recordingResourceAccess{events: &events})

		_, err := handler.createOrUpdateResource(newCrudRequest(http.MethodPut, "/meshes/mesh-1", "mesh-1", `{"type":"Mesh","name":"other"}`))

		expectTitledError(g, err, "Failed to find a resource")
		g.Expect(errors.Is(err, lookupErr)).To(BeTrue())
		g.Expect(events).To(Equal([]string{"lookup"}))
	})

	t.Run("request validation precedes create authorization", func(t *testing.T) {
		g := NewWithT(t)
		events := []string{}
		resManager := &recordingResourceManager{
			events: &events,
			getErr: store.ErrorResourceNotFound(core_mesh.MeshType, "mesh-1", core_model.NoMesh),
		}
		handler := newContractCrudHandler(resManager, &recordingResourceAccess{events: &events, createErr: accessErr})

		_, err := handler.createOrUpdateResource(newCrudRequest(http.MethodPut, "/meshes/mesh-1", "mesh-1", `{"type":"Mesh","name":"other"}`))

		expectTitledError(g, err, "Could not process a resource")
		g.Expect(events).To(Equal([]string{"lookup"}))
	})

	t.Run("denied create does not write", func(t *testing.T) {
		g := NewWithT(t)
		events := []string{}
		resManager := &recordingResourceManager{
			events: &events,
			getErr: store.ErrorResourceNotFound(core_mesh.MeshType, "mesh-1", core_model.NoMesh),
		}
		handler := newContractCrudHandler(resManager, &recordingResourceAccess{events: &events, createErr: accessErr})

		_, err := handler.createOrUpdateResource(newCrudRequest(http.MethodPut, "/meshes/mesh-1", "mesh-1", validBody))

		expectTitledError(g, err, "Access Denied")
		g.Expect(errors.Is(err, accessErr)).To(BeTrue())
		g.Expect(events).To(Equal([]string{"lookup", "authorize-create"}))
	})

	t.Run("request validation precedes update authorization", func(t *testing.T) {
		g := NewWithT(t)
		events := []string{}
		resManager := &recordingResourceManager{events: &events, existing: existingMesh()}
		handler := newContractCrudHandler(resManager, &recordingResourceAccess{events: &events, updateErr: accessErr})

		_, err := handler.createOrUpdateResource(newCrudRequest(http.MethodPut, "/meshes/mesh-1", "mesh-1", `{"type":"Mesh","name":"other"}`))

		expectTitledError(g, err, "Could not process a resource")
		g.Expect(events).To(Equal([]string{"lookup"}))
	})

	t.Run("denied update does not write", func(t *testing.T) {
		g := NewWithT(t)
		events := []string{}
		resManager := &recordingResourceManager{events: &events, existing: existingMesh()}
		handler := newContractCrudHandler(resManager, &recordingResourceAccess{events: &events, updateErr: accessErr})

		_, err := handler.createOrUpdateResource(newCrudRequest(http.MethodPut, "/meshes/mesh-1", "mesh-1", validBody))

		expectTitledError(g, err, "Access Denied")
		g.Expect(errors.Is(err, accessErr)).To(BeTrue())
		g.Expect(events).To(Equal([]string{"lookup", "authorize-update"}))
	})
}

func TestDeleteResourcePrecedence(t *testing.T) {
	lookupErr := errors.New("lookup failed")
	accessErr := errors.New("access denied")

	t.Run("parent Mesh lookup precedes resource lookup", func(t *testing.T) {
		g := NewWithT(t)
		events := []string{}
		resManager := &recordingResourceManager{events: &events, getErr: lookupErr}
		handler := newContractCrudHandler(resManager, &recordingResourceAccess{events: &events, deleteErr: accessErr})
		handler.descriptor = core_mesh.DataplaneResourceTypeDescriptor
		request := newCrudRequest(http.MethodDelete, "/meshes/missing/dataplanes/dp-1", "dp-1", "")
		request.PathParameters()["mesh"] = "missing"

		_, err := handler.deleteResource(request)

		expectTitledError(g, err, "Failed to retrieve Mesh")
		g.Expect(errors.Is(err, lookupErr)).To(BeTrue())
		g.Expect(events).To(Equal([]string{"lookup"}))
	})

	t.Run("lookup failure precedes authorization", func(t *testing.T) {
		g := NewWithT(t)
		events := []string{}
		resManager := &recordingResourceManager{events: &events, getErr: lookupErr}
		handler := newContractCrudHandler(resManager, &recordingResourceAccess{events: &events, deleteErr: accessErr})

		_, err := handler.deleteResource(newCrudRequest(http.MethodDelete, "/meshes/mesh-1", "mesh-1", ""))

		expectTitledError(g, err, "Could not delete a resource")
		g.Expect(errors.Is(err, lookupErr)).To(BeTrue())
		g.Expect(events).To(Equal([]string{"lookup"}))
	})

	t.Run("denied delete does not write", func(t *testing.T) {
		g := NewWithT(t)
		events := []string{}
		resManager := &recordingResourceManager{events: &events, existing: existingMesh()}
		handler := newContractCrudHandler(resManager, &recordingResourceAccess{events: &events, deleteErr: accessErr})

		_, err := handler.deleteResource(newCrudRequest(http.MethodDelete, "/meshes/mesh-1", "mesh-1", ""))

		expectTitledError(g, err, "Access Denied")
		g.Expect(errors.Is(err, accessErr)).To(BeTrue())
		g.Expect(events).To(Equal([]string{"lookup", "authorize-delete"}))
	})

	t.Run("origin validation precedes authorization", func(t *testing.T) {
		g := NewWithT(t)
		events := []string{}
		mesh := existingMesh()
		mesh.SetMeta(&test_model.ResourceMeta{
			Name: "mesh-1",
			Mesh: core_model.NoMesh,
			Labels: map[string]string{
				mesh_proto.ResourceOriginLabel: string(mesh_proto.ZoneResourceOrigin),
			},
		})
		resManager := &recordingResourceManager{events: &events, existing: mesh}
		handler := newContractCrudHandler(resManager, &recordingResourceAccess{events: &events, deleteErr: accessErr})
		handler.mode = config_core.Global
		handler.disableOriginLabelValidation = false

		_, err := handler.deleteResource(newCrudRequest(http.MethodDelete, "/meshes/mesh-1", "mesh-1", ""))

		expectTitledError(g, err, "Could not delete a resource")
		g.Expect(events).To(Equal([]string{"lookup"}))
	})
}

func TestReadResourcePrecedence(t *testing.T) {
	t.Run("GET parent Mesh lookup precedes authorization", func(t *testing.T) {
		g := NewWithT(t)
		events := []string{}
		lookupErr := errors.New("lookup failed")
		accessErr := errors.New("access denied")
		resManager := &recordingResourceManager{events: &events, getErr: lookupErr}
		handler := newContractCrudHandler(resManager, &recordingResourceAccess{events: &events, getErr: accessErr})
		handler.descriptor = core_mesh.DataplaneResourceTypeDescriptor
		request := newCrudRequest(http.MethodGet, "/meshes/missing/dataplanes/dp-1", "dp-1", "")
		request.PathParameters()["mesh"] = "missing"

		_, err := handler.findResource(false)(request)

		expectTitledError(g, err, "Failed to retrieve Mesh")
		g.Expect(errors.Is(err, lookupErr)).To(BeTrue())
		g.Expect(events).To(Equal([]string{"lookup"}))
	})

	t.Run("GET authorization precedes lookup", func(t *testing.T) {
		g := NewWithT(t)
		events := []string{}
		lookupErr := errors.New("lookup failed")
		accessErr := errors.New("access denied")
		resManager := &recordingResourceManager{events: &events, getErr: lookupErr}
		handler := newContractCrudHandler(resManager, &recordingResourceAccess{events: &events, getErr: accessErr})

		_, err := handler.findResource(false)(newCrudRequest(http.MethodGet, "/meshes/mesh-1", "mesh-1", ""))

		expectTitledError(g, err, "Access Denied")
		g.Expect(errors.Is(err, accessErr)).To(BeTrue())
		g.Expect(events).To(Equal([]string{"authorize-get"}))
	})

	t.Run("LIST query validation precedes authorization", func(t *testing.T) {
		g := NewWithT(t)
		events := []string{}
		accessErr := errors.New("access denied")
		resManager := &recordingResourceManager{events: &events}
		handler := newContractCrudHandler(resManager, &recordingResourceAccess{events: &events, listErr: accessErr})

		_, err := handler.listResources(false)(newCrudRequest(http.MethodGet, "/meshes?size=invalid", "", ""))

		expectTitledError(g, err, "Could not retrieve resources")
		g.Expect(events).To(BeEmpty())
	})

	t.Run("LIST authorization precedes storage", func(t *testing.T) {
		g := NewWithT(t)
		events := []string{}
		accessErr := errors.New("access denied")
		resManager := &recordingResourceManager{events: &events}
		handler := newContractCrudHandler(resManager, &recordingResourceAccess{events: &events, listErr: accessErr})

		_, err := handler.listResources(false)(newCrudRequest(http.MethodGet, "/meshes", "", ""))

		expectTitledError(g, err, "Access Denied")
		g.Expect(errors.Is(err, accessErr)).To(BeTrue())
		g.Expect(events).To(Equal([]string{"filter", "authorize-list"}))
	})

	t.Run("LIST parent Mesh lookup precedes authorization", func(t *testing.T) {
		g := NewWithT(t)
		events := []string{}
		lookupErr := errors.New("lookup failed")
		accessErr := errors.New("access denied")
		resManager := &recordingResourceManager{events: &events, getErr: lookupErr}
		handler := newContractCrudHandler(resManager, &recordingResourceAccess{events: &events, listErr: accessErr})
		handler.descriptor = core_mesh.DataplaneResourceTypeDescriptor
		request := newCrudRequest(http.MethodGet, "/meshes/missing/dataplanes", "", "")
		request.PathParameters()["mesh"] = "missing"

		_, err := handler.listResources(false)(request)

		expectTitledError(g, err, "Failed to retrieve Mesh")
		g.Expect(errors.Is(err, lookupErr)).To(BeTrue())
		g.Expect(events).To(Equal([]string{"filter", "lookup"}))
	})
}

type recordingResourceManager struct {
	events   *[]string
	getErr   error
	existing core_model.Resource
}

func (r *recordingResourceManager) Get(_ context.Context, resource core_model.Resource, _ ...store.GetOptionsFunc) error {
	*r.events = append(*r.events, "lookup")
	if r.getErr != nil {
		return r.getErr
	}
	if r.existing != nil {
		resource.SetMeta(r.existing.GetMeta())
		return resource.SetSpec(r.existing.GetSpec())
	}
	return nil
}

func (r *recordingResourceManager) List(context.Context, core_model.ResourceList, ...store.ListOptionsFunc) error {
	*r.events = append(*r.events, "list")
	return nil
}

func (r *recordingResourceManager) Create(context.Context, core_model.Resource, ...store.CreateOptionsFunc) error {
	*r.events = append(*r.events, "create")
	return nil
}

func (r *recordingResourceManager) Update(context.Context, core_model.Resource, ...store.UpdateOptionsFunc) error {
	*r.events = append(*r.events, "update")
	return nil
}

func (r *recordingResourceManager) Delete(context.Context, core_model.Resource, ...store.DeleteOptionsFunc) error {
	*r.events = append(*r.events, "delete")
	return nil
}

func (r *recordingResourceManager) DeleteAll(context.Context, core_model.ResourceList, ...store.DeleteAllOptionsFunc) error {
	*r.events = append(*r.events, "delete-all")
	return nil
}

type recordingResourceAccess struct {
	events    *[]string
	createErr error
	updateErr error
	deleteErr error
	listErr   error
	getErr    error
}

func (r *recordingResourceAccess) ValidateCreate(context.Context, core_model.ResourceKey, core_model.ResourceSpec, core_model.ResourceTypeDescriptor, core_user.User) error {
	*r.events = append(*r.events, "authorize-create")
	return r.createErr
}

func (r *recordingResourceAccess) ValidateUpdate(context.Context, core_model.ResourceKey, core_model.ResourceSpec, core_model.ResourceSpec, core_model.ResourceTypeDescriptor, core_user.User) error {
	*r.events = append(*r.events, "authorize-update")
	return r.updateErr
}

func (r *recordingResourceAccess) ValidateDelete(context.Context, core_model.ResourceKey, core_model.ResourceSpec, core_model.ResourceTypeDescriptor, core_user.User) error {
	*r.events = append(*r.events, "authorize-delete")
	return r.deleteErr
}

func (r *recordingResourceAccess) ValidateList(context.Context, string, core_model.ResourceTypeDescriptor, core_user.User) error {
	*r.events = append(*r.events, "authorize-list")
	return r.listErr
}

func (r *recordingResourceAccess) ValidateGet(context.Context, core_model.ResourceKey, core_model.ResourceTypeDescriptor, core_user.User) error {
	*r.events = append(*r.events, "authorize-get")
	return r.getErr
}

func newContractCrudHandler(resManager *recordingResourceManager, resourceAccess *recordingResourceAccess) *resourceCrudHandler {
	return &resourceCrudHandler{
		mode:                         config_core.Zone,
		resManager:                   resManager,
		descriptor:                   core_mesh.MeshResourceTypeDescriptor,
		resourceAccess:               resourceAccess,
		disableOriginLabelValidation: true,
		filter: func(*restful.Request) (store.ListFilterFunc, error) {
			*resManager.events = append(*resManager.events, "filter")
			return nil, nil
		},
	}
}

func newCrudRequest(method, target, name, body string) *restful.Request {
	request := restful.NewRequest(httptest.NewRequestWithContext(context.Background(), method, target, strings.NewReader(body)))
	request.PathParameters()["name"] = name
	return request
}

func existingMesh() core_model.Resource {
	mesh := core_mesh.NewMeshResource()
	mesh.SetMeta(&test_model.ResourceMeta{Name: "mesh-1", Mesh: core_model.NoMesh})
	return mesh
}

func expectTitledError(g *WithT, err error, title string) {
	g.Expect(err).To(HaveOccurred())
	var titled *titledError
	g.Expect(errors.As(err, &titled)).To(BeTrue())
	g.Expect(titled.title).To(Equal(title))
}
