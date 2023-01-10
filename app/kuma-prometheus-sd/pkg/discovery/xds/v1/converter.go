package v1

import (
	"fmt"

	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/discovery/targetgroup"
	"github.com/prometheus/prometheus/util/strutil"

	observability_v1 "github.com/kumahq/kuma/api/observability/v1"
)

const (
	// meshLabel is the name of the label that holds the mesh name.
	meshLabel = model.LabelName("mesh")
	// serviceLabel is the name of the label that holds the service name.
	serviceLabel = model.LabelName("service")
	// dataplaneLabel is the name of the label that holds the dataplane name.
	dataplaneLabel = model.LabelName("dataplane")
)

type Converter struct{}

func (c Converter) ConvertAll(assignments []*observability_v1.MonitoringAssignment) []*targetgroup.Group {
	var groups []*targetgroup.Group
	for _, assignment := range assignments {
		groups = append(groups, c.Convert(assignment)...)
	}
	return groups
}

// Convert translates MADS resources into Prometheus TargetGroups
//
// Beware of the following constraints when it comes to integration with Prometheus:
//
//  1. Prometheus model for all `sd`s except for `file_sd` looks like this:
//
//     // Group is a set of targets with a common label set (production, test, staging etc.).
//     type Group struct {
//     // Targets is a list of targets identified by a label set. Each target is
//     // uniquely identifiable in the group by its address label.
//     Targets []model.LabelSet
//     // Labels is a set of labels that is common across all targets in the group.
//     Labels model.LabelSet
//
//     // Source is an identifier that describes a group of targets.
//     Source string
//     }
//
//     That is why Kuma's MonitoringAssignment was designed to be close to that model.
//
//  2. However, `file_sd` uses different model for reading data from a file:
//
//     struct {
//     Targets []string       `yaml:"targets"`
//     Labels  model.LabelSet `yaml:"labels"`
//     }
//
//     Notice that Targets is just a list of addresses rather than a list of model.LabelSet.
//
//  3. Because of that mismatch, some form of conversion is unavoidable on client side,
//     e.g. inside `kuma-prometheus-sd`
//
//  4. The next component that imposes its constraints is `custom-sd`- adapter
//     (https://github.com/prometheus/prometheus/tree/master/documentation/examples/custom-sd)
//     that is recommended for use by all `file_sd`-based `sd`s.
//
//     This adapter is doing conversion from Prometheus model into `file_sd` model
//     and it expects that `Targets` field has only 1 label - `__address__` -
//     and the rest of the labels must be a part of `Labels` field.
//
//  5. Therefore, we need to convert MonitoringAssignment into a model that `custom-sd` expects.
//
//     In practice, it means that generated MonitoringAssignment will be mapped to a set of groups, one per target.
//     In the Prometheus native SD, this will not be the case and there will be a 1-1 mapping between assignments and groups.
func (c Converter) Convert(assignment *observability_v1.MonitoringAssignment) []*targetgroup.Group {
	commonLabels := convertLabels(assignment.Labels)

	commonLabels[meshLabel] = model.LabelValue(assignment.Mesh)
	commonLabels[serviceLabel] = model.LabelValue(assignment.Service)

	var groups []*targetgroup.Group

	for i, target := range assignment.Targets {
		targetLabels := convertLabels(target.Labels).Merge(commonLabels)

		targetLabels[dataplaneLabel] = model.LabelValue(target.Name)
		targetLabels[model.InstanceLabel] = model.LabelValue(target.Name)
		targetLabels[model.SchemeLabel] = model.LabelValue(target.Scheme)
		targetLabels[model.MetricsPathLabel] = model.LabelValue(target.MetricsPath)
		targetLabels[model.JobLabel] = model.LabelValue(assignment.Service)

		group := &targetgroup.Group{
			Source: sourceName(assignment, target, i),
			Targets: []model.LabelSet{{
				model.AddressLabel: model.LabelValue(target.Address),
			}},
			Labels: targetLabels,
		}
		groups = append(groups, group)
	}

	return groups
}

func convertLabels(labels map[string]string) model.LabelSet {
	labelSet := model.LabelSet{}
	for key, value := range labels {
		name := strutil.SanitizeLabelName(key)
		labelSet[model.LabelName(name)] = model.LabelValue(value)
	}
	return labelSet
}

func sourceName(assignment *observability_v1.MonitoringAssignment, target *observability_v1.MonitoringAssignment_Target, i int) string {
	// unique name, e.g. REST API uri
	return fmt.Sprintf("/meshes/%s/targets/%s/%d", assignment.GetMesh(), target.GetName(), i)
}
