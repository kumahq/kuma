package remote_test

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_rest "github.com/kumahq/kuma/pkg/core/resources/model/rest"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	errors_types "github.com/kumahq/kuma/pkg/core/rest/errors/types"
	"github.com/kumahq/kuma/pkg/plugins/resources/remote"
	sample_api "github.com/kumahq/kuma/pkg/test/apis/sample/v1alpha1"
	sample_core "github.com/kumahq/kuma/pkg/test/resources/apis/sample"
	"github.com/kumahq/kuma/pkg/test/resources/model"
)

var _ = Describe("RemoteStore", func() {

	creationTime, _ := time.Parse(time.RFC3339, "2018-07-17T16:05:36.995Z")
	modificationTime, _ := time.Parse(time.RFC3339, "2019-07-17T16:05:36.995Z")
	type RequestAssertion = func(req *http.Request)
	setupStore := func(file string, assertion RequestAssertion) core_store.ResourceStore {
		client := &http.Client{
			Transport: RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
				assertion(req)

				file, err := os.Open(filepath.Join("testdata", file))
				if err != nil {
					return nil, err
				}
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bufio.NewReader(file)),
				}, nil
			}),
		}
		apis := &core_rest.ApiDescriptor{
			Resources: map[core_model.ResourceType]core_rest.ResourceApi{
				sample_core.TrafficRouteType: core_rest.NewResourceApi(core_model.ScopeMesh, "traffic-routes"),
				mesh.MeshType:                core_rest.NewResourceApi(core_model.ScopeGlobal, "meshes"),
			},
		}
		return remote.NewStore(client, apis)
	}

	setupErrorStore := func(code int, errorMsg string) core_store.ResourceStore {
		client := &http.Client{
			Transport: RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: code,
					Body:       io.NopCloser(strings.NewReader(errorMsg)),
				}, nil
			}),
		}
		apis := &core_rest.ApiDescriptor{
			Resources: map[core_model.ResourceType]core_rest.ResourceApi{
				sample_core.TrafficRouteType: core_rest.NewResourceApi(core_model.ScopeMesh, "traffic-routes"),
				mesh.MeshType:                core_rest.NewResourceApi(core_model.ScopeMesh, "meshes"),
			},
		}
		return remote.NewStore(client, apis)
	}
	Describe("Get()", func() {
		It("should get resource", func() {
			// setup
			name := "res-1"
			store := setupStore("get.json", func(req *http.Request) {
				Expect(req.URL.Path).To(Equal(fmt.Sprintf("/meshes/default/traffic-routes/%s", name)))
			})

			// when
			resource := sample_core.NewTrafficRouteResource()
			err := store.Get(context.Background(), resource, core_store.GetByKey(name, "default"))

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(resource.Spec.Path).To(Equal("/example"))

			Expect(resource.GetMeta().GetName()).To(Equal("res-1"))
			Expect(resource.GetMeta().GetMesh()).To(Equal("default"))
			Expect(resource.GetMeta().GetCreationTime()).Should(Equal(creationTime))
			Expect(resource.GetMeta().GetModificationTime()).Should(Equal(modificationTime))
		})

		It("should get mesh resource", func() {
			meshName := "someMesh"
			store := setupStore("get-mesh.json", func(req *http.Request) {
				Expect(req.URL.Path).To(Equal(fmt.Sprintf("/meshes/%s", meshName)))
			})

			// when
			resource := mesh.NewMeshResource()
			err := store.Get(context.Background(), resource, core_store.GetByKey(meshName, core_model.NoMesh))

			// then
			Expect(err).ToNot(HaveOccurred())

			Expect(resource.GetMeta().GetName()).To(Equal(meshName))
			Expect(resource.GetMeta().GetMesh()).To(Equal(core_model.NoMesh))
			Expect(resource.GetMeta().GetCreationTime()).Should(Equal(creationTime))
			Expect(resource.GetMeta().GetModificationTime()).Should(Equal(modificationTime))
		})

		It("should parse kuma api server error", func() {
			json := `
			{
				"title": "Could not get resource",
				"details": "Internal Server Error"
			}
		`
			store := setupErrorStore(400, json)

			// when
			resource := mesh.NewMeshResource()
			err := store.Get(context.Background(), resource, core_store.GetByKey("test", "test"))

			// then
			Expect(err).To(HaveOccurred())

			Expect(err).To(Equal(&errors_types.Error{
				Title:   "Could not get resource",
				Details: "Internal Server Error",
			}))
		})

		It("should map 404 error to ResourceNotFound", func() {
			// given
			json := `
			{
				"title": "Could not get a resource",
				"details": "Not found"
			}`
			store := setupErrorStore(404, json)

			// when
			resource := mesh.NewMeshResource()
			err := store.Get(context.Background(), resource, core_store.GetByKey("test", "test"))

			// then
			Expect(core_store.IsResourceNotFound(err)).To(BeTrue())
		})
	})

	Describe("Create()", func() {
		It("should send proper json", func() {
			// setup
			name := "res-1"
			store := setupStore("create_update.json", func(req *http.Request) {
				Expect(req.URL.Path).To(Equal(fmt.Sprintf("/meshes/default/traffic-routes/%s", name)))
				bytes, err := io.ReadAll(req.Body)
				Expect(err).ToNot(HaveOccurred())
				Expect(bytes).To(MatchJSON(`{"mesh":"default","name":"res-1","path":"/some-path","type":"SampleTrafficRoute","creationTime": "0001-01-01T00:00:00Z","modificationTime": "0001-01-01T00:00:00Z"}`))
			})

			// when
			resource := sample_core.TrafficRouteResource{
				Spec: &sample_api.TrafficRoute{
					Path: "/some-path",
				},
			}
			err := store.Create(context.Background(), &resource, core_store.CreateByKey(name, "default"))

			// then
			Expect(err).ToNot(HaveOccurred())
		})

		It("should send proper mesh json", func() {
			// setup
			meshName := "someMesh"
			store := setupStore("create_update.json", func(req *http.Request) {
				Expect(req.URL.Path).To(Equal(fmt.Sprintf("/meshes/%s", meshName)))
				bytes, err := io.ReadAll(req.Body)
				Expect(err).ToNot(HaveOccurred())
				Expect(bytes).To(MatchJSON(`{"name":"someMesh","type":"Mesh","creationTime": "0001-01-01T00:00:00Z","modificationTime": "0001-01-01T00:00:00Z"}`))
			})

			// when
			resource := mesh.MeshResource{
				Spec: &v1alpha1.Mesh{},
			}
			err := store.Create(context.Background(), &resource, core_store.CreateByKey(meshName, core_model.NoMesh))

			// then
			Expect(err).ToNot(HaveOccurred())
		})

		It("should parse kuma api server error", func() {
			json := `
			{
				"title": "Could not process resource",
				"details": "Resource is not valid",
				"causes": [
					{
						"field": "mtls",
						"message": "cannot be empty"
					}
				]
			}
		`
			store := setupErrorStore(400, json)

			// when
			err := store.Create(context.Background(), mesh.NewMeshResource(), core_store.CreateByKey("test", core_model.NoMesh))

			// then
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(&errors_types.Error{
				Title:   "Could not process resource",
				Details: "Resource is not valid",
				Causes: []errors_types.Cause{
					{
						Field:   "mtls",
						Message: "cannot be empty",
					},
				},
			}))
		})
	})

	Describe("Update()", func() {
		It("should send proper json", func() {
			// setup
			name := "res-1"
			store := setupStore("create_update.json", func(req *http.Request) {
				Expect(req.URL.Path).To(Equal(fmt.Sprintf("/meshes/default/traffic-routes/%s", name)))
				bytes, err := io.ReadAll(req.Body)
				Expect(err).ToNot(HaveOccurred())
				Expect(bytes).To(MatchJSON(`{"mesh":"default","name":"res-1","path":"/some-path","type":"SampleTrafficRoute","creationTime": "0001-01-01T00:00:00Z","modificationTime": "0001-01-01T00:00:00Z"}`))
			})

			// when
			resource := sample_core.TrafficRouteResource{
				Spec: &sample_api.TrafficRoute{
					Path: "/some-path",
				},
				Meta: &model.ResourceMeta{
					Mesh: "default",
					Name: name,
				},
			}
			err := store.Update(context.Background(), &resource)

			// then
			Expect(err).ToNot(HaveOccurred())
		})

		It("should send proper mesh json", func() {
			// setup
			meshName := "someMesh"
			store := setupStore("create_update.json", func(req *http.Request) {
				Expect(req.URL.Path).To(Equal(fmt.Sprintf("/meshes/%s", meshName)))
				bytes, err := io.ReadAll(req.Body)
				Expect(err).ToNot(HaveOccurred())
				Expect(bytes).To(MatchJSON(`{"name":"someMesh","mtls":{"enabledBackend":"builtin","backends":[{"name":"builtin","type":"builtin"}]},"name":"someMesh","type":"Mesh","creationTime": "0001-01-01T00:00:00Z","modificationTime": "0001-01-01T00:00:00Z"}`))
			})

			// when
			resource := mesh.MeshResource{
				Spec: &v1alpha1.Mesh{
					Mtls: &v1alpha1.Mesh_Mtls{
						EnabledBackend: "builtin",
						Backends: []*v1alpha1.CertificateAuthorityBackend{
							{
								Name: "builtin",
								Type: "builtin",
							},
						},
					},
				},
				Meta: &model.ResourceMeta{
					Name: meshName,
				},
			}
			err := store.Update(context.Background(), &resource)

			// then
			Expect(err).ToNot(HaveOccurred())
		})

		It("should return error from the api server", func() {
			// given
			store := setupErrorStore(400, "some error from the server")

			// when
			resource := mesh.MeshResource{
				Spec: &v1alpha1.Mesh{},
				Meta: &model.ResourceMeta{
					Name: "default",
				},
			}
			err := store.Create(context.Background(), &resource)

			// then
			Expect(err).To(MatchError("(400): some error from the server"))
		})

		It("should parse kuma api server error", func() {
			json := `
			{
				"title": "Could not process resource",
				"details": "Resource is not valid",
				"causes": [
					{
						"field": "mtls",
						"message": "cannot be empty"
					},
					{
						"field": "mesh",
						"message": "cannot be empty"
					}
				]
			}
		`
			store := setupErrorStore(400, json)

			// when
			resource := mesh.MeshResource{
				Meta: &model.ResourceMeta{
					Name: "test",
				},
				Spec: &v1alpha1.Mesh{},
			}
			err := store.Update(context.Background(), &resource)

			// then
			Expect(err).To(HaveOccurred())

			Expect(err).To(Equal(&errors_types.Error{
				Title:   "Could not process resource",
				Details: "Resource is not valid",
				Causes: []errors_types.Cause{
					{
						Field:   "mtls",
						Message: "cannot be empty",
					},
					{
						Field:   "mesh",
						Message: "cannot be empty",
					},
				},
			}))
		})
	})

	Describe("List()", func() {
		It("should successfully list known resources", func() {
			// given
			store := setupStore("list.json", func(req *http.Request) {
				Expect(req.URL.Path).To(Equal("/meshes/demo/traffic-routes"))
			})

			// when
			rs := sample_core.TrafficRouteResourceList{}
			err := store.List(context.Background(), &rs, core_store.ListByMesh("demo"))

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(rs.Items).To(HaveLen(2))
			// and
			Expect(rs.Items[0].Meta.GetName()).To(Equal("one"))
			Expect(rs.Items[0].Meta.GetMesh()).To(Equal("default"))
			Expect(rs.Items[0].Meta.GetVersion()).To(Equal(""))
			Expect(rs.Items[0].Spec.Path).To(Equal("/example"))
			Expect(rs.Items[0].Meta.GetCreationTime()).Should(Equal(creationTime))
			Expect(rs.Items[0].Meta.GetModificationTime()).Should(Equal(modificationTime))
			// and
			Expect(rs.Items[1].Meta.GetName()).To(Equal("two"))
			Expect(rs.Items[1].Meta.GetMesh()).To(Equal("demo"))
			Expect(rs.Items[1].Meta.GetVersion()).To(Equal(""))
			Expect(rs.Items[1].Spec.Path).To(Equal("/another"))
			Expect(rs.Items[1].Meta.GetCreationTime()).Should(Equal(creationTime))
			Expect(rs.Items[1].Meta.GetModificationTime()).Should(Equal(modificationTime))
		})

		It("should list known resources using pagination", func() {
			// given
			store := setupStore("list-pagination.json", func(req *http.Request) {
				Expect(req.URL.Path).To(Equal("/meshes/demo/traffic-routes"))
				Expect(req.URL.Query().Get("size")).To(Equal("1"))
				Expect(req.URL.Query().Get("offset")).To(Equal("2"))
			})

			// when
			rs := sample_core.TrafficRouteResourceList{}
			err := store.List(context.Background(), &rs, core_store.ListByMesh("demo"), core_store.ListByPage(1, "2"))

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(rs.Items).To(HaveLen(1))
			// and
			Expect(rs.Items[0].Meta.GetName()).To(Equal("one"))
			Expect(rs.Items[0].Meta.GetMesh()).To(Equal("default"))
			Expect(rs.Items[0].Meta.GetVersion()).To(Equal(""))
			Expect(rs.Items[0].Spec.Path).To(Equal("/example"))
			Expect(rs.Items[0].Meta.GetCreationTime()).Should(Equal(creationTime))
			Expect(rs.Items[0].Meta.GetModificationTime()).Should(Equal(modificationTime))
		})

		It("should list meshes", func() {
			// given
			store := setupStore("list-meshes.json", func(req *http.Request) {
				Expect(req.URL.Path).To(Equal("/meshes"))
			})

			// when
			meshes := mesh.MeshResourceList{}
			err := store.List(context.Background(), &meshes)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(meshes.Items).To(HaveLen(2))

			Expect(meshes.Items[0].Meta.GetName()).To(Equal("mesh-1"))
			Expect(meshes.Items[0].Meta.GetMesh()).To(Equal(core_model.NoMesh))
			Expect(meshes.Items[0].Meta.GetCreationTime()).Should(Equal(creationTime))
			Expect(meshes.Items[0].Meta.GetModificationTime()).Should(Equal(modificationTime))

			Expect(meshes.Items[1].Meta.GetName()).To(Equal("mesh-2"))
			Expect(meshes.Items[1].Meta.GetMesh()).To(Equal(core_model.NoMesh))
			Expect(meshes.Items[1].Meta.GetCreationTime()).Should(Equal(creationTime))
			Expect(meshes.Items[1].Meta.GetModificationTime()).Should(Equal(modificationTime))
		})

		It("should return error from the api server", func() {
			// given
			store := setupErrorStore(400, "some error from the server")

			// when
			meshes := mesh.MeshResourceList{}
			err := store.List(context.Background(), &meshes)

			// then
			Expect(err).To(MatchError("(400): some error from the server"))
		})

		It("should parse kuma api server error", func() {
			json := `
			{
				"title": "Could not list resource",
				"details": "Internal Server Error"
			}
		`
			store := setupErrorStore(400, json)

			// when
			meshes := mesh.MeshResourceList{}
			err := store.List(context.Background(), &meshes)

			// then
			Expect(err).To(HaveOccurred())

			Expect(err).To(Equal(&errors_types.Error{
				Title:   "Could not list resource",
				Details: "Internal Server Error",
			}))
		})
	})

	Describe("Delete()", func() {
		It("should delete the resource", func() {
			// given
			name := "tr-1"
			meshName := "mesh-1"
			store := setupStore("delete.json", func(req *http.Request) {
				Expect(req.URL.Path).To(Equal(fmt.Sprintf("/meshes/%s/traffic-routes/%s", meshName, name)))
			})

			// when
			resource := sample_core.NewTrafficRouteResource()
			err := store.Delete(context.Background(), resource, core_store.DeleteByKey(name, meshName))

			// then
			Expect(err).ToNot(HaveOccurred())
		})

		It("should delete mesh resource", func() {
			// given
			meshName := "mesh-1"
			store := setupStore("delete.json", func(req *http.Request) {
				Expect(req.URL.Path).To(Equal(fmt.Sprintf("/meshes/%s", meshName)))
			})

			// when
			resource := mesh.NewMeshResource()
			err := store.Delete(context.Background(), resource, core_store.DeleteByKey(meshName, meshName))

			// then
			Expect(err).ToNot(HaveOccurred())
		})

		It("should return error from the api server", func() {
			// given
			store := setupErrorStore(400, "some error from the server")

			// when
			resource := sample_core.NewTrafficRouteResource()
			err := store.Delete(context.Background(), resource, core_store.DeleteByKey("tr-1", "mesh-1"))

			// then
			Expect(err).To(MatchError("(400): some error from the server"))
		})

		It("should map 404 error to ResourceNotFound", func() {
			// given
			json := `
			{
				"title": "Could not get a resource",
				"details": "Not found"
			}`
			store := setupErrorStore(404, json)

			// when
			resource := sample_core.NewTrafficRouteResource()
			err := store.Delete(context.Background(), resource, core_store.DeleteByKey("tr-1", "mesh-1"))

			// then
			Expect(core_store.IsResourceNotFound(err)).To(BeTrue())
		})

		It("should parse kuma api server error", func() {
			json := `
			{
				"title": "Could not delete resource",
				"details": "Internal Server Error"
			}`
			store := setupErrorStore(400, json)

			// when
			resource := sample_core.NewTrafficRouteResource()
			err := store.Delete(context.Background(), resource, core_store.DeleteByKey("tr-1", "mesh-1"))

			// then
			Expect(err).To(HaveOccurred())

			Expect(err).To(Equal(&errors_types.Error{
				Title:   "Could not delete resource",
				Details: "Internal Server Error",
			}))
		})
	})

})

type RoundTripperFunc func(*http.Request) (*http.Response, error)

func (f RoundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
