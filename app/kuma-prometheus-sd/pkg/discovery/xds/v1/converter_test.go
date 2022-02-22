package v1_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/discovery/targetgroup"

	observability_v1 "github.com/kumahq/kuma/api/observability/v1"
	v1 "github.com/kumahq/kuma/app/kuma-prometheus-sd/pkg/discovery/xds/v1"
)

var _ = Describe("Converter", func() {

	Describe("Convert()", func() {

		type testCase struct {
			input    *observability_v1.MonitoringAssignment
			expected []*targetgroup.Group
		}

		DescribeTable("should convert Kuma MonitoringAssignments into Prometheus Target Groups",
			func(given testCase) {
				// setup
				c := v1.Converter{}
				// when
				actual := c.Convert(given.input)
				// then
				Expect(actual).To(Equal(given.expected))
			},
			Entry("1 Dataplane per assignment, in the format expected by `custom-sd` adapter", testCase{
				input: &observability_v1.MonitoringAssignment{
					Mesh:    "demo",
					Service: "backend",
					Targets: []*observability_v1.MonitoringAssignment_Target{{
						Name:        "backend-01",
						Address:     "192.168.0.1:1234",
						Scheme:      "http",
						MetricsPath: "/non-standard-path",
						Labels: map[string]string{
							"env":              "prod",
							"envs":             ",prod,",
							"kuma_io_service":  "backend",
							"kuma_io_services": ",backend,backend-https,", // must have multiple values
							"version":          "v1",
							"versions":         ",v1,v2,", // must have multiple values
						},
					}},
				},
				expected: []*targetgroup.Group{{
					Source: "/meshes/demo/targets/backend-01/0",
					Targets: []model.LabelSet{
						{
							"__address__": "192.168.0.1:1234",
						},
					},
					Labels: model.LabelSet{
						"mesh":             "demo",
						"dataplane":        "backend-01",
						"service":          "backend",
						"__scheme__":       "http",
						"__metrics_path__": "/non-standard-path",
						"job":              "backend",
						"instance":         "backend-01",
						// custom labels
						"env":              "prod",
						"envs":             ",prod,",
						"kuma_io_service":  "backend",
						"kuma_io_services": ",backend,backend-https,", // must have multiple values
						"version":          "v1",
						"versions":         ",v1,v2,", // must have multiple values
					},
				}},
			}),
		)
	})
})
