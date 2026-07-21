package unversioned_test

import (
	"encoding/json"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
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
					Spec: &mesh_proto.Mesh{
						Mtls: &mesh_proto.Mesh_Mtls{
							EnabledBackend: "ca-1",
						},
					},
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
  "mtls": {
    "enabledBackend": "ca-1"
  }
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
				expected := `{"type":"Mesh","name":"one","creationTime":"2018-07-17T16:05:36.995Z","modificationTime":"2019-07-17T16:05:36.995Z"}`
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
					"name": "one",
					"mtls": {
					  "enabledBackend": "ca-1"
					}
				 },
				 {
					"type": "Mesh",
					"name": "two",
					"mtls": {
					  "enabledBackend": "ca-2"
					}
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
				Expect(rs.Items[0].GetSpec()).To(matchers.MatchProto(&mesh_proto.Mesh{
					Mtls: &mesh_proto.Mesh_Mtls{
						EnabledBackend: "ca-1",
					},
				}))
				Expect(rs.Items[1].GetMeta()).To(Equal(v1alpha1.ResourceMeta{
					Type: "Mesh",
					Name: "two",
				}))
				Expect(rs.Items[1].GetSpec()).To(matchers.MatchProto(&mesh_proto.Mesh{
					Mtls: &mesh_proto.Mesh_Mtls{
						EnabledBackend: "ca-2",
					},
				}))
				Expect(*rs.Next).To(Equal("http://localhost:5681/meshes?offset=1"))
			})
		})
	})
})
