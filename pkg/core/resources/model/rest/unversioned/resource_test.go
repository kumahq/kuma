package unversioned_test

import (
	"encoding/json"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/model/rest"
	"github.com/kumahq/kuma/pkg/core/resources/model/rest/unversioned"
	"github.com/kumahq/kuma/pkg/core/resources/model/rest/v1alpha1"
	sample_proto "github.com/kumahq/kuma/pkg/test/apis/sample/v1alpha1"
	"github.com/kumahq/kuma/pkg/test/matchers"
	sample_core "github.com/kumahq/kuma/pkg/test/resources/apis/sample"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
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
				res := &unversioned.Resource{
					Meta: v1alpha1.ResourceMeta{
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
				Expect(rs.Items[0].GetMeta()).To(Equal(v1alpha1.ResourceMeta{
					Type: "TrafficRoute",
					Mesh: "default",
					Name: "one",
				}))
				Expect(rs.Items[0].GetSpec()).To(matchers.MatchProto(&sample_proto.TrafficRoute{
					Path: "/example",
				}))
				Expect(rs.Items[1].GetMeta()).To(Equal(v1alpha1.ResourceMeta{
					Type: "TrafficRoute",
					Mesh: "demo",
					Name: "two",
				}))
				Expect(rs.Items[1].GetSpec()).To(matchers.MatchProto(&sample_proto.TrafficRoute{
					Path: "/another",
				}))
				Expect(*rs.Next).To(Equal("http://localhost:5681/meshes/default/traffic-routes?offset=1"))
			})
			It("it should be possible to unmarshal JSON response with protobuf field types", func() {
				// given
				content := `
{
	"items": [
	 {
		"type": "CircuitBreaker",
		"mesh": "default",
		"name": "one",
		"sources": [
			{
				"match": {
					"kuma.io/service": "*"
			  }
			}
		],
		"destinations": [
			{
				"match": {
					"kuma.io/service": "*"
				}
			}
		],
		"conf": {
			"interval": "99s",
			"detectors": {
				"totalErrors": {
					"consecutive": 3
				}
			}
		}
	},
	{
		"type": "CircuitBreaker",
		"mesh": "default",
		"name": "two",
		"sources": [
			{
				"match": {
					"kuma.io/service": "*"
			  }
			}
		],
		"destinations": [
			{
				"match": {
					"kuma.io/service": "*"
				}
			}
		],
		"conf": {
			"interval": "11s",
			"detectors": {
				"totalErrors": {
					"consecutive": 10
				}
			}
		}	
	}
	],
	"next": "http://localhost:5681/meshes/default/circuit-breakers?offset=1"
}`

				// when
				rsr := &rest.ResourceListReceiver{
					NewResource: func() model.Resource {
						return mesh.NewCircuitBreakerResource()
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
					Type: "CircuitBreaker",
					Mesh: "default",
					Name: "one",
				}))
				Expect(rs.Items[0].GetSpec()).To(matchers.MatchProto(&mesh_proto.CircuitBreaker{
					Sources: []*mesh_proto.Selector{
						{Match: map[string]string{"kuma.io/service": "*"}},
					},
					Destinations: []*mesh_proto.Selector{
						{Match: map[string]string{"kuma.io/service": "*"}},
					},
					Conf: &mesh_proto.CircuitBreaker_Conf{
						Interval: util_proto.Duration(99 * time.Second),
						Detectors: &mesh_proto.CircuitBreaker_Conf_Detectors{
							TotalErrors: &mesh_proto.CircuitBreaker_Conf_Detectors_Errors{
								Consecutive: util_proto.UInt32(3),
							},
						},
					},
				}))
				Expect(rs.Items[1].GetMeta()).To(Equal(v1alpha1.ResourceMeta{
					Type: "CircuitBreaker",
					Mesh: "default",
					Name: "two",
				}))
				Expect(rs.Items[1].GetSpec()).To(matchers.MatchProto(&mesh_proto.CircuitBreaker{
					Sources: []*mesh_proto.Selector{
						{Match: map[string]string{"kuma.io/service": "*"}},
					},
					Destinations: []*mesh_proto.Selector{
						{Match: map[string]string{"kuma.io/service": "*"}},
					},
					Conf: &mesh_proto.CircuitBreaker_Conf{
						Interval: util_proto.Duration(11 * time.Second),
						Detectors: &mesh_proto.CircuitBreaker_Conf_Detectors{
							TotalErrors: &mesh_proto.CircuitBreaker_Conf_Detectors_Errors{
								Consecutive: util_proto.UInt32(10),
							},
						},
					},
				}))
				Expect(*rs.Next).To(Equal("http://localhost:5681/meshes/default/circuit-breakers?offset=1"))
			})
		})
	})
})
