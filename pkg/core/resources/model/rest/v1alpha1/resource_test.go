package v1alpha1_test

import (
	"encoding/json"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	meshexternalservice_api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshexternalservice/api/v1alpha1"
	meshmultizoneservice_api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshmultizoneservice/api/v1alpha1"
	meshservice_api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshservice/api/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/core/resources/model/rest/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/test/kds/samples"
	"github.com/kumahq/kuma/v2/pkg/util/pointer"
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
  "kri": "kri_mtp_default___one_",
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

			It("should include port-sorted snis for MeshService with multiple ports", func() {
				res := &v1alpha1.Resource{
					ResourceMeta: v1alpha1.ResourceMeta{
						Type:             "MeshService",
						Mesh:             "default",
						Name:             "backend",
						CreationTime:     t1,
						ModificationTime: t2,
						Labels: map[string]string{
							mesh_proto.ZoneTag:             "east",
							mesh_proto.KubeNamespaceTag:    "demo",
							mesh_proto.ResourceOriginLabel: string(mesh_proto.ZoneResourceOrigin),
						},
					},
					Spec: &meshservice_api.MeshService{
						Ports: []meshservice_api.Port{
							{Port: 8080, Name: pointer.To("http")},
							{Port: 6379},
						},
					},
				}

				bytes, err := json.Marshal(res)
				Expect(err).ToNot(HaveOccurred())

				var got struct {
					KRI  string              `json:"kri"`
					SNIs []v1alpha1.SNIEntry `json:"snis"`
				}
				Expect(json.Unmarshal(bytes, &got)).To(Succeed())
				Expect(got.KRI).To(Equal("kri_msvc_default_east_demo_backend_"))
				Expect(got.SNIs).To(Equal([]v1alpha1.SNIEntry{
					{Port: 6379, Section: "6379", SNI: "sni.msvc.default.east.demo.backend.6379"},
					{Port: 8080, Section: "http", SNI: "sni.msvc.default.east.demo.backend.http"},
				}))
			})

			It("should include a single sni for MeshExternalService", func() {
				res := &v1alpha1.Resource{
					ResourceMeta: v1alpha1.ResourceMeta{
						Type:             "MeshExternalService",
						Mesh:             "default",
						Name:             "google",
						CreationTime:     t1,
						ModificationTime: t2,
					},
					Spec: &meshexternalservice_api.MeshExternalService{
						Match: meshexternalservice_api.Match{Port: 443},
					},
				}

				bytes, err := json.Marshal(res)
				Expect(err).ToNot(HaveOccurred())

				var got struct {
					SNIs []v1alpha1.SNIEntry `json:"snis"`
				}
				Expect(json.Unmarshal(bytes, &got)).To(Succeed())
				Expect(got.SNIs).To(Equal([]v1alpha1.SNIEntry{
					{Port: 443, Section: "443", SNI: "sni.extsvc.default.google.443"},
				}))
			})

			It("should include snis for MeshMultiZoneService", func() {
				res := &v1alpha1.Resource{
					ResourceMeta: v1alpha1.ResourceMeta{
						Type:             "MeshMultiZoneService",
						Mesh:             "default",
						Name:             "agg",
						CreationTime:     t1,
						ModificationTime: t2,
					},
					Spec: &meshmultizoneservice_api.MeshMultiZoneService{
						Ports: []meshmultizoneservice_api.Port{
							{Port: 9090, Name: pointer.To("grpc")},
						},
					},
				}

				bytes, err := json.Marshal(res)
				Expect(err).ToNot(HaveOccurred())

				var got struct {
					SNIs []v1alpha1.SNIEntry `json:"snis"`
				}
				Expect(json.Unmarshal(bytes, &got)).To(Succeed())
				Expect(got.SNIs).To(Equal([]v1alpha1.SNIEntry{
					{Port: 9090, Section: "grpc", SNI: "sni.mzsvc.default.agg.grpc"},
				}))
			})

			It("should omit snis for MeshService with no ports", func() {
				res := &v1alpha1.Resource{
					ResourceMeta: v1alpha1.ResourceMeta{
						Type:             "MeshService",
						Mesh:             "default",
						Name:             "empty",
						CreationTime:     t1,
						ModificationTime: t2,
					},
					Spec: &meshservice_api.MeshService{},
				}

				bytes, err := json.Marshal(res)
				Expect(err).ToNot(HaveOccurred())
				Expect(bytes).ToNot(ContainSubstring(`"snis"`))
			})

			It("should not include snis for non-destination resources", func() {
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

				bytes, err := json.Marshal(res)
				Expect(err).ToNot(HaveOccurred())
				Expect(bytes).ToNot(ContainSubstring(`"snis"`))
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
  "modificationTime": "2019-07-17T16:05:36.995Z",
  "kri": "kri_mtp_default___one_"
}
`
				Expect(string(bytes)).To(MatchJSON(expected))
			})
		})
	})
})
