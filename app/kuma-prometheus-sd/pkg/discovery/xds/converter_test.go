package xds_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	. "github.com/Kong/kuma/app/kuma-prometheus-sd/pkg/discovery/xds"

	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/discovery/targetgroup"

	observability_proto "github.com/Kong/kuma/api/observability/v1alpha1"
)

var _ = Describe("Converter", func() {

	Describe("Convert()", func() {

		type testCase struct {
			input    *observability_proto.MonitoringAssignment
			expected []*targetgroup.Group
		}

		DescribeTable("should convert Kuma MonitoringAssignments into Prometheus Target Groups",
			func(given testCase) {
				// setup
				c := Converter{}
				// when
				actual := c.Convert(given.input)
				// then
				Expect(actual).To(Equal(given.expected))
			},
			Entry("1 Dataplane per assignment, in the format expected by `custom-sd` adapter", testCase{
				input: &observability_proto.MonitoringAssignment{
					Name: "/meshes/default/dataplanes/backend-01",
					Targets: []*observability_proto.MonitoringAssignment_Target{{
						Labels: map[string]string{
							"__address__": "192.168.0.1:8080",
						},
					}},
					Labels: map[string]string{
						"__scheme__":       "http",
						"__metrics_path__": "/metrics",
						"job":              "backend",
						"instance":         "backend-01",
						"mesh":             "default",
						"dataplane":        "backend-01",
						"env":              "prod",
						"service":          "backend",
					},
				},
				expected: []*targetgroup.Group{{
					Source: "/meshes/default/dataplanes/backend-01/0",
					Targets: []model.LabelSet{
						{
							"__address__": "192.168.0.1:8080",
						},
					},
					Labels: model.LabelSet{
						"mesh":             "default",
						"dataplane":        "backend-01",
						"env":              "prod",
						"service":          "backend",
						"__scheme__":       "http",
						"__metrics_path__": "/metrics",
						"job":              "backend",
						"instance":         "backend-01",
					},
				}},
			}),
			Entry("1 Dataplane per assignment, in a free format", testCase{
				input: &observability_proto.MonitoringAssignment{
					Name: "/meshes/default/dataplanes/backend-01",
					Targets: []*observability_proto.MonitoringAssignment_Target{{
						Labels: map[string]string{
							"__address__":      "192.168.0.1:8080",
							"__scheme__":       "http",
							"__metrics_path__": "/metrics",
							"instance":         "backend-01",
							"dataplane":        "backend-01",
						},
					}},
					Labels: map[string]string{
						"job":     "backend",
						"mesh":    "default",
						"service": "backend",
						"env":     "prod",
					},
				},
				expected: []*targetgroup.Group{{
					Source: "/meshes/default/dataplanes/backend-01/0",
					Targets: []model.LabelSet{
						{
							"__address__": "192.168.0.1:8080",
						},
					},
					Labels: model.LabelSet{
						"mesh":             "default",
						"dataplane":        "backend-01",
						"env":              "prod",
						"service":          "backend",
						"__scheme__":       "http",
						"__metrics_path__": "/metrics",
						"job":              "backend",
						"instance":         "backend-01",
					},
				}},
			}),
			Entry("N Dataplanes per assignment, in a free format", testCase{
				input: &observability_proto.MonitoringAssignment{
					Name: "/meshes/default/services/backend",
					Targets: []*observability_proto.MonitoringAssignment_Target{
						{
							Labels: map[string]string{
								"__address__":      "192.168.0.1:8080",
								"__scheme__":       "http",
								"__metrics_path__": "/metrics",
								"instance":         "backend-01",
								"dataplane":        "backend-01",
								"env":              "prod",
							},
						},
						{
							Labels: map[string]string{
								"__address__":      "192.168.0.2:8081",
								"__scheme__":       "http",
								"__metrics_path__": "/metrics",
								"instance":         "backend-02",
								"dataplane":        "backend-02",
								"env":              "test",
							},
						},
					},
					Labels: map[string]string{
						"job":     "backend",
						"mesh":    "default",
						"service": "backend",
					},
				},
				expected: []*targetgroup.Group{
					{
						Source: "/meshes/default/services/backend/0",
						Targets: []model.LabelSet{
							{
								"__address__": "192.168.0.1:8080",
							},
						},
						Labels: model.LabelSet{
							"mesh":             "default",
							"dataplane":        "backend-01",
							"env":              "prod",
							"service":          "backend",
							"__scheme__":       "http",
							"__metrics_path__": "/metrics",
							"job":              "backend",
							"instance":         "backend-01",
						},
					},
					{
						Source: "/meshes/default/services/backend/1",
						Targets: []model.LabelSet{
							{
								"__address__": "192.168.0.2:8081",
							},
						},
						Labels: model.LabelSet{
							"mesh":             "default",
							"dataplane":        "backend-02",
							"env":              "test",
							"service":          "backend",
							"__scheme__":       "http",
							"__metrics_path__": "/metrics",
							"job":              "backend",
							"instance":         "backend-02",
						},
					},
				},
			}),
		)
	})
})
