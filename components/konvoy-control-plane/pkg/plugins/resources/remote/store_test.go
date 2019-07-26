package remote_test

import (
	"bufio"
	"context"
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
	Describe("List()", func() {
		It("should successfully list known resources", func() {
			// setup
			client := &http.Client{
				Transport: RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
					Expect(req.URL.Path).To(Equal("/meshes/pilot/trafficroutes"))

					file, err := os.Open(filepath.Join("testdata", "list.json"))
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

			// given
			store := remote.NewStore(client, apis)

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
