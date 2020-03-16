package rest_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/Kong/kuma/pkg/core/resources/model"
	"github.com/Kong/kuma/pkg/core/resources/model/rest"
	sample_proto "github.com/Kong/kuma/pkg/test/apis/sample/v1alpha1"
	sample_core "github.com/Kong/kuma/pkg/test/resources/apis/sample"
)

var _ = Describe("Rest Resource", func() {
	Describe("Resource", func() {
		Describe("MarshalJSON", func() {
			It("should marshal JSON with proper field order", func() {
				// given
				res := &rest.Resource{
					Meta: rest.ResourceMeta{
						Type: "TrafficRoute",
						Mesh: "default",
						Name: "one",
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
				expected := `{"type":"TrafficRoute","mesh":"default","name":"one","path":"/example"}`
				Expect(string(bytes)).To(Equal(expected))
			})

			It("should marshal JSON with proper field order and empty spec", func() {
				// given
				res := &rest.Resource{
					Meta: rest.ResourceMeta{
						Type: "TrafficRoute",
						Mesh: "default",
						Name: "one",
					},
				}

				// when
				bytes, err := json.Marshal(res)

				// then
				Expect(err).ToNot(HaveOccurred())

				// and
				expected := `{"type":"TrafficRoute","mesh":"default","name":"one"}`
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
				]
			}`

				// when
				rsr := &rest.ResourceListReceiver{
					NewResource: func() model.Resource {
						return &sample_core.TrafficRouteResource{}
					},
				}
				err := json.Unmarshal([]byte(content), rsr)

				// then
				Expect(err).ToNot(HaveOccurred())

				// when
				rs := rsr.ResourceList
				// then
				Expect(rs.Items).To(ConsistOf(
					&rest.Resource{
						Meta: rest.ResourceMeta{
							Type: "TrafficRoute",
							Mesh: "default",
							Name: "one",
						},
						Spec: &sample_proto.TrafficRoute{
							Path: "/example",
						},
					},
					&rest.Resource{
						Meta: rest.ResourceMeta{
							Type: "TrafficRoute",
							Mesh: "demo",
							Name: "two",
						},
						Spec: &sample_proto.TrafficRoute{
							Path: "/another",
						},
					}))
			})
		})
	})
})
