package rest_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/model"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/model/rest"

	sample_proto "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/test/apis/sample/v1alpha1"
	sample_core "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/test/resources/apis/sample"
)

var _ = Describe("ResourceListReceiver", func() {
	Describe("UnmarshalJSON", func() {
		It("it should be possible to unmarshal JSON response from Konvoy API Server", func() {
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
					"mesh": "pilot",
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
						Mesh: "pilot",
						Name: "two",
					},
					Spec: &sample_proto.TrafficRoute{
						Path: "/another",
					},
				}))
		})
	})
})
