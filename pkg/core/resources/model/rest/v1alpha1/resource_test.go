package v1alpha1_test

import (
	"encoding/json"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/core/resources/model/rest/v1alpha1"
	"github.com/kumahq/kuma/pkg/test/kds/samples"
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
				res := &v1alpha1.Resource{
					ResourceMeta: v1alpha1.ResourceMeta{
						Type:             "MeshTrafficPermission",
						Mesh:             "default",
						Name:             "one",
						CreationTime:     t1,
						ModificationTime: t2,
					},
					Spec: samples.MeshTrafficPermission,
				}

				// when
				bytes, err := json.Marshal(res)

				// then
				Expect(err).ToNot(HaveOccurred())

				// and
				expected := `
{
  "type": "MeshTrafficPermission",
  "mesh": "default",
  "name": "one",
  "creationTime": "2018-07-17T16:05:36.995Z",
  "modificationTime": "2019-07-17T16:05:36.995Z",
  "spec": {
    "targetRef": {
      "kind": "Mesh"
    },
    "from": [
      {
        "targetRef": {
          "kind": "Mesh"
        },
        "default": {
          "action": "Allow"
        }
      }
    ]
  }
}`
				Expect(string(bytes)).To(MatchJSON(expected))
			})

			It("should marshal JSON with proper field order and empty spec", func() {
				// given
				res := &v1alpha1.Resource{
					ResourceMeta: v1alpha1.ResourceMeta{
						Type:             "MeshTrafficPermission",
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
				expected := `
{
  "type": "MeshTrafficPermission",
  "mesh": "default",
  "name": "one",
  "creationTime": "2018-07-17T16:05:36.995Z",
  "modificationTime": "2019-07-17T16:05:36.995Z"
}
`
				Expect(string(bytes)).To(MatchJSON(expected))
			})
		})
	})
})
