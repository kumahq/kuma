package mesh_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v3/api/system/v1alpha1"
	. "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/v3/pkg/plugins/ca/provided/config"
	util_proto "github.com/kumahq/kuma/v3/pkg/util/proto"
)

var _ = Describe("MeshResource", func() {
	Describe("HasPrometheusMetricsEnabled", func() {
		type testCase struct {
			mesh     *MeshResource
			expected bool
		}

		DescribeTable("should correctly determine whether Prometheus metrics has been enabled on that Mesh",
			func(given testCase) {
				Expect(given.mesh.HasPrometheusMetricsEnabled()).To(Equal(given.expected))
			},
			Entry("mesh == nil", testCase{
				mesh:     nil,
				expected: false,
			}),
			Entry("mesh.metrics == nil", testCase{
				mesh:     NewMeshResource(),
				expected: false,
			}),
			Entry("mesh.metrics.prometheus == nil", testCase{
				mesh: &MeshResource{
					Spec: &mesh_proto.Mesh{
						Metrics: &mesh_proto.Metrics{},
					},
				},
				expected: false,
			}),
			Entry("mesh.metrics.prometheus != nil", testCase{
				mesh: &MeshResource{
					Spec: &mesh_proto.Mesh{
						Metrics: &mesh_proto.Metrics{
							EnabledBackend: "prometheus-1",
							Backends: []*mesh_proto.MetricsBackend{
								{
									Name: "prometheus-1",
									Type: mesh_proto.MetricsPrometheusType,
								},
							},
						},
					},
				},
				expected: true,
			}),
		)
	})

	Describe("ParseDuration", func() {
		type testCase struct {
			input  string
			output time.Duration
		}

		DescribeTable("should return the correct duration",
			func(given testCase) {
				duration, _ := ParseDuration(given.input)
				Expect(given.output).To(Equal(duration))
			},
			Entry("should return 0 if seconds is 0", testCase{
				input:  "0s",
				output: 0,
			}),
			Entry("should return minute", testCase{
				input:  "5m",
				output: 5 * time.Minute,
			}),
			Entry("should return day", testCase{
				input:  "4d",
				output: 4 * 24 * time.Hour,
			}),
			Entry("should return year", testCase{
				input:  "5y",
				output: 5 * 365 * 24 * time.Hour,
			}),
		)
	})
	Describe("MarshalLog", func() {
		It("should mask the sensitive information when marshaling", func() {
			// given
			conf, _ := util_proto.ToStruct(&config.ProvidedCertificateAuthorityConfig{
				Cert: &v1alpha1.DataSource{Type: &v1alpha1.DataSource_Inline{Inline: util_proto.Bytes([]byte("secret1"))}},
				Key:  &v1alpha1.DataSource{Type: &v1alpha1.DataSource_Inline{Inline: util_proto.Bytes([]byte("secret2"))}},
			})
			meshResourceList := MeshResourceList{
				Items: []*MeshResource{
					{
						Spec: &mesh_proto.Mesh{
							Mtls: &mesh_proto.Mesh_Mtls{
								Backends: []*mesh_proto.CertificateAuthorityBackend{
									{
										Conf: conf,
									},
								},
							},
						},
					},
				},
			}

			// when
			masked := meshResourceList.MarshalLog().(MeshResourceList)

			// then
			cfg := &config.ProvidedCertificateAuthorityConfig{}
			err := util_proto.ToTyped(masked.Items[0].Spec.Mtls.Backends[0].Conf, cfg)
			Expect(err).ToNot(HaveOccurred())
			Expect(cfg.Key.String()).To(Equal(`inline:{value:"***"}`))
			Expect(cfg.Cert.String()).To(Equal(`inline:{value:"***"}`))
		})
	})
})
