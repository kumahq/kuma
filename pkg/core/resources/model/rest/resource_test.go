package rest_test

import (
	"encoding/json"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/model/rest"
	sample_proto "github.com/kumahq/kuma/pkg/test/apis/sample/v1alpha1"
	"github.com/kumahq/kuma/pkg/test/matchers"
	sample_core "github.com/kumahq/kuma/pkg/test/resources/apis/sample"
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
				res := &rest.Resource{
					Meta: rest.ResourceMeta{
						Type:             "TrafficRoute",
						Mesh:             "default",
						Name:             "one",
						CreationTime:     t1,
						ModificationTime: t2,
					},
					Spec: &sample_proto.TrafficRoute{
						Path: "/example",
					},
				}

				// when
				bytes, err := json.Marshal(res)

				// then
				Expect(err).ToNot(HaveOccurred())

				// and
				expected := `{"type":"TrafficRoute","mesh":"default","name":"one","creationTime":"2018-07-17T16:05:36.995Z","modificationTime":"2019-07-17T16:05:36.995Z","path":"/example"}`
				Expect(string(bytes)).To(Equal(expected))
			})

			It("should marshal JSON with proper field order and empty spec", func() {
				// given
				res := &rest.Resource{
					Meta: rest.ResourceMeta{
						Type:             "TrafficRoute",
						Mesh:             "default",
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
				expected := `{"type":"TrafficRoute","mesh":"default","name":"one","creationTime":"2018-07-17T16:05:36.995Z","modificationTime":"2019-07-17T16:05:36.995Z"}`
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
					"type": "TrafficRoute",
					"mesh": "default",
					"name": "one",
					"path": "/example"
				 },
				 {
					"type": "TrafficRoute",
					"mesh": "demo",
					"name": "two",
					"path": "/another"
				 }
				],
				"next": "http://localhost:5681/meshes/default/traffic-routes?offset=1"
			}`

				// when
				rsr := &rest.ResourceListReceiver{
					NewResource: func() model.Resource {
						return sample_core.NewTrafficRouteResource()
					},
				}
				err := json.Unmarshal([]byte(content), rsr)

				// then
				Expect(err).ToNot(HaveOccurred())

				// when
				rs := rsr.ResourceList
				// then
				Expect(rs.Items).To(HaveLen(2))
				Expect(rs.Items[0].Meta).To(Equal(rest.ResourceMeta{
					Type: "TrafficRoute",
					Mesh: "default",
					Name: "one",
				}))
				Expect(rs.Items[0].Spec).To(matchers.MatchProto(&sample_proto.TrafficRoute{
					Path: "/example",
				}))
				Expect(rs.Items[1].Meta).To(Equal(rest.ResourceMeta{
					Type: "TrafficRoute",
					Mesh: "demo",
					Name: "two",
				}))
				Expect(rs.Items[1].Spec).To(matchers.MatchProto(&sample_proto.TrafficRoute{
					Path: "/another",
				}))
				Expect(*rs.Next).To(Equal("http://localhost:5681/meshes/default/traffic-routes?offset=1"))
			})
		})
	})
})
