package remote_test

import (
	"bufio"
	"context"
	"fmt"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/test/apis/sample/v1alpha1"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/test/resources/model"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	core_model "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/model"
	core_rest "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/model/rest"
	core_store "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/store"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/plugins/resources/remote"

	sample_core "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/test/resources/apis/sample"
)

var _ = Describe("RemoteStore", func() {

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
					Body:       ioutil.NopCloser(bufio.NewReader(file)),
				}, nil
			}),
		}
		apis := &core_rest.ApiDescriptor{
			Resources: map[core_model.ResourceType]core_rest.ResourceApi{
				sample_core.TrafficRouteType: core_rest.ResourceApi{CollectionPath: "trafficroutes"},
			},
		}
		return remote.NewStore(client, apis)
	}

	Describe("Get()", func() {
		It("should get resource", func() {
			// setup
			name := "res-1"
			store := setupStore("get.json", func(req *http.Request) {
				Expect(req.URL.Path).To(Equal(fmt.Sprintf("/meshes/default/trafficroutes/%s", name)))
			})

			// when
			resource := sample_core.TrafficRouteResource{}
			err := store.Get(context.Background(), &resource, core_store.GetByKey("", name, "default"))

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(resource.Spec.Path).To(Equal("/example"))

			Expect(resource.GetMeta().GetName()).To(Equal("res-1"))
			Expect(resource.GetMeta().GetMesh()).To(Equal("default"))
			Expect(resource.GetMeta().GetNamespace()).To(Equal(""))
		})
	})

	Describe("Create()", func() {
		It("should send proper json", func() {
			// setup
			name := "res-1"
			store := setupStore("create_update.json", func(req *http.Request) {
				Expect(req.URL.Path).To(Equal(fmt.Sprintf("/meshes/default/trafficroutes/%s", name)))
				bytes, err := ioutil.ReadAll(req.Body)
				Expect(err).ToNot(HaveOccurred())
				Expect(string(bytes)).To(Equal(`{"mesh":"default","name":"res-1","path":"/some-path","type":"TrafficRoute"}`))
			})

			// when
			resource := sample_core.TrafficRouteResource{
				Spec: v1alpha1.TrafficRoute{
					Path: "/some-path",
				},
			}
			err := store.Create(context.Background(), &resource, core_store.CreateByKey("", name, "default"))

			// then
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe("Update()", func() {
		It("should send proper json", func() {
			// setup
			name := "res-1"
			store := setupStore("create_update.json", func(req *http.Request) {
				Expect(req.URL.Path).To(Equal(fmt.Sprintf("/meshes/default/trafficroutes/%s", name)))
				bytes, err := ioutil.ReadAll(req.Body)
				Expect(err).ToNot(HaveOccurred())
				Expect(string(bytes)).To(Equal(`{"mesh":"default","name":"res-1","path":"/some-path","type":"TrafficRoute"}`))
			})

			// when
			resource := sample_core.TrafficRouteResource{
				Spec: v1alpha1.TrafficRoute{
					Path: "/some-path",
				},
				Meta: &model.ResourceMeta{
					Mesh:      "default",
					Name:      name,
					Namespace: "",
				},
			}
			err := store.Update(context.Background(), &resource)

			// then
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe("List()", func() {
		It("should successfully list known resources", func() {
			// given
			store := setupStore("list.json", func(req *http.Request) {
				Expect(req.URL.Path).To(Equal(fmt.Sprintf("/meshes/pilot/trafficroutes")))
			})

			// when
			rs := sample_core.TrafficRouteResourceList{}
			err := store.List(context.Background(), &rs, core_store.ListByMesh("pilot"))

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(rs.Items).To(HaveLen(2))
			// and
			Expect(rs.Items[0].Meta.GetNamespace()).To(Equal(""))
			Expect(rs.Items[0].Meta.GetName()).To(Equal("one"))
			Expect(rs.Items[0].Meta.GetMesh()).To(Equal("default"))
			Expect(rs.Items[0].Meta.GetVersion()).To(Equal(""))
			Expect(rs.Items[0].Spec.Path).To(Equal("/example"))
			// and
			Expect(rs.Items[1].Meta.GetNamespace()).To(Equal(""))
			Expect(rs.Items[1].Meta.GetName()).To(Equal("two"))
			Expect(rs.Items[1].Meta.GetMesh()).To(Equal("pilot"))
			Expect(rs.Items[1].Meta.GetVersion()).To(Equal(""))
			Expect(rs.Items[1].Spec.Path).To(Equal("/another"))
		})
	})
})

type RoundTripperFunc func(*http.Request) (*http.Response, error)

func (f RoundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
