package unversioned_test

import (
	"encoding/json"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	meshtrust_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshtrust/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/core/resources/model/rest"
	"github.com/kumahq/kuma/v3/pkg/core/resources/model/rest/unversioned"
	"github.com/kumahq/kuma/v3/pkg/core/resources/model/rest/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/test/matchers"
)

var _ = Describe("Rest Resource", func() {
	var t1, t2 time.Time
	BeforeEach(func() {
		t1, _ = time.Parse(time.RFC3339, "2018-07-17T16:05:36.995+00:00")
		t2, _ = time.Parse(time.RFC3339, "2019-07-17T16:05:36.995+00:00")
	})
	Describe("Resource", func() {
		Describe("MarshalJSON", func() {
			It("should marshal JSON with proper field order", func() {
				// given
				res := &unversioned.Resource{
					Meta: v1alpha1.ResourceMeta{
						Type:             "Mesh",
						Name:             "one",
						CreationTime:     t1,
						ModificationTime: t2,
					},
					Spec: &mesh_proto.Mesh{},
				}

				// when
				bytes, err := json.Marshal(res)

				// then
				Expect(err).ToNot(HaveOccurred())

				// and
				expected := `
{
  "type": "Mesh",
  "name": "one",
  "creationTime": "2018-07-17T16:05:36.995Z",
  "modificationTime": "2019-07-17T16:05:36.995Z",
  "kri": "kri_m____one_"
}`
				Expect(bytes).To(MatchJSON(expected))
			})

			It("should marshal JSON with proper field order and empty spec", func() {
				// given
				res := &unversioned.Resource{
					Meta: v1alpha1.ResourceMeta{
						Type:             "Mesh",
						Name:             "one",
						CreationTime:     t1,
						ModificationTime: t2,
					},
				}

				// when
				bytes, err := json.Marshal(res)

				// then
				Expect(err).ToNot(HaveOccurred())

				// and
				expected := `{"type":"Mesh","name":"one","creationTime":"2018-07-17T16:05:36.995Z","modificationTime":"2019-07-17T16:05:36.995Z","kri":"kri_m____one_"}`
				Expect(string(bytes)).To(Equal(expected))
			})
		})
	})

	Describe("ResourceListReceiver", func() {
		Describe("UnmarshalJSON", func() {
			It("it should be possible to unmarshal JSON response from Kuma API Server", func() {
				// given
				content := `
			{
				"items": [
				 {
					"type": "Mesh",
					"name": "one"
				 },
				 {
					"type": "Mesh",
					"name": "two"
				 }
				],
				"next": "http://localhost:5681/meshes?offset=1"
			}`

				// when
				rsr := &rest.ResourceListReceiver{
					NewResource: func() model.Resource {
						return mesh.NewMeshResource()
					},
				}
				err := json.Unmarshal([]byte(content), rsr)

				// then
				Expect(err).ToNot(HaveOccurred())

				// when
				rs := rsr.ResourceList
				// then
				Expect(rs.Items).To(HaveLen(2))
				Expect(rs.Items[0].GetMeta()).To(Equal(v1alpha1.ResourceMeta{
					Type: "Mesh",
					Name: "one",
				}))
				Expect(rs.Items[0].GetSpec()).To(matchers.MatchProto(&mesh_proto.Mesh{}))
				Expect(rs.Items[1].GetMeta()).To(Equal(v1alpha1.ResourceMeta{
					Type: "Mesh",
					Name: "two",
				}))
				Expect(rs.Items[1].GetSpec()).To(matchers.MatchProto(&mesh_proto.Mesh{}))
				Expect(*rs.Next).To(Equal("http://localhost:5681/meshes?offset=1"))
			})
		})
	})
})

var _ = Describe("Rest Resource with a spec that is not a protobuf message", func() {
	It("should round trip through the inlined representation, so converting a core resource away from protobuf does not move its endpoint to the nested spec form", func() {
		// given
		res := &unversioned.Resource{
			Meta: v1alpha1.ResourceMeta{Type: "MeshTrust", Name: "one", Mesh: "default"},
			Spec: &meshtrust_api.MeshTrust{
				TrustDomain: "default.mesh.local",
				CABundles: []meshtrust_api.CABundle{{
					Type: meshtrust_api.PemCABundleType,
					PEM:  &meshtrust_api.PEM{Value: "cert"},
				}},
			},
		}

		// when
		bytes, err := json.Marshal(res)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(string(bytes)).To(ContainSubstring(`"trustDomain":"default.mesh.local"`))
		Expect(string(bytes)).ToNot(ContainSubstring(`"spec"`))

		// when
		back := &unversioned.Resource{Spec: &meshtrust_api.MeshTrust{}}
		err = json.Unmarshal(bytes, back)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(back.Meta).To(Equal(res.Meta))
		Expect(back.Spec).To(Equal(res.Spec))
	})
})
